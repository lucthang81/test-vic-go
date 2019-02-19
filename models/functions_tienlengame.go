package models

import (
	"errors"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/game/tienlen"
	"github.com/vic/vic_go/utils"
)

func tienlenPlayCards(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	playCards := utils.GetStringSliceAtPath(data, "cards")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	tienlenGame := gameInstance.(*tienlen.TienLenGame)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	if !room.ContainsPlayer(player) {
		return nil, errors.New("err:player_not_in_room")
	}

	if room.Session() == nil {
		return nil, errors.New("err:game_not_start_yet")
	}

	if len(playCards) == 0 {
		return nil, errors.New("err:no_cards")
	}

	err = tienlenGame.PlayCards(room.Session(), player, playCards)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func tienlenSkipTurn(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	tienlenGame := gameInstance.(*tienlen.TienLenGame)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	if !room.ContainsPlayer(player) {
		return nil, errors.New("err:player_not_in_room")
	}

	if room.Session() == nil {
		return nil, errors.New("err:game_not_start_yet")
	}

	err = tienlenGame.SkipTurn(room.Session(), player)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
