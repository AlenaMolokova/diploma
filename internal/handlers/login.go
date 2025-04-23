package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"database/sql"

	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	store     models.UserStorage
	jwtSecret string
}

func NewLoginHandler(store models.UserStorage, jwtSecret string) *LoginHandler {
	return &LoginHandler{store: store, jwtSecret: jwtSecret}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode login request: %v", err)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Login == "" || req.Password == "" {
		log.Printf("Empty login or password")
		utils.WriteJSONError(w, http.StatusBadRequest, "Login and password are required")
		return
	}

	user, err := h.store.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", req.Login)
			utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid login or password")
		} else {
			log.Printf("Failed to get user %s: %v", req.Login, err)
			utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Invalid password for user %s", req.Login)
		utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid login or password")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		log.Printf("Failed to sign token: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
	log.Printf("User %s authenticated", req.Login)
}