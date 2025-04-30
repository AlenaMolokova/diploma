package router

import (
	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/AlenaMolokova/diploma/internal/usecase"
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

	balanceUC := usecase.NewBalanceUseCase(store)
	withdrawalUC := usecase.NewWithdrawalUseCase(store, balanceUC)
	orderUC := usecase.NewOrderUseCase(store, loyaltyClient, balanceUC)

	r.Post(UserPrefix+RegisterPath, handlers.NewRegisterHandler(store, jwtSecret).ServeHTTP)
	r.Post(UserPrefix+LoginPath, handlers.NewLoginHandler(store, jwtSecret).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Post(UserPrefix+OrdersPath, handlers.NewOrderHandler(orderUC).ServeHTTP)
		r.Get(UserPrefix+OrdersPath, handlers.NewOrderGetHandler(store).ServeHTTP)
		r.Get(UserPrefix+BalancePath, handlers.NewBalanceHandler(balanceUC).ServeHTTP)
		r.Post(UserPrefix+WithdrawPath, handlers.NewWithdrawHandler(withdrawalUC).ServeHTTP)
		r.Get(UserPrefix+WithdrawalsPath, handlers.NewWithdrawalsHandler(withdrawalUC).ServeHTTP)
	})

	
	return r
}