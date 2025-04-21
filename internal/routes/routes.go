package router

import (
	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/go-chi/chi/v5"
)

const (
	UserPrefix      = "/api/user"
	OrdersPath      = "/orders"
	RegisterPath    = "/register"
	LoginPath       = "/login"
	BalancePath     = "/balance"
	WithdrawPath    = "/balance/withdraw"
	WithdrawalsPath = "/withdrawals"
)

func SetupRoutes(store *storage.Storage, jwtSecret, loyaltyURL string) *chi.Mux {
	r := chi.NewRouter()
	loyaltyClient := loyalty.NewClient(loyaltyURL)

	r.Post(UserPrefix+RegisterPath, handlers.NewRegisterHandler(store, jwtSecret).ServeHTTP)
	r.Post(UserPrefix+LoginPath, handlers.NewLoginHandler(store, jwtSecret).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Post(UserPrefix+OrdersPath, handlers.NewOrderHandler(store, store, loyaltyClient).ServeHTTP)
		r.Get(UserPrefix+OrdersPath, handlers.NewOrderGetHandler(store).ServeHTTP)
		r.Get(UserPrefix+BalancePath, handlers.NewBalanceHandler(store).ServeHTTP)
		r.Post(UserPrefix+WithdrawPath, handlers.NewWithdrawHandler(store, store).ServeHTTP)
		r.Get(UserPrefix+WithdrawalsPath, handlers.NewWithdrawalsHandler(store).ServeHTTP)
	})

	return r
}