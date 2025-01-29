package utils

import (
	"io"
	"log"
	"os"
)

func LoggingSettings(logFile string){
	// ログファイルの読み込み
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// エラー処理
	if err != nil {
		log.Fatalf("file=logFile err=%s", err.Error())
	}
	// ログファイルがなければ、作成する。
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	// 出力フラグの設定
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// 出力先の設定
	log.SetOutput(multiLogFile)
}