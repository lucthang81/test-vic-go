package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slotpoker"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckSlotpoker(models *Models, data map[string]interface{}, playerId int64) (
	*slotpoker.SlotpokerGame, *player.Player, error) {

	gameCode := slotpoker.SLOTPOKER_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	slotpokerGame, isOk := gameInstance.(*slotpoker.SlotpokerGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return slotpokerGame, player, nil
}

func SlotpokerChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotpokerGame, player, err := GeneralCheckSlotpoker(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotpokerGame.ChooseMoneyPerLine(player, moneyPerLine)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotpokerGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotpokerGame, player, err := GeneralCheckSlotpoker(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotpokerGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotpokerSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotpokerGame, player, err := GeneralCheckSlotpoker(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotpokerGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
