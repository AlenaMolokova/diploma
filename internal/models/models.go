package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    pgtype.Float8
	UploadedAt pgtype.Timestamptz
}

type User struct {
	ID        int64
	Login     string
	Password  string
	Balance   pgtype.Float8
	Withdrawn pgtype.Float8
}

type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Sum         pgtype.Float8
	ProcessedAt pgtype.Timestamptz
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, order Order) error
	GetOrderByNumber(ctx context.Context, number string) (Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
	UpdateOrder(ctx context.Context, order Order) error
	GetOrdersByUserID(ctx context.Context, userID int64) ([]Order, error)
}

type BalanceStorage interface {
	GetBalance(ctx context.Context, userID int64) (current pgtype.Float8, withdrawn pgtype.Float8, err error)
	UpdateBalance(ctx context.Context, userID int64, current float64) error
	UpdateWithdrawn(ctx context.Context, userID int64, withdrawn float64) error
	CreateWithdrawal(ctx context.Context, withdrawal Withdrawal) error
}

type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, withdrawal Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]Withdrawal, error)
}

type UserStorage interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
}