package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithdrawHandlerServeHTTP(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()
	validOrderNumber := "4532015112830366"

	tests := []struct {
		name           string
		body           string
		userID         interface{}
		setupMocks     func(*testutils.MockWithdrawalStorage, *testutils.MockBalanceStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "успешное снятие средств",
			body:   `{"order":"` + validOrderNumber + `","sum":100.0}`,
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 200.0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
				bs.On("UpdateWithdrawn", mock.Anything, userID, 100.0).Return(nil)
				ws.On("CreateWithdrawal", mock.Anything, mock.AnythingOfType("models.Withdrawal")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "неавторизованный запрос",
			body:           `{"order":"` + validOrderNumber + `","sum":100.0}`,
			userID:         nil,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
		{
			name:           "неверный формат запроса",
			body:           `invalid json`,
			userID:         userID,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request format"}`,
		},
		{
			name:           "пустой order или положительная сумма",
			body:           `{"order":"","sum":0}`,
			userID:         userID,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Order and positive sum are required"}`,
		},
		{
			name:           "номер заказа не из цифр",
			body:           `{"order":"123abc","sum":100.0}`,
			userID:         userID,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"Order number must be digits"}`,
		},
		{
			name:           "номер заказа не проходит проверку Луна",
			body:           `{"order":"4532015112830367","sum":100.0}`,
			userID:         userID,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"Invalid order number"}`,
		},
		{
			name:   "недостаточный баланс",
			body:   `{"order":"` + validOrderNumber + `","sum":100.0}`,
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 50.0, Valid: true}, pgtype.Float8{}, nil)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   `{"error":"Insufficient balance"}`,
		},
		{
			name:   "внутренняя ошибка",
			body:   `{"order":"` + validOrderNumber + `","sum":100.0}`,
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage, bs *testutils.MockBalanceStorage) {
				bs.On("GetBalance", mock.Anything, userID).Return(pgtype.Float8{Float64: 200.0, Valid: true}, pgtype.Float8{}, nil)
				bs.On("UpdateBalance", mock.Anything, userID, 100.0).Return(nil)
				bs.On("UpdateWithdrawn", mock.Anything, userID, 100.0).Return(nil)
				ws.On("CreateWithdrawal", mock.Anything, mock.AnythingOfType("models.Withdrawal")).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutils.MockWithdrawalStorage{}
			bs := &testutils.MockBalanceStorage{}
			tt.setupMocks(ws, bs)

			balanceUC := usecase.NewBalanceUseCase(bs)
			uc := usecase.NewWithdrawalUseCase(ws, balanceUC)
			handler := NewWithdrawHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(tt.body))
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

			ws.AssertExpectations(t)
			bs.AssertExpectations(t)
		})
	}
}
