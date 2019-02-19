package tienlen

type PlayerData struct {
	id           int64
	order        int
	money        int64
	moneyOnTable int64
	bet          int64
	turnTime     float64
}

func (playerInstance *PlayerData) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = playerInstance.id
	data["order"] = playerInstance.order
	data["money"] = playerInstance.money
	data["turn_time"] = playerInstance.turnTime
	data["money_on_table"] = playerInstance.moneyOnTable
	data["bet"] = playerInstance.bet
	return data
}
