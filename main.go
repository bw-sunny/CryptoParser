package main

import "fmt"

func main() {
	fmt.Println("Запуск парсера...")
	RequestTicker()
	fmt.Println("Запуск редиса")
	StartRedis()
}
