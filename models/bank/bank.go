package bank

import (
	"fmt"
	"sync"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/record"
)

var mutex sync.Mutex
var dataCenter *datacenter.DataCenter
var banks []*Bank

func init() {
	_ = fmt.Println
	banks = make([]*Bank, 0)
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

func RegisterGame(gameCode string) {
	for _, currencyType := range []string{currency.Money, currency.TestMoney} {
		dataCenter.Db().Exec("INSERT INTO bank (game_code, currency_type) VALUES ($1,$2)", gameCode, currencyType)
	}

}

type Bank struct {
	id           int64
	gameCode     string
	value        int64
	currencyType string
}

func (bank *Bank) AddMoney(amount int64, matchId int64) {
	mutex.Lock()
	defer mutex.Unlock()

	oldMoney := bank.value

	_, err := dataCenter.Db().Exec("UPDATE bank SET value = value + $1 WHERE game_code = $2 AND currency_type = $3",
		amount, bank.gameCode, bank.currencyType)
	if err != nil {
		log.LogSerious("error add value to bank %s, amount %d,currency type %s, error %v", bank.gameCode, amount, bank.currencyType, err)
	}

	bank.value += amount
	record.LogBankRecord(matchId, bank.gameCode, bank.currencyType, oldMoney, bank.value)
}

func (bank *Bank) AddMoneyByBot(amount int64, playerId int64) {
	mutex.Lock()
	defer mutex.Unlock()
	oldMoney := bank.value

	_, err := dataCenter.Db().Exec("UPDATE bank SET value = value + $1 WHERE game_code = $2 AND currency_type = $3",
		amount, bank.gameCode, bank.currencyType)
	if err != nil {
		log.LogSerious("error add value to bank %s, amount %d,currency %s, error %v", bank.gameCode, amount, bank.currencyType, err)
	}
	bank.value += amount
	record.LogBankRecordByBot(playerId, bank.gameCode, bank.currencyType, oldMoney, bank.value)
}

func (bank *Bank) Value() int64 {
	return bank.value
}

func GetBank(gameCode string, currencyType string) *Bank {
	var bank *Bank
	for _, bankInList := range banks {
		if bankInList.gameCode == gameCode && bankInList.currencyType == currencyType {
			bank = bankInList
		}
	}

	if bank == nil {
		//		fmt.Println(gameCode, currencyType)
		row := dataCenter.Db().QueryRow("SELECT id, value FROM bank WHERE game_code = $1 AND currency_type = $2", gameCode, currencyType)
		var id, value int64
		err := row.Scan(&id, &value)
		if err != nil {
			log.LogSerious("err fetch bank %v %v", gameCode, err)
			return nil
		}
		bank = &Bank{
			id:           id,
			value:        value,
			gameCode:     gameCode,
			currencyType: currencyType,
		}
		banks = append(banks, bank)

	}
	return bank
}
