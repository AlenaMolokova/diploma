package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlenaMolokova/diploma/internal/config"
	"github.com/AlenaMolokova/diploma/internal/routes"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, err := migrate.New("file://migrations", cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	log.Println("Migrations applied successfully")

	pool, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	store, err := storage.NewStorage(pool)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-gophermart-secret-key-1234567890"
		log.Println("JWT_SECRET not set, using default value")
	}
	log.Printf("JWT_SECRET loaded: %q", jwtSecret)

	r := router.SetupRoutes(store, jwtSecret, cfg.AccrualAddr)

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	go func() {
		log.Printf("Starting Gophermart server on %s", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	pool.Close()
	log.Println("Server stopped")
}