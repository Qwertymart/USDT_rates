package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type ExchangeClient interface {
	GetRates(ctx context.Context, market string) (ask, bid string, err error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Order struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
}

type DepthResponse struct {
	Asks []Order `json:"asks"`
	Bids []Order `json:"bids"`
}

func NewClient(url string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	
	jar, _ := cookiejar.New(nil)
	
	return &Client{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
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

	// Add User-Agent to prevent 403 Forbidden by anti-bot systems
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

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

	askPrice := "0"
	if len(data.Asks) > 0 {
		askPrice = data.Asks[0].Price
	}

	bidPrice := "0"
	if len(data.Bids) > 0 {
		bidPrice = data.Bids[0].Price
	}

	if askPrice == "0" && bidPrice == "0" {
		return "", "", fmt.Errorf("orderbook is completely empty for market %s", market)
	}

	return askPrice, bidPrice, nil
}