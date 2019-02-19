package models

import (
	"errors"

	"github.com/vic/vic_go/models/game/xocdia2"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

// player and game instance in function is exist
func GeneralCheckXocdia(models *Models, data map[string]interface{}, playerId int64) (
	*xocdia2.XocdiaGame, *player.Player, error) {

	gameCode := "xocdia2"
	currencyType := utils.GetStringAtPath(data, "currency_type")

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	xocdiaGame, isOk := gameInstance.(*xocdia2.XocdiaGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return xocdiaGame, player, nil
}

func Xocdia2AddBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	betSelection := utils.GetStringAtPath(data, "betSelection")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	gameI, player, err := GeneralCheckXocdia(models, data, playerId)
	err = gameI.AddBet(player, roomId, betSelection, moneyValue)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Xocdia2BetEqualLast(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckXocdia(models, data, playerId)
	err = gameI.BetEqualLast(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Xocdia2BetDoubleLast(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckXocdia(models, data, playerId)
	err = gameI.BetDoubleLast(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Xocdia2AcceptBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	betSelection := utils.GetStringAtPath(data, "betSelection")
	ratio := utils.GetFloat64AtPath(data, "ratio")

	gameI, player, err := GeneralCheckXocdia(models, data, playerId)
	err = gameI.AcceptBet(player, roomId, betSelection, ratio)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Xocdia2BecomeHost(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckXocdia(models, data, playerId)
	err = gameI.BecomeHost(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
