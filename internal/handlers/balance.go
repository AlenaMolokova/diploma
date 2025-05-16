package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/AlenaMolokova/diploma/internal/utils"
)

type BalanceHandler struct {
	balanceUC usecase.BalanceUseCase
}

func NewBalanceHandler(balanceUC usecase.BalanceUseCase) *BalanceHandler {
	return &BalanceHandler{balanceUC: balanceUC}
}

func (h *BalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	current, withdrawn, err := h.balanceUC.GetUserBalance(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get balance for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := map[string]float64{
		"current":   current,
		"withdrawn": withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode balance response: %v", err)
	}
	log.Printf("Returned balance current=%.2f, withdrawn=%.2f for user %d", current, withdrawn, userID)
}
