package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ratesv1 "github.com/Qwertymart/USDT_rates/pkg/api/rates/v1"
)

// MockRatesService is a mock for service.RatesService
type MockRatesService struct {
	mock.Mock
}

func (m *MockRatesService) GetAndSaveRates(ctx context.Context) (ask, bid string, err error) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.Error(2)
}

func TestServerAPI_GetRates(t *testing.T) {
	ctx := context.Background()

	t.Run("success: returns correct protobuf response", func(t *testing.T) {
		mockService := new(MockRatesService)
		api := &ServerAPI{service: mockService}

		expectedAsk := "90.5"
		expectedBid := "89.5"

		// Setup expectation
		mockService.On("GetAndSaveRates", ctx).Return(expectedAsk, expectedBid, nil)

		// Execute
		resp, err := api.GetRates(ctx, &ratesv1.GetRatesRequest{})

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedAsk, resp.Ask)
		assert.Equal(t, expectedBid, resp.Bid)
		
		// Ensure timestamp is roughly correct (within 1 second of now)
		assert.WithinDuration(t, time.Now(), resp.Timestamp.AsTime(), time.Second)

		mockService.AssertExpectations(t)
	})

	t.Run("failure: service returns error translates to grpc Internal error", func(t *testing.T) {
		mockService := new(MockRatesService)
		api := &ServerAPI{service: mockService}

		expectedErr := errors.New("service error")

		// Setup expectation
		mockService.On("GetAndSaveRates", ctx).Return("", "", expectedErr)

		// Execute
		resp, err := api.GetRates(ctx, &ratesv1.GetRatesRequest{})

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to get rates: service error")

		mockService.AssertExpectations(t)
	})
}
