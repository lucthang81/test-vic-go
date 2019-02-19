package models

import (
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/sql"
	"github.com/vic/vic_go/utils"
	"time"
)

func (models *Models) startAllScheduleTasks() {
	if feature.IsLeaderboardAvailable() {
		models.startScheduleForWeeklyLeaderboardCalculation()
		models.startScheduleForDailyLeaderboardCalculation()
	}
	// models.startScheduleToBackup()
	models.startScheduleForCCURecord()
	notification.StartScheduleForPushNotification()
	models.startScheduleToCleanIpAddressMap()

}

func (models *Models) startScheduleForWeeklyLeaderboardCalculation() {
	go models.actuallyWaitAndRunWeeklyLeaderboardCalculation()
}
func (models *Models) startScheduleForDailyLeaderboardCalculation() {
	go models.actuallyWaitAndRunDailyLeaderboardCalculation()
}
func (models *Models) startScheduleForCCURecord() {
	go models.actuallyStartScheduleForCCURecord()
}

func (models *Models) startScheduleToCleanIpAddressMap() {
	go models.actuallyStartScheduleToCleanIpAddressMap()
}

func (models *Models) startScheduleToBackup() {
	go func() {
		timeDuringDay := utils.TimeFromVietnameseTimeString("23-12-2016", "03:00:00")
		targetTime := utils.NextTimeFromTimeOnly(timeDuringDay)
		timeout := utils.NewTimeOut(targetTime.Sub(time.Now()))
		if timeout.Start() {
			sql.BackupTable("casino_x1_db", []string{"player", "card", "currency"})
		}
		defer models.startScheduleToBackup()
	}()
}

func (models *Models) actuallyWaitAndRunWeeklyLeaderboardCalculation() {
	// wait
	// by using time.Now() we are using the time of current server
	// for vn time we should use
	// zone, offset := cur.Zone()
	durationUntilEndOfWeek := utils.TimeDurationUntilEndOfWeek(utils.CurrentTimeInVN()) // hardcode time, this is for end of week
	utils.DelayInDuration(durationUntilEndOfWeek)

	// run
	player.CalculateWeeklyLeaderboardAndDistributePrize()
}

func (models *Models) actuallyWaitAndRunDailyLeaderboardCalculation() {
	// wait
	// by using time.Now() we are using the time of current server
	// for vn time we should use
	// zone, offset := cur.Zone()
	durationUntilEndOfDay := utils.TimeDurationUntilEndOfDay(utils.CurrentTimeInVN()) // hardcode time, this is for end of week
	utils.DelayInDuration(durationUntilEndOfDay)

	// run
	player.CalculateDailyLeaderboardAndDistributePrize()
}

func (models *Models) actuallyStartScheduleForCCURecord() {
	defer func() {
		if r := recover(); r != nil {
			// send crash report
			log.SendMailWithCurrentStack("schedule ccu crash")
		}
	}()
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			totalOnlineCount := 0
			totalNormalOnlineCount := 0
			totalBotOnlineCount := 0

			rawData := make(map[string]interface{})
			for _, currencyType := range []string{currency.Money, currency.TestMoney} {
				data := make(map[string]map[string]int)
				for _, game := range models.games {
					if game.CurrencyType() == currencyType {
						data[game.GameCode()] = make(map[string]int)
						data[game.GameCode()]["online_total_count"] = 0
						data[game.GameCode()]["online_bot_count"] = 0
						data[game.GameCode()]["online_normal_count"] = 0
					}
				}

				for _, gameInstance := range models.games {
					if gameInstance.CurrencyType() == currencyType {
						var numberOfRoom int
						var botCount int
						var normalCount int
						for _, roomInstance := range gameInstance.GameData().Rooms().Copy() {
							if roomInstance.IsRoomOnline() {
								numberOfRoom++
								for _, player := range roomInstance.Players().Copy() {
									if player.PlayerType() == "bot" {
										botCount++
									} else {
										normalCount++
									}
								}
							}
						}
						data[gameInstance.GameCode()]["number_of_rooms"] = gameInstance.GameData().Rooms().Len()
						data[gameInstance.GameCode()]["online_total_count"] = botCount + normalCount
						data[gameInstance.GameCode()]["online_bot_count"] = botCount
						data[gameInstance.GameCode()]["online_normal_count"] = normalCount
					}
				}

				rawData[currencyType] = data
			}

			for _, player := range models.onlinePlayers.Copy() {
				totalOnlineCount++
				if player.PlayerType() == "bot" {
					totalBotOnlineCount++
				} else {
					totalNormalOnlineCount++
				}
			}
			record.LogCCU(totalOnlineCount, totalNormalOnlineCount, totalBotOnlineCount, rawData)
		}
	}

}

func (models *Models) getCCUDataForRecord() (data map[string]interface{}) {
	totalOnlineCount := 0
	totalNormalOnlineCount := 0
	totalBotOnlineCount := 0
	data = make(map[string]interface{})
	for _, currencyType := range []string{currency.Money, currency.TestMoney} {
		gameData := make(map[string]map[string]int)
		for _, game := range models.games {
			if game.CurrencyType() == currencyType {
				gameData[game.GameCode()] = make(map[string]int)
				gameData[game.GameCode()]["online_total_count"] = 0
				gameData[game.GameCode()]["online_bot_count"] = 0
				gameData[game.GameCode()]["online_normal_count"] = 0
				gameData[game.GameCode()]["number_of_rooms"] = 0
				gameData[game.GameCode()]["no_room"] = 0
			}
		}

		for _, gameInstance := range models.games {
			if gameInstance.CurrencyType() == currencyType {
				var numberOfRoom int
				var botCount int
				var normalCount int
				for _, roomInstance := range gameInstance.GameData().Rooms().Copy() {
					if roomInstance.IsRoomOnline() {
						numberOfRoom++
						for _, player := range roomInstance.Players().Copy() {
							if player.PlayerType() == "bot" {
								botCount++
							} else {
								normalCount++
							}
						}
					}
				}
				gameData[gameInstance.GameCode()]["number_of_rooms"] = gameInstance.GameData().Rooms().Len()
				gameData[gameInstance.GameCode()]["online_total_count"] = botCount + normalCount
				gameData[gameInstance.GameCode()]["online_bot_count"] = botCount
				gameData[gameInstance.GameCode()]["online_normal_count"] = normalCount
			}
		}

		data[currencyType] = gameData
	}

	for _, player := range models.onlinePlayers.Copy() {
		totalOnlineCount++
		if player.PlayerType() == "bot" {
			totalBotOnlineCount++
		} else {
			totalNormalOnlineCount++
		}
	}
	data["online_total_count"] = totalOnlineCount
	data["online_bot_count"] = totalBotOnlineCount
	data["online_normal_count"] = totalNormalOnlineCount

	return data
}

func (models *Models) GetNHumanOnline() int {
	totalOnlineCount := 0
	totalNormalOnlineCount := 0
	totalBotOnlineCount := 0
	for _, player := range models.onlinePlayers.Copy() {
		totalOnlineCount++
		if player.PlayerType() == "bot" {
			totalBotOnlineCount++
		} else {
			totalNormalOnlineCount++
		}
	}
	return totalNormalOnlineCount
}

func (models *Models) getCCUDataEachGameForRecord(gameCode string, currencyType string) (data map[string]interface{}) {
	data = make(map[string]interface{})

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance != nil {
		var numberOfRooms, onlineTotalCount, onlineBotCount, onlineNormalCount int64
		rooms := make([]map[string]interface{}, 0)
		numberOfRooms = int64(gameInstance.GameData().Rooms().Len())
		for _, room := range gameInstance.GameData().Rooms().Copy() {
			roomData := room.SerializedDataWithFields(nil, []string{"players_id",
				"bets",
				"ready_players_id",
				"session",
				"password",
				"last_match_results",
				"moneys_on_table"})

			playerList := make([]map[string]interface{}, 0)
			for _, playerInstance := range room.Players().Copy() {
				playerData := playerInstance.SerializedDataMinimal()
				playerData["money"] = utils.FormatWithComma(playerInstance.GetMoney(currencyType))
				playerData["player_type"] = playerInstance.PlayerType()
				if playerInstance.PlayerType() == "bot" {
					onlineBotCount++
				} else {
					playerData["ip_address"] = playerInstance.IpAddress()
					onlineNormalCount++
				}
				onlineTotalCount++
				playerList = append(playerList, playerData)
			}
			roomData["player_list"] = playerList
			if room.Session() != nil && gameCode != "roulette" {
				roomData["is_playing"] = true
			} else {
				roomData["is_playing"] = false
			}
			rooms = append(rooms, roomData)
		}

		data["game_code"] = gameCode
		data["online_total_count"] = onlineTotalCount
		data["online_bot_count"] = onlineBotCount
		data["online_normal_count"] = onlineNormalCount
		data["number_of_rooms"] = numberOfRooms
		data["rooms"] = rooms
		data["is_minigame"] = false
		return data
	} else {
		return data
	}
}
