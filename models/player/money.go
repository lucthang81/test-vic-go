package player

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	x "github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/record"
)

func (player *Player) LockMoney(currencyType string) {
	player.currencyGroup.Lock(currencyType)
}

func (player *Player) UnlockMoney(currencyType string) {
	player.currencyGroup.Unlock(currencyType)
}

func (player *Player) IncreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	temp(currencyType, player, money)
	newMoney, err = player.currencyGroup.IncreaseMoney(money, currencyType, shouldLock)
	if err == nil && currencyType != x.SlotSpin100 {
		player.notifyPlayerDataChange()
	}
	return
}

func (player *Player) DecreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	temp(currencyType, player, -money)
	newMoney, err = player.currencyGroup.DecreaseMoney(money, currencyType, shouldLock)
	if err == nil {
		player.notifyPlayerDataChange()

	}
	return
}

// almost only use for currency.CustomMoney,
// include log currency_record
func (player *Player) SetMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	valueBefore := player.GetMoney(currencyType)
	valueAfter := money
	newMoney, err = player.currencyGroup.SetMoney(money, currencyType, shouldLock)
	if err == nil {
		player.notifyPlayerDataChange()
	}
	record.LogCurrencyRecord(
		player.Id(), record.ACTION_CUSTOM_ROOM_SET_MONEY, "",
		map[string]interface{}{}, currencyType,
		valueBefore, valueAfter, valueAfter-valueBefore,
	)
	return
}

func (player *Player) FreezeMoney(money int64, currencyType string, reasonString string, shouldLock bool) (err error) {
	return player.currencyGroup.FreezeValue(currencyType, reasonString, money, shouldLock)
}

func (player *Player) IncreaseFreezeMoney(increaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	return player.currencyGroup.IncreaseFreezeValue(increaseAmount, currencyType, reasonString, shouldLock)
}

// amount is a positive number
func (player *Player) DecreaseFromFreezeMoney(decreaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	temp(currencyType, player, -decreaseAmount)
	return player.currencyGroup.DecreaseFromFreezeValue(decreaseAmount, currencyType, reasonString, shouldLock)
}

// moneyValue is the changed amount,
// action is reason why the money changed, ussualy it is a player command,
// include log currency_record
func (player *Player) ChangeMoneyAndLog(
	moneyValue int64, currencyType string,
	isDecreaseFreeze bool, roomStringForFreeze string,
	action string, gameCode string, matchId string,
) error {
	//
	valueBefore := player.GetMoney(currencyType)
	var valueAfter int64
	var err error
	if !isDecreaseFreeze {
		if moneyValue >= 0 {
			valueAfter, err = player.IncreaseMoney(moneyValue, currencyType, true)
		} else {
			valueAfter, err = player.DecreaseMoney(-moneyValue, currencyType, true)
		}
	} else {
		valueAfter, err = player.DecreaseFromFreezeMoney(-moneyValue, currencyType, roomStringForFreeze, true)
	}
	if player.playerType == "normal" {
		record.LogCurrencyRecord(
			player.Id(), action, gameCode,
			map[string]interface{}{"match_record_id": matchId}, currencyType,
			valueBefore, valueAfter, moneyValue,
		)
	}
	return err
}

func (player *Player) GetFreezeValue(currencyType string) int64 {
	return player.currencyGroup.TotalFreezeValue(currencyType)
}

func (player *Player) GetFreezeValueForReason(currencyType string, reasonString string) int64 {
	return player.currencyGroup.GetFreezeValue(currencyType, reasonString)
}

func (player *Player) GetAvailableMoney(currencyType string) int64 {
	return player.currencyGroup.TotalAvailableValue(currencyType)
}

func (player *Player) GetMoney(currencyType string) int64 {
	return player.currencyGroup.GetValue(currencyType)
}

// call when money change,
// update moneyChangeCounter and EVENT_EARNING_
func temp(currencyType string, player *Player, moneyValue int64) {
	// counter money change times for unlock CardPay
	if currencyType == x.SlotSpin100 {

	} else {
		player.IncreaseMoney(1, x.SlotSpin100, false)
		if player.GetMoney(x.SlotSpin100) >= 20 &&
			player.GetMoney(x.Money) <= 0 {
			record.SetIapAndCardPay(player.Id(), false, true)
		}
	}
	// event top earning money
	if currencyType == x.TestMoney && player.playerType == "normal" {
		top.GlobalMutex.Lock()
		event := top.MapEvents[top.EVENT_EARNING_TEST_MONEY]
		top.GlobalMutex.Unlock()
		if event != nil {
			event.ChangeBestValue(player.Id(), moneyValue)
		}
	} else if currencyType == x.Money && player.playerType == "normal" {
		top.GlobalMutex.Lock()
		event := top.MapEvents[top.EVENT_EARNING_MONEY]
		top.GlobalMutex.Unlock()
		if event != nil {
			event.ChangeBestValue(player.Id(), moneyValue)
		}
	}
}

//
func (player *Player) TransferMoney(targetPlayerId int64, moneyAmount int64, note string) error {
	target, _ := GetPlayer(targetPlayerId)
	if target == nil {
		return errors.New(l.Get(l.M0042))
	}
	//	if moneyAmount < 1000000 {
	//		return errors.New("Cần chuyển ít nhất 1000000 Kim")
	//	}
	currencyType := x.Money
	player.LockMoney(currencyType)
	if player.GetAvailableMoney(currencyType) < moneyAmount {
		player.UnlockMoney(currencyType)
		return errors.New(l.Get(l.M0016))
	} else {
		valueBefore := player.GetAvailableMoney(currencyType)
		valueAfter, _ := player.DecreaseMoney(moneyAmount, currencyType, false)
		player.UnlockMoney(currencyType)
		rowid := record.LogTransfer(player.Id(), targetPlayerId, moneyAmount)
		msg1 := fmt.Sprintf(l.Get(l.M0043),
			cardgame.HumanFormatNumber(moneyAmount), target.DisplayName(),
			target.Id(), rowid)
		player.CreatePopUp(msg1)
		player.CreateRawMessage(l.Get(l.M0044), msg1)
		record.LogCurrencyRecord(
			player.Id(), record.ACTION_SEND_MONEY, "",
			map[string]interface{}{}, currencyType,
			valueBefore, valueAfter, -moneyAmount,
		)
		target.ChangeMoneyAndLog(
			moneyAmount, currencyType, false, "",
			record.ACTION_RECEIVE_MONEY, "", "")
		msg2 := fmt.Sprintf(l.Get(l.M0045)+
			l.Get(l.M0046),
			cardgame.HumanFormatNumber(moneyAmount), player.DisplayName(),
			player.Id(), rowid, note)
		target.CreatePopUp(msg2)
		target.CreateRawMessage(l.Get(l.M0044), msg2)
		return nil
	}
}
