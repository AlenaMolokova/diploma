package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/AlenaMolokova/diploma/internal/utils"
)

type WithdrawHandler struct {
	withdrawalUC *usecase.WithdrawalUseCase
}

func NewWithdrawHandler(withdrawalUC *usecase.WithdrawalUseCase) *WithdrawHandler {
	return &WithdrawHandler{
		withdrawalUC: withdrawalUC,
	}
}

func (h *WithdrawHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
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

	err := h.withdrawalUC.ProcessWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err.Error() == "insufficient balance" {
			log.Printf("Insufficient balance for user %d: requested=%.2f", userID, req.Sum)
			utils.WriteJSONError(w, http.StatusPaymentRequired, "Insufficient balance")
			return
		}
		if err.Error() == "invalid order number" {
			log.Printf("Invalid order number: '%s'", req.Order)
			utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid order number")
			return
		}
		log.Printf("Failed to process withdrawal for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Withdrawal of %.2f for order %s by user %d successful", req.Sum, req.Order, userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
