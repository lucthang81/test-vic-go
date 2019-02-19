package models

import (
	"errors"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/taixiu2"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

// player and game instance in function is exist
func GeneralCheckBaucua(models *Models, data map[string]interface{}, playerId int64) (
	*baucua.TaixiuGame, *player.Player, error) {

	gameCode := "baucua"
	currencyType := currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	taixiuGame, isOk := gameInstance.(*baucua.TaixiuGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return taixiuGame, player, nil
}

func BaucuaGetInfo(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	taixiuGame, player, err := GeneralCheckBaucua(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = taixiuGame.GetInfo(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaucuaAddBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	selection := utils.GetStringAtPath(data, "selection")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	taixiuGame, player, err := GeneralCheckBaucua(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = taixiuGame.AddBet(player, selection, moneyValue)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaucuaChat(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	message := utils.GetStringAtPath(data, "message")

	taixiuGame, player, err := GeneralCheckBaucua(models, data, playerId)
	if err != nil {
		return nil, err
	}

	err = taixiuGame.Chat(player, message)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaucuaGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	taixiuGame, player, err := GeneralCheckBaucua(models, data, playerId)
	if err != nil {
		return nil, err
	}

	err = taixiuGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
