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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/middleware/access_control"
	auth "github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/api-moose/company-earnings/internal/utils/logging"
	"google.golang.org/api/option"
)

const version = "0.1.0"
const notFoundMessage = "404 page not found"

type FirebaseAuthWrapper struct {
	client *firebaseAuth.Client
}

func (f *FirebaseAuthWrapper) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	return f.client.VerifyIDToken(ctx, idToken)
}

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Financial Data Platform API")

	// Load Firebase credentials
	credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	if credentialsJSON == "" {
		log.Println("Warning: FIREBASE_CREDENTIALS_FILE environment variable is not set")
	}

	// Initialize Firebase app
	var app *firebase.App
	var err error
	if credentialsJSON != "" {
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))
		app, err = firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Printf("Error initializing Firebase app: %v", err)
			// Continue without Firebase for now
		}
	}

	var wrappedAuthClient *FirebaseAuthWrapper
	if app != nil {
		// Get Firebase Auth client
		authClient, err := app.Auth(context.Background())
		if err != nil {
			log.Printf("Error getting Firebase Auth client: %v", err)
			// Continue without Firebase Auth for now
		} else {
			wrappedAuthClient = &FirebaseAuthWrapper{client: authClient}
		}
	}

	// Set up router
	r := setupRouter(wrappedAuthClient)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("PORT not set, using default port 8080")
	}

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRouter(authClient auth.FirebaseAuthClient) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(logging.LoggingMiddleware)

	if authClient != nil {
		r.Use(tenancy.NewTenantMiddleware(authClient).Middleware)
		r.Use(auth.NewAuthMiddleware(authClient).Middleware)
		r.Use(access_control.RBACMiddleware)
	} else {
		log.Println("Warning: Running without authentication middleware")
	}

	r.Get("/", mainHandler)
	r.Get("/health", healthCheckHandler)
	r.Get("/version", versionHandler)

	r.NotFound(notFoundHandler)

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, notFoundMessage, http.StatusNotFound)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the Financial Data Platform API")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "healthy"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"version": version}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding version response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
