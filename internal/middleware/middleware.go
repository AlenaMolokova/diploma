package middleware

  import (
      "context"
      "log"
      "net/http"
      "strings"
      "github.com/AlenaMolokova/diploma/internal/utils"
      "github.com/golang-jwt/jwt/v5"
  )

  const UserIDKey = "user_id"

  func Auth(secret string) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              tokenString := r.Header.Get("Authorization")
              log.Printf("Middleware: received Authorization header: %s", tokenString)
              if !strings.HasPrefix(tokenString, "Bearer ") {
                  log.Printf("Middleware: missing or invalid Authorization header")
                  utils.WriteJSONError(w, http.StatusUnauthorized, "Missing or invalid token")
                  return
              }
              tokenString = strings.TrimPrefix(tokenString, "Bearer ")
              log.Printf("Middleware: extracted token: %s", tokenString)

              token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                  if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                      log.Printf("Middleware: unexpected signing method: %v", token.Header["alg"])
                      return nil, jwt.ErrSignatureInvalid
                  }
                  return []byte(secret), nil
              })
              if err != nil {
                  log.Printf("Middleware: failed to parse token: %v", err)
                  utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token")
                  return
              }
              if !token.Valid {
                  log.Printf("Middleware: token is invalid")
                  utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token")
                  return
              }

              claims, ok := token.Claims.(jwt.MapClaims)
              if !ok {
                  log.Printf("Middleware: invalid claims format")
                  utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token claims")
                  return
              }
              userID, ok := claims["user_id"].(float64)
              if !ok {
                  log.Printf("Middleware: invalid user_id in claims")
                  utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid user_id")
                  return
              }

              log.Printf("Middleware: authenticated user_id=%d", int64(userID))
              ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
              next.ServeHTTP(w, r.WithContext(ctx))
          })
      }
  }