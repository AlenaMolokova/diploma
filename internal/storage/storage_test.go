package storage

import (
	"context"
	"testing"
	"time"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/AlenaMolokova/diploma/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type StorageTester interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
	CreateOrder(ctx context.Context, order models.Order) error
}

type mockStorage struct {
	*testutils.MockBalanceStorage
	*testutils.MockOrderStorage
}

func (m *mockStorage) CreateUser(ctx context.Context, login, password string) (int64, error) {
	return 1, nil
}

func TestStorage(t *testing.T) {
	ctx := context.Background()
	mockBalanceStorage := &testutils.MockBalanceStorage{}
	mockOrderStorage := &testutils.MockOrderStorage{}
	store := &mockStorage{
		MockBalanceStorage: mockBalanceStorage,
		MockOrderStorage:   mockOrderStorage,
	}

	userID := int64(1)
	login := "testuser"
	password := "securepass"
	balance := pgtype.Float8{Float64: 100.0, Valid: true}
	withdrawn := pgtype.Float8{Float64: 20.0, Valid: true}
	orderNumber := "4532015112830366"
	uploadedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	tests := []struct {
		name        string
		setupMocks  func()
		testFunc    func() error
		expectedErr error
	}{
		{
			name: "успешное создание пользователя",
			setupMocks: func() {
			},
			testFunc: func() error {
				_, err := store.CreateUser(ctx, login, password)
				return err
			},
			expectedErr: nil,
		},
		{
			name: "успешное получение баланса",
			setupMocks: func() {
				mockBalanceStorage.On("GetBalance", ctx, userID).
					Return(balance, withdrawn, nil)
			},
			testFunc: func() error {
				bal, with, err := store.GetBalance(ctx, userID)
				if err != nil {
					return err
				}
				assert.Equal(t, balance, bal)
				assert.Equal(t, withdrawn, with)
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "успешное обновление баланса",
			setupMocks: func() {
				mockBalanceStorage.On("UpdateBalance", ctx, userID, 100.0).
					Return(nil)
			},
			testFunc: func() error {
				return store.UpdateBalance(ctx, userID, 100.0)
			},
			expectedErr: nil,
		},
		{
			name: "успешное создание заказа",
			setupMocks: func() {
				mockOrderStorage.On("CreateOrder", ctx, models.Order{
					UserID:     userID,
					Number:     orderNumber,
					Status:     "NEW",
					UploadedAt: uploadedAt,
				}).Return(nil)
			},
			testFunc: func() error {
				return store.CreateOrder(ctx, models.Order{
					UserID:     userID,
					Number:     orderNumber,
					Status:     "NEW",
					UploadedAt: uploadedAt,
				})
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := tt.testFunc()
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockBalanceStorage.AssertExpectations(t)
			mockOrderStorage.AssertExpectations(t)
		})
	}
}
