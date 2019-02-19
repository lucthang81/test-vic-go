package models

import (
	"errors"
	"github.com/vic/vic_go/models/gsingleplayer/tangkasqu"
	"github.com/vic/vic_go/utils"
)

func TangkasquChooseBaseMoney(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	game := models.GameTangkasqu
	if game == nil {
		return nil, errors.New("M035GameInvalidGameCode")
	}
	err := game.ChooseBaseMoney(playerId, utils.GetFloat64AtPath(data, "BaseMoney"))
	return nil, err
}

func TangkasquCreateMatch(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	game := models.GameTangkasqu
	if game == nil {
		return nil, errors.New("M035GameInvalidGameCode")
	}
	match := &tangkasqu.EggMatch{}
	err := game.InitMatch(playerId, match)
	if err != nil {
		return nil, err
	}
	err = match.SendMove(data)
	return map[string]interface{}{"MatchId": match.MatchId}, nil
}

func TangkasquGetPlayingMatch(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	game := models.GameTangkasqu
	if game == nil {
		return nil, errors.New("M035GameInvalidGameCode")
	}
	match := game.GetPlayingMatch(playerId)
	if match == nil {
		return nil, errors.New("match == nil")
	}
	return match.ToMap(), nil
}

func TangkasquSendMove(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	game := models.GameTangkasqu
	if game == nil {
		return nil, errors.New("M035GameInvalidGameCode")
	}
	match := game.GetPlayingMatch(playerId)
	if match == nil {
		return nil, errors.New("M037GameInvalidMatchId")
	}
	err := match.SendMove(data)
	return nil, err
}

func DragontigerGetCurrentMatch(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	if models.GameDragontiger.SharedMatch != nil &&
		!models.GameDragontiger.SharedMatch.IsFinished {
		models.GameDragontiger.SharedMatch.AddUserId(playerId)
		return models.GameDragontiger.SharedMatch.ToMapForUid(playerId), nil
	}
	return models.GameDragontiger.GetCurrentMatch()
}

func DragontigerMatchesHistory(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	r := []string{}
	if models.GameDragontiger.MatchesHistory != nil {
		for _, e := range models.GameDragontiger.MatchesHistory.Elements {
			r = append(r, e)
		}
	}
	return map[string]interface{}{"Rows": r}, nil
}

func DragontigerSendMove(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	data["UserId"] = playerId
	match := models.GameDragontiger.GetPlayingMatch(playerId)
	if match == nil {
		return nil, errors.New("match == nil")
	}
	err := match.SendMove(data)
	return nil, err
}
