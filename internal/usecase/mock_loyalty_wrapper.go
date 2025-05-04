package usecase

import (
	"context"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
)

type MockLoyaltyClientWrapper struct {
	Client *testutils.MockLoyaltyClient
}

func (m *MockLoyaltyClientWrapper) CheckOrder(ctx context.Context, orderNumber string) (*models.LoyaltyResponse, error) {
	return m.Client.CheckOrder(ctx, orderNumber)
}
