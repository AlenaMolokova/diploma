package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrOrderNotFound   = fmt.Errorf("order not found")
	ErrRateLimit       = fmt.Errorf("rate limit exceeded")
	ErrOrderProcessing = fmt.Errorf("order is still processing")
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (c *Client) CheckOrder(ctx context.Context, orderNumber string) (*accrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/orders/"+orderNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for order %s: %v", orderNumber, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order %s: %v", orderNumber, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
			return nil, fmt.Errorf("failed to decode response for order %s: %v", orderNumber, err)
		}
		return &accrual, nil
	case http.StatusNotFound:
		return nil, ErrOrderNotFound
	case http.StatusTooManyRequests:
		return nil, ErrRateLimit
	case http.StatusNoContent:
		return nil, ErrOrderProcessing
	default:
		return nil, fmt.Errorf("unexpected status code for order %s: %d", orderNumber, resp.StatusCode)
	}
}

type balanceUpdater interface {
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
}

func (c *Client) StartOrderProcessing(ctx context.Context, store models.OrderStorage) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	balanceStore, ok := store.(balanceUpdater)
	if !ok {
		log.Println("Store does not implement balance update interface")
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Order processing stopped")
			return
		case <-ticker.C:
			orders, err := store.GetAllOrders(ctx)
			if err != nil {
				log.Printf("Failed to get orders: %v", err)
				continue
			}

			for _, order := range orders {
				if order.Status == "PROCESSED" || order.Status == "INVALID" {
					continue
				}

				resp, err := c.CheckOrder(ctx, order.Number)
				if err != nil {
					if errors.Is(err, ErrOrderProcessing) {
						continue
					}
					log.Printf("Failed to check order %s: %v", order.Number, err)
					continue
				}

				prevStatus := order.Status
				updatedOrder := models.Order{
					Number:     order.Number,
					Status:     resp.Status,
					Accrual:    pgtype.Float8{Float64: resp.Accrual, Valid: resp.Accrual > 0},
					UploadedAt: order.UploadedAt,
				}

				if err := store.UpdateOrder(ctx, updatedOrder); err != nil {
					log.Printf("Failed to update order %s: %v", order.Number, err)
					continue
				}

				log.Printf("Updated order %s: status=%s, accrual=%.2f", order.Number, resp.Status, resp.Accrual)

				if ok && resp.Status == "PROCESSED" && prevStatus != "PROCESSED" && resp.Accrual > 0 {
					if err := updateUserBalance(ctx, balanceStore, order.UserID, resp.Accrual); err != nil {
						log.Printf("Failed to update balance for user %d: %v", order.UserID, err)
					}
				}
			}
		}
	}
}

func updateUserBalance(ctx context.Context, store balanceUpdater, userID int64, accrual float64) error {
	current, _, err := store.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	newBalance := accrual
	if current.Valid {
		newBalance += current.Float64
	}

	if err := store.UpdateBalance(ctx, userID, newBalance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	log.Printf("Updated balance for user %d: added %.2f, new balance: %.2f", userID, accrual, newBalance)
	return nil
}