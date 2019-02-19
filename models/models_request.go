package models

func (models *Models) graduallySendRequestToAllOnlinePlayers(method string, data map[string]interface{}) {
	keys := make([]int64, 0, models.onlinePlayers.Len())
	for _, player := range models.onlinePlayers.Copy() {
		keys = append(keys, player.Id())
	}
	server.SendRequests(method, data, keys)
}
