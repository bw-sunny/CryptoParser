package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// на вход TickerName, на выход поиск цены в памяти, ин дип фьючер
func StartRedis() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := client.Set(ctx, "LastRequest", "BTC", 0).Err()
	if err != nil {
		fmt.Printf("!!! \tОшибка \t!!!")
	}
	crypto, err := client.Get(ctx, "LastRequest").Result()
	if err != nil {
		fmt.Println("Error!!!")
	}
	fmt.Printf("Name of Ticker: %s", crypto)
}
