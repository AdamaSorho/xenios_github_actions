package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/adapter/handler"
	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/usecase"
)

var startTime = time.Now()

func main() {
	server, cleanup := setupServer()
	defer cleanup()
	runServer(server)
}

// setupServer creates and configures the HTTP server.
// Returns the server and a cleanup function to close resources.
func setupServer() (*http.Server, func()) {
	port := getPort()

	// Set up health handler with optional database connectivity
	healthHandler, cleanup := setupHealthHandler()
	versionHandler := handler.NewVersionHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /version", versionHandler.Version)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, cleanup
}

// setupHealthHandler creates a health handler with optional database connectivity.
// If DATABASE_URL is not set, returns a handler with no use case (legacy behavior).
// If DATABASE_URL is set, wires up the full health checking infrastructure.
// Returns the handler and a cleanup function to close the database pool.
func setupHealthHandler() (*handler.HealthHandler, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, health check will not verify database connectivity")
		return handler.NewHealthHandler(), func() {}
	}

	// Set up database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Warning: failed to create database connection pool: %v", err)
		log.Println("Health check will not verify database connectivity")
		return handler.NewHealthHandler(), func() {}
	}

	// Verify initial connection
	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
		log.Println("Health check will not verify database connectivity")
		pool.Close()
		return handler.NewHealthHandler(), func() {}
	}

	return wireHealthHandler(pool)
}

// wireHealthHandler sets up the health checking infrastructure with an established database pool.
// This is separated from setupHealthHandler for testability.
func wireHealthHandler(pool *pgxpool.Pool) (*handler.HealthHandler, func()) {
	log.Println("Database connection pool established successfully")

	// Wire up the health checking infrastructure (Clean Architecture)
	healthChecker := repository.NewPostgresHealthChecker(pool, 3*time.Second)
	healthUseCase := usecase.NewGetHealthStatusUseCase(healthChecker, startTime)

	// Return handler and cleanup function to close the pool
	cleanup := func() {
		log.Println("Closing database connection pool...")
		pool.Close()
		log.Println("Database connection pool closed")
	}

	return handler.NewHealthHandlerWithUseCase(healthUseCase), cleanup
}

// getPort returns the port to listen on
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// runServer starts the server and handles graceful shutdown
func runServer(server *http.Server) {
	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
