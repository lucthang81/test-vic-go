package models

import (
	"errors"
	"fmt"

	//	"github.com/vic/vic_go/models/currency"
	//	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/slotax1to5"
	//	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

func Slotax1to5ChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotax1to5.SLOTAX1TO5_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotGame.ChooseMoneyPerLine(player, moneyPerLine)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Slotax1to5ChoosePaylines(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	paylineIndexs := []int{}
	paylineIndexsSI, isOk := data["paylineIndexs"].([]interface{})
	if !isOk {
		return nil, errors.New("1 paylineIndexs type must be []int")
	}
	for _, e := range paylineIndexsSI {
		// fmt.Printf("haha %T %v\n", e, e)
		eFloat64, isOkL1 := e.(float64)
		if !isOkL1 {
			return nil, errors.New("2 paylineIndexs type must be []int")
		}
		eInt := int(eFloat64)
		paylineIndexs = append(paylineIndexs, eInt)
	}

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotax1to5.SLOTAX1TO5_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotGame.ChoosePaylines(player, paylineIndexs)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Slotax1to5GetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotax1to5.SLOTAX1TO5_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Slotax1to5Spin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotax1to5.SLOTAX1TO5_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Slotax1to5Choose(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotax1to5.SLOTAX1TO5_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotax1to5.Choose(slotGame, player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
