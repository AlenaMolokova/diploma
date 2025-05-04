package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrOrderAlreadyExists      = errors.New("order already exists for this user")
	ErrOrderBelongsToOtherUser = errors.New("order belongs to another user")
	ErrInvalidOrderNumber      = errors.New("invalid order number")
)

type OrderStorage interface {
	CreateOrder(ctx context.Context, order models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
}

type OrderUseCase struct {
	storage      OrderStorage
	loyaltyCheck LoyaltyChecker
	balanceUC    *BalanceUseCase
}

func NewOrderUseCase(storage OrderStorage, loyaltyCheck LoyaltyChecker, balanceUC *BalanceUseCase) *OrderUseCase {
	return &OrderUseCase{
		storage:      storage,
		loyaltyCheck: loyaltyCheck,
		balanceUC:    balanceUC,
	}
}

func (uc *OrderUseCase) ProcessNewOrder(ctx context.Context, userID int64, orderNumber string) error {
	existingOrder, err := uc.storage.GetOrderByNumber(ctx, orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			return ErrOrderAlreadyExists
		}
		return ErrOrderBelongsToOtherUser
	}

	loyaltyResp, err := uc.loyaltyCheck.CheckOrder(ctx, orderNumber)

	order := models.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     constants.StatusNew,
		UploadedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if err == nil && loyaltyResp != nil {
		switch loyaltyResp.Status {
		case constants.StatusRegistered, constants.StatusProcessing:
			order.Status = loyaltyResp.Status
		case constants.StatusProcessed:
			order.Status = loyaltyResp.Status
			order.Accrual = pgtype.Float8{Float64: loyaltyResp.Accrual, Valid: true}
		case constants.StatusInvalid:
			order.Status = loyaltyResp.Status
		}
	}

	if err := uc.storage.CreateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	if loyaltyResp != nil && loyaltyResp.Status == constants.StatusProcessed && loyaltyResp.Accrual > 0 {
		if err := uc.balanceUC.AddToBalance(ctx, userID, loyaltyResp.Accrual); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}
	}

	return nil
}

func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	return uc.storage.GetOrdersByUserID(ctx, userID)
}

func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, order models.Order, prevStatus string) error {
	err := uc.storage.UpdateOrder(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if order.Status == constants.StatusProcessed &&
		prevStatus != constants.StatusProcessed &&
		order.Accrual.Valid &&
		order.Accrual.Float64 > 0 {
		if err := uc.balanceUC.AddToBalance(ctx, order.UserID, order.Accrual.Float64); err != nil {
			return fmt.Errorf("failed to update balance for processed order: %w", err)
		}
	}

	return nil
}
