package currency

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

const (
	Money         = "money"
	TestMoney     = "test_money"
	WheelSpin     = "WheelSpin"
	CustomMoney   = "CustomMoney"
	Wheel2Spin    = "Wheel2Spin"
	ChargingBonus = "ChargingBonus"

	// using for counter money change times
	SlotSpin100 = "SlotSpin100"

	// SlotSpin1000  = "SlotSpin1000"  // dont use anymore
	// SlotSpin10000 = "SlotSpin10000" // dont use anymore
	VipPoint = "vip_point" // dont use anymore

)

var currencyTypesString = []string{
	Money, TestMoney,
	WheelSpin, CustomMoney, Wheel2Spin,
	SlotSpin100,
	//	SlotSpin1000, SlotSpin10000,
	VipPoint, ChargingBonus,
}
var currencyTypes = make([]*CurrenctyType, 0)
var dataCenter *datacenter.DataCenter

func init() {
	fmt.Println("")
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	fetchCurrenciesType()
}

type CurrenctyType struct {
	currencyType string
	initialValue int64
}

type CurrencyGroup struct {
	playerId   int64
	currencies *StringCurrencyMap
}

type Currency struct {
	playerId     int64
	currencyType string
	value        int64

	freezeValueMap *utils.StringInt64Map

	mutex sync.Mutex
}

func NewCurrency() *Currency {
	return &Currency{
		freezeValueMap: utils.NewStringInt64Map(),
	}
}

func NewCurrencyGroup(playerId int64) *CurrencyGroup {
	//	fmt.Println("checkpoint in NewCurrencyGroup")

	currencyGroup := &CurrencyGroup{}
	currencyGroup.playerId = playerId
	currencies := NewStringCurrencyMap()

	queryString := "SELECT currency_type, value FROM currency WHERE player_id = $1"
	rows, err := dataCenter.Db().Query(queryString, playerId)
	if err != nil {
		log.LogSerious("fetch currency err %v, playerId %d", err, playerId)
		return nil
		if err == sql.ErrNoRows {
			// create new
			for _, currencyType := range currencyTypes {
				queryString = "INSERT INTO currency (currency_type, player_id, value) VALUES ($1,$2,$3)"
				_, err := dataCenter.Db().Exec(queryString, currencyType.currencyType, playerId, currencyType.initialValue)

				//				fmt.Println("hihi", currencyType.currencyType, currencyType.initialValue)

				if err != nil {
					log.LogSerious("insert new currency err %v, playerId %d", err, playerId)
					continue
				}

				currency := NewCurrency()
				currency.playerId = playerId
				currency.currencyType = currencyType.currencyType
				currency.value = currencyType.initialValue
				currencies.set(currencyType.currencyType, currency)

			}
		} else {
		}
	} else {
		for rows.Next() {
			var currencyType string
			var value int64
			err := rows.Scan(&currencyType, &value)
			if err != nil {
				log.LogSerious("fetch currency err %v, playerId %d", err, playerId)
				rows.Close()
				return nil
			}

			currency := NewCurrency()
			currency.playerId = playerId
			currency.currencyType = currencyType
			currency.value = value
			currencies.set(currencyType, currency)
		}
		rows.Close()

		if currencies.len() < len(currencyTypes) {
			for _, currencyType := range currencyTypes {
				if currencies.get(currencyType.currencyType) == nil {
					queryString = "INSERT INTO currency (currency_type, player_id, value) VALUES ($1,$2,$3)"
					_, err := dataCenter.Db().Exec(queryString, currencyType.currencyType, playerId, currencyType.initialValue)
					if err != nil {
						log.LogSerious("insert new currency err %v, playerId %d", err, playerId)
						continue
					}

					currency := NewCurrency()
					currency.playerId = playerId
					currency.currencyType = currencyType.currencyType
					currency.value = currencyType.initialValue
					currencies.set(currencyType.currencyType, currency)
				}
			}
		}
	}

	currencyGroup.currencies = currencies
	return currencyGroup
}

func fetchCurrenciesType() {
	// try insert first
	for _, currencyType := range currencyTypesString {
		queryString := "INSERT INTO currency_type (currency_type, initial_value) VALUES ($1,$2)"
		dataCenter.Db().Exec(queryString, currencyType, 0)
	}

	queryString := "SELECT currency_type, initial_value FROM currency_type"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("err fetch currencty type %v", err)
		return
	}
	for rows.Next() {
		var currencyType string
		var initialValue int64

		err := rows.Scan(&currencyType, &initialValue)
		if err != nil {
			rows.Close()
			log.LogSerious("err fetch currencty type %v", err)
			return
		}
		currencyTypeObject := &CurrenctyType{
			currencyType: currencyType,
			initialValue: initialValue,
		}
		currencyTypes = append(currencyTypes, currencyTypeObject)
	}
	rows.Close()
}

/*
freeze
*/

func (currency *Currency) totalFreezeValue() int64 {
	var total int64
	for _, value := range currency.freezeValueMap.Copy() {
		total += value
	}
	return total
}

func (currency *Currency) totalAvailableValue() int64 {
	return currency.value - currency.totalFreezeValue()
}

func (currency *Currency) freezeValue(reasonString string, value int64) (err error) {
	alreadyFreezeValue := currency.freezeValueMap.Get(reasonString)
	if currency.totalAvailableValue()+alreadyFreezeValue < value {
		return errors.New(l.Get(l.M0016))
	}
	currency.freezeValueMap.Set(reasonString, value)
	return nil
}

func (currency *Currency) getFreezeValue(reasonString string) int64 {
	return currency.freezeValueMap.Get(reasonString)
}
