package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode"

	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterHandler struct {
	store  UserStorage
	secret string
}

func NewRegisterHandler(store UserStorage, secret string) *RegisterHandler {
	return &RegisterHandler{store: store, secret: secret}
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		log.Printf("Password validation failed: length %d < 8", len(password))
		return false
	}
	hasLetter := false
	for i, c := range password {
		if unicode.IsLetter(c) {
			hasLetter = true
		}
		log.Printf("Password char %d: %c, isLetter=%v", i, c, unicode.IsLetter(c))
	}
	if !hasLetter {
		log.Printf("Password validation failed: no letters")
		return false
	}
	log.Printf("Password validation passed: %s", password)
	return true
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

	if !isValidPassword(req.Password) {
		log.Printf("Invalid password for login %s: must be >=8 chars with letters", req.Login)
		utils.WriteJSONError(w, http.StatusBadRequest, "Password must be at least 8 characters long and contain letters")
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