package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BybitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			Symbol    string `json:"symbol"`
			LastPrice string `json:"lastPrice"`
		} `json:"list"`
	} `json:"result"`
}

func RequestTicker(tickerName string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}                                                     // Тайм-аут
	url := fmt.Sprintf("https://api.bybit.com/v5/market/tickers?category=spot&symbol=%sUSDT", tickerName) // request URL+TickerName

	resp, err := client.Get(url) // method get
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result BybitResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { // json decode to struct BybitResponse
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if result.RetCode != 0 {
		return "", fmt.Errorf("bybit API error: %s", result.RetMsg)
	}

	if len(result.Result.List) == 0 {
		return "", fmt.Errorf("no ticker data found")
	}

	return result.Result.List[0].LastPrice, nil
}
