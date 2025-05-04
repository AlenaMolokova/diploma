package usecase

import (
	"context"

	"github.com/AlenaMolokova/diploma/internal/models"
)

type LoyaltyChecker interface {
	CheckOrder(ctx context.Context, orderNumber string) (*models.LoyaltyResponse, error)
}
