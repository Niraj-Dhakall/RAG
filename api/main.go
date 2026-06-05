package main

import (
	"net/http"
	"encodings/json"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request){
	
}
func main() {

	http.HandleFunc("/health", handleHealthCheck)
}