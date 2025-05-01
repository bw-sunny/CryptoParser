package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type TickerResponse struct {
	RetCode int    `json:"ret_code"` // ищем в ответе ret_code
	RetMsg  string `json:"ret_msg"`
	Result  Result `json:"result"`
}

type Result struct {
	List []Ticker `json:"list"` // будет состоять из полей Ticker
}

type Ticker struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
	HighPrice string `json:"highPrice24h"`
	LowPrice  string `json:"lowPrice24h"`
}

//на вход NameTicker, на выход массив атрибутов

func RequestTicker() {

	baseURL := "https://api-testnet.bybit.com/v5/market/tickers"

	params := url.Values{}
	params.Add("category", "spot")
	params.Add("baseCoin", "BTC")
	params.Add("expDate", "30APR22") // Срок годности
	params.Add("symbol", "BTCUSDT")

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод JSON для отладки

	var tickerResponse TickerResponse

	if err := json.Unmarshal(body, &tickerResponse); err != nil {
		fmt.Println(err)
		return
	}

	// Вывод только нужных значений
	for _, ticker := range tickerResponse.Result.List {
		fmt.Printf("Name: %s, \nCurrent Price: %s, \nHigh Price for 24h: %s, \nLow Price for 24h: %s.\n",
			ticker.Symbol, ticker.LastPrice, ticker.HighPrice, ticker.LowPrice)
	}

}
