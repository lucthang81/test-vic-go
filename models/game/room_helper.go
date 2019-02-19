package game

import (
	"errors"

	"github.com/vic/vic_go/language"
)

func (room *Room) unfreezeMoneyFromPlayer(playerInstance GamePlayer) {
	playerInstance.FreezeMoney(0, room.currencyType, room.GetRoomIdentifierString(), true)
}

func (room *Room) GetTotalPlayerMoney(playerId int64) int64 {
	playerInstance := room.getPlayer(playerId)
	if playerInstance == nil {
		return 0
	}
	return playerInstance.GetMoney(room.currencyType)
}

func (room *Room) GetMoneyOnTable(playerId int64) int64 {
	playerInstance := room.getPlayer(playerId)
	if playerInstance == nil {
		return 0
	}
	return playerInstance.GetFreezeValueForReason(room.currencyType, room.GetRoomIdentifierString())
}

func (room *Room) SetMoneyOnTable(playerId int64, value int64, shouldLock bool) (err error) {
	playerInstance := room.getPlayer(playerId)
	if playerInstance == nil {
		return errors.New(l.Get(l.M0065))
	}
	return playerInstance.FreezeMoney(value, room.currencyType, room.GetRoomIdentifierString(), shouldLock)
}
