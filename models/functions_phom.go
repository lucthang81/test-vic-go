package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/game/phom"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckPhom(models *Models, data map[string]interface{}, playerId int64) (
	*phom.PhomGame, *player.Player, error) {

	// gameCode := "phom"
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err: invalid game_code or currency_type")
	}
	phomGame, isOk := gameInstance.(*phom.PhomGame)
	if !isOk {
		return nil, nil, errors.New("err: this command only for phom or phomSolo")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return phomGame, player, nil
}

func PhomDrawCard(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.DrawCard(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomEatCard(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.EatCard(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomPopCard(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	cardString := utils.GetStringAtPath(data, "cardString")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.PopCard(player, roomId, cardString)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomHangCard(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")
	cardString := utils.GetStringAtPath(data, "cardString")
	targetPlayerId := utils.GetInt64AtPath(data, "targetPlayerId")
	comboId := utils.GetStringAtPath(data, "comboId")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.HangCard(player, roomId, cardString, targetPlayerId, comboId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomAutoHangCards(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.AutoHangCards(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomAutoShowCombos(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.AutoShowCombos(player, roomId)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func PhomShowComboByUser(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	roomId := utils.GetInt64AtPath(data, "room_id")

	cardsToShow := []string{}
	temp, isOk := data["cardsToShow"].([]interface{})
	if !isOk {
		return nil, errors.New("err:comboTypeMustBe[]string")
	}
	for _, eL1 := range temp {
		tempL1, isOkL1 := eL1.(string)
		if !isOkL1 {
			return nil, errors.New("err:cardsToShowTypeMustBe[]string")
		}
		cardsToShow = append(cardsToShow, tempL1)
	}

	gameI, player, err := GeneralCheckPhom(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = gameI.ShowComboByUser(player, roomId, cardsToShow)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
