package main

import (
	"encoding/json"
	"log/slog"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// debtCheckTask represents items for handlers
type debtCheckTask struct {
	ApplicationID int `json:"application_id"`
	Client string `json:"client"`
	Amount int `json:"amount"`
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

	logger.Info("worker started, waiting for tasks")
	
	// NOTE: Read messages from the channel one by one
	for msg := range deliveries {
		var task debtCheckTask
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			logger.Error("bad task payload", "error", err)
			msg.Nack(false, false) // WHY: malformed message — reject without return
			continue
		}

		logger.Info("processed debt check", "app_id", task.ApplicationID, "client", task.Client)

		// plag
		decision := "approved"
		reason := "no debts found"
		if task.Amount > 500000 {
			decision = "rejected"
			reason = "amount exceeds limit"
		}

		logger.Info("debt check done",
			"app_id", task.ApplicationID,
			"decision", decision,
			"reason", reason,
		)
		msg.Ack(false) // NOTE: acknowledging: processed, remove from queue
	}
}