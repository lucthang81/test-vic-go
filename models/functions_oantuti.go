package models

import (
	"errors"
	"fmt"

	//	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemm/oantuti"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckOantuti(models *Models, data map[string]interface{}, playerId int64) (
	*oantuti.OantutiGame, *player.Player, error) {

	gameCode := oantuti.GAME_CODE_OANTUTI

	gameInstance := models.GetGameMM(gameCode)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_gamecode")
	}
	oantutiG, isOk := gameInstance.(*oantuti.OantutiGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return oantutiG, player, nil
}

// chọn mức tiền 1000 || 2000 || 5000
func OantutiChooseRule(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	RequirementMoney := utils.GetInt64AtPath(data, "RequirementMoney")

	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.ChooseRule(player, RequirementMoney)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiFindMatch(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.FindMatch(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiStopFindingMatch(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.StopFindingMatch(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiChooseHandPaper(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.ChooseHandPaper(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiChooseHandRock(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.ChooseHandRock(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiChooseHandScissors(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.ChooseHandScissors(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiGetUserInfo(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.GetUserInfo(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

//
func OantutiGetTop(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	oantutiG, player, err := GeneralCheckOantuti(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = oantutiG.GetTop(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
