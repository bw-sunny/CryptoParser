package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BybitRatioResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			Symbol         string `json:"symbol"`
			BuyRatio       string `json:"buyRatio"`
			SellRatio      string `json:"sellRatio"`
			LongShortRatio string `json:"longShortRatio"`
			Timestamp      string `json:"timestamp"`
		} `json:"list"`
	} `json:"result"`
}

func RequestRatio(tickerName string) (string, string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.bybit.com/v5/market/account-ratio?category=linear&symbol=%sUSDT&period=1h", tickerName)

	resp, err := client.Get(url)
	if err != nil {
		return "", "", "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result BybitRatioResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if result.RetCode != 0 {
		return "", "", "", fmt.Errorf("bybit API error: %s", result.RetMsg)
	}

	if len(result.Result.List) == 0 {
		return "", "", "", fmt.Errorf("no ratio data found")
	}

	// buy ratio, sell ratio и long/short ratio
	ratioData := result.Result.List[0]
	return ratioData.BuyRatio, ratioData.SellRatio, ratioData.LongShortRatio, nil
}
