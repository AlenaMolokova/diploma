package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/AlenaMolokova/diploma/internal/config"
	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	mux := http.NewServeMux()

	mux.Handle("/api/user/register", handlers.NewRegisterHandler(store, jwtSecret))
	mux.Handle("/api/user/login", handlers.NewLoginHandler(store, jwtSecret))
	mux.Handle("/api/user/orders", middleware.AuthMiddleware(jwtSecret)(handlers.NewOrderHandler(store, store, loyaltyClient)))
	mux.Handle("/api/user/balance", middleware.AuthMiddleware(jwtSecret)(handlers.NewBalanceHandler(store)))
	mux.Handle("/api/user/balance/withdraw", middleware.AuthMiddleware(jwtSecret)(handlers.NewWithdrawHandler(store, store)))
	mux.Handle("/api/user/withdrawals", middleware.AuthMiddleware(jwtSecret)(handlers.NewWithdrawalsHandler(store)))

	go loyaltyClient.StartOrderProcessing(context.Background(), store)

	log.Printf("Starting Gophermart server on %s", cfg.RunAddr)
	if err := http.ListenAndServe(cfg.RunAddr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}