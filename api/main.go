package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}
type server struct {
	pool *pgxpool.Pool
}

// Healthcheck handler. Returns "Status": "OK" if everything is working
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allowed-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "OK",
	})
}


// Handles query searches. Takes in a request with a Query and a Top_K value and returns the result from the DB.
func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allowed-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var request SearchRequest
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	embedding, err := embeddingGenerator(r.Context(), request.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	results, err := similaritySearch(r.Context(), s.pool, embedding, request.TopK)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"Results": results})
}
func main() {
	pool, err := connectDB(context.Background())
	if err != nil {
		log.Fatalf("Error connecting with DB: %v", err)
	}

	s := &server{pool: pool}
	http.HandleFunc("/health", handleHealthCheck)
	http.HandleFunc("/search", s.handleSearch)
	fmt.Print("Now running on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %s", err)
	}

}
