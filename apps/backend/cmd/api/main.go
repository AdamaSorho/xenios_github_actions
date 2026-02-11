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

func main() {
	server := setupServer()
	runServer(server)
}

// setupServer creates and configures the HTTP server
func setupServer() *http.Server {
	port := getPort()
	startTime := time.Now()

	// Initialize database connection pool (optional)
	var healthHandler *handler.HealthHandler
	dbPool := initDatabasePool()
	if dbPool != nil {
		defer dbPool.Close()

		// Wire up dependencies: Infrastructure -> Use Case -> Handler
		healthChecker := repository.NewPostgresHealthChecker(dbPool, 3*time.Second)
		getHealthStatusUseCase := usecase.NewGetHealthStatusUseCase(healthChecker, startTime)
		healthHandler = handler.NewHealthHandlerWithUseCase(getHealthStatusUseCase)
	} else {
		// Fallback to simple health handler if no database configured
		healthHandler = handler.NewHealthHandler()
	}

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
	}
}

// initDatabasePool initializes the PostgreSQL connection pool.
// Returns nil if DATABASE_URL is not set (database is optional for health endpoint).
// Logs a warning if database connection fails but continues without it.
func initDatabasePool() *pgxpool.Pool {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL not set, running without database health checks")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Printf("Warning: Failed to parse DATABASE_URL: %v", err)
		return nil
	}

	// Configure connection pool
	config.MaxConns = 10
	config.MinConns = 2
	config.ConnConfig.ConnectTimeout = 3 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Warning: Failed to create database pool: %v", err)
		return nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: Failed to ping database: %v", err)
		pool.Close()
		return nil
	}

	log.Println("Database connection pool initialized successfully")
	return pool
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
