package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/validation"
	"github.com/jackc/pgx/v5/pgtype"
)

type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type WithdrawalUseCase struct {
	storage   WithdrawalStorage
	balanceUC BalanceUseCase
	validator validation.OrderValidator
}

func NewWithdrawalUseCase(storage WithdrawalStorage, balanceUC BalanceUseCase) *WithdrawalUseCase {
	return &WithdrawalUseCase{
		storage:   storage,
		balanceUC: balanceUC,
		validator: validation.NewLuhnValidator(),
	}
}

func (uc *WithdrawalUseCase) ProcessWithdrawal(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive")
	}

	if !uc.validator.ValidateOrderNumber(orderNumber) {
		return fmt.Errorf("invalid order number")
	}

	if err := uc.balanceUC.WithdrawFromBalance(ctx, userID, amount, orderNumber); err != nil {
		return err
	}

	withdrawal := models.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         pgtype.Float8{Float64: amount, Valid: true},
		ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if err := uc.storage.CreateWithdrawal(ctx, withdrawal); err != nil {
		return fmt.Errorf("failed to record withdrawal: %w", err)
	}

	return nil
}

func (uc *WithdrawalUseCase) GetUserWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	withdrawals, err := uc.storage.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}
