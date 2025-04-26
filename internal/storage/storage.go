package storage

import (
	"context"
	"errors"

	"github.com/AlenaMolokova/diploma/internal/models"
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

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	user, err := s.queries.GetUserByLogin(ctx, login)
	if err != nil {
		return models.User{}, err
	}
	return models.User{
		ID:        user.ID,
		Login:     user.Login,
		Password:  user.Password,
		Balance:   user.Balance,
		Withdrawn: user.Withdrawn,
	}, nil
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

func (s *Storage) UpdateBalanceWithWithdrawn(ctx context.Context, userID int64, balance, withdrawn float64) error {
	return s.queries.UpdateBalance(ctx, UpdateBalanceParams{
		ID:        userID,
		Balance:   pgtype.Float8{Float64: balance, Valid: true},
		Withdrawn: pgtype.Float8{Float64: withdrawn, Valid: true},
	})
}

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) error {
	return s.queries.CreateOrder(ctx, CreateOrderParams{
		UserID:     pgtype.Int8{Int64: order.UserID, Valid: true},
		Number:     order.Number,
		Status:     order.Status,
		UploadedAt: order.UploadedAt,
	})
}

func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (models.Order, error) {
	order, err := s.queries.GetOrderByNumber(ctx, number)
	if err != nil {
		return models.Order{}, err
	}
	return models.Order{
		ID:         order.ID,
		UserID:     order.UserID.Int64,
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt,
	}, nil
}

func (s *Storage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	rows, err := s.queries.GetOrdersByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	orders := make([]models.Order, len(rows))
	for i, row := range rows {
		orders[i] = models.Order{
			Number:     row.Number,
			Status:     row.Status,
			Accrual:    row.Accrual,
			UploadedAt: row.UploadedAt,
		}
	}
	return orders, nil
}

func (s *Storage) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := s.queries.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	orders := make([]models.Order, len(rows))
	for i, row := range rows {
		orders[i] = models.Order{
			ID:         row.ID,
			UserID:     row.UserID.Int64,
			Number:     row.Number,
			Status:     row.Status,
			Accrual:    row.Accrual,
			UploadedAt: row.UploadedAt,
		}
	}
	return orders, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	return s.queries.CreateWithdrawal(ctx, CreateWithdrawalParams{
		UserID:      pgtype.Int8{Int64: withdrawal.UserID, Valid: true},
		OrderNumber: withdrawal.OrderNumber,
		Sum:         withdrawal.Sum.Float64,
		ProcessedAt: withdrawal.ProcessedAt,
	})
}

func (s *Storage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	rows, err := s.queries.GetWithdrawalsByUser(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	withdrawals := make([]models.Withdrawal, len(rows))
	for i, row := range rows {
		withdrawals[i] = models.Withdrawal{
			OrderNumber: row.OrderNumber,
			Sum:         pgtype.Float8{Float64: row.Sum, Valid: true},
			ProcessedAt: row.ProcessedAt,
		}
	}
	return withdrawals, nil
}

func (s *Storage) UpdateOrder(ctx context.Context, order models.Order) error {
	return s.queries.UpdateOrder(ctx, UpdateOrderParams{
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt,
	})
}