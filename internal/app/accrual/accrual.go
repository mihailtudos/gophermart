package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
)

const httpClientTimeout = time.Minute

type Client struct {
	*http.Client
	Address string
}

func New(accrualAddress string) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: httpClientTimeout,
		},
		Address: accrualAddress,
	}
}
func (c *Client) GetOrderInfo(ctx context.Context, order domain.Order) (domain.Order, error) {
	URL, err := url.Parse(fmt.Sprintf(c.Address+"/api/orders/%s", order.OrderNumber))
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	logger.Log.InfoContext(ctx, "handling order number "+URL.String())

	req, err := http.NewRequest(http.MethodGet, URL.String(), http.NoBody)
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Order{}, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Decode response body into order struct
	var orderResponse domain.Order
	if err = json.NewDecoder(resp.Body).Decode(&orderResponse); err != nil {
		return domain.Order{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return orderResponse, nil
}
