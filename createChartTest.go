package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func GetLastTenDays() []time.Time {
	today := time.Now()
	dates := make([]time.Time, 0)

	for i := 10; i > 0; i-- {
		dates = append(dates, today.AddDate(0, 0, -i))
	}

	return dates
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("Программа выполнена за: %v\n", time.Since(start))
	}()

	dates := GetLastTenDays()
	valuesEMA, err := calculateEMA("XRP")

	if err != nil {
		log.Fatal("Error calculating EMA:", err)
	}

	// Проверяем, что values содержит достаточно данных
	if len(valuesEMA) < 10 {
		log.Fatalf("Not enough EMA values: got %d, expected 10", len(valuesEMA))
	}

	data := make([]opts.LineData, 0)
	xData := make([]string, 0)

	// Используем только 10 элементов
	for i := 0; i < 10; i++ {
		xData = append(xData, dates[i].Format("2006/01/02"))
		data = append(data, opts.LineData{Value: valuesEMA[i]})
	}

	line := charts.NewLine()
	line.SetXAxis(xData).
		AddSeries("EMA Values", data)

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "XRP EMA Time Series",
			Subtitle: "Last 10 days including today",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Date",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "EMA Value",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "grey",
		}),
	)

	f, err := os.Create("timeseries.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := line.Render(f); err != nil {
		log.Fatal(err)
	}

	log.Println("Chart saved to timeseries.html")
}
