package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderHandler struct {
	store        models.OrderStorage
	balance      models.BalanceStorage
	loyaltyCheck *loyalty.Client
}

func NewOrderHandler(store models.OrderStorage, balance models.BalanceStorage, loyaltyCheck *loyalty.Client) *OrderHandler {
	return &OrderHandler{store: store, balance: balance, loyaltyCheck: loyaltyCheck}
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Unauthorized: missing user_id in context")
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		log.Printf("Empty order number")
		utils.WriteJSONError(w, http.StatusBadRequest, "Order number is required")
		return
	}

	if !utils.LuhnCheck(orderNumber) {
		log.Printf("Order number '%s' failed Luhn check", orderNumber)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}

	existingOrder, err := h.store.GetOrderByNumber(r.Context(), orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			log.Printf("Order %s already exists for user %d", orderNumber, userID)
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Printf("Order %s belongs to another user", orderNumber)
		utils.WriteJSONError(w, http.StatusConflict, "Order already taken by another user")
		return
	}

	loyaltyResp, err := h.loyaltyCheck.CheckOrder(r.Context(), orderNumber)
	if err != nil {
		if errors.Is(err, loyalty.ErrOrderNotFound) {
			log.Printf("Loyalty check failed for order %s: %v", orderNumber, err)
		} else if errors.Is(err, loyalty.ErrRateLimit) {
			log.Printf("Rate limit exceeded for order %s: %v", orderNumber, err)
			utils.WriteJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		} else {
			log.Printf("Failed to check loyalty for order %s: %v", orderNumber, err)
			utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	order := models.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     "NEW",
		UploadedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if loyaltyResp != nil {
		switch loyaltyResp.Status {
		case "REGISTERED", "PROCESSING":
			order.Status = loyaltyResp.Status
		case "PROCESSED":
			order.Status = loyaltyResp.Status
			order.Accrual = pgtype.Float8{Float64: loyaltyResp.Accrual, Valid: true}
		case "INVALID":
			order.Status = loyaltyResp.Status
		}

		current, _, err := h.balance.GetBalance(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to get balance for user %d: %v", userID, err)
			utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		newBalance := current.Float64 + loyaltyResp.Accrual
		if err := h.balance.UpdateBalance(r.Context(), userID, newBalance); err != nil {
			log.Printf("Failed to update balance for user %d: %v", userID, err)
		 utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		log.Printf("Accrued %.2f points for user %d for order %s, new balance: %.2f", loyaltyResp.Accrual, userID, orderNumber, newBalance)
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		log.Printf("Failed to create order %s for user %d: %v", orderNumber, userID, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Order %s created for user %d with status %s", orderNumber, userID, order.Status)
	w.WriteHeader(http.StatusAccepted)
}