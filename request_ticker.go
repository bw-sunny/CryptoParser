package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type TickerResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	Result  Result `json:"result"`
}

type Result struct {
	List []Ticker `json:"list"`
}

type Ticker struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
	HighPrice string `json:"highPrice24h"`
	LowPrice  string `json:"lowPrice24h"`
}

func RequestTicker(TickerName string) string {
	pair := TickerName + "USDT"
	baseURL := "https://api-testnet.bybit.com/v5/market/tickers"

	params := url.Values{}
	params.Add("category", "spot")
	params.Add("baseCoin", TickerName)
	params.Add("symbol", pair)

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Вывод JSON для отладки

	var tickerResponse TickerResponse
	var tickerPrice string

	if err := json.Unmarshal(body, &tickerResponse); err != nil {
		fmt.Println(err)
	}
	for _, ticker := range tickerResponse.Result.List {
		tickerPrice = ticker.LastPrice
	}
	return tickerPrice
}
