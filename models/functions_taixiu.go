package models

import (
	"errors"

	"github.com/vic/vic_go/models/gamemini/taixiu"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/utils"
)

// player and game instance in function is exist
func GeneralCheckTaixiu(models *Models, data map[string]interface{}, playerId int64) (
	*taixiu.TaixiuGame, *player.Player, error) {

	gameCode := "taixiu"
	currencyType := utils.GetStringAtPath(data, "currency_type")

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	taixiuGame, isOk := gameInstance.(*taixiu.TaixiuGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return taixiuGame, player, nil
}

func TaixiuGetInfo(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	taixiuGame, player, err := GeneralCheckTaixiu(models, data, playerId)
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

func TaixiuAddBet(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	selection := utils.GetStringAtPath(data, "selection")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	taixiuGame, player, err := GeneralCheckTaixiu(models, data, playerId)
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

func TaixiuBetAsLast(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	taixiuGame, player, err := GeneralCheckTaixiu(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = taixiuGame.BetAsLast(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func TaixiuBetX2Last(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	taixiuGame, player, err := GeneralCheckTaixiu(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = taixiuGame.BetX2Last(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func TaixiuChat(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	senderName := utils.GetStringAtPath(data, "senderName")
	message := utils.GetStringAtPath(data, "message")

	taixiuGame, player, err := GeneralCheckTaixiu(models, data, playerId)
	if err != nil {
		return nil, err
	}

	err = taixiuGame.Chat(player, message, senderName)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func TopNetWorth(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return map[string]interface{}{
		"TopNetWorth": models.TopNetWorth,
	}, nil
}
func TopNumberOfWins(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	rows := rank.GetLeaderboard(rank.RANK_NUMBER_OF_WINS)
	userRows := make([]map[string]interface{}, 0)
	for _, row := range rows {
		user, _ := player.GetPlayer(row.UserId)
		if user == nil {
			continue
		}
		userRow := map[string]interface{}{}
		userRow["Username"] = user.Username()
		userRow["DisplayName"] = user.DisplayName()
		userRow["NumberOfWins"] = row.RKey
		userRows = append(userRows, userRow)
	}
	return map[string]interface{}{"Rows": userRows}, nil
}
func TopTaixiu(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	topDate := utils.GetStringAtPath(data, "TopDate")
	isDesc := utils.GetBoolAtPath(data, "IsDesc")
	r, e := taixiu.TopLoadLeaderboard(topDate, isDesc)
	return map[string]interface{}{"Rows": r}, e
}
