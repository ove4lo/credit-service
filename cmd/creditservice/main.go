package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ove4lo/credit-service/internal/application"
)

func main() {
	/**
		The logger and the handler are deliberately separated: 
		the logger determines *what* to record, while the handler determines the format and destination
	*/
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	/**
		NOTE: 
		NewJSONHandler - each record will be output as a JSON object
		os.Stdout - Where to write. The application writes to stdout (standard output, the console) 
		            and doesn't concern itself with files
	*/ 
		Level: slog.LevelInfo,
	}))

	dsn := os.Getenv("DATABASE_DSN")

	if dsn == "" {
		// NOTE: slog doesn't have a Fatal method
		logger.Error("DATABASE_DSN is not set")
		os.Exit(1)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Error("JWT_SECRET isn't set")
		os.Exit(1)
	}

	// Connecting to the database
	store, err := application.NewStore(dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Building the `server` with dependencies
	srv := &server{
		logger: logger,
		store:  store,
		jwtSecret: []byte(jwtSecret), // WHY: []byte, because the signature library works with bytes
	}

	// Routing: path → server method
	http.HandleFunc("POST /login", srv.handleLogin) // simple auth, open
	http.HandleFunc("POST /applications", srv.requiredAuth(srv.handleCreateApplication)) // close

	// NOTE: Create the server instance instead of using global http.ListenAndServe
	// WHY: We need the object itself to call the .Shutdown() method later
	httpServer := &http.Server{
		Addr : ":4000",
		Handler: nil, // Handles requests via default multiplexer
	}

	// NOTE: Create a context that automatically cancels when the OS tells the server to stop
	// WHY: signal.NotifyContext replaces manual channels, converting Ctrl+C (SIGINT/SIGTERM) into a context cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	/** WHY: srv.ListenAndServe blocks the thread
		We run it in a background goroutine (our "intern") 
		so the main code can continue down to listen for the shutdown signal 
	*/
	go func() {
		logger.Info("starting server", "addr", ":4000")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			stop() // WHY: If the server fails to start, tell the main goroutine to wake up and quit
		}
	}()

	// NOTE: This arrow is a brake lever. The main goroutine sleeps here
	// WHY: It will wake up only when ctx.Done() closes (meaning a stop signal or error happened)
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// NOTE: Give the server a hard deadline of 10 seconds to finish existing requests
	// WHY: If a request hangs for 2 minutes, we don't want to wait forever; we force shut it after 10s
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// NOTE: Tell the server to stop accepting new requests and wait for current clients to finish
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}