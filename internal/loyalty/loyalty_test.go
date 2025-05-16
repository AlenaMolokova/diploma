package loyalty

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockOrderStorage struct {
	orders []models.Order
}

func (m *mockOrderStorage) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	return m.orders, nil
}

func (m *mockOrderStorage) UpdateOrder(ctx context.Context, order models.Order) error {
	for i, o := range m.orders {
		if o.ID == order.ID {
			m.orders[i] = order
			return nil
		}
	}
	return errors.New("order not found")
}

type mockBalanceUpdater struct {
	balances map[int64]pgtype.Float8
}

func (m *mockBalanceUpdater) GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error) {
	if balance, ok := m.balances[userID]; ok {
		return balance, pgtype.Float8{}, nil
	}
	return pgtype.Float8{}, pgtype.Float8{}, nil
}

func (m *mockBalanceUpdater) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	m.balances[userID] = pgtype.Float8{Float64: amount, Valid: true}
	return nil
}

func TestCheckOrder(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "Order processed",
			statusCode:     http.StatusOK,
			responseBody:   `{"order":"123","status":"PROCESSED","accrual":100.0}`,
			expectedStatus: "PROCESSED",
			expectError:    false,
		},
		{
			name:        "Order not found",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
		{
			name:        "Rate limit exceeded",
			statusCode:  http.StatusTooManyRequests,
			expectError: true,
		},
		{
			name:        "Order processing",
			statusCode:  http.StatusNoContent,
			expectError: true,
		},
		{
			name:        "Unexpected status code",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.CheckOrder(context.Background(), "123")

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if resp != nil && resp.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, resp.Status)
			}
		})
	}
}

func TestProcessOrder(t *testing.T) {
	order := models.Order{
		ID:     1,
		UserID: 1,
		Number: "123",
		Status: constants.StatusNew,
	}

	orderStorage := &mockOrderStorage{
		orders: []models.Order{order},
	}

	balanceUpdater := &mockBalanceUpdater{
		balances: make(map[int64]pgtype.Float8),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"123","status":"PROCESSED","accrual":100.0}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.processOrder(context.Background(), struct {
		OrderStorage
		BalanceUpdater
	}{orderStorage, balanceUpdater}, order)

	updatedOrder := orderStorage.orders[0]
	if updatedOrder.Status != constants.StatusProcessed {
		t.Errorf("Expected status %s, got %s", constants.StatusProcessed, updatedOrder.Status)
	}

	balance, _, _ := balanceUpdater.GetBalance(context.Background(), order.UserID)
	if !balance.Valid || balance.Float64 != 100.0 {
		t.Errorf("Expected balance 100.0, got %v", balance.Float64)
	}
}
