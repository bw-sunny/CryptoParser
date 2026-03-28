package main

import (
	"context"
	"encoding/csv"
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

type CryptoClient struct {
	conn            *grpc.ClientConn
	client          pb.CryptoClient
	statsHandler    *RequestCounterStatsHandler
	csvFilename     string
	requestInterval time.Duration
}

type RequestCounterStatsHandler struct {
	requestCount atomic.Int64
}

type Crypto struct {
	Time   string
	Ticker string
	Price  string
}

type CryptoRatio struct {
	Time           string
	Ticker         string
	BuyRatio       string
	SellRatio      string
	LongShortRatio string
}

func NewCryptoClient(serverAddr, csvFilename string, requestInterval time.Duration) (*CryptoClient, error) {
	statsHandler := &RequestCounterStatsHandler{}

	conn, err := grpc.Dial(serverAddr,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(statsHandler))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewCryptoClient(conn)

	return &CryptoClient{
		conn:            conn,
		client:          client,
		statsHandler:    statsHandler,
		csvFilename:     csvFilename,
		requestInterval: requestInterval,
	}, nil
}

func (cc *CryptoClient) Close() error {
	if cc.conn != nil {
		return cc.conn.Close()
	}
	return nil
}

func (cc *CryptoClient) GetRequestCount() int64 {
	return cc.statsHandler.requestCount.Load()
}

func (cc *CryptoClient) GetCryptoPrice(ctx context.Context, ticker string) (Crypto, error) {
	response, err := cc.client.CryptoPrice(ctx, &pb.PriceRequest{Name: ticker})
	if err != nil {
		return Crypto{}, fmt.Errorf("could not get price for %s: %v", ticker, err)
	}

	crypto, err := parsePriceMessage(response.Message)
	if err != nil {
		return Crypto{}, fmt.Errorf("failed to parse message for %s: %v", ticker, err)
	}

	return crypto, nil
}

func (cc *CryptoClient) GetCryptoRatio(ctx context.Context, ticker string) (CryptoRatio, error) {
	response, err := cc.client.CryptoRatio(ctx, &pb.RatioRequest{Name: ticker})
	if err != nil {
		return CryptoRatio{}, fmt.Errorf("could not get ratio for %s: %v", ticker, err)
	}

	ratio, err := parseRatioMessage(response.Message)
	if err != nil {
		return CryptoRatio{}, fmt.Errorf("failed to parse ratio message for %s: %v", ticker, err)
	}

	return ratio, nil
}

func (cc *CryptoClient) SaveCryptoToCSV(crypto Crypto) error {
	return appendPriceToCSVFile(crypto, cc.csvFilename)
}

func (cc *CryptoClient) SaveRatioToCSV(ratio CryptoRatio) error {
	return appendRatioToCSVFile(ratio, "crypto_ratio.csv")
}

func (cc *CryptoClient) InitializePriceCSV() error {
	return initializeCSVFile(cc.csvFilename, []string{"Time", "Ticker", "Price"})
}

func (cc *CryptoClient) InitializeRatioCSV() error {
	return initializeCSVFile("crypto_ratio.csv", []string{"Time", "Ticker", "BuyRatio", "SellRatio", "LongShortRatio"})
}

func (cc *CryptoClient) RunInteractiveMode(ctx context.Context) error {

	if err := cc.InitializePriceCSV(); err != nil {
		return fmt.Errorf("failed to initialize price CSV file: %v", err)
	}
	if err := cc.InitializeRatioCSV(); err != nil {
		return fmt.Errorf("failed to initialize ratio CSV file: %v", err)
	}

	fmt.Println("=== Crypto Data Fetcher ===")
	fmt.Println("Введите названия криптовалют (по одному в строке)")
	fmt.Println("Для завершения введите 'quit' или 'exit'")
	fmt.Println("===========================")

	for {
		fmt.Print("Введите тикер криптовалюты: ")
		var ticker string
		fmt.Scanln(&ticker)

		ticker = strings.TrimSpace(ticker)
		ticker = strings.ToUpper(ticker)

		if ticker == "QUIT" || ticker == "EXIT" || ticker == "" {
			break
			return nil
		}

		crypto, err := cc.GetCryptoPrice(ctx, ticker)
		if err != nil {
			fmt.Printf("Ошибка при получении цены: %v\n", err)
		} else {

			if err := cc.SaveCryptoToCSV(crypto); err != nil {
				fmt.Printf("Ошибка при сохранении цены: %v\n", err)
			} else {
				fmt.Printf("✓ Цена %s: $%s\n", crypto.Ticker, crypto.Price)
			}
		}

		ratio, err := cc.GetCryptoRatio(ctx, ticker)
		if err != nil {
			fmt.Printf("Ошибка при получении ratio: %v\n", err)
		} else {

			if err := cc.SaveRatioToCSV(ratio); err != nil {
				fmt.Printf("Ошибка при сохранении ratio: %v\n", err)
			} else {
				fmt.Printf("✓ Ratio %s: Buy=%s%%, Sell=%s%%, L/S=%s\n",
					ratio.Ticker, ratio.BuyRatio, ratio.SellRatio, ratio.LongShortRatio)
			}
		}

		fmt.Printf("✓ Данные сохранены в CSV файлы\n")
		fmt.Println("---")

		if cc.requestInterval > 0 {
			time.Sleep(cc.requestInterval)
		}
	}

	return nil
}

func parsePriceMessage(msg string) (Crypto, error) {
	re := regexp.MustCompile(`Цена (.+): \$(.+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		return Crypto{}, fmt.Errorf("invalid message format")
	}

	return Crypto{
		Time:   time.Now().Format(time.RFC3339),
		Ticker: strings.TrimSpace(matches[1]),
		Price:  strings.TrimSpace(matches[2]),
	}, nil
}

func parseRatioMessage(msg string) (CryptoRatio, error) {
	re := regexp.MustCompile(`Long/Short Ratio (.+): Buy=(.+)%, Sell=(.+)%, L/S=(.*)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) >= 5 {
		longShortRatio := strings.TrimSpace(matches[4])
		if longShortRatio == "" {
			buyRatio := parseFloatSafe(matches[2])
			sellRatio := parseFloatSafe(matches[3])
			if sellRatio > 0 {
				longShortRatio = fmt.Sprintf("%.4f", buyRatio/sellRatio)
			} else {
				longShortRatio = "0"
			}
		}

		return CryptoRatio{
			Time:           time.Now().Format(time.RFC3339),
			Ticker:         strings.TrimSpace(matches[1]),
			BuyRatio:       strings.TrimSpace(matches[2]),
			SellRatio:      strings.TrimSpace(matches[3]),
			LongShortRatio: longShortRatio,
		}, nil
	}

	return CryptoRatio{}, fmt.Errorf("invalid ratio message format: %s", msg)
}

func parseFloatSafe(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func initializeCSVFile(filename string, headers []string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		writer := csv.NewWriter(file)
		defer writer.Flush()
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	return nil
}

func appendPriceToCSVFile(crypto Crypto, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{crypto.Time, crypto.Ticker, crypto.Price}
	return writer.Write(record)
}

func appendRatioToCSVFile(ratio CryptoRatio, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{ratio.Time, ratio.Ticker, ratio.BuyRatio, ratio.SellRatio, ratio.LongShortRatio}
	return writer.Write(record)
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

	cryptoClient, err := NewCryptoClient(
		"localhost:50052",
		"crypto_data.csv",
		100*time.Millisecond,
	)
	if err != nil {
		log.Fatalf("Failed to create crypto client: %v", err)
	}
	defer cryptoClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := cryptoClient.RunInteractiveMode(ctx); err != nil {
		log.Fatalf("Interactive mode failed: %v", err)
	}

	fmt.Printf("\nTotal gRPC requests made: %d\n", cryptoClient.GetRequestCount())
	fmt.Println("Данные о ценах сохранены в crypto_data.csv")
	fmt.Println("Данные о ratio сохранены в crypto_ratio.csv")
}
