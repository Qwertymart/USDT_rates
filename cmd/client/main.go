package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ratesv1 "github.com/Qwertymart/USDT_rates/pkg/api/rates/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// connect
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := ratesv1.NewRatesServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := client.GetRates(ctx, &ratesv1.GetRatesRequest{})
	if err != nil {
		log.Fatalf("could not get rates: %v", err)
	}

	fmt.Printf("Success!\nAsk: %s\nBid: %s\nTime: %s\n", 
		r.GetAsk(), r.GetBid(), r.GetTimestamp().AsTime().Local())
}
