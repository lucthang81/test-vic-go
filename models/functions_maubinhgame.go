package models

import (
	"errors"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/game/maubinh"
	"github.com/vic/vic_go/utils"
)

func maubinhUploadCards(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	roomId := utils.GetInt64AtPath(data, "room_id")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	cardsData := utils.GetMapAtPath(data, "cards_data")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	maubinhGame := gameInstance.(*maubinh.MauBinhGame)
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

	if len(cardsData) == 0 {
		return nil, errors.New("err:no_cards")
	}

	err = maubinhGame.UploadCards(room.Session(), player, cardsData)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func maubinhFinishOrganizeCards(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	roomId := utils.GetInt64AtPath(data, "room_id")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	cardsData := utils.GetMapAtPath(data, "cards_data")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	maubinhGame := gameInstance.(*maubinh.MauBinhGame)
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

	if len(cardsData) == 0 {
		return nil, errors.New("err:no_cards")
	}

	err = maubinhGame.FinishOrganizeCards(room.Session(), player, cardsData)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func maubinhStartOrganizeCardsAgain(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	roomId := utils.GetInt64AtPath(data, "room_id")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	maubinhGame := gameInstance.(*maubinh.MauBinhGame)
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

	err = maubinhGame.StartOrganizeCardsAgain(room.Session(), player)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
