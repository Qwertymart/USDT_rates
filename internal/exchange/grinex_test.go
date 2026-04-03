package exchange

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetRates(t *testing.T) {
	ctx := context.Background()

	t.Run("success: parses valid orderbook", func(t *testing.T) {
		// Создаем фейковый HTTP сервер
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v2/depth?market=usdtrub", r.URL.String())
			assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla")

			w.WriteHeader(http.StatusOK)
			// Эмулируем структуру JSON биржи Grinex
			_, _ = w.Write([]byte(`{
				"asks": [{"price": "90.0", "volume": "10"}],
				"bids": [{"price": "89.0", "volume": "20"}]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, 2*time.Second)
		ask, bid, err := client.GetRates(ctx, "usdtrub")

		assert.NoError(t, err)
		assert.Equal(t, "90.0", ask)
		assert.Equal(t, "89.0", bid)
	})

	t.Run("success: handles empty asks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Эмулируем пустой стакан продаж
			_, _ = w.Write([]byte(`{
				"asks": [],
				"bids": [{"price": "89.0", "volume": "20"}]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, 2*time.Second)
		ask, bid, err := client.GetRates(ctx, "usdtrub")

		assert.NoError(t, err)
		assert.Equal(t, "0", ask) // ask должен быть "0"
		assert.Equal(t, "89.0", bid)
	})

	t.Run("failure: empty orderbook completely", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Эмулируем полностью пустой стакан
			_, _ = w.Write([]byte(`{"asks": [], "bids": []}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, 2*time.Second)
		ask, bid, err := client.GetRates(ctx, "usdtrub")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "orderbook is completely empty")
		assert.Equal(t, "", ask)
		assert.Equal(t, "", bid)
	})

	t.Run("failure: bad HTTP status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL, 2*time.Second)
		ask, bid, err := client.GetRates(ctx, "usdtrub")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange returned bad status: 500")
		assert.Equal(t, "", ask)
		assert.Equal(t, "", bid)
	})

	t.Run("failure: bad JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		client := NewClient(server.URL, 2*time.Second)
		ask, bid, err := client.GetRates(ctx, "usdtrub")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
		assert.Equal(t, "", ask)
		assert.Equal(t, "", bid)
	})
}
