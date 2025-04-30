package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

type UserID string

type userKey struct{}

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			log.Printf("Middleware: received Authorization header: %s", authHeader)

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				log.Printf("Middleware: missing or invalid Authorization header")
				utils.WriteJSONError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				log.Printf("Middleware: invalid token: %v", err)
				utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("Middleware: invalid claims")
				utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			exp, ok := claims["exp"].(float64)
			if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
				log.Printf("Middleware: token expired or invalid exp claim")
				utils.WriteJSONError(w, http.StatusUnauthorized, "Token expired or invalid")
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				log.Printf("Middleware: user_id not found in claims")
				utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}
			userID := int64(userIDFloat)

			userData := map[UserID]interface{}{
				UserID("id"): userID,
			}
			ctx := context.WithValue(r.Context(), userKey{}, userData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (int64, bool) {
	userData, ok := r.Context().Value(userKey{}).(map[UserID]interface{})
	if !ok {
		return 0, false
	}
	userID, ok := userData[UserID("id")].(int64)
	return userID, ok
}