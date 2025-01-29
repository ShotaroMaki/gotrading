package main

import (
	"gotrading/app/controllers"
	"gotrading/config"
	"gotrading/utils"
)

func main() {
    //df, _ := models.GetAllCandle(config.Config.ProductCode, time.Minute, 365)
	//fmt.Printf("%+v¥n", df.OptimizeParams())
	
	// ログファイル用の設定値読み込み
	utils.LoggingSettings(config.Config.LogFile)
	controllers.StreamIngestionData()
	controllers.StartWebServer()
}
