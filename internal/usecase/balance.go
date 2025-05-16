package usecase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type BalanceStorage interface {
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
	UpdateWithdrawn(ctx context.Context, userID int64, withdrawn float64) error
}

type BalanceUseCase interface {
	GetUserBalance(ctx context.Context, userID int64) (current, withdrawn float64, err error)
	AddToBalance(ctx context.Context, userID int64, amount float64) error
	WithdrawFromBalance(ctx context.Context, userID int64, amount float64, orderNumber string) error
}

type balanceUseCase struct {
	storage BalanceStorage
}

func NewBalanceUseCase(storage BalanceStorage) BalanceUseCase {
	return &balanceUseCase{storage: storage}
}

func (u *balanceUseCase) GetUserBalance(ctx context.Context, userID int64) (current, withdrawn float64, err error) {
	currentPgx, withdrawnPgx, err := u.storage.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get balance: %w", err)
	}

	current = 0
	if currentPgx.Valid {
		current = currentPgx.Float64
	}

	withdrawn = 0
	if withdrawnPgx.Valid {
		withdrawn = withdrawnPgx.Float64
	}

	return current, withdrawn, nil
}

func (u *balanceUseCase) AddToBalance(ctx context.Context, userID int64, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	current, _, err := u.storage.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	newBalance := amount
	if current.Valid {
		newBalance += current.Float64
	}

	return u.storage.UpdateBalance(ctx, userID, newBalance)
}

func (u *balanceUseCase) WithdrawFromBalance(ctx context.Context, userID int64, amount float64, orderNumber string) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive")
	}

	current, withdrawn, err := u.storage.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	currentBalance := 0.0
	if current.Valid {
		currentBalance = current.Float64
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance")
	}

	newBalance := currentBalance - amount
	newWithdrawn := amount
	if withdrawn.Valid {
		newWithdrawn += withdrawn.Float64
	}

	if err := u.storage.UpdateBalance(ctx, userID, newBalance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := u.storage.UpdateWithdrawn(ctx, userID, newWithdrawn); err != nil {
		return fmt.Errorf("failed to update withdrawn amount: %w", err)
	}

	return nil
}
