package models

import (
	"github.com/vic/vic_go/utils"
)

func claimGift(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	giftId := utils.GetInt64AtPath(data, "gift_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return player.ClaimGift(giftId)
}

func declineGift(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	giftId := utils.GetInt64AtPath(data, "gift_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return player.DeclineGift(giftId)
}

func ReceiveLoginsGift3(
	models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	return nil, ReceiveLoginsGift(3, playerId)
}

func ReceiveLoginsGift7(
	models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	return nil, ReceiveLoginsGift(7, playerId)
}
