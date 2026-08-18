package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ove4lo/credit-service/internal/application"
)

// server represents items for handlers
type server struct {
	logger *slog.Logger
	store *application.Store
}

type createApplicationRequest struct {
	Client string `json:"client"`
	Amount int `json:"amount"`
	Term int `json:"term"`
}

// handleCreateApplication handles the POST request to create a new application
func (s *server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var req createApplicationRequest

	// NOTE: Read and decode the incoming JSON body into the app structure
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("invalid request body", "error", err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return  // WHY: Stop execution immediately if the input data is invalid
	}

	// NOTE: Build the domain model from allowed input fields only
	app := application.Application{
		Client: req.Client,
		Amount: req.Amount,
		Term:   req.Term,
	}

	// NOTE: Save the application to the database using the store dependency
	saved, err := s.store.Add(app)
	if err != nil {
		s.logger.Error("couldn't create application", "error", err, "client", req.Client)
		// WHY: Hide internal DB errors from the client for security reasons
		http.Error(w, "internal error", http.StatusInternalServerError)
		return 
	}

	s.logger.Info("application created", "app_id", saved.ID, "client", saved.Client, "amount", saved.Amount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // HTTP 201: Successfully created
	json.NewEncoder(w).Encode(saved) // WHY: Return the updated object (with its new ID) back to the client
}