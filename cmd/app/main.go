package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/api-moose/company-earnings/internal/config"
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
	// Initialize PocketBase
	pbConfig := config.NewPocketBaseConfig()

	// Start PocketBase in a goroutine
	go func() {
		if err := pbConfig.Start(); err != nil {
			log.Fatalf("Failed to start PocketBase: %v", err)
		}
	}()

	router := setupRouter()

	// Create a server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-stop

	log.Println("Shutting down server...")

	// Shutdown the server
	if err := server.Shutdown(nil); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	pbConfig.App.ResetBootstrapState()

	log.Println("Server and PocketBase stopped gracefully")
}
