package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceHandler(t *testing.T) {
	mockStorage := new(testutils.MockBalanceStorage)
	uc := usecase.NewBalanceUseCase(mockStorage)
	handler := handlers.NewBalanceHandler(uc)

	t.Run("success", func(t *testing.T) {
		mockStorage.On("GetBalance", mock.Anything, int64(1)).Return(
			pgtype.Float8{Float64: 100, Valid: true},
			pgtype.Float8{Float64: 20, Valid: true},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		userData := map[middleware.UserID]interface{}{
			middleware.UserID("id"): int64(1),
		}
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey{}, userData))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		mockStorage.On("GetBalance", mock.Anything, int64(2)).Return(
			pgtype.Float8{},
			pgtype.Float8{},
			errors.New("DB down"),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		userData := map[middleware.UserID]interface{}{
			middleware.UserID("id"): int64(2),
		}
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey{}, userData))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
