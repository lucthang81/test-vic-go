package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/wheel"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckWheel(models *Models, data map[string]interface{}, playerId int64) (
	*wheel.WheelGame, *player.Player, error) {

	gameCode := wheel.WHEEL_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	wheelGame, isOk := gameInstance.(*wheel.WheelGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return wheelGame, player, nil
}

func WheelReceiveFreeSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	wheelGame, player, err := GeneralCheckWheel(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = wheelGame.ReceiveFreeSpin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func WheelGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	wheelGame, player, err := GeneralCheckWheel(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = wheelGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func WheelSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	wheelGame, player, err := GeneralCheckWheel(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = wheelGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
