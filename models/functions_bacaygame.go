package models

import (
	"errors"

	"github.com/vic/vic_go/models/game/bacay2"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

// player and game instance in function is exist
func GeneralCheckBacay(models *Models, data map[string]interface{}, playerId int64) (
	*bacay2.BaCayGame, *player.Player, error) {

	gameCode := "bacay2"
	currencyType := utils.GetStringAtPath(data, "currency_type")

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	baCayGame, isOk := gameInstance.(*bacay2.BaCayGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return baCayGame, player, nil
}

func BaCayMandatoryBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.MandatoryBet(player, roomId, moneyValue)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaCayJoinGroupBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.JoinGroupBet(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaCayJoinPairBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")
	enemyId := utils.GetInt64AtPath(data, "enemyId")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.JoinPairBet(player, roomId, moneyValue, enemyId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaCayJoinAllPairBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.JoinAllPairBet(player, roomId, moneyValue)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaCayJoinAllPairBet2(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.JoinAllPairBet2(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func BaCayBecomeOwner(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	choice := utils.GetBoolAtPath(data, "choice")

	baCayGame, player, err := GeneralCheckBacay(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = baCayGame.BecomeOwner(player, roomId, choice)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
