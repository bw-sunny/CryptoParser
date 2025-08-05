package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	pb "CryptoParser/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type Cryrptocurrency struct {
	Name  string `json: name`
	Value string `json: value`
}

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

	filename := "example.txt"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии/создании файла:", err)
		return
	}
	defer file.Close()

	names := []string{"WIF", "ETH", "BTC", "SUI", "TRUMP"}
	for _, name := range names {
		r, err := c.CryptoPrice(ctx, &pb.PriceRequest{Name: name})
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%s", r.Message)
		_, err = file.WriteString(r.Message + "\n")
		if err != nil {
			fmt.Println("Ошибка при записи в файл:", err)
			return
		}
	}
	fmt.Printf("Total gRPC requests made: %d\n", handler.requestCount.Load())
}
