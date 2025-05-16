package main

import (
	"context"
	"log"
	"net/http"

	"github.com/AlenaMolokova/diploma/internal/config"
	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/migrations"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := migrations.Apply(cfg.DatabaseURI); err != nil {
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
	loyaltyClient.SetPollInterval(cfg.PollIntervalSec)

	balanceUC := usecase.NewBalanceUseCase(store)
	withdrawalUC := usecase.NewWithdrawalUseCase(store, balanceUC)
	orderUC := usecase.NewOrderUseCase(store, loyaltyClient, balanceUC)

	registerHandler := handlers.NewRegisterHandler(store, cfg.JWTSecret)
	loginHandler := handlers.NewLoginHandler(store, cfg.JWTSecret)
	orderHandler := handlers.NewOrderHandler(orderUC)
	orderGetHandler := handlers.NewOrderGetHandler(store)
	balanceHandler := handlers.NewBalanceHandler(balanceUC)
	withdrawHandler := handlers.NewWithdrawHandler(withdrawalUC)
	withdrawalsHandler := handlers.NewWithdrawalsHandler(withdrawalUC)

	r := chi.NewRouter()

	r.Post("/api/user/register", registerHandler.ServeHTTP)
	r.Post("/api/user/login", loginHandler.ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		r.Post("/api/user/orders", orderHandler.ServeHTTP)
		r.Get("/api/user/orders", orderGetHandler.ServeHTTP)
		r.Get("/api/user/balance", balanceHandler.ServeHTTP)
		r.Post("/api/user/balance/withdraw", withdrawHandler.ServeHTTP)
		r.Get("/api/user/withdrawals", withdrawalsHandler.ServeHTTP)
	})

	go loyaltyClient.StartOrderProcessing(context.Background(), store)

	log.Printf("Starting Gophermart server on %s", cfg.RunAddr)
	if err := http.ListenAndServe(cfg.RunAddr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
