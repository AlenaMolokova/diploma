package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
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
	resp, err := c.client.Get(c.baseURL + "/api/orders/" + orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order %s: %v", orderNumber, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code for order %s: %d", orderNumber, resp.StatusCode)
	}

	var accrual accrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
		return nil, fmt.Errorf("failed to decode response for order %s: %v", orderNumber, err)
	}

	return &accrual, nil
}

func (c *Client) StartOrderProcessing(ctx context.Context, store models.OrderStorage) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

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
					log.Printf("Failed to check order %s: %v", order.Number, err)
					continue
				}

				updatedOrder := models.Order{
					Number:     order.Number,
					Status:     resp.Status,
					Accrual:    pgtype.Float8{Float64: resp.Accrual, Valid: resp.Accrual != 0},
					UploadedAt: order.UploadedAt,
				}

				if err := store.UpdateOrder(ctx, updatedOrder); err != nil {
					log.Printf("Failed to update order %s: %v", order.Number, err)
					continue
				}

				log.Printf("Updated order %s: status=%s, accrual=%f", order.Number, resp.Status, resp.Accrual)
			}
		}
	}
}