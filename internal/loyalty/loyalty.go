package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/AlenaMolokova/diploma/internal/storage"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (c *Client) CheckOrder(ctx context.Context, number string) (OrderResponse, error) {
	url := c.addr + "/api/orders/" + number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return OrderResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OrderResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return OrderResponse{}, errors.New("order not found")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return OrderResponse{}, errors.New("too many requests")
	}
	if resp.StatusCode != http.StatusOK {
		return OrderResponse{}, errors.New("unexpected status code")
	}

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return OrderResponse{}, err
	}

	return orderResp, nil
}

func (c *Client) StartOrderProcessing(ctx context.Context, store *storage.Storage) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			orders, err := store.GetOrdersByUserID(ctx, 0) // 0 для всех пользователей
			if err != nil {
				log.Printf("Failed to get orders: %v", err)
				continue
			}
			for _, order := range orders {
				if order.Status == "PROCESSED" {
					continue
				}
				resp, err := c.CheckOrder(ctx, order.Number)
				if err != nil {
					log.Printf("Failed to check order %s: %v", order.Number, err)
					continue
				}
				if resp.Status != order.Status {
					err = store.UpdateOrder(ctx, order.Number, resp.Status, resp.Accrual)
					if err != nil {
						log.Printf("Failed to update order %s: %v", order.Number, err)
						continue
					}
					if resp.Status == "PROCESSED" && resp.Accrual > 0 {
						err = store.UpdateBalance(ctx, order.UserID.Int64, resp.Accrual)
						if err != nil {
							log.Printf("Failed to update balance for user %d: %v", order.UserID.Int64, err)
						}
					}
					log.Printf("Order %s updated to status %s, accrual=%.2f", order.Number, resp.Status, resp.Accrual)
				}
			}
		}
	}
}