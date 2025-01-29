package controllers

import (
	"log"

	"gotrading/app/models"
	"gotrading/bitflyer"
	"gotrading/config"
)

/*
関数名　：
機能　　：
引数　　：[IN] , [OUT]
その他　：
*/ 
func StreamIngestionData() {
	c := config.Config
	ai := NewAI(c.ProductCode, c.TradeDuration, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)
	// 初期化
	var tickerChannel = make(chan bitflyer.Ticker)

	// bitflyer認証情報を渡して、API権利を獲得
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)

	// 取引対象銘柄の現在情報を取得
	go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)
	go func() {
		// 取得したティッカー情報を格納
		for ticker := range tickerChannel {
			// ログを記録
			log.Printf("action=StreamIngestionData, %v", ticker)
			// 存続期間をiniファイルより取得しキャンドルスティックを作成
			for _, duration := range config.Config.Durations {
				// キャンドルスティックを作成
				isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
				if isCreated == true && duration == config.Config.TradeDuration {
					ai.Trade()
				}
			}
		}
	}()
}