package main

import (
	"encoding/json"
	"fmt"
	"math"
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

func getHistoricalPrice(tickerName string, daysAgo int) (float64, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	targetTime := time.Now().AddDate(0, 0, -daysAgo)
	startTime := targetTime.UnixMilli()
	endTime := targetTime.Add(24 * time.Hour).UnixMilli()

	url := fmt.Sprintf("https://api.bybit.com/v5/market/kline?category=spot&symbol=%sUSDT&interval=D&start=%d&end=%d",
		tickerName, startTime, endTime)

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var klineResp KlineResponse
	if err := json.NewDecoder(resp.Body).Decode(&klineResp); err != nil {
		return 0, err
	}

	if klineResp.RetCode != 0 {
		return 0, fmt.Errorf("API error: %s", klineResp.RetMsg)
	}

	if len(klineResp.Result.List) == 0 {
		return 0, fmt.Errorf("no historical data found")
	}

	// Данные возвращаются в формате: [timestamp, open, high, low, close, volume, turnover]
	closePrice, ok := klineResp.Result.List[0][4].(string)
	if !ok {
		return 0, fmt.Errorf("invalid price format")
	}

	var price float64
	fmt.Sscanf(closePrice, "%f", &price)

	return price, nil
}
func SMA(tickerName string) (float64, error) {
	var sumPrices float64
	prices := make([]float64, 0)
	for i := 21; i > 11; i-- {
		price, err := getHistoricalPrice(tickerName, i)

		if err != nil {
			fmt.Println(err)
		}
		sumPrices += price
		prices = append(prices, price)
	}
	return sumPrices / 10, nil
}
func calculateEMA(tickerName string) ([]float64, error) {
	K := 2.0 / 11.0
	arrayEMA := make([]float64, 0)

	for i := 11; i > 1; i-- {
		closedPrice, _ := getHistoricalPrice(tickerName, i)

		// fmt.Printf("Price %v days ago: %.2f \n", i-1, closedPrice)

		if i == 11 {
			// fmt.Println("calculatinig sma...")
			lastSMA, _ := SMA(tickerName)
			// fmt.Printf("Appended SMA of %s: %.2f \n", tickerName, lastSMA)
			arrayEMA = append(arrayEMA, math.Round(lastSMA*100)/100)
		} else {
			lenght := len(arrayEMA)
			currentEMA := (closedPrice-arrayEMA[lenght-1])*K + arrayEMA[lenght-1]
			// fmt.Printf("Calculate new EMA \n Last Closed(%v) - LastEMA(%v) * K(%.2f) + LastEMA = CurrentEMA(%.2f)\n", closedPrice, arrayEMA[lenght-1], K, currentEMA)
			arrayEMA = append(arrayEMA, math.Round(currentEMA*100)/100)
		}

	}
	// fmt.Printf("EMA's: %v", EMAs)
	// return Array SMA + EMAs
	return arrayEMA, nil
}
