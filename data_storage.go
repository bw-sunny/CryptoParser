package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func DeleteCSV(filename string) error {
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("ошибка удаления файла %s: %v", filename, err)
	}
	return nil
}

func CheckLastDate(filename string) (date string, err error) {
	return "", nil
}

func CreateCSV(tickerName string) error {
	filename := fmt.Sprintf("prices%s.csv", tickerName)

	if fileExists(filename) {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Date", "Ticker", "Price"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("ошибка записи заголовка: %v", err)
	}

	return nil
}

func AddDataToCSV(filename, date, tickerName, price string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{date, tickerName, price}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("ошибка записи данных: %v", err)
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
