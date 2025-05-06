package testutils

import (
	"context"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

type MockOrderStorage struct {
	mock.Mock
}

func (m *MockOrderStorage) CreateOrder(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderStorage) GetOrderByNumber(ctx context.Context, number string) (models.Order, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockOrderStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderStorage) UpdateOrder(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderStorage) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Order), args.Error(1)
}

type MockBalanceStorage struct {
	mock.Mock
}

func (m *MockBalanceStorage) GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(pgtype.Float8), args.Get(1).(pgtype.Float8), args.Error(2)
}

func (m *MockBalanceStorage) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockBalanceStorage) UpdateWithdrawn(ctx context.Context, userID int64, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

type MockWithdrawalStorage struct {
	mock.Mock
}

func (m *MockWithdrawalStorage) CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	args := m.Called(ctx, withdrawal)
	return args.Error(0)
}

func (m *MockWithdrawalStorage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

type MockLoyaltyClient struct {
	mock.Mock
}

func (m *MockLoyaltyClient) CheckOrder(ctx context.Context, orderNumber string) (*models.LoyaltyResponse, error) {
	args := m.Called(ctx, orderNumber)
	return args.Get(0).(*models.LoyaltyResponse), args.Error(1)
}
