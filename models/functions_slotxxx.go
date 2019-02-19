package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	//	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/slotxxx"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckSlotxxx(models *Models, data map[string]interface{}, playerId int64) (
	*slotxxx.SlotxxxGame, *player.Player, error) {

	gameCode := slotxxx.SLOTXXX_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	slotxxxGame, isOk := gameInstance.(*slotxxx.SlotxxxGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return slotxxxGame, player, nil
}

func SlotxxxChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.ChooseMoneyPerLine(player, moneyPerLine)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxGetMatchInfo(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.GetMatchInfo(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxStopPlaying(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.StopPlaying(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxSelectSmall(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.SelectSmall(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotxxxSelectBig(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotxxxGame, player, err := GeneralCheckSlotxxx(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = slotxxxGame.SelectBig(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
