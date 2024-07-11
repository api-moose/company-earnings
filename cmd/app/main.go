package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/api-moose/company-earnings/internal/utils/logging"
)

const version = "0.1.0"

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": version})
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Financial Data Platform API")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func setupRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/version", versionHandler)

	return logging.LoggingMiddleware(mux)
}

func main() {
	router := setupRouter()
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
