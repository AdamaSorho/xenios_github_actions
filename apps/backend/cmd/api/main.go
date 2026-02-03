package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xenios/backend/internal/adapter/handler"
	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/infrastructure/config"
	"github.com/xenios/backend/internal/infrastructure/database"
	"github.com/xenios/backend/internal/usecase"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Wire up dependencies (Clean Architecture: outer layers depend on inner)
	userRepo := repository.NewPostgresUserRepository(db)
	getUserUseCase := usecase.NewGetUserUseCase(userRepo)
	createUserUseCase := usecase.NewCreateUserUseCase(userRepo)

	// HTTP handlers
	userHandler := handler.NewUserHandler(getUserUseCase, createUserUseCase)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
