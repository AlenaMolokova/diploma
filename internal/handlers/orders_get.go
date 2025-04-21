package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderGetHandler struct {
	store OrderStorage
}

func NewOrderGetHandler(store OrderStorage) *OrderGetHandler {
	return &OrderGetHandler{store: store}
}

type OrderResponse struct {
	Number     string              `json:"number"`
	Status     string              `json:"status"`
	Accrual    float64             `json:"accrual,omitempty"`
	UploadedAt pgtype.Timestamptz  `json:"uploaded_at"`
}

func (h *OrderGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.store.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get orders for user %d: %v", userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(orders) == 0 {
		log.Printf("No orders found for user %d", userID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, order := range orders {
		resp := OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}
		if order.Accrual.Valid {
			resp.Accrual = order.Accrual.Float64
		}
		response[i] = resp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode orders response: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	log.Printf("Returned %d orders for user %d", len(orders), userID)
}