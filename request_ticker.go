package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type KlineResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Category string          `json:"category"`
		List     [][]interface{} `json:"list"`
	} `json:"result"`
}

func currentDate() string {
	now := time.Now()
	date := now.Format("02.01.2006")
	return date
}

func yesterdayDate() string {
	yesterday := time.Now().AddDate(0, 0, -1) // вычесть 1 день
	return yesterday.Format("02.01.2006")
}

func dateToTimestamps(date string) (startTimestamp, endTimestamp int64) {
	parsedDate, err := time.Parse("02.01.2006", date)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}

	startOfDay := parsedDate.UTC().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Millisecond)

	startTimestampMs := startOfDay.UnixMilli()
	endTimestampMs := endOfDay.UnixMilli()

	return startTimestampMs, endTimestampMs
}

func RequestTicker(tickerName string) (string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second} // Тайм-аут
	date := yesterdayDate()
	start, end := dateToTimestamps(date)
	url := fmt.Sprintf("https://api.bybit.com/v5/market/kline?category=spot&symbol=%s&interval=D&start=%d&end=%d", tickerName, start, end) // request URL+TickerName

	resp, err := client.Get(url) // method get
	if err != nil {
		return "", "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result KlineResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { // json decode to struct BybitResponse
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if result.RetCode != 0 {
		return "", "", fmt.Errorf("bybit API error: %s", result.RetMsg)
	}

	if len(result.Result.List) == 0 {
		return "", "", fmt.Errorf("no ticker data found")
	}

	lastPrice, ok := result.Result.List[0][4].(string)
	if !ok {
		return "", "", fmt.Errorf("price is not a string: %T", result.Result.List[0][4])
	}
	return lastPrice, date, nil
}

func main() {
	price, date, _ := RequestTicker("BTCUSDT")
	fmt.Printf("%s | Closed price of BTC: %v", date, price)
}
