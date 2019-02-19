package game

import (
	"github.com/vic/vic_go/utils"
)

func MoneyAfterApplyMultiplier(money int64, multiplier float64) int64 {
	return int64(utils.Round((float64(money) * multiplier)))
}

func ContainPlayer(players []GamePlayer, player GamePlayer) bool {
	for _, playerInList := range players {
		if playerInList.Id() == player.Id() {
			return true
		}
	}
	return false
}

func RemovePlayer(players []GamePlayer, player GamePlayer) []GamePlayer {
	temp := make([]GamePlayer, 0)
	for _, playerInList := range players {
		if playerInList.Id() != player.Id() {
			temp = append(temp, playerInList)
		}
	}
	return temp
}

func GetIdFromPlayers(players []GamePlayer) []int64 {
	keys := make([]int64, 0, len(players))
	for _, player := range players {
		keys = append(keys, player.Id())
	}
	return keys
}

func RemoveIds(ids []int64, idsToRemove []int64) []int64 {
	results := make([]int64, 0)
	for _, id := range ids {
		if !utils.ContainsByInt64(idsToRemove, id) {
			results = append(results, id)
		}
	}
	return results
}

func ClonePlayers(players []GamePlayer) []GamePlayer {
	temp := make([]GamePlayer, len(players))
	copy(temp, players)
	return temp
}

func GetPlayer(players []GamePlayer, playerId int64) GamePlayer {
	for _, playerInList := range players {
		if playerInList.Id() == playerId {
			return playerInList
		}
	}
	return nil
}
