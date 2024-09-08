package controllers

import (
	"gotrading/bitflyer"
	"gotrading/config"

	"gotrading/app/models"
)

func StreamIngestionData() {
	var tickerCannel = make(chan bitflyer.Ticker)
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerCannel)
	for ticker := range tickerCannel {
		for _, duration := range config.Config.Durations {
			isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
			if isCreated == true && duration == config.Config.TradeDuration {
				// TODO
			}
		}
	}
}