package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/AlenaMolokova/diploma/internal/config"
	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/router"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func applyMigrations(databaseURI string) error {
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
		log.Printf("DatabaseURI was empty, using default: %s", cfg.DatabaseURI)
	}

	log.Printf("Using DatabaseURI: %s", cfg.DatabaseURI)

	if err := applyMigrations(cfg.DatabaseURI); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store, err := storage.NewStorage(db)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	loyaltyClient := loyalty.NewClient(cfg.AccrualAddr)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
		log.Println("JWT_SECRET not set, using default")
	}

	r := router.SetupRoutes(store, jwtSecret, cfg.AccrualAddr)

	var orderStore models.OrderStorage = store
	go loyaltyClient.StartOrderProcessing(context.Background(), orderStore)

	log.Printf("Starting Gophermart server on %s", cfg.RunAddr)
	if err := http.ListenAndServe(cfg.RunAddr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}