package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/utils"
)

type WithdrawalUseCase interface {
	GetUserWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type WithdrawalsHandler struct {
	withdrawalUC WithdrawalUseCase
}

func NewWithdrawalsHandler(withdrawalUC WithdrawalUseCase) *WithdrawalsHandler {
	return &WithdrawalsHandler{withdrawalUC: withdrawalUC}
}

type WithdrawalResponse struct {
	Order       string `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}

func (h *WithdrawalsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	withdrawals, err := h.withdrawalUC.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get withdrawals for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(withdrawals) == 0 {
		log.Printf("No withdrawals found for user %d", userID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]WithdrawalResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		response[i] = WithdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum.Float64,
			ProcessedAt: withdrawal.ProcessedAt.Time.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode withdrawals response: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	log.Printf("Returned %d withdrawals for user %d", len(withdrawals), userID)
}