package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ove4lo/credit-service/internal/application"
	amqp "github.com/rabbitmq/amqp091-go"
)

// debtCheckTask represents items for handlers
type debtCheckTask struct {
	ApplicationID int `json:"application_id"`
	Client string `json:"client"`
	Amount int `json:"amount"`
}

type worker struct {
	logger *slog.Logger
	store *application.Store
}

func (wk *worker) processMessage(msg amqp.Delivery) {
	var task debtCheckTask
	if err := json.Unmarshal(msg.Body, &task); err != nil {
		wk.logger.Error("bad task payload", "error", err)
		msg.Nack(false, false) // WHY: malformed message — reject without return
		return
	}

	wk.logger.Info("processed debt check", "app_id", task.ApplicationID, "client", task.Client)

	// plag
	decision := "approved"
	reason := "no debts found"
	if task.Amount > 500000 {
		decision = "rejected"
		reason = "amount exceeds limit"
	}

	if err := wk.store.UpdateStatus(context.Background(), task.ApplicationID, decision); err != nil {
		wk.logger.Error("failed to update status", "error", err)
		msg.Nack(false, false) // unable to record the solution — rejecting
		return
	}

	wk.logger.Info("debt check done",
		"app_id", task.ApplicationID,
		"decision", decision,
		"reason", reason,
	
	)
	msg.Ack(false) // NOTE: acknowledging: processed, remove from queue
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		logger.Error("RABBITMQ_URL isn't set")
		os.Exit(1)
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		logger.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open channel", "error", err)
		os.Exit(1)
	}
	defer ch.Close()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		logger.Error("DATABASE_DSN isn't set")
		os.Exit(1)
	}

	store, err := application.NewStore(dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// NOTE: Declare the same queue — in case the worker started first
	_, err = ch.QueueDeclare(
		"debt_check", 
		true, 
		false, 
		false, 
		false, 
		nil,
	)
	if err != nil {
		logger.Error("failed to declare queue", "error", err)
		os.Exit(1)
	}

	deliveries, err := ch.Consume( // NOTE: A Go channel, RabbitMQ sends messages to it
		"debt_check", // queue
		"", // consumer tag (empty — will be auto-generated)
		false, // auto-ack = false → manual ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil, // args
	)
	if err != nil {
		logger.Error("failed to register consumer", "error", err)
		os.Exit(1)
	}

	wk := &worker{logger: logger, store: store}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	const workerCount = 3

	var wg sync.WaitGroup
	logger.Info("workers started, waiting for tasks", "workers", workerCount)

	for i := 0; i < workerCount; i++ {
		// WHY: `wg.Add(1)` before each goroutine — we register with the `WaitGroup` that we’ve launched another one
		wg.Add(1)
		go func() {
			defer wg.Done() // NOTE: when the goroutine finishes, it will decrement the counter
			for msg := range deliveries { // NOTE: All three goroutines read from the same channel
				wk.processMessage(msg)
			}
		}()
	}
	
	// NOTE: Waiting for a stop signal
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// NOTE: Cancel the subscription → the deliveries channel will close → the goroutine loops will terminate
	if err := ch.Close(); err != nil {
		logger.Error("failed to close channel", "error", err)
	}

	wg.Wait() // The main blocks here until all three goroutines finish
	logger.Info("all workers stopped")
}