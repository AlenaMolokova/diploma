package handlers

import (
	"context"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserStorage interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (storage.User, error)
}

type BalanceStorage interface {
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
}

type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, withdrawal storage.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]storage.Withdrawal, error)
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, order storage.Order) error
	GetOrderByNumber(ctx context.Context, number string) (storage.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]storage.Order, error)
}