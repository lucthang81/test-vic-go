package models

func getOtherNotificationList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = player.GetNotFriendRequestNotificationList()
	return responseData, nil
}

func getFriendRequestNotificationList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = player.GetFriendRequestNotificationList()
	return responseData, nil
}

func getTotalNumberOfNotifications(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	responseData = player.GetTotalNumberOfNotifications()
	return responseData, nil
}
