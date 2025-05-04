package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserCreator struct {
	mock.Mock
}

func (m *mockUserCreator) CreateUser(ctx context.Context, login, password string) (int64, error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	mockStore := new(mockUserCreator)
	secret := "testsecret"
	handler := handlers.NewRegisterHandler(mockStore, secret)

	t.Run("successful registration", func(t *testing.T) {
		mockStore.On("CreateUser", mock.Anything, "newuser", mock.Anything).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"login":"newuser", "password":"securepass"}`))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Authorization"), "Bearer ")
		mockStore.AssertExpectations(t)
	})

	t.Run("weak password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"login":"newuser", "password":"123"}`))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate user", func(t *testing.T) {
		mockStore.On("CreateUser", mock.Anything, "exists", mock.Anything).Return(int64(0), errors.New("login already exists"))

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"login":"exists", "password":"securepass"}`))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}
