package usecase

import (
	"context"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserBalance(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := NewBalanceUseCase(mockStorage)

	ctx := context.Background()
	userID := int64(1)
	current := pgtype.Float8{Float64: 100, Valid: true}
	withdrawn := pgtype.Float8{Float64: 20, Valid: true}

	mockStorage.On("GetBalance", mock.Anything, userID).Return(current, withdrawn, nil)

	curr, wd, err := uc.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, curr)
	assert.Equal(t, 20.0, wd)
}

func TestWithdrawFromBalance_Insufficient(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := NewBalanceUseCase(mockStorage)

	ctx := context.Background()
	userID := int64(1)
	mockStorage.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 10, Valid: true}, pgtype.Float8{Float64: 5, Valid: true}, nil)

	err := uc.WithdrawFromBalance(ctx, userID, 100, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestWithdrawFromBalance_Success(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := NewBalanceUseCase(mockStorage)

	ctx := context.Background()
	userID := int64(1)
	mockStorage.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 100, Valid: true}, pgtype.Float8{Float64: 5, Valid: true}, nil)
	mockStorage.On("UpdateBalance", mock.Anything, userID, 50.0).Return(nil)
	mockStorage.On("UpdateWithdrawn", mock.Anything, userID, 55.0).Return(nil)

	err := uc.WithdrawFromBalance(ctx, userID, 50, "123")
	assert.NoError(t, err)
}

func TestAddToBalance_Success(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := NewBalanceUseCase(mockStorage)

	ctx := context.Background()
	userID := int64(1)
	mockStorage.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 50, Valid: true}, pgtype.Float8{}, nil)
	mockStorage.On("UpdateBalance", mock.Anything, userID, 150.0).Return(nil)

	err := uc.AddToBalance(ctx, userID, 100)
	assert.NoError(t, err)
}

func TestAddToBalance_NegativeAmount(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := NewBalanceUseCase(mockStorage)

	ctx := context.Background()
	userID := int64(1)

	err := uc.AddToBalance(ctx, userID, -100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
}
