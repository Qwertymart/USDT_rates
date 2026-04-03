package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Qwertymart/USDT_rates/internal/service"
	ratesv1 "github.com/Qwertymart/USDT_rates/pkg/api/rates/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServerAPI struct {
	ratesv1.UnimplementedRatesServiceServer
	service service.RatesService
}

func Register(gRPC *grpc.Server, service service.RatesService) {
	ratesv1.RegisterRatesServiceServer(gRPC, &ServerAPI{service: service})
}

func (s *ServerAPI) GetRates(
	ctx context.Context,
	in *ratesv1.GetRatesRequest,
) (*ratesv1.GetRatesResponse, error) {
	ask, bid, err := s.service.GetAndSaveRates(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get rates: %v", err))
	}

	return &ratesv1.GetRatesResponse{
		Ask:       ask,
		Bid:       bid,
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}
