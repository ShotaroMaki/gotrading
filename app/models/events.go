package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gotrading/config"
)

// DBのデータ格納用構造体
type SignalEvent struct {
	Time        time.Time `json:"time"`
	ProductCode string    `json:"product_code"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
}

/*
関数名　：イベント情報登録
機能　　：取引情報をDBに登録する
引数　　：[IN] 取引情報の
        [OUT] DB登録の実行結果
その他　：
*/ 
func (s *SignalEvent) Save() bool {
	// SQLを生成
	cmd := fmt.Sprintf("INSERT INTO %s (time, product_code, side, price, size) VALUES (?, ?, ?, ?, ?)", tableNameSignalEvents)
	
	// SQL実行し、実行結果をエラー変数に格納
	_, err := DbConnection.Exec(cmd, s.Time.Format(time.RFC3339), s.ProductCode, s.Side, s.Price, s.Size)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Println(err)
			return true
		}
		return false
	}
	return true
}

// 最終取引が売買のどちらか
type SignalEvents struct {
	Signals []SignalEvent `json:"signals,omitempty"`
}

/*
関数名　：コンストラクタ
機能　　：シグナルイベンツの構造体を生成
引数　　：[IN] なし
        [OUT] なし
その他　：
*/ 
func NewSignalEvents() *SignalEvents {
	return &SignalEvents{}
}

/*
関数名　：シグナルイベンツ取得処理
機能　　：シグナルイベンツの情報を昇順で取得し、その格納先アドレスを返す
引数　　：[IN] 取得したシグナルイベンツ情報
        [OUT]　シグナルイベンツの格納アドレス
その他　：
*/ 
func GetSignalEventsByCount(loadEvents int) *SignalEvents {
	// SQLを生成
	cmd := fmt.Sprintf(`SELECT * FROM (
        SELECT time, product_code, side, price, size FROM %s WHERE product_code = ? ORDER BY time DESC LIMIT ? )
        ORDER BY time ASC;`, tableNameSignalEvents)
	
	// SQL実行
	rows, err := DbConnection.Query(cmd, config.Config.ProductCode, loadEvents)
	// エラー処理
	if err != nil {
		return nil
	}

	// 遅行処理として、
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	err = rows.Err()
	if err != nil {
		return nil
	}
	return &signalEvents
}

/*
関数名　：指定時間別シグナルイベンツ取得処理
機能　　：指定した時間以降のシグナルイベンツの情報を昇順で取得し、その格納先アドレスを返す
引数　　：[IN] 取得したシグナルイベンツ情報
    　　[OUT]　シグナルイベンツの格納アドレス
その他　：
*/ 
func GetSignalEventsAfterTime(timeTime time.Time) *SignalEvents {
	cmd := fmt.Sprintf(`SELECT * FROM (
                SELECT time, product_code, side, price, size FROM %s
                WHERE DATETIME(time) >= DATETIME(?)
                ORDER BY time DESC
            ) ORDER BY time ASC;`, tableNameSignalEvents)
	rows, err := DbConnection.Query(cmd, timeTime.Format(time.RFC3339))
	if err != nil {
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	return &signalEvents
}

func (s *SignalEvents) CanBuy(time time.Time) bool {
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return true
	}

	lastSignal := s.Signals[lenSignals-1]
	if lastSignal.Side == "SELL" && lastSignal.Time.Before(time) {
		return true
	}
	return false
}

func (s *SignalEvents) CanSell(time time.Time) bool {
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return false
	}

	lastSignal := s.Signals[lenSignals-1]
	if lastSignal.Side == "BUY" && lastSignal.Time.Before(time) {
		return true
	}
	return false
}

func (s *SignalEvents) Buy(ProductCode string, time time.Time, price, size float64, save bool) bool {
	if !s.CanBuy(time) {
		return false
	}
	signalEvent := SignalEvent{
		ProductCode: ProductCode,
		Time:        time,
		Side:        "BUY",
		Price:       price,
		Size:        size,
	}
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

func (s *SignalEvents) Sell(productCode string, time time.Time, price, size float64, save bool) bool {

	if !s.CanSell(time) {
		return false
	}

	signalEvent := SignalEvent{
		ProductCode: productCode,
		Time:        time,
		Side:        "SELL",
		Price:       price,
		Size:        size,
	}
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

func (s *SignalEvents) Profit() float64 {
	total := 0.0
	beforeSell := 0.0
	isHolding := false
	for i, signalEvent := range s.Signals {
		if i == 0 && signalEvent.Side == "SELL" {
			continue
		}
		if signalEvent.Side == "BUY" {
			total -= signalEvent.Price * signalEvent.Size
			isHolding = true
		}
		if signalEvent.Side == "SELL" {
			total += signalEvent.Price * signalEvent.Size
			isHolding = false
			beforeSell = total
		}
	}
	if isHolding == true {
		return beforeSell
	}
	return total
}

func (s SignalEvents) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(&struct {
		Signals []SignalEvent `json:"signals,omitempty"`
		Profit  float64       `json:"profit,omitempty"`
	}{
		Signals: s.Signals,
		Profit:  s.Profit(),
	})
	if err != nil {
		return nil, err
	}
	return value, err
}

func (s *SignalEvents) CollectAfter(time time.Time) *SignalEvents {
	for i, signal := range s.Signals {
		if time.After(signal.Time) {
			continue
		}
		return &SignalEvents{Signals: s.Signals[i:]}
	}
	return nil
}
