package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) CreateUser(ctx context.Context, login, password string) (int64, error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(models.User), args.Error(1)
}

func TestLoginHandler_ServeHTTP(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()
	jwtSecret := "test-secret"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		body           string
		setupMocks     func(*MockUserStorage)
		expectedStatus int
		expectedBody   string
		expectedToken  bool
	}{
		{
			name: "успешный логин",
			body: `{"login":"testuser","password":"testpass"}`,
			setupMocks: func(us *MockUserStorage) {
				us.On("GetUserByLogin", mock.Anything, "testuser").Return(models.User{
					ID:       userID,
					Login:    "testuser",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			expectedToken:  true,
		},
		{
			name: "неверный логин",
			body: `{"login":"testuser","password":"testpass"}`,
			setupMocks: func(us *MockUserStorage) {
				us.On("GetUserByLogin", mock.Anything, "testuser").Return(models.User{}, sql.ErrNoRows)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid login or password"}`,
			expectedToken:  false,
		},
		{
			name: "неверный пароль",
			body: `{"login":"testuser","password":"wrongpass"}`,
			setupMocks: func(us *MockUserStorage) {
				us.On("GetUserByLogin", mock.Anything, "testuser").Return(models.User{
					ID:       userID,
					Login:    "testuser",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid login or password"}`,
			expectedToken:  false,
		},
		{
			name:           "невалидный JSON",
			body:           `{"login":"testuser","password":"testpass"`,
			setupMocks:     func(us *MockUserStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request format"}`,
			expectedToken:  false,
		},
		{
			name:           "пустой логин",
			body:           `{"login":"","password":"testpass"}`,
			setupMocks:     func(us *MockUserStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Login and password are required"}`,
			expectedToken:  false,
		},
		{
			name:           "пустой пароль",
			body:           `{"login":"testuser","password":""}`,
			setupMocks:     func(us *MockUserStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Login and password are required"}`,
			expectedToken:  false,
		},
		{
			name: "внутренняя ошибка хранилища",
			body: `{"login":"testuser","password":"testpass"}`,
			setupMocks: func(us *MockUserStorage) {
				us.On("GetUserByLogin", mock.Anything, "testuser").Return(models.User{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
			expectedToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &MockUserStorage{}
			tt.setupMocks(us)

			handler := NewLoginHandler(us, jwtSecret)
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}

			if tt.expectedToken {
				tokenHeader := w.Header().Get("Authorization")
				assert.NotEmpty(t, tokenHeader)
				assert.True(t, strings.HasPrefix(tokenHeader, "Bearer "))
				tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)
				claims, ok := token.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, float64(userID), claims["user_id"])
			} else {
				assert.Empty(t, w.Header().Get("Authorization"))
			}

			us.AssertExpectations(t)
		})
	}
}
