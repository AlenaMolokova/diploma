package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrOrderNotFound   = fmt.Errorf("order not found")
	ErrRateLimit       = fmt.Errorf("rate limit exceeded")
	ErrOrderProcessing = fmt.Errorf("order is still processing")
)

type OrderStorage interface {
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
}

type BalanceUpdater interface {
	GetBalance(ctx context.Context, userID int64) (pgtype.Float8, pgtype.Float8, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
}

type Client struct {
	baseURL      string
	client       *http.Client
	pollInterval time.Duration
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		pollInterval: time.Duration(constants.DefaultPollInterval) * time.Second,
	}
}

func (c *Client) SetPollInterval(seconds int) {
	c.pollInterval = time.Duration(seconds) * time.Second
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

func (c *Client) StartOrderProcessing(ctx context.Context, store OrderStorage) {
	log.Printf("Starting order processing with interval: %v", c.pollInterval)
	
	for ticker := time.NewTicker(c.pollInterval); ; {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Println("Order processing stopped")
			return
		case <-ticker.C:
			c.processOrders(ctx, store)
		}
	}
}

func (c *Client) processOrders(ctx context.Context, store OrderStorage) {
	orders, err := store.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
		return
	}

	for _, order := range orders {
		c.processOrder(ctx, store, order)
	}
}

func (c *Client) processOrder(ctx context.Context, store OrderStorage, order models.Order) {
	if order.Status == constants.StatusProcessed || order.Status == constants.StatusInvalid {
		return
	}

	resp, err := c.CheckOrder(ctx, order.Number)
	if err != nil {
		if errors.Is(err, ErrOrderProcessing) {
			return
		}
		log.Printf("Failed to check order %s: %v", order.Number, err)
		return
	}

	prevStatus := order.Status
	updatedOrder := models.Order{
		ID:         order.ID,
		UserID:     order.UserID,
		Number:     order.Number,
		Status:     resp.Status,
		Accrual:    pgtype.Float8{Float64: resp.Accrual, Valid: resp.Accrual > 0},
		UploadedAt: order.UploadedAt,
	}

	if err := store.UpdateOrder(ctx, updatedOrder); err != nil {
		log.Printf("Failed to update order %s: %v", order.Number, err)
		return
	}

	log.Printf("Updated order %s: status=%s, accrual=%.2f", order.Number, resp.Status, resp.Accrual)

	balanceStore, ok := store.(BalanceUpdater)
	if ok && resp.Status == constants.StatusProcessed && prevStatus != constants.StatusProcessed && resp.Accrual > 0 {
		c.updateUserBalance(ctx, balanceStore, order.UserID, resp.Accrual)
	}
}

func (c *Client) updateUserBalance(ctx context.Context, store BalanceUpdater, userID int64, accrual float64) {
	current, _, err := store.GetBalance(ctx, userID)
	if err != nil {
		log.Printf("Failed to get current balance: %v", err)
		return
	}

	newBalance := accrual
	if current.Valid {
		newBalance += current.Float64
	}

	if err := store.UpdateBalance(ctx, userID, newBalance); err != nil {
		log.Printf("Failed to update balance: %v", err)
		return
	}

	log.Printf("Updated balance for user %d: added %.2f, new balance: %.2f", userID, accrual, newBalance)
}