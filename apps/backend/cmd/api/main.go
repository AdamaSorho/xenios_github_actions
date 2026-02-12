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
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/worker"
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
	healthHandler, pool, cleanup := setupHealthHandler()

	mux, jobWorker := configureRoutes(healthHandler, pool)

	// Compose cleanup function
	fullCleanup := func() {
		if jobWorker != nil {
			log.Println("Stopping job worker...")
			jobWorker.Stop()
			log.Println("Job worker stopped")
		}
		cleanup()
	}

	return &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, fullCleanup
}

// configureRoutes sets up the HTTP mux with all routes.
// If pool is non-nil, job queue endpoints are registered.
func configureRoutes(healthHandler *handler.HealthHandler, pool *pgxpool.Pool) (*http.ServeMux, *worker.Worker) {
	versionHandler := handler.NewVersionHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /version", versionHandler.Version)

	// Set up job queue infrastructure if database is available
	var jobWorker *worker.Worker
	if pool != nil {
		queueHandler, w := setupJobQueue(pool)
		jobWorker = w
		mux.HandleFunc("POST /jobs", queueHandler.EnqueueJob)
		mux.HandleFunc("GET /jobs/status", queueHandler.GetQueueStatus)
	}

	return mux, jobWorker
}

// setupHealthHandler creates a health handler with optional database connectivity.
// If DATABASE_URL is not set, returns a handler with no use case (legacy behavior).
// If DATABASE_URL is set, wires up the full health checking infrastructure.
// Returns the handler, the database pool (nil if no database), and a cleanup function.
func setupHealthHandler() (*handler.HealthHandler, *pgxpool.Pool, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, health check will not verify database connectivity")
		return handler.NewHealthHandler(), nil, func() {}
	}

	// Set up database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Warning: failed to create database connection pool: %v", err)
		log.Println("Health check will not verify database connectivity")
		return handler.NewHealthHandler(), nil, func() {}
	}

	// Verify initial connection
	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
		log.Println("Health check will not verify database connectivity")
		pool.Close()
		return handler.NewHealthHandler(), nil, func() {}
	}

	healthHandler, cleanup := wireHealthHandler(pool)
	return healthHandler, pool, cleanup
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

// setupJobQueue wires up the job queue infrastructure and starts the worker.
func setupJobQueue(pool *pgxpool.Pool) (*handler.QueueHandler, *worker.Worker) {
	// Infrastructure layer: PostgreSQL-backed job queue
	jobQueue := repository.NewPostgresJobQueue(pool)

	// Application layer: use cases
	enqueueUC := usecase.NewEnqueueJobUseCase(jobQueue)
	statusUC := usecase.NewGetQueueStatusUseCase(jobQueue)

	// Presentation layer: HTTP handler
	queueHandler := handler.NewQueueHandler(enqueueUC, statusUC)

	// Worker infrastructure
	w := worker.NewWorker(jobQueue, 5*time.Second, 5*time.Minute)

	// Register placeholder handlers for all job types.
	// These will be replaced with real implementations as features are built.
	allJobTypes := []entities.JobType{
		entities.JobTypeTranscription,
		entities.JobTypeDocumentExtraction,
		entities.JobTypeInsightGeneration,
		entities.JobTypeAnalyticsAggregation,
		entities.JobTypeRiskDetection,
		entities.JobTypeAudioCleanup,
	}
	for _, jt := range allJobTypes {
		w.RegisterHandler(jt, func(ctx context.Context, job *entities.Job) error {
			log.Printf("Processing %s job %s (placeholder handler)", jt, job.ID)
			return nil
		})
	}

	// Start worker
	ctx := context.Background()
	w.Start(ctx)
	log.Println("Job worker started with handlers for all job types")

	return queueHandler, w
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
