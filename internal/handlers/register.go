package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/AlenaMolokova/diploma/internal/validation"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserCreator interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
}

type RegisterHandler struct {
	store     UserCreator
	secret    string
	validator validation.PasswordValidator
}

func NewRegisterHandler(store UserCreator, secret string) *RegisterHandler {
	return &RegisterHandler{
		store:     store,
		secret:    secret,
		validator: validation.NewDefaultPasswordValidator(),
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode register request: %v", err)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	log.Printf("Register request: login=%s", req.Login)

	if req.Login == "" || req.Password == "" {
		log.Printf("Empty login or password")
		utils.WriteJSONError(w, http.StatusBadRequest, "Login and password are required")
		return
	}

	if !h.validator.ValidatePassword(req.Password) {
		log.Printf("Invalid password for login %s: must be >=8 chars", req.Login)
		utils.WriteJSONError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for login %s: %v", req.Login, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userID, err := h.store.CreateUser(r.Context(), req.Login, string(hashedPassword))
	if err != nil {
		if err.Error() == "login already exists" {
			log.Printf("Login %s already exists", req.Login)
			utils.WriteJSONError(w, http.StatusConflict, "Login already exists")
			return
		}
		log.Printf("Failed to create user %s: %v", req.Login, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.secret))
	if err != nil {
		log.Printf("Failed to sign token for user %s: %v", req.Login, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
	log.Printf("User %s registered successfully, user_id: %d", req.Login, userID)
}
