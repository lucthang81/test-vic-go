package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slotbacay"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckSlotbacay(models *Models, data map[string]interface{}, playerId int64) (
	*slotbacay.SlotbacayGame, *player.Player, error) {

	gameCode := slotbacay.SLOTBACAY_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	slotbacayGame, isOk := gameInstance.(*slotbacay.SlotbacayGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return slotbacayGame, player, nil
}

func SlotbacayChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotbacayGame, player, err := GeneralCheckSlotbacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotbacayGame.ChooseMoneyPerLine(player, moneyPerLine)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotbacayGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotbacayGame, player, err := GeneralCheckSlotbacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotbacayGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotbacaySpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotbacayGame, player, err := GeneralCheckSlotbacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotbacayGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
