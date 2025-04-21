package loyalty

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "time"
)

type Client struct {
    baseURL string
    client  *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

type LoyaltyResponse struct {
    Order   string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual,omitempty"`
}

func (c *Client) CheckOrder(ctx context.Context, number string) (LoyaltyResponse, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/orders/"+number, nil)
    if err != nil {
        log.Printf("Failed to create loyalty request: %v", err)
        return LoyaltyResponse{}, err
    }

    resp, err := c.client.Do(req)
    if err != nil {
        log.Printf("Failed to query loyalty system for order %s: %v", number, err)
        return LoyaltyResponse{}, err
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusOK:
        var loyaltyResp LoyaltyResponse
        if err := json.NewDecoder(resp.Body).Decode(&loyaltyResp); err != nil {
            log.Printf("Failed to decode loyalty response: %v", err)
            return LoyaltyResponse{}, err
        }
        log.Printf("Loyalty system returned for order %s: status=%s, accrual=%.2f", number, loyaltyResp.Status, loyaltyResp.Accrual)
        return loyaltyResp, nil
    case http.StatusTooManyRequests:
        retryAfter := resp.Header.Get("Retry-After")
        log.Printf("Loyalty system rate limit exceeded for order %s, retry after %s", number, retryAfter)
        return LoyaltyResponse{}, errors.New("rate limit exceeded")
    case http.StatusNoContent:
        log.Printf("Order %s not registered in loyalty system", number)
        return LoyaltyResponse{}, errors.New("order not registered")
    default:
        log.Printf("Loyalty system returned status %d for order %s", resp.StatusCode, number)
        return LoyaltyResponse{}, errors.New("loyalty system error")
    }
}