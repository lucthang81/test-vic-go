package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slot"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckSlot(models *Models, data map[string]interface{}, playerId int64) (
	*slot.SlotGame, *player.Player, error) {

	gameCode := slot.SLOT_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	slotGame, isOk := gameInstance.(*slot.SlotGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return slotGame, player, nil
}

func SlotChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotGame, player, err := GeneralCheckSlot(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotGame.ChooseMoneyPerLine(player, moneyPerLine)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotChoosePaylines(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	paylineIndexs := []int{}
	paylineIndexsSI, isOk := data["paylineIndexs"].([]interface{})
	if !isOk {
		return nil, errors.New("1 paylineIndexs type must be []int")
	}
	for _, e := range paylineIndexsSI {
		// fmt.Printf("haha %T %v\n", e, e)
		eFloat64, isOkL1 := e.(float64)
		if !isOkL1 {
			return nil, errors.New("2 paylineIndexs type must be []int")
		}
		eInt := int(eFloat64)
		paylineIndexs = append(paylineIndexs, eInt)
	}

	slotGame, player, err := GeneralCheckSlot(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotGame.ChoosePaylines(player, paylineIndexs)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckSlot(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckSlot(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
