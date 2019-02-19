// Package zglobal contains global variables,
// these values update once per minute from database.
package zglobal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/vic/vic_go/record"
)

// map productId to Kim
var MapIapAndroid map[string]int64
var CashOutRate float64
var ManualLobbyPricePerHour int64
var SmsSender string // onewaysms, messagebird

func init() {
	// init values
	MapIapAndroid0 := map[string]int64{
		"goi1do":                 18000,
		"goi2do":                 36000,
		"goi3do":                 54000,
		"goi4do":                 72000,
		"android.test.purchased": 1000,
	}
	CashOutRate0 := float64(1.8)
	ManualLobbyPricePerHour0 := int64(5000)
	SmsSender0 := "messagebird"
	// loop update values
	go func() {
		time.Sleep(5 * time.Second) // waiting for init record.dbPool
		for {
			var key, value string
			var err error
			//
			key = "MapIapAndroid"
			value = record.PsqlLoadGlobal(key)
			err = json.Unmarshal([]byte(value), &MapIapAndroid)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MapIapAndroid = MapIapAndroid0
				temp, _ := json.Marshal(MapIapAndroid0)
				record.PsqlSaveGlobal(key, string(temp))
			}
			//
			key = "CashOutRate"
			value = record.PsqlLoadGlobal(key)
			CashOutRate, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				CashOutRate = CashOutRate0
				temp := fmt.Sprintf("%v", CashOutRate0)
				record.PsqlSaveGlobal(key, temp)
			}
			//
			key = "ManualLobbyPricePerHour"
			value = record.PsqlLoadGlobal(key)
			ManualLobbyPricePerHour, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				ManualLobbyPricePerHour = ManualLobbyPricePerHour0
				temp := fmt.Sprintf("%v", ManualLobbyPricePerHour0)
				record.PsqlSaveGlobal(key, temp)
			}
			//
			key = "SmsSender"
			value = record.PsqlLoadGlobal(key)
			SmsSender = value
			if value == "" {
				fmt.Println("zglobal err", key, err)
				SmsSender = SmsSender0
				temp := fmt.Sprintf("%v", SmsSender0)
				record.PsqlSaveGlobal(key, temp)
			}
			//
			time.Sleep(5 * time.Second)
		}
	}()
}
