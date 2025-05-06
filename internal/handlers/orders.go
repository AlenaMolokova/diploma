package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/AlenaMolokova/diploma/internal/middleware"
	"github.com/AlenaMolokova/diploma/internal/usecase"
	"github.com/AlenaMolokova/diploma/internal/utils"
	"github.com/AlenaMolokova/diploma/internal/validation"
)

type OrderHandler struct {
	orderUC   *usecase.OrderUseCase
	validator validation.OrderValidator
}

func NewOrderHandler(orderUC *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		orderUC:   orderUC,
		validator: validation.NewLuhnValidator(),
	}
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

	if !h.validator.ValidateOrderNumber(orderNumber) {
		log.Printf("Order number '%s' failed validation", orderNumber)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}

	err = h.orderUC.ProcessNewOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderAlreadyExists) {
			log.Printf("Order %s already exists for user %d", orderNumber, userID)
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, usecase.ErrOrderBelongsToOtherUser) {
			log.Printf("Order %s belongs to another user", orderNumber)
			utils.WriteJSONError(w, http.StatusConflict, "Order already taken by another user")
			return
		}
		log.Printf("Failed to process order %s: %v", orderNumber, err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Order %s created for user %d", orderNumber, userID)
	w.WriteHeader(http.StatusAccepted)
}
