package handlers

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AlenaMolokova/diploma/internal/loyalty"
	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/storage"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func LuhnCheck(number string) bool {
	var sum int
	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		digit := int(number[i] - '0')
		if (len(number)-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

type OrderHandler struct {
	store   OrderStorage
	balance BalanceStorage
	loyalty *loyalty.Client
}

func NewOrderHandler(store OrderStorage, balance BalanceStorage, loyalty *loyalty.Client) *OrderHandler {
	return &OrderHandler{store: store, balance: balance, loyalty: loyalty}
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
		utils.WriteJSONError(w, http.StatusBadRequest, "Cannot read request body")
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		log.Printf("Empty order number")
		utils.WriteJSONError(w, http.StatusBadRequest, "Order number is required")
		return
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(number) {
		log.Printf("Invalid order number format: '%s'", number)
		utils.WriteJSONError(w, http.StatusBadRequest, "Order number must be digits")
		return
	}

	if !LuhnCheck(number) {
		log.Printf("Order number '%s' failed Luhn check", number)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}

	existingOrder, err := h.store.GetOrderByNumber(r.Context(), number)
	if err == nil {
		if existingOrder.UserID.Valid && existingOrder.UserID.Int64 == userID {
			log.Printf("Order %s already exists for user %d", number, userID)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Printf("Order %s belongs to another user", number)
		utils.WriteJSONError(w, http.StatusConflict, "Order belongs to another user")
		return
	}
	if err != pgx.ErrNoRows {
		log.Printf("Failed to check order %s: %v", number, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	order := storage.Order{
		UserID:     pgtype.Int8{Int64: userID, Valid: true},
		Number:     number,
		Status:     "NEW",
		UploadedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			existingOrder, err := h.store.GetOrderByNumber(r.Context(), number)
			if err == nil {
				if existingOrder.UserID.Valid && existingOrder.UserID.Int64 == userID {
					log.Printf("Order %s already exists for user %d after retry", number, userID)
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					return
				}
				log.Printf("Order %s belongs to another user after retry", number)
				utils.WriteJSONError(w, http.StatusConflict, "Order belongs to another user")
				return
			}
			log.Printf("Unexpected error after duplicate: %v", err)
		}
		log.Printf("Failed to create order %s: %v", number, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	loyaltyResp, err := h.loyalty.CheckOrder(r.Context(), number)
	if err == nil {
		order.Status = loyaltyResp.Status
		if loyaltyResp.Accrual > 0 {
			order.Accrual = pgtype.Float8{Float64: loyaltyResp.Accrual, Valid: true}
			if order.Status == "PROCESSED" {
				current, _, err := h.balance.GetBalance(r.Context(), userID)
				if err != nil {
					log.Printf("Failed to get balance for user %d: %v", userID, err)
					utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
					return
				}
				newBalance := current.Float64 + loyaltyResp.Accrual
				if err := h.balance.UpdateBalance(r.Context(), userID, loyaltyResp.Accrual); err != nil {
					log.Printf("Failed to update balance for user %d: %v", userID, err)
					utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
					return
				}
				log.Printf("Accrued %.2f points for user %d for order %s, new balance: %.2f", loyaltyResp.Accrual, userID, number, newBalance)
			}
		}
	} else {
		log.Printf("Loyalty check failed for order %s: %v, proceeding with NEW status", number, err)
	}

	log.Printf("Order %s created for user %d with status %s", number, userID, order.Status)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
}