package currency

import (
	"errors"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

func (currencyGroup *CurrencyGroup) GetValue(currencyType string) int64 {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0
	}
	return currency.value
}

func (currencyGroup *CurrencyGroup) Lock(currencyType string) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return
	}
	currency.mutex.Lock()
}

func (currencyGroup *CurrencyGroup) Unlock(currencyType string) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return
	}
	currency.mutex.Unlock()
}

func (currencyGroup *CurrencyGroup) IncreaseMoney(amount int64, currencyType string, shouldLock bool) (newValue int64, err error) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0, errors.New("err:currency_not_found")
	}
	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock()
	}
	if currency.value+amount < 0 {
		log.LogSeriousWithStack("increase money player negative, player id %d, type %s, money %d, decrease %d",
			currencyGroup.playerId, currencyType, currency.value, amount)
	}
	newValue = utils.MaxInt64(currency.value+amount, int64(0))
	err = currency.updateValue(newValue)
	return newValue, err
}

func (currencyGroup *CurrencyGroup) DecreaseMoney(amount int64, currencyType string, shouldLock bool) (newValue int64, err error) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0, errors.New("err:currency_not_found")
	}
	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock()
	}

	if currency.totalAvailableValue() < amount {
		log.LogSeriousWithStack("decrease money player negative, player id %d, type %s, money %d, decrease %d",
			currencyGroup.playerId, currencyType, currency.totalAvailableValue(), amount)
		amount = currency.totalAvailableValue()
	}
	newValue = utils.MaxInt64(currency.value-amount, int64(0))
	err = currency.updateValue(newValue)
	return newValue, err
}

func (currencyGroup *CurrencyGroup) SetMoney(amount int64, currencyType string, shouldLock bool) (newValue int64, err error) {
	currency := currencyGroup.currencies.get(currencyType)
	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock()
	}
	newValue = amount
	err = currency.updateValue(amount)
	return newValue, err
}

func (currency *Currency) updateValue(value int64) (err error) {
	_, err = dataCenter.Db().Exec("UPDATE currency SET value = $1 WHERE player_id = $2 AND currency_type = $3", value, currency.playerId, currency.currencyType)
	if err != nil {
		return err
	}
	currency.value = value
	return nil
}

func (currencyGroup *CurrencyGroup) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	for _, currency := range currencyGroup.currencies.copy() {
		data[currency.currencyType] = currency.value
	}
	return data
}

/*
FREEZE
*/

func (currencyGroup *CurrencyGroup) TotalAvailableValue(currencyType string) int64 {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0
	}
	return currency.totalAvailableValue()
}

func (currencyGroup *CurrencyGroup) TotalFreezeValue(currencyType string) int64 {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0
	}
	return currency.totalFreezeValue()
}

func (currencyGroup *CurrencyGroup) FreezeValue(currencyType string, reasonString string, value int64, shouldLock bool) (err error) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return errors.New("err:currency_not_found")
	}

	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock() // stupid 2 locks for a piece of code => deadlock
	}

	return currency.freezeValue(reasonString, value)
}

// amount is a positive number
func (currencyGroup *CurrencyGroup) DecreaseFromFreezeValue(amount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0, errors.New("err:currency_not_found")
	}
	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock()
	}

	value := currency.getFreezeValue(reasonString)
	if value < amount {
		log.LogSeriousWithStack("decrease freeze money player negative, player id %d, type %s, money %d, decrease %d",
			currencyGroup.playerId, currencyType, value, amount)
		return 0, errors.New(l.Get(l.M0016))
		amount = value
	}
	newValue = utils.MaxInt64(value-amount, int64(0))
	err = currency.freezeValue(reasonString, newValue)
	return currencyGroup.DecreaseMoney(amount, currencyType, false)
}

func (currencyGroup *CurrencyGroup) IncreaseFreezeValue(amount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	if amount < 0 {
		return 0, errors.New("err:value_negative")
	}

	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0, errors.New("err:currency_not_found")
	}
	if shouldLock {
		currency.mutex.Lock()
		defer currency.mutex.Unlock()
	}

	value := currency.getFreezeValue(reasonString)
	newValue = value + amount
	err = currency.freezeValue(reasonString, newValue)
	return newValue, err
}

func (currencyGroup *CurrencyGroup) GetFreezeValue(currencyType string, reasonString string) int64 {
	currency := currencyGroup.currencies.get(currencyType)
	if currency == nil {
		return 0
	}
	return currency.getFreezeValue(reasonString)
}
