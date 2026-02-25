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
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/auth"
	"github.com/xenios/backend/internal/infrastructure/config"
	"github.com/xenios/backend/internal/infrastructure/worker"
	"github.com/xenios/backend/internal/usecase"
)

var startTime = time.Now()

func main() {
	server, cleanup := setupServer()
	defer cleanup()
	runServer(server)
}

// setupServer creates and configures the HTTP server with Chi router,
// middleware chain, and all route handlers.
func setupServer() (*http.Server, func()) {
	cfg := config.Load()

	healthHandler, pool, cleanup := setupHealthHandler()
	router, jobWorker, asyncAudit := configureRoutes(cfg, healthHandler, pool)

	fullCleanup := func() {
		if jobWorker != nil {
			log.Println("Stopping job worker...")
			jobWorker.Stop()
			log.Println("Job worker stopped")
		}
		if asyncAudit != nil {
			log.Println("Stopping async audit logger...")
			asyncAudit.Stop()
			log.Println("Async audit logger stopped")
		}
		cleanup()
	}

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, fullCleanup
}

// configureRoutes sets up the Chi router with middleware and all routes.
func configureRoutes(cfg *config.Config, healthHandler *handler.HealthHandler, pool *pgxpool.Pool) (chi.Router, *worker.Worker, *repository.AsyncAuditRepository) {
	r := chi.NewRouter()

	// Middleware chain
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AuditContextMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Set up audit repository (async for non-blocking logging)
	auditRepo, asyncAudit := setupAuditRepository(pool)

	// Public routes
	r.Get("/api/health", healthHandler.Health)
	r.Get("/version", handler.NewVersionHandler().Version)

	// Auth routes (public — no JWT required)
	authHandler := setupAuthHandler(cfg, auditRepo)
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/refresh", authHandler.Refresh)

	// Protected API routes with JWT auth and versioned prefix
	var jobWorker *worker.Worker
	r.Route("/api/v1", func(api chi.Router) {
		api.Use(middleware.JWTAuth(cfg.JWTSecret))

		// Auth logout (requires valid JWT)
		api.Post("/auth/logout", authHandler.Logout)

		// Coach-client management endpoints
		ccRepo := repository.NewInMemoryCoachClientRepository()
		createUC := usecase.NewCreateCoachClientUseCase(ccRepo)
		listUC := usecase.NewListCoachClientsUseCase(ccRepo)
		ccHandler := handler.NewCoachClientHandler(createUC, listUC)

		api.Post("/coaches/{coachID}/clients", ccHandler.Create)
		api.Get("/coaches/{coachID}/clients", ccHandler.List)

		// Audit log query endpoint (admin-only)
		queryAuditUC := usecase.NewQueryAuditLogUseCase(auditRepo)
		auditHandler := handler.NewAuditHandler(queryAuditUC)
		api.Get("/admin/audit", auditHandler.Query)

		// File upload/download endpoints
		uploadHandler := setupUploadHandler()
		api.Post("/uploads/presign", uploadHandler.RequestPresignedURL)
		api.Post("/uploads/{artifactID}/confirm", uploadHandler.ConfirmUpload)
		api.Post("/uploads/{artifactID}/download", uploadHandler.RequestDownloadURL)

		// Job queue endpoints (if database is available)
		if pool != nil {
			queueHandler, w := setupJobQueue(pool, auditRepo)
			jobWorker = w
			api.Post("/jobs", queueHandler.EnqueueJob)
			api.Get("/jobs/status", queueHandler.GetQueueStatus)
		}
	})

	return r, jobWorker, asyncAudit
}

// setupAuditRepository creates the audit repository with async wrapper.
// Uses PostgreSQL when database is available, in-memory otherwise.
func setupAuditRepository(pool *pgxpool.Pool) (domainrepo.AuditRepository, *repository.AsyncAuditRepository) {
	var inner repository.AuditRepositoryInterface
	if pool != nil {
		inner = repository.NewPostgresAuditRepository(pool)
		log.Println("Audit logging configured with PostgreSQL backend")
	} else {
		inner = repository.NewInMemoryAuditRepository()
		log.Println("Audit logging configured with in-memory backend (no database)")
	}

	asyncRepo := repository.NewAsyncAuditRepository(inner, 1000)
	asyncRepo.Start()
	log.Println("Async audit logger started with buffer size 1000")

	return asyncRepo, asyncRepo
}

// setupAuthHandler wires up auth dependencies and returns the handler.
func setupAuthHandler(cfg *config.Config, auditRepo domainrepo.AuditRepository) *handler.AuthHandler {
	userRepo := repository.NewInMemoryUserRepository()
	tokenRepo := repository.NewInMemoryRefreshTokenRepository()

	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "dev-secret-do-not-use-in-production"
	}
	tokenService := auth.NewJWTTokenService(jwtSecret, 15*time.Minute)
	hasher := auth.NewBcryptHasher(12)

	registerUC := usecase.NewRegisterUserUseCase(userRepo, tokenRepo, tokenService, auditRepo, hasher)
	loginUC := usecase.NewLoginUserUseCase(userRepo, tokenRepo, tokenService, auditRepo, hasher)
	refreshUC := usecase.NewRefreshTokenUseCase(userRepo, tokenRepo, tokenService, auditRepo)
	logoutUC := usecase.NewLogoutUserUseCase(tokenRepo, auditRepo)

	return handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC)
}

// setupHealthHandler creates a health handler with optional database connectivity.
func setupHealthHandler() (*handler.HealthHandler, *pgxpool.Pool, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, health check will not verify database connectivity")
		return handler.NewHealthHandler(), nil, func() {}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Warning: failed to create database connection pool: %v", err)
		log.Println("Health check will not verify database connectivity")
		return handler.NewHealthHandler(), nil, func() {}
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
		log.Println("Health check will not verify database connectivity")
		pool.Close()
		return handler.NewHealthHandler(), nil, func() {}
	}

	healthHandler, cleanup := wireHealthHandler(pool)
	return healthHandler, pool, cleanup
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

// setupJobQueue wires up the job queue infrastructure and starts the worker.
func setupJobQueue(pool *pgxpool.Pool, auditRepo domainrepo.AuditRepository) (*handler.QueueHandler, *worker.Worker) {
	jobQueue := repository.NewPostgresJobQueue(pool)
	enqueueUC := usecase.NewEnqueueJobUseCase(jobQueue)
	statusUC := usecase.NewGetQueueStatusUseCase(jobQueue)
	queueHandler := handler.NewQueueHandler(enqueueUC, statusUC)

	w := worker.NewWorker(jobQueue, 5*time.Second, 5*time.Minute)

	// Register placeholder handlers for job types without real implementations yet
	placeholderTypes := []entities.JobType{
		entities.JobTypeTranscription,
		entities.JobTypeDocumentExtraction,
		entities.JobTypeAnalyticsAggregation,
		entities.JobTypeRiskDetection,
		entities.JobTypeAudioCleanup,
	}
	for _, jt := range placeholderTypes {
		w.RegisterHandler(jt, func(ctx context.Context, job *entities.Job) error {
			log.Printf("Processing %s job %s (placeholder handler)", jt, job.ID)
			return nil
		})
	}

	// Register insight generation handler with real use case
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measureRepo := repository.NewInMemoryMeasurementRepository()
	wearableRepo := repository.NewInMemoryWearableSummaryRepository()
	generateInsightsUC := usecase.NewGenerateInsightsUseCase(insightRepo, measureRepo, wearableRepo, auditRepo)
	w.RegisterHandler(entities.JobTypeInsightGeneration, worker.NewInsightGenerationHandler(generateInsightsUC))

	ctx := context.Background()
	w.Start(ctx)
	log.Println("Job worker started with handlers for all job types (insight_generation uses real handler)")

	return queueHandler, w
}

// setupUploadHandler wires up file upload/download dependencies and returns the handler.
func setupUploadHandler() *handler.UploadHandler {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	requestUploadUC := usecase.NewRequestUploadUseCase(artifactRepo, fileStorage, auditRepo)
	confirmUploadUC := usecase.NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo)
	requestDownloadUC := usecase.NewRequestDownloadUseCase(artifactRepo, fileStorage, auditRepo)

	return handler.NewUploadHandler(requestUploadUC, confirmUploadUC, requestDownloadUC)
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
