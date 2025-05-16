package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLoyaltyClientWrapper struct {
	*testutils.MockLoyaltyClient
}

func TestOrderUseCaseProcessNewOrder(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()
	validOrderNumber := "4532015112830366"

	tests := []struct {
		name        string
		orderNumber string
		setupMocks  func(*testutils.MockOrderStorage, *testutils.MockLoyaltyClient, *testutils.MockBalanceStorage)
		expectedErr error
	}{
		{
			name:        "успешное создание заказа - StatusNew",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("not found"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "успешное создание заказа - StatusRegistered",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Status: constants.StatusRegistered,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "успешное создание заказа - StatusProcessing",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Status: constants.StatusProcessing,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "успешное создание заказа - StatusProcessed",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Status:  constants.StatusProcessed,
					Accrual: 100.0,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "успешное создание заказа - StatusInvalid",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Status: constants.StatusInvalid,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "ошибка проверки заказа",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("rate limit"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "заказ уже существует",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{UserID: userID}, nil)
			},
			expectedErr: usecase.ErrOrderAlreadyExists,
		},
		{
			name:        "заказ принадлежит другому пользователю",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{UserID: 2}, nil)
			},
			expectedErr: usecase.ErrOrderBelongsToOtherUser,
		},
		{
			name:        "ошибка создания заказа",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("not found"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(errors.New("db error"))
			},
			expectedErr: errors.New("failed to create order: db error"),
		},
		{
			name:        "ошибка обновления баланса",
			orderNumber: validOrderNumber,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Status:  constants.StatusProcessed,
					Accrual: 100.0,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(errors.New("db error"))
			},
			expectedErr: errors.New("failed to update balance: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &testutils.MockOrderStorage{}
			lc := &testutils.MockLoyaltyClient{}
			bs := &testutils.MockBalanceStorage{}
			tt.setupMocks(os, lc, bs)

			balanceUC := usecase.NewBalanceUseCase(bs)
			uc := usecase.NewOrderUseCase(os, &MockLoyaltyClientWrapper{lc}, balanceUC)

			err := uc.ProcessNewOrder(ctx, userID, tt.orderNumber)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			os.AssertExpectations(t)
			lc.AssertExpectations(t)
			bs.AssertExpectations(t)
		})
	}
}

func TestOrderUseCaseGetUserOrders(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()

	tests := []struct {
		name           string
		setupMocks     func(*testutils.MockOrderStorage)
		expectedOrders []models.Order
		expectedErr    error
	}{
		{
			name: "успешное получение заказов",
			setupMocks: func(os *testutils.MockOrderStorage) {
				os.On("GetOrdersByUserID", mock.Anything, userID).Return([]models.Order{
					{UserID: userID, Number: "123"},
				}, nil)
			},
			expectedOrders: []models.Order{{UserID: userID, Number: "123"}},
			expectedErr:    nil,
		},
		{
			name: "нет заказов",
			setupMocks: func(os *testutils.MockOrderStorage) {
				os.On("GetOrdersByUserID", mock.Anything, userID).Return([]models.Order{}, nil)
			},
			expectedOrders: []models.Order{},
			expectedErr:    nil,
		},
		{
			name: "ошибка хранилища",
			setupMocks: func(os *testutils.MockOrderStorage) {
				os.On("GetOrdersByUserID", mock.Anything, userID).Return([]models.Order{}, errors.New("db error"))
			},
			expectedOrders: nil,
			expectedErr:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &testutils.MockOrderStorage{}
			tt.setupMocks(os)

			uc := usecase.NewOrderUseCase(os, nil, nil)
			orders, err := uc.GetUserOrders(ctx, userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrders, orders)
			}

			os.AssertExpectations(t)
		})
	}
}

func TestOrderUseCaseUpdateOrderStatus(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()

	tests := []struct {
		name        string
		order       models.Order
		prevStatus  string
		setupMocks  func(*testutils.MockOrderStorage, *testutils.MockBalanceStorage)
		expectedErr error
	}{
		{
			name: "успешное обновление заказа",
			order: models.Order{
				UserID:  userID,
				Number:  "123",
				Status:  constants.StatusProcessed,
				Accrual: pgtype.Float8{Float64: 100.0, Valid: true},
			},
			prevStatus: constants.StatusProcessing,
			setupMocks: func(os *testutils.MockOrderStorage, bs *testutils.MockBalanceStorage) {
				os.On("UpdateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "ошибка обновления заказа",
			order: models.Order{
				UserID: userID,
				Number: "123",
				Status: constants.StatusProcessed,
			},
			prevStatus: constants.StatusProcessing,
			setupMocks: func(os *testutils.MockOrderStorage, bs *testutils.MockBalanceStorage) {
				os.On("UpdateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(errors.New("db error"))
			},
			expectedErr: errors.New("failed to update order: db error"),
		},
		{
			name: "ошибка обновления баланса",
			order: models.Order{
				UserID:  userID,
				Number:  "123",
				Status:  constants.StatusProcessed,
				Accrual: pgtype.Float8{Float64: 100.0, Valid: true},
			},
			prevStatus: constants.StatusProcessing,
			setupMocks: func(os *testutils.MockOrderStorage, bs *testutils.MockBalanceStorage) {
				os.On("UpdateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(errors.New("db error"))
			},
			expectedErr: errors.New("failed to update balance for processed order: db error"),
		},
		{
			name: "без обновления баланса (тот же статус)",
			order: models.Order{
				UserID:  userID,
				Number:  "123",
				Status:  constants.StatusProcessed,
				Accrual: pgtype.Float8{Float64: 100.0, Valid: true},
			},
			prevStatus: constants.StatusProcessed,
			setupMocks: func(os *testutils.MockOrderStorage, bs *testutils.MockBalanceStorage) {
				os.On("UpdateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &testutils.MockOrderStorage{}
			bs := &testutils.MockBalanceStorage{}
			tt.setupMocks(os, bs)

			balanceUC := usecase.NewBalanceUseCase(bs)
			uc := usecase.NewOrderUseCase(os, &MockLoyaltyClientWrapper{nil}, balanceUC)

			err := uc.UpdateOrderStatus(ctx, tt.order, tt.prevStatus)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			os.AssertExpectations(t)
			bs.AssertExpectations(t)
		})
	}
}
