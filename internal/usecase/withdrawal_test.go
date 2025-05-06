package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessWithdrawal(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()
	validOrderNumber := "4532015112830366"

	tests := []struct {
		name        string
		amount      float64
		orderNumber string
		setupMocks  func(*testutils.MockWithdrawalStorage, *testutils.MockBalanceStorage)
		expectedErr error
	}{
		{
			name:        "успешное снятие средств",
			amount:      100.0,
			orderNumber: validOrderNumber,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 200.0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
				bs.On("UpdateWithdrawn", mock.Anything, userID, 100.0).Return(nil)
				ws.On("CreateWithdrawal", mock.Anything, mock.AnythingOfType("models.Withdrawal")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "неположительная сумма",
			amount:      0,
			orderNumber: validOrderNumber,
			setupMocks:  func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedErr: errors.New("withdrawal amount must be positive"),
		},
		{
			name:        "недостаточный баланс",
			amount:      100.0,
			orderNumber: validOrderNumber,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 50.0, Valid: true}, pgtype.Float8{}, nil)
			},
			expectedErr: errors.New("insufficient balance"),
		},
		{
			name:        "неверный номер заказа",
			amount:      100.0,
			orderNumber: "4532015112830367",
			setupMocks:  func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedErr: errors.New("invalid order number"),
		},
		{
			name:        "ошибка записи списания",
			amount:      100.0,
			orderNumber: validOrderNumber,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 200.0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
				bs.On("UpdateWithdrawn", mock.Anything, userID, 100.0).Return(nil)
				ws.On("CreateWithdrawal", mock.Anything, mock.AnythingOfType("models.Withdrawal")).Return(errors.New("db error"))
			},
			expectedErr: errors.New("failed to record withdrawal: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutils.MockWithdrawalStorage{}
			bs := &testutils.MockBalanceStorage{}
			tt.setupMocks(ws, bs)

			balanceUC := NewBalanceUseCase(bs)
			uc := NewWithdrawalUseCase(ws, balanceUC)

			err := uc.ProcessWithdrawal(ctx, userID, tt.orderNumber, tt.amount)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			ws.AssertExpectations(t)
			bs.AssertExpectations(t)
		})
	}
}

func TestGetUserWithdrawals(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()

	tests := []struct {
		name                string
		setupMocks          func(*testutils.MockWithdrawalStorage)
		expectedWithdrawals []models.Withdrawal
		expectedErr         error
	}{
		{
			name: "успешное получение списаний",
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{
					{UserID: userID, OrderNumber: "123", Sum: pgtype.Float8{Float64: 100.0, Valid: true}},
				}, nil)
			},
			expectedWithdrawals: []models.Withdrawal{
				{UserID: userID, OrderNumber: "123", Sum: pgtype.Float8{Float64: 100.0, Valid: true}},
			},
			expectedErr: nil,
		},
		{
			name: "нет списаний",
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{}, nil)
			},
			expectedWithdrawals: []models.Withdrawal{},
			expectedErr:         nil,
		},
		{
			name: "ошибка хранилища",
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{}, errors.New("db error"))
			},
			expectedWithdrawals: nil,
			expectedErr:         errors.New("failed to get withdrawals: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutils.MockWithdrawalStorage{}
			tt.setupMocks(ws)

			uc := NewWithdrawalUseCase(ws, nil)
			withdrawals, err := uc.GetUserWithdrawals(ctx, userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWithdrawals, withdrawals)
			}

			ws.AssertExpectations(t)
		})
	}
}
