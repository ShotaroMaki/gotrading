package models

import (
	"fmt"
	"time"

	"gotrading/bitflyer"
)

// Candle：キャンドルスティックの構造体
type Candle struct {
	ProductCode string        `json:"product_code"`
	Duration    time.Duration `json:"duration"`
	Time        time.Time     `json:"time"`
	Open        float64       `json:"open"`
	Close       float64       `json:"close"`
	High        float64       `json:"high"`
	Low         float64       `json:"low"`
	Volume      float64       `json:"volume"`
}

/*
関数名　：新規キャンドルスティック作成
機能　　：受け取った引数から、新規キャンドルスティックを作成し、そのアドレスを戻す
引数　　：[IN] 認証情報
        [IN] 経過時間
        [IN] 日時情報
        [IN] 開始時点値
        [IN] 終了時点値
        [IN] 最高値
        [IN] 最安値
        [IN] 取引量
		[OUT]　キャンドルスティックのオブジェクト
その他　：
*/ 
func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open, close, high, low, volume float64) *Candle {
	return &Candle{
		productCode,
		duration,
		timeDate,
		open,
		close,
		high,
		low,
		volume,
	}
}

/*
関数名　：新規キャンドルスティック作成
機能　　：受け取った引数から、新規キャンドルスティックを作成し、そのアドレスを戻す
引数　　：[IN] 認証情報
		[OUT]　キャンドルスティックのオブジェクト
その他　：
*/ 
func (c *Candle) TableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

func (c *Candle) Create() error {
	cmd := fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?)", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Time.Format(time.RFC3339), c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		return err
	}
	return err
}

func (c *Candle) Save() error {
	cmd := fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return err
}

/*
関数名　：キャンドルスティック取得処理
機能　　：DBからキャンドルスティックの情報を抽出する
引数　　：[IN] 認証情報
        [IN] 経過時間
		[IN] 日時情報
　　　　 [OUT] 対象のキャンドルスティックのアドレスに格納されている数値
その他　：
*/ 
func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *Candle {
	// DBテーブルの名称取得
	tableName := GetCandleTableName(productCode, duration)
	// データ抽出SQLを作成
	cmd := fmt.Sprintf("SELECT time, open, close, high, low, volume FROM  %s WHERE time = ?", tableName)
	// SQLを実行しレコードを取得
	row := DbConnection.QueryRow(cmd, dateTime.Format(time.RFC3339))
	var candle Candle
	// エラー処理
	err := row.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil {
		return nil
	}
	// 
	return NewCandle(productCode, duration, candle.Time, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}

/*
関数名　：キャンドルスティック作成
機能　　：iniファイルのDurationの存続期間でキャンドルスティックを作成
引数　　：[IN] tiker = bitflyerの銘柄情報
        [IN] productCode = 取引対象
　　　　 [OUT]
その他　：
*/ 
func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	// 対象のキャンドルスティックを取得
	currentCandle := GetCandle(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	if currentCandle == nil {
		candle := NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		candle.Create()
		return true
	}

	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	currentCandle.Volume += ticker.Volume
	currentCandle.Close = price
	currentCandle.Save()
	return false
}

func GetAllCandle(productCode string, duration time.Duration, limit int) (dfCandle *DataFrameCandle, err error) {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf(`SELECT * FROM (
	SELECT time, open, close, high, low, volume FROM %s ORDER BY time DESC LIMIT ?
	) ORDER BY TIME ASC;`, tableName)
	rows, err := DbConnection.Query(cmd, limit) 
	if err != nil {
		return
	}
	defer rows.Close()

	dfCandle = &DataFrameCandle{}
	dfCandle.ProductCode = productCode
	dfCandle.Duration = duration
	for rows.Next() {
		var candle Candle
		candle.ProductCode = productCode
		candle.Duration = duration
		rows.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return dfCandle, nil
}
