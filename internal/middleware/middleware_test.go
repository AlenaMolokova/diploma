package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	userID := int64(1)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			utils.WriteJSONError(w, http.StatusInternalServerError, "UserID not found")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"user_id":%d}`, userID)))
	})

	validToken := generateTestToken(t, userID, secret, time.Now().Add(time.Hour))
	expiredToken := generateTestToken(t, userID, secret, time.Now().Add(-time.Hour))

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "валидный токен",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user_id":1}`,
		},
		{
			name:           "отсутствует заголовок Authorization",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing or invalid Authorization header"}`,
		},
		{
			name:           "неверный формат заголовка",
			authHeader:     "Basic token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing or invalid Authorization header"}`,
		},
		{
			name:           "истёкший токен",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:           "невалидный токен",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware := AuthMiddleware(secret)(nextHandler)
			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetUserID(t *testing.T) {
	userID := int64(1)
	ctx := context.Background()

	tests := []struct {
		name       string
		ctxValue   interface{}
		expectedID int64
		expectedOK bool
	}{
		{
			name: "валидный userID",
			ctxValue: map[UserID]interface{}{
				UserID("id"): userID,
			},
			expectedID: userID,
			expectedOK: true,
		},
		{
			name:       "отсутствует userID",
			ctxValue:   nil,
			expectedID: 0,
			expectedOK: false,
		},
		{
			name:       "неверный тип данных",
			ctxValue:   "not a map",
			expectedID: 0,
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.ctxValue != nil {
				req = req.WithContext(context.WithValue(ctx, UserKey{}, tt.ctxValue))
			}
			userID, ok := GetUserID(req)
			assert.Equal(t, tt.expectedID, userID)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func generateTestToken(t *testing.T, userID int64, secret string, exp time.Time) string {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     float64(exp.Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	return tokenString
}
