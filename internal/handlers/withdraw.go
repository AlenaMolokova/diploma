package handlers

import (
	
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type WithdrawHandler struct {
	store   WithdrawalStorage
	balance BalanceStorage
}

func NewWithdrawHandler(store WithdrawalStorage, balance BalanceStorage) *WithdrawHandler {
	return &WithdrawHandler{store: store, balance: balance}
}

func (h *WithdrawHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode withdraw request: %v", err)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		log.Printf("Invalid withdraw request: order=%s, sum=%.2f", req.Order, req.Sum)
		utils.WriteJSONError(w, http.StatusBadRequest, "Order and positive sum are required")
		return
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(req.Order) {
		log.Printf("Invalid order number format: %s", req.Order)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Order number must be digits")
		return
	}

	if !LuhnCheck(req.Order) {
		log.Printf("Order number %s failed Luhn check", req.Order)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}

	current, _, err := h.balance.GetBalance(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get balance for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !current.Valid || current.Float64 < req.Sum {
		log.Printf("Insufficient balance for user %d: current=%.2f, requested=%.2f", userID, current.Float64, req.Sum)
		utils.WriteJSONError(w, http.StatusPaymentRequired, "Insufficient balance")
		return
	}

	withdrawal := storage.Withdrawal{
		UserID:      pgtype.Int8{Int64: userID, Valid: true},
		OrderNumber: req.Order,
		Sum:         req.Sum,
		ProcessedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if err := h.balance.UpdateBalance(r.Context(), userID, -req.Sum); err != nil {
		log.Printf("Failed to update balance for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.store.CreateWithdrawal(r.Context(), withdrawal); err != nil {
		log.Printf("Failed to create withdrawal for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Withdrawal of %.2f for order %s by user %d successful", req.Sum, req.Order, userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}