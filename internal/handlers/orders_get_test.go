package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderGetter struct {
	mock.Mock
}

func (m *mockOrderGetter) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}

func TestOrderGetHandler(t *testing.T) {
	mockGetter := new(mockOrderGetter)
	handler := handlers.NewOrderGetHandler(mockGetter)

	now := time.Now()
	mockOrder := models.Order{
		Number:     "1234567890",
		Status:     "PROCESSED",
		Accrual:    pgtype.Float8{Float64: 120.5, Valid: true},
		UploadedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	t.Run("orders found", func(t *testing.T) {
		mockGetter.On("GetOrdersByUserID", mock.Anything, int64(1)).Return([]models.Order{mockOrder}, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		userData := map[middleware.UserID]interface{}{
			middleware.UserID("id"): int64(1),
		}
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey{}, userData))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp []handlers.OrderResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1, "expected one order in response")
		assert.Equal(t, "1234567890", resp[0].Number)
	})
}
