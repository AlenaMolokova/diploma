package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID        int64
	Login     string
	Password  string
	Balance   pgtype.Float8
	Withdrawn pgtype.Float8
}

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    pgtype.Float8
	UploadedAt pgtype.Timestamptz
}

type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Sum         pgtype.Float8
	ProcessedAt pgtype.Timestamptz
}

type UserStorage interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
}

type BalanceStorage interface {
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
}

type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, withdrawal Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]Withdrawal, error)
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, order Order) error
	GetOrderByNumber(ctx context.Context, number string) (Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
	UpdateOrder(ctx context.Context, order Order) error
}