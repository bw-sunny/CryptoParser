package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	pb "CryptoParser/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type Crypto struct {
	Name  string `json: name`
	Value string `json: value`
}

func parseMessage(msg string) (Crypto, error) {
	// Регулярное выражение для ответа сервера
	re := regexp.MustCompile(`Цена (.+): \$(.+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		return Crypto{}, fmt.Errorf("неверный формат сообщения")
	}

	return Crypto{
		Name:  strings.TrimSpace(matches[1]),
		Value: strings.TrimSpace(matches[2]),
	}, nil
}
func appendToJSONFile(crypto Crypto, filename string) error {
	data, err := json.Marshal(crypto) //преобразование данных в структуру crypto
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close() // открытие или создание файла

	if _, err := f.Write(append(data, '\n')); err != nil { // запись в файл
		return err
	}

	return nil
}

// Counter of requests
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

	const jsonFile = "crypto_data.json"

	names := []string{"WIF", "ETH", "BTC", "SUI", "TRUMP"}
	for _, name := range names {

		r, err := c.CryptoPrice(ctx, &pb.PriceRequest{Name: name})
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		crypto, err := parseMessage(r.Message)
		if err != nil {
			log.Printf("Ошибка парсинга: %v", err)
			continue
		}

		if err := appendToJSONFile(crypto, jsonFile); err != nil {
			log.Printf("Ошибка записи в файл: %v", err)
		}
		log.Printf("%s", r.Message)

	}
	fmt.Printf("Total gRPC requests made: %d\n", handler.requestCount.Load())
}
