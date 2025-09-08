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

// CryptoClient представляет gRPC клиент для получения данных о криптовалютах
type CryptoClient struct {
	conn            *grpc.ClientConn
	client          pb.CryptoClient
	statsHandler    *RequestCounterStatsHandler
	csvFilename     string
	requestInterval time.Duration
}

// RequestCounterStatsHandler счетчик запросов
type RequestCounterStatsHandler struct {
	requestCount atomic.Int64
}

// Crypto структура для данных о криптовалюте
type Crypto struct {
	Time   string
	Ticker string
	Price  string
}

// NewCryptoClient создает новый экземпляр CryptoClient
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

// Close закрывает соединение с сервером
func (cc *CryptoClient) Close() error {
	if cc.conn != nil {
		return cc.conn.Close()
	}
	return nil
}

// GetRequestCount возвращает количество выполненных запросов
func (cc *CryptoClient) GetRequestCount() int64 {
	return cc.statsHandler.requestCount.Load()
}

// GetCryptoPrice получает цену для конкретной криптовалюты
func (cc *CryptoClient) GetCryptoPrice(ctx context.Context, ticker string) (Crypto, error) {
	response, err := cc.client.CryptoPrice(ctx, &pb.PriceRequest{Name: ticker})
	if err != nil {
		return Crypto{}, fmt.Errorf("could not get price for %s: %v", ticker, err)
	}

	crypto, err := parseMessage(response.Message)
	if err != nil {
		return Crypto{}, fmt.Errorf("failed to parse message for %s: %v", ticker, err)
	}

	return crypto, nil
}

// SaveCryptoToCSV сохраняет данные о криптовалюте в CSV файл
func (cc *CryptoClient) SaveCryptoToCSV(crypto Crypto) error {
	return appendToCSVFile(crypto, cc.csvFilename)
}

// InitializeCSVFile создает CSV файл с заголовками
func (cc *CryptoClient) InitializeCSVFile() error {
	file, err := os.OpenFile(cc.csvFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Проверяем, пуст ли файл
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Если файл пустой, добавляем заголовки
	if info.Size() == 0 {
		writer := csv.NewWriter(file)
		defer writer.Flush()

		headers := []string{"Time", "Ticker", "Price"}
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	return nil
}

// RunInteractiveMode запускает интерактивный режим для ввода тикеров
func (cc *CryptoClient) RunInteractiveMode(ctx context.Context) error {
	// Инициализируем CSV файл
	if err := cc.InitializeCSVFile(); err != nil {
		return fmt.Errorf("failed to initialize CSV file: %v", err)
	}

	fmt.Println("=== Crypto Price Fetcher ===")
	fmt.Println("Введите названия криптовалют (по одному в строке)")
	fmt.Println("Для завершения введите 'quit' или 'exit'")
	fmt.Println("============================")

	for {
		fmt.Print("Введите тикер криптовалюты: ")
		var ticker string
		fmt.Scanln(&ticker)

		ticker = strings.TrimSpace(ticker)
		ticker = strings.ToUpper(ticker)

		if ticker == "QUIT" || ticker == "EXIT" || ticker == "" {
			break
		}

		// Получаем данные о криптовалюте
		crypto, err := cc.GetCryptoPrice(ctx, ticker)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		// Сохраняем в CSV
		if err := cc.SaveCryptoToCSV(crypto); err != nil {
			fmt.Printf("Ошибка при сохранении: %v\n", err)
			continue
		}

		fmt.Printf("✓ Получена цена для %s: $%s\n", crypto.Ticker, crypto.Price)
		fmt.Printf("✓ Данные сохранены в %s\n", cc.csvFilename)
		fmt.Println("---")

		// Добавляем задержку между запросами
		if cc.requestInterval > 0 {
			time.Sleep(cc.requestInterval)
		}
	}

	return nil
}

// Вспомогательные функции

func parseMessage(msg string) (Crypto, error) {
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

func appendToCSVFile(crypto Crypto, filename string) error {
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

// Методы для RequestCounterStatsHandler
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

// Main функция клиента
func main() {
	// Создаем клиент только с CSV
	cryptoClient, err := NewCryptoClient(
		"localhost:50052",
		"crypto_data.csv",
		100*time.Millisecond,
	)
	if err != nil {
		log.Fatalf("Failed to create crypto client: %v", err)
	}
	defer cryptoClient.Close()

	// Настраиваем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Запускаем интерактивный режим
	if err := cryptoClient.RunInteractiveMode(ctx); err != nil {
		log.Fatalf("Interactive mode failed: %v", err)
	}

	// Выводим статистику
	fmt.Printf("\nTotal gRPC requests made: %d\n", cryptoClient.GetRequestCount())
	fmt.Println("Данные сохранены в crypto_data.csv")
}
