package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLoyaltyChecker struct {
	mock *testutils.MockLoyaltyClient
}

func (m *MockLoyaltyChecker) CheckOrder(ctx context.Context, orderNumber string) (*models.LoyaltyResponse, error) {
	return m.mock.CheckOrder(ctx, orderNumber)
}

func TestOrderHandlerServeHTTP(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()
	validOrderNumber := "4532015112830366"

	tests := []struct {
		name           string
		body           string
		userID         interface{}
		setupMocks     func(*testutils.MockOrderStorage, *testutils.MockLoyaltyClient, *testutils.MockBalanceStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "успешное создание заказа - StatusNew",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("not found"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "успешное создание заказа - StatusRegistered",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Order:  validOrderNumber,
					Status: constants.StatusRegistered,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "успешное создание заказа - StatusProcessing",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Order:  validOrderNumber,
					Status: constants.StatusProcessing,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "успешное создание заказа - StatusProcessed",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Order:   validOrderNumber,
					Status:  constants.StatusProcessed,
					Accrual: 100.0,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "успешное создание заказа - StatusInvalid",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Order:  validOrderNumber,
					Status: constants.StatusInvalid,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "ошибка проверки заказа",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("rate limit"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "",
		},
		{
			name:   "неавторизованный запрос",
			body:   validOrderNumber,
			userID: nil,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
		{
			name:   "пустой номер заказа",
			body:   "",
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Order number is required"}`,
		},
		{
			name:   "неверный Luhn номер",
			body:   "4532015112830367",
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"Invalid order number"}`,
		},
		{
			name:   "заказ уже существует",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{UserID: userID}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:   "заказ принадлежит другому пользователю",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{UserID: 2}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"Order already taken by another user"}`,
		},
		{
			name:   "внутренняя ошибка создания заказа",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return((*models.LoyaltyResponse)(nil), errors.New("not found"))
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name:   "ошибка обновления баланса",
			body:   validOrderNumber,
			userID: userID,
			setupMocks: func(os *testutils.MockOrderStorage, lc *testutils.MockLoyaltyClient, bs *testutils.MockBalanceStorage) {
				os.On("GetOrderByNumber", mock.Anything, validOrderNumber).Return(models.Order{}, errors.New("not found"))
				lc.On("CheckOrder", mock.Anything, validOrderNumber).Return(&models.LoyaltyResponse{
					Order:   validOrderNumber,
					Status:  constants.StatusProcessed,
					Accrual: 100.0,
				}, nil)
				os.On("CreateOrder", mock.Anything, mock.AnythingOfType("models.Order")).Return(nil)
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &testutils.MockOrderStorage{}
			lc := &testutils.MockLoyaltyClient{}
			bs := &testutils.MockBalanceStorage{}
			tt.setupMocks(os, lc, bs)

			loyaltyChecker := &MockLoyaltyChecker{mock: lc}
			balanceUC := usecase.NewBalanceUseCase(bs)
			uc := usecase.NewOrderUseCase(os, loyaltyChecker, balanceUC)
			handler := NewOrderHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.body))
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(ctx, middleware.UserKey{}, map[middleware.UserID]interface{}{
					middleware.UserID("id"): tt.userID,
				}))
			} else {
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			os.AssertExpectations(t)
			lc.AssertExpectations(t)
			bs.AssertExpectations(t)
		})
	}
}
