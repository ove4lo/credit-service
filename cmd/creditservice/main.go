package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ove4lo/credit-service/internal/application"
)

func main() {
	store := application.NewStore()

	// Registering a handler function for the POST /applications route
	http.HandleFunc("POST /applications", func(w http.ResponseWriter, r *http.Request) {
		var app application.Application // NOTE: creating an empty variable `app` to subsequently store data from the request
		
		// WHY: JSON decoder that reads the request body and attempts to unpack data into the `app`
		if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		saved := store.Add(app)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(saved) // NOTE: converting the `saved` object back into a JSON string
	})

	log.Println("listening on :4000")
	http.ListenAndServe(":4000", nil) // WHY: nil means using the standard router
}