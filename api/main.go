package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request){
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

func handleSearch(w http.ResponseWriter, r *http.Request){
	
}
func main() {
	http.HandleFunc("/health", handleHealthCheck)
	http.HandleFunc("/search", handleSearch)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %s", err)
	}
	
}