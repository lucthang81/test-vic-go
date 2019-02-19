package game

import (
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/utils"
)

// room method

func IsRoomRequirementValid(gameInstance GameInterface, requirement int64) bool {
	for _, entry := range gameInstance.BetData().Entries() {
		minBet := entry.Min()
		if requirement == minBet {
			return true
		}
	}
	return false
}
func IsRoomMaxPlayersValid(gameInstance GameInterface, maxPlayer int, roomRequirement int64) bool {
	if maxPlayer >= gameInstance.MinNumberOfPlayers() && maxPlayer <= gameInstance.MaxNumberOfPlayers() {
		return true
	}
	return false
}
func IsPlayerMoneyValidToJoinRoom(gameInstance GameInterface, playerMoney int64, roomRequirement int64) (err error) {
	requireMoney := utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
	if playerMoney < requireMoney {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need":       requireMoney,
			"multiplier": gameInstance.RequirementMultiplier(),
		})
	}
	return nil
}

func IsPlayerMoneyValidToStayInRoom(gameInstance GameInterface, playerMoney int64, roomRequirement int64) (err error) {
	requireMoney := utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
	if playerMoney < requireMoney {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need":       requireMoney,
			"multiplier": gameInstance.RequirementMultiplier(),
		})
	}
	return nil
}

func IsPlayerMoneyValidToCreateRoom(gameInstance GameInterface, playerMoney int64, roomRequirement int64, maxNumberOfPlayers int) (err error) {
	requireMoney := utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
	if playerMoney < requireMoney {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need":       requireMoney,
			"multiplier": gameInstance.RequirementMultiplier(),
		})
	}
	return nil
}

func IsPlayerMoneyValidToBecomeOwner(gameInstance GameInterface, playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	requireMoney := utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
	if playerMoney < requireMoney {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need":       requireMoney,
			"multiplier": gameInstance.RequirementMultiplier(),
		})
	}
	return nil
}

func IsPlayerMoneyValidToStayOwner(gameInstance GameInterface, playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	requireMoney := utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
	if playerMoney < requireMoney {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need":       requireMoney,
			"multiplier": gameInstance.RequirementMultiplier(),
		})
	}
	return nil
}

func MoneyOnTable(gameInstance GameInterface, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
}
func MoneyOnTableForOwner(gameInstance GameInterface, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.RequirementMultiplier())
}
