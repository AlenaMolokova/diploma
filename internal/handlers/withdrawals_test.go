package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithdrawalsHandlerServeHTTP(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()

	tests := []struct {
		name           string
		userID         interface{}
		setupMocks     func(*testutils.MockWithdrawalStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "успешное получение списка списаний",
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{
					{
						UserID:      userID,
						OrderNumber: "79927398713",
						Sum:         pgtype.Float8{Float64: 100.0, Valid: true},
						ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"order":"79927398713","sum":100.0,"processed_at":"` + time.Now().Format(time.RFC3339) + `"}]`,
		},
		{
			name:   "нет списаний",
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "неавторизованный запрос",
			userID:         nil,
			setupMocks:     func(ws *testutils.MockWithdrawalStorage) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
		{
			name:   "внутренняя ошибка",
			userID: userID,
			setupMocks: func(ws *testutils.MockWithdrawalStorage) {
				ws.On("GetWithdrawalsByUserID", mock.Anything, userID).Return([]models.Withdrawal{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutils.MockWithdrawalStorage{}
			tt.setupMocks(ws)

			uc := usecase.NewWithdrawalUseCase(ws, nil)
			handler := NewWithdrawalsHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance/withdrawals", nil)
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
		})
	}
}
