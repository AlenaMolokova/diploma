package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type WithdrawalsHandler struct {
	store models.WithdrawalStorage
}

func NewWithdrawalsHandler(store models.WithdrawalStorage) *WithdrawalsHandler {
	return &WithdrawalsHandler{store: store}
}

type WithdrawalResponse struct {
	Order       string             `json:"order"`
	Sum         float64            `json:"sum"`
	ProcessedAt pgtype.Timestamptz `json:"processed_at"`
}

func (h *WithdrawalsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	withdrawals, err := h.store.GetWithdrawalsByUserID(r.Context(), userID)
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
	for i, w := range withdrawals {
		response[i] = WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         w.Sum.Float64,
			ProcessedAt: w.ProcessedAt,
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