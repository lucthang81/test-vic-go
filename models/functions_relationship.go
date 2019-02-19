package models

import (
	"github.com/vic/vic_go/utils"
)

func getFriendList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["friends"] = player.GetFriendListData()
	return responseData, nil
}
func acceptFriendRequest(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	fromPlayerId := utils.GetInt64AtPath(data, "from_player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	err = player.AcceptFriendRequest(fromPlayerId)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["from_player_id"] = fromPlayerId
	return responseData, nil
}
func declineFriendRequest(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	fromPlayerId := utils.GetInt64AtPath(data, "from_player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	err = player.DeclineFriendRequest(fromPlayerId)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["from_player_id"] = fromPlayerId
	return responseData, nil
}
func sendFriendRequest(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	toPlayerId := utils.GetInt64AtPath(data, "to_player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	becomeFriendInstantly, err := player.SendFriendRequest(toPlayerId)
	if err != nil {
		return nil, err
	}

	responseData = make(map[string]interface{})
	responseData["become_friend_instantly"] = becomeFriendInstantly
	return responseData, nil
}

func unfriend(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	toPlayerId := utils.GetInt64AtPath(data, "to_player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	err = player.Unfriend(toPlayerId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getNumberOfFriends(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	responseData = make(map[string]interface{})
	responseData["number_of_friends"] = player.GetNumberOfFriends()

	return responseData, nil
}
