package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slotbongda"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckSlotbongda(models *Models, data map[string]interface{}, playerId int64) (
	*slotbongda.SlotGame, *player.Player, error) {

	gameCode := slotbongda.SLOT_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	slotGame, isOk := gameInstance.(*slotbongda.SlotGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return slotGame, player, nil
}

func SlotbongdaChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotGame, player, err := GeneralCheckSlotbongda(models, data, playerId)
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

func SlotbongdaChoosePaylines(models *Models, data map[string]interface{}, playerId int64) (
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

	slotGame, player, err := GeneralCheckSlotbongda(models, data, playerId)
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

func SlotbongdaGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckSlotbongda(models, data, playerId)
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

func SlotbongdaSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckSlotbongda(models, data, playerId)
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
