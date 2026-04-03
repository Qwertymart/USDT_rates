package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ExchangeClient interface {
	GetRates(ctx context.Context, market string) (ask, bid string, err error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type DepthResponse struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}

func NewClient(url string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetRates fetches top ask and bid prices from Grinex for a given market
func (c *Client) GetRates(ctx context.Context, market string) (ask, bid string, err error) {
	url := fmt.Sprintf("%s/api/v2/depth?market=%s", c.baseURL, market)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("exchange returned bad status: %d", resp.StatusCode)
	}

	var data DepthResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(data.Asks) == 0 || len(data.Bids) == 0 {
		return "", "", fmt.Errorf("empty orderbook for market %s", market)
	}

	return data.Asks[0][0], data.Bids[0][0], nil
}