package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockExchangeClient is a mock for exchange.ExchangeClient
type MockExchangeClient struct {
	mock.Mock
}

func (m *MockExchangeClient) GetRates(ctx context.Context, market string) (ask, bid string, err error) {
	args := m.Called(ctx, market)
	return args.String(0), args.String(1), args.Error(2)
}

// MockRepository is a mock for postgres.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveRate(ctx context.Context, ask, bid string, ts time.Time) (uuid.UUID, error) {
	// We use mock.Anything for timestamp since time.Now() is non-deterministic
	args := m.Called(ctx, ask, bid, mock.Anything)
	
	// Convert the first argument safely to uuid.UUID if it's not nil
	if id, ok := args.Get(0).(uuid.UUID); ok {
		return id, args.Error(1)
	}
	return uuid.Nil, args.Error(1)
}

func (m *MockRepository) Close() {
	m.Called()
}

func TestService_GetAndSaveRates(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	testMarket := "usdtrub"

	t.Run("success: returns rates and saves to db", func(t *testing.T) {
		mockExchange := new(MockExchangeClient)
		mockRepo := new(MockRepository)
		svc := New(mockExchange, mockRepo, logger, testMarket)

		expectedAsk := "90.5"
		expectedBid := "89.5"
		expectedID := uuid.MustParse("018e0000-0000-7000-8000-000000000000")

		// Setup expectations
		mockExchange.On("GetRates", ctx, testMarket).Return(expectedAsk, expectedBid, nil)
		mockRepo.On("SaveRate", ctx, expectedAsk, expectedBid, mock.Anything).Return(expectedID, nil)

		// Execute
		ask, bid, err := svc.GetAndSaveRates(ctx)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, expectedAsk, ask)
		assert.Equal(t, expectedBid, bid)

		// Ensure all mocks were called as expected
		mockExchange.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failure: exchange returns error", func(t *testing.T) {
		mockExchange := new(MockExchangeClient)
		mockRepo := new(MockRepository)
		svc := New(mockExchange, mockRepo, logger, testMarket)

		expectedErr := errors.New("exchange network timeout")

		// Setup expectations
		mockExchange.On("GetRates", ctx, testMarket).Return("", "", expectedErr)
		// Repo should not be called if exchange fails

		// Execute
		ask, bid, err := svc.GetAndSaveRates(ctx)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service failed to get rates")
		assert.Equal(t, "", ask)
		assert.Equal(t, "", bid)

		mockExchange.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "SaveRate")
	})

	t.Run("failure: database save returns error", func(t *testing.T) {
		mockExchange := new(MockExchangeClient)
		mockRepo := new(MockRepository)
		svc := New(mockExchange, mockRepo, logger, testMarket)

		expectedAsk := "90.5"
		expectedBid := "89.5"
		dbErr := errors.New("database connection lost")

		// Setup expectations
		mockExchange.On("GetRates", ctx, testMarket).Return(expectedAsk, expectedBid, nil)
		mockRepo.On("SaveRate", ctx, expectedAsk, expectedBid, mock.Anything).Return(uuid.Nil, dbErr)

		// Execute
		ask, bid, err := svc.GetAndSaveRates(ctx)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service failed to save rates")
		assert.Equal(t, "", ask)
		assert.Equal(t, "", bid)

		mockExchange.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
