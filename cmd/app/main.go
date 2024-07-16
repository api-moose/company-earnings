package main

import (
	"context"
	"encoding/json"
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
	credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	if credentialsJSON == "" {
		log.Fatal("FIREBASE_CREDENTIALS_FILE environment variable is not set")
	}

	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	wrappedAuthClient := &FirebaseAuthWrapper{client: authClient}

	r := setupRouter(wrappedAuthClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func setupRouter(authClient auth.FirebaseAuthClient) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(logging.LoggingMiddleware)
	r.Use(tenancy.NewTenantMiddleware(authClient).Middleware)
	r.Use(auth.NewAuthMiddleware(authClient).Middleware)
	r.Use(access_control.RBACMiddleware)

	r.Get("/", mainHandler)
	r.Get("/health", healthCheckHandler)
	r.Get("/version", versionHandler)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 page not found", http.StatusNotFound)
	})

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, notFoundMessage, http.StatusNotFound)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the Financial Data Platform API"))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "healthy"}
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"version": version}
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}
