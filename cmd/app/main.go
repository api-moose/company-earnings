package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/api-moose/company-earnings/internal/config"
	"github.com/api-moose/company-earnings/internal/middleware/access_control"
	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/api-moose/company-earnings/internal/utils/logging"
)

const version = "0.1.0"

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": version})
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Welcome to the Financial Data Platform API")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func setupRouter(pbConfig *config.PocketBaseConfig, authMiddleware func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/version", versionHandler)

	// Apply middleware in the correct order
	handler := logging.LoggingMiddleware(mux)
	handler = tenancy.TenantMiddleware(handler)

	if authMiddleware == nil {
		authMiddleware = auth.AuthMiddleware(pbConfig.Adapter)
	}
	handler = authMiddleware(handler)

	handler = access_control.RBACMiddleware(handler)

	return handler
}

func main() {
	pbConfig := config.NewPocketBaseConfig()

	// Start PocketBase in a separate goroutine
	go func() {
		if err := pbConfig.Start(); err != nil {
			log.Fatalf("Failed to start PocketBase: %v", err)
		}
	}()

	// Setup the HTTP router with middleware
	router := setupRouter(pbConfig, nil)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Listen for interrupt signals to gracefully shutdown the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down server...")

	// Gracefully shutdown the server with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	// Reset PocketBase bootstrap state
	pbConfig.App.ResetBootstrapState()

	log.Println("Server and PocketBase stopped gracefully")
}
