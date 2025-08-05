package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	pb "CryptoParser/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type RequestCounterStatsHandler struct {
	requestCount atomic.Int64
}

func (h *RequestCounterStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *RequestCounterStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	if _, ok := s.(*stats.End); ok {
		h.requestCount.Add(1)
	}
}

func (h *RequestCounterStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

func (h *RequestCounterStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {}

func main() {
	handler := &RequestCounterStatsHandler{} //new handler for valid count of requests in a second
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure(), grpc.WithStatsHandler(handler))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewCryptoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	names := []string{"BTC", "ETH", "TON", "SOL", "TRUMP"}
	for _, name := range names {
		r, err := c.CryptoPrice(ctx, &pb.PriceRequest{Name: name})
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%s", r.Message)
	}
	fmt.Printf("Total gRPC requests made: %d\n", handler.requestCount.Load())
}
