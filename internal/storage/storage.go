package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
)

type Storage struct {
	db      *pgxpool.Pool
	queries *Queries
}

func NewStorage(db *pgxpool.Pool) (*Storage, error) {
	if db == nil {
		return nil, errors.New("database pool is nil")
	}
	queries := New(db)
	return &Storage{db: db, queries: queries}, nil
}

func (s *Storage) CreateUser(ctx context.Context, login, password string) (int64, error) {
	id, err := s.queries.CreateUser(ctx, CreateUserParams{
		Login:    login,
		Password: password,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, errors.New("login already exists")
		}
		return 0, err
	}
	return id, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (User, error) {
	return s.queries.GetUserByLogin(ctx, login)
}

func (s *Storage) GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error) {
	bal, err := s.queries.GetUserBalance(ctx, userID)
	if err != nil {
		return pgtype.Float8{}, pgtype.Float8{}, err
	}
	return bal.Balance, bal.Withdrawn, nil
}

func (s *Storage) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	return s.queries.UpdateBalance(ctx, UpdateBalanceParams{
		ID:        userID,
		Balance:   pgtype.Float8{Float64: amount, Valid: true},
		Withdrawn: pgtype.Float8{Float64: 0, Valid: true},
	})
}

func (s *Storage) CreateOrder(ctx context.Context, order Order) error {
	return s.queries.CreateOrder(ctx, CreateOrderParams{
		UserID:     order.UserID,
		Number:     order.Number,
		Status:     order.Status,
		UploadedAt: order.UploadedAt,
	})
}

func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (Order, error) {
	return s.queries.GetOrderByNumber(ctx, number)
}

func (s *Storage) GetOrdersByUserID(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := s.queries.GetOrdersByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	orders := make([]Order, len(rows))
	for i, row := range rows {
		orders[i] = Order{
			UserID:     pgtype.Int8{Int64: userID, Valid: true},
			Number:     row.Number,
			Status:     row.Status,
			Accrual:    row.Accrual,
			UploadedAt: row.UploadedAt,
		}
	}
	return orders, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, withdrawal Withdrawal) error {
	return s.queries.CreateWithdrawal(ctx, CreateWithdrawalParams{
		UserID:      withdrawal.UserID,
		OrderNumber: withdrawal.OrderNumber,
		Sum:         withdrawal.Sum,
		ProcessedAt: withdrawal.ProcessedAt,
	})
}

func (s *Storage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]Withdrawal, error) {
	rows, err := s.queries.GetWithdrawalsByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	withdrawals := make([]Withdrawal, len(rows))
	for i, row := range rows {
		withdrawals[i] = Withdrawal{
			UserID:      pgtype.Int8{Int64: userID, Valid: true},
			OrderNumber: row.OrderNumber,
			Sum:         row.Sum,
			ProcessedAt: row.ProcessedAt,
		}
	}
	return withdrawals, nil
}