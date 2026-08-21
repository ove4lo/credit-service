package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ove4lo/credit-service/internal/application"
)

// server represents items for handlers
type server struct {
	logger *slog.Logger
	store *application.Store
	jwtSecret []byte
	amqpCh *amqp091.Channel
}

type loginRequest struct {
	Username string `json:"username"`
}

type createApplicationRequest struct {
	Client string `json:"client"`
	Amount int `json:"amount"`
	Term int `json:"term"`
}

// handleLogin verifies the token
func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Username) == "" {
		http.Error(w, "username is required", http.StatusUnprocessableEntity)
		return
	}

	// WHY: The password is not verified—a token is simply issued based on the name
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{ // NOTE: jwt.SigningMethodHS256 is the standard algorithm used to encrypt the signature
		"sub": req.Username, // Who is the token owner
		"iat": time.Now().Unix(), // Date of issue
		"exp": time.Now().Add(time.Hour).Unix(), // When it goes bad
	})

	// NOTE: Signing the token with a secret key
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Error("couldn't sign token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": signed})
}

func (s *server) requiredAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Retrieve the header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized) // NOTE: No header → nothing to check → 401
			return
		}

		// 2. Expect the format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Decipher and verify the signature
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			/** Recalculates the signature and compares: 
				if it matches → the token is authentic; 
				if it doesn't match (someone tampered with the payload) → error
			*/
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { // WHY: an explicit check that the algorithm is the expected one (HMAC)
				/** NOTE: Without this check, there is a known attack:
					the attacker slips in a token with the none algorithm (“without signature”), 
					and the naive code accepts it
				*/ 
				return nil, errors.New("unexpected signing method")
			}
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			s.logger.Warn("invalid token", "error", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 4. Token is valid — proceed to the handler
		next(w, r)
	}
}

// validate reviews the application
func (r createApplicationRequest) validate() error {
	if strings.TrimSpace(r.Client) == "" { // WHY: trims whitespace
		return errors.New("client is required")
	}

	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if r.Term <= 0 {
		return  errors.New("term must be positive")
	}

	return nil
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

	// NOTE: validate input before touching the database
	if err := req.validate(); err != nil {
		s.logger.Warn("validation failed", "error", err)
		/** WHY:  422 — the JSON is valid and parsed successfully, 
			but the data is logically invalid (e.g., a negative amount)
		*/
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) 
		// NOTE: In other words: "I understood it, but I can't accept it"
		return
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