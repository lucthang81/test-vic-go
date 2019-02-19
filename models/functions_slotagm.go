package models

import (
	"errors"
	"fmt"

	//	"github.com/vic/vic_go/models/currency"
	//	"github.com/vic/vic_go/models/gamemini"
	slotagm "github.com/vic/vic_go/models/gamemini/slotagm"
	//	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

func SlotagmChooseMoneyPerLine(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	moneyPerLine := utils.GetInt64AtPath(data, "moneyPerLine")

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
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

func SlotagmChoosePaylines(models *Models, data map[string]interface{}, playerId int64) (
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
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
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

func SlotagmGetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
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

func SlotagmSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
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

func SlotagmChooseGoldPotIndex(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	potIndex := utils.GetIntAtPath(data, "potIndex")
	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotagm.ChooseGoldPotIndex(slotGame, player, potIndex)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func SlotagmStopPlaying(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	slotGame, player, err := GeneralCheckGslot(
		models, data, playerId, slotagm.SLOTAGM_GAME_CODE)
	if err != nil {
		return nil, err
	}
	err = slotagm.StopPlaying(slotGame, player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
