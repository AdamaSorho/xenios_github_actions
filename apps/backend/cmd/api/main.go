package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/adapter/handler"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/infrastructure/config"
	"github.com/xenios/backend/internal/usecase"
)

var startTime = time.Now()

func main() {
	cfg := config.Load()
	server, cleanup := setupServer(cfg)
	defer cleanup()
	runServer(server)
}

// setupServer creates and configures the HTTP server with chi router,
// middleware chain, and all route handlers.
func setupServer(cfg *config.Config) (*http.Server, func()) {
	// Set up health handler with optional database connectivity
	healthHandler, dbCleanup := setupHealthHandler()
	versionHandler := handler.NewVersionHandler()

	// Set up coach-client handlers with in-memory repo (skeleton)
	ccRepo := repository.NewInMemoryCoachClientRepository()
	createCCUseCase := usecase.NewCreateCoachClientUseCase(ccRepo)
	listCCUseCase := usecase.NewListCoachClientsUseCase(ccRepo)
	ccHandler := handler.NewCoachClientHandler(createCCUseCase, listCCUseCase)

	// Build chi router with middleware chain
	r := chi.NewRouter()

	// Global middleware (applied to all routes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(os.Stdout))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes (no JWT required)
	r.Get("/health", healthHandler.Health)
	r.Get("/version", versionHandler.Version)

	// API v1 routes (JWT protected)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret))

		// Coach-client management
		r.Post("/coaches/{coachID}/clients", ccHandler.Create)
		r.Get("/coaches/{coachID}/clients", ccHandler.List)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, dbCleanup
}

// setupHealthHandler creates a health handler with optional database connectivity.
func setupHealthHandler() (*handler.HealthHandler, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, health check will not verify database connectivity")
		return handler.NewHealthHandler(), func() {}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Warning: failed to create database connection pool: %v", err)
		log.Println("Health check will not verify database connectivity")
		return handler.NewHealthHandler(), func() {}
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
		log.Println("Health check will not verify database connectivity")
		pool.Close()
		return handler.NewHealthHandler(), func() {}
	}

	return wireHealthHandler(pool)
}

// wireHealthHandler sets up the health checking infrastructure.
func wireHealthHandler(pool *pgxpool.Pool) (*handler.HealthHandler, func()) {
	log.Println("Database connection pool established successfully")

	healthChecker := repository.NewPostgresHealthChecker(pool, 3*time.Second)
	healthUseCase := usecase.NewGetHealthStatusUseCase(healthChecker, startTime)

	cleanup := func() {
		log.Println("Closing database connection pool...")
		pool.Close()
		log.Println("Database connection pool closed")
	}

	return handler.NewHealthHandlerWithUseCase(healthUseCase), cleanup
}

// runServer starts the server and handles graceful shutdown.
func runServer(server *http.Server) {
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
