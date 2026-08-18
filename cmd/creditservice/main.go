package main

import (
	"log/slog"
	"net/http"
	"os"

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
	}

	// Routing: path → server method
	http.HandleFunc("POST /applications", srv.handleCreateApplication)

	// WHY: "starting server" is the message (the "what"), while "addr" and ":4000" are the context fields (the details)
	logger.Info("starting server", "addr", ":4000")
	http.ListenAndServe(":4000", nil) // WHY: nil means using the standard router
}