package player

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/utils"
	"strings"
	"time"
)

type Prize struct {
	id       int64
	imageUrl string
	fromRank int64
	toRank   int64
	prize    int64
	gameCode string
}

const TotalGainWeeklyPrizeDatabaseTableName string = "total_gain_weekly_prize"
const BiggestWinWeeklyPrizeDatabaseTableName string = "biggest_win_weekly_prize"

func (prize *Prize) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = prize.id
	data["image_url"] = prize.imageUrl
	data["from_rank"] = prize.fromRank
	data["to_rank"] = prize.toRank
	data["prize"] = prize.prize
	return data
}

func CalculateWeeklyLeaderboardAndDistributePrize() {
	// for _, gameInstance := range games {
	// 	calculateBiggestWinPrize(gameInstance)
	// 	calculateTotalGainPrize(gameInstance)
	// }
	cleanupWeeklyStats()

	timeout := utils.NewTimeOut(24 * 7 * time.Hour)
	if timeout.Start() {
		go CalculateWeeklyLeaderboardAndDistributePrize()
	}
}

func CalculateDailyLeaderboardAndDistributePrize() {
	// for _, gameInstance := range games {
	// 	calculateBiggestWinPrize(gameInstance)
	// 	calculateTotalGainPrize(gameInstance)
	// }
	cleanupDailyStats()

	timeout := utils.NewTimeOut(24 * time.Hour)
	if timeout.Start() {
		go CalculateDailyLeaderboardAndDistributePrize()
	}
}

func calculateBiggestWinPrize(gameInstance game.GameInterface, currencyType string) {
	// biggest win
	biggestWinPrizeList := getBiggestWinPrizeList(gameInstance)
	rangesToFetch := getFetchRangeFromPrizeList(biggestWinPrizeList)

	for _, rangeToFetch := range rangesToFetch {
		achievements := fetchAchievementsByRange(rangeToFetch, "biggest_win_this_week", gameInstance.GameCode(), currencyType)

		rank := rangeToFetch[0]
		var lastColumnValue int64
		for _, achievement := range achievements {

			if lastColumnValue == 0 {
				lastColumnValue = achievement.biggestWinThisWeek
			} else if lastColumnValue != achievement.biggestWinThisWeek {
				rank++
				lastColumnValue = achievement.biggestWinThisWeek
			}

			player, err := GetPlayer(achievement.playerId)
			if err != nil {
				log.LogSerious("error when get player %d in calculate leaderboard %s", achievement.playerId, err.Error())
				continue
			}
			for _, prize := range biggestWinPrizeList {
				if rank >= prize.fromRank && rank <= prize.toRank {
					additionalData := make(map[string]interface{})
					additionalData["rank"] = rank
					additionalData["game_code"] = gameInstance.GameCode()
					additionalData["currency_type"] = gameInstance.CurrencyType()
					expiredDate := time.Now().Add(7 * time.Hour * 24)
					_, err = player.giftManager.createGift("leaderboard_biggest_win_weekly", currencyType, prize.prize, additionalData, expiredDate)
					if err != nil {
						log.LogSerious("err when create gift %s", err.Error())
					}
				}
			}
		}
	}
}

func calculateTotalGainPrize(gameInstance game.GameInterface, currencyType string) {
	totalGainPrizeList := getTotalGainPrizeList(gameInstance)
	rangesToFetch := getFetchRangeFromPrizeList(totalGainPrizeList)

	for _, rangeToFetch := range rangesToFetch {
		achievements := fetchAchievementsByRange(rangeToFetch, "total_gain_this_week", gameInstance.GameCode(), currencyType)
		rank := rangeToFetch[0]
		var lastColumnValue int64
		for _, achievement := range achievements {
			if lastColumnValue == 0 {
				lastColumnValue = achievement.totalGainThisWeek
			} else if lastColumnValue != achievement.totalGainThisWeek {
				rank++
				lastColumnValue = achievement.totalGainThisWeek
			}
			player, err := GetPlayer(achievement.playerId)
			if err != nil {
				log.LogSerious("error when get player %d in calculate leaderboard %s", achievement.playerId, err.Error())
				continue
			}
			for _, prize := range totalGainPrizeList {
				if rank >= prize.fromRank && rank <= prize.toRank {
					additionalData := make(map[string]interface{})
					additionalData["rank"] = rank
					additionalData["game_code"] = gameInstance.GameCode()
					additionalData["currency_type"] = gameInstance.CurrencyType()
					expiredDate := time.Now().Add(7 * time.Hour * 24)
					player.giftManager.createGift("leaderboard_total_gain_weekly", currencyType, prize.prize, additionalData, expiredDate)
				}
			}
		}
	}
}

func cleanupWeeklyStats() {
	for _, playerInstance := range players.Copy() {
		for _, achi := range playerInstance.achievementManager.achievements {
			achi.biggestWinThisWeek = 0
			achi.totalGainThisWeek = 0
		}
	}

	// queryString := fmt.Sprintf("UPDATE %s SET biggest_win_this_week = 0, total_gain_this_week = 0", AchievementDatabaseTableName)
	// _, err := dataCenter.Db().Exec(queryString)
	// if err != nil {
	// 	log.LogSerious("err when cleanup weekly stats %s", err.Error())
	// }
}

func cleanupDailyStats() {
	for _, playerInstance := range players.Copy() {
		for _, achi := range playerInstance.achievementManager.achievements {
			achi.biggestWinThisDay = 0
			achi.totalGainThisDay = 0
		}
	}

	// queryString := fmt.Sprintf("UPDATE %s SET biggest_win_this_day = 0, total_gain_this_day = 0", AchievementDatabaseTableName)
	// _, err := dataCenter.Db().Exec(queryString)
	// if err != nil {
	// 	log.LogSerious("err when cleanup daily stats %s", err.Error())
	// }
}

func getBiggestWinPrizeList(gameInstance game.GameInterface) (list []*Prize) {
	queryPrizeString := fmt.Sprintf("SELECT id, image_url, from_rank, to_rank, prize, game_code FROM %s WHERE game_code = $1 ORDER BY from_rank", BiggestWinWeeklyPrizeDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryPrizeString, gameInstance.GameCode())
	if err != nil {
		log.LogSerious("err when fetch biggestwinprizelist %s", err.Error())
		return nil
	}
	list = make([]*Prize, 0)
	for rows.Next() {
		var id int64
		var imageUrl []byte
		var fromRank int64
		var toRank int64
		var prize int64
		var gameCode []byte
		err = rows.Scan(&id, &imageUrl, &fromRank, &toRank, &prize, &gameCode)
		if err != nil {
			rows.Close()
			log.LogSerious("err when fetch biggestwinprizelist %s", err.Error())
			return nil
		}
		prizeObject := &Prize{
			id:       id,
			imageUrl: string(imageUrl),
			fromRank: fromRank,
			toRank:   toRank,
			prize:    prize,
			gameCode: string(gameCode),
		}
		list = append(list, prizeObject)
	}
	rows.Close()
	return list
}

func getBiggestWinPrizeListData(gameInstance game.GameInterface) (data []map[string]interface{}) {
	list := getBiggestWinPrizeList(gameInstance)
	data = make([]map[string]interface{}, 0)
	for _, prize := range list {
		data = append(data, prize.SerializedData())
	}
	return data
}

func getTotalGainPrizeList(gameInstance game.GameInterface) (list []*Prize) {
	queryPrizeString := fmt.Sprintf("SELECT id, image_url, from_rank, to_rank, prize, game_code FROM %s WHERE game_code = $1 ORDER BY from_rank", TotalGainWeeklyPrizeDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryPrizeString, gameInstance.GameCode())
	if err != nil {
		log.LogSerious("err when fetch biggestwinprizelist %s", err.Error())
		return nil
	}
	list = make([]*Prize, 0)
	for rows.Next() {
		var id int64
		var imageUrl []byte
		var fromRank int64
		var toRank int64
		var prize int64
		var gameCode []byte
		err = rows.Scan(&id, &imageUrl, &fromRank, &toRank, &prize, &gameCode)
		if err != nil {
			rows.Close()
			log.LogSerious("err when fetch totalgainprizelist %s", err.Error())
			return nil
		}
		prizeObject := &Prize{
			id:       id,
			imageUrl: string(imageUrl),
			fromRank: fromRank,
			toRank:   toRank,
			prize:    prize,
			gameCode: string(gameCode),
		}
		list = append(list, prizeObject)
	}
	rows.Close()
	return list
}

func createWeeklyReward(imageUrl string, fromRank int64, toRank int64, prize int64, rewardType string, gameCode string) (err error) {
	databaseName := TotalGainWeeklyPrizeDatabaseTableName
	if rewardType == "biggest_win" {
		databaseName = BiggestWinWeeklyPrizeDatabaseTableName
	}
	queryPrizeString := fmt.Sprintf("INSERT INTO %s (image_url, from_rank, to_rank, prize, game_code) VALUES ($1,$2,$3,$4,$5)", databaseName)
	_, err = dataCenter.Db().Exec(queryPrizeString, imageUrl, fromRank, toRank, prize, gameCode)
	return err
}

func editWeeklyReward(id int64, imageUrl string, fromRank int64, toRank int64, prize int64, rewardType string, gameCode string) (err error) {
	databaseName := TotalGainWeeklyPrizeDatabaseTableName
	if rewardType == "biggest_win" {
		databaseName = BiggestWinWeeklyPrizeDatabaseTableName
	}
	queryPrizeString := fmt.Sprintf("UPDATE %s SET image_url = $1, from_rank = $2, to_rank = $3, prize = $4 WHERE id = $5", databaseName)
	_, err = dataCenter.Db().Exec(queryPrizeString, imageUrl, fromRank, toRank, prize, id)
	return err
}

func deleteWeeklyReward(id int64, rewardType string) (err error) {
	databaseName := TotalGainWeeklyPrizeDatabaseTableName
	if rewardType == "biggest_win" {
		databaseName = BiggestWinWeeklyPrizeDatabaseTableName
	}
	queryPrizeString := fmt.Sprintf("DELETE FROM %s WHERE id = $1", databaseName)
	_, err = dataCenter.Db().Exec(queryPrizeString, id)
	return err
}

func getBiggestTotalGainPlayersData(gameInstance game.GameInterface) (data []map[string]interface{}) {
	list := getTotalGainPrizeList(gameInstance)
	data = make([]map[string]interface{}, 0)
	for _, prize := range list {
		data = append(data, prize.SerializedData())
	}
	return data
}

func getTotalGainPrizeListData(gameInstance game.GameInterface) (data []map[string]interface{}) {
	list := getTotalGainPrizeList(gameInstance)
	data = make([]map[string]interface{}, 0)
	for _, prize := range list {
		data = append(data, prize.SerializedData())
	}
	return data
}

// helper

func getFetchRangeFromPrizeList(prizes []*Prize) (rangesToFetch [][]int64) {
	rangesToFetch = make([][]int64, 0)
	currentRange := make([]int64, 0) // index 1 will be from, index 2 will be to (included)
	for _, prize := range prizes {
		if len(currentRange) == 0 {
			currentRange = append(currentRange, prize.fromRank)
			currentRange = append(currentRange, prize.toRank)
		} else {
			if prize.fromRank <= currentRange[1]+1 {
				currentRange[1] = prize.toRank
			} else {
				rangesToFetch = append(rangesToFetch, currentRange)
				currentRange = make([]int64, 0)
				currentRange = append(currentRange, prize.fromRank)
				currentRange = append(currentRange, prize.toRank)
			}
		}
	}
	if len(currentRange) != 0 {
		rangesToFetch = append(rangesToFetch, currentRange)
	}
	return rangesToFetch
}

func fetchAchievementsByRange(rangeToFetch []int64, column string, currencyType string, gameCode string) (achievements []*Achievement) {
	var startDate, endDate time.Time
	if strings.Contains(column, "day") {
		startDate = utils.StartOfDayFromTime(time.Now())
		endDate = utils.EndOfDayFromTime(time.Now())
	} else {
		startDate = utils.StartOfWeekFromTime(time.Now())
		endDate = utils.EndOfWeekFromTime(time.Now())
	}
	// get achievement
	queryString := fmt.Sprintf("SELECT DISTINCT ON (%s) %s FROM %s WHERE game_code = $1 AND currency_type = $2 AND %s > 0"+
		" AND updated_at >= $3 AND updated_at <= $4 ORDER BY %s DESC LIMIT %d OFFSET %d ",
		column,
		column,
		AchievementDatabaseTableName,
		column,
		column,
		rangeToFetch[1]-rangeToFetch[0]+1,
		rangeToFetch[0]-1)
	fmt.Println(queryString)
	rows, err := dataCenter.Db().Query(queryString, gameCode, currencyType, startDate, endDate)
	if err != nil {
		log.LogSerious("error fetch achievement %s", err.Error())
		return nil
	}
	values := make([]string, 0)
	for rows.Next() {
		var value int64
		err = rows.Scan(&value)
		if err != nil {
			rows.Close()
			log.LogSerious("error fetch achievement %s", err.Error())
			return nil
		}
		values = append(values, fmt.Sprintf("%d", value))
	}
	rows.Close()

	achievements = make([]*Achievement, 0)
	if len(values) == 0 {
		return achievements
	}

	paramsString := strings.Join(values, ",")
	// second query
	queryString = fmt.Sprintf("SELECT id, player_id, %s FROM %s WHERE game_code = $1 AND currency_type = $2 AND %s IN (%s) "+
		" AND updated_at >= $3 AND updated_at <= $4 ORDER BY %s DESC",
		column,
		AchievementDatabaseTableName,
		column,
		paramsString,
		column)
	fmt.Println(queryString)
	rows, err = dataCenter.Db().Query(queryString, gameCode, currencyType, startDate, endDate)
	if err != nil {
		log.LogSerious("error fetch achievement %s", err.Error())
		return nil
	}

	for rows.Next() {
		var id int64
		var playerId int64
		var value int64
		err = rows.Scan(&id, &playerId, &value)

		if err != nil {
			rows.Close()
			log.LogSerious("error fetch achievement %s", err.Error())
			return nil
		}
		achievement := &Achievement{
			playerId: playerId,
		}
		achievement.id = id
		if column == "total_gain_this_week" {
			achievement.totalGainThisWeek = value
		} else {
			achievement.biggestWinThisWeek = value
		}

		achievements = append(achievements, achievement)
	}
	rows.Close()
	return achievements
}

func fetchPlayersInLeaderboard(limit int64, offset int64, column string, currencyType string, gameCode string) (playersData []map[string]interface{}, total int64, err error) {
	var startDate, endDate time.Time
	if strings.Contains(column, "day") {
		startDate = utils.StartOfDayFromTime(time.Now())
		endDate = utils.EndOfDayFromTime(time.Now())
	} else {
		startDate = utils.StartOfWeekFromTime(time.Now())
		endDate = utils.EndOfWeekFromTime(time.Now())
	}
	// get achievement
	total = dataCenter.GetInt64FromQuery(fmt.Sprintf("SELECT COUNT(player_id) FROM %s WHERE game_code = $1 AND currency_type = $2"+
		" AND %s > 0 AND updated_at >= $3 AND updated_at <= $4",
		AchievementDatabaseTableName, column), gameCode, currencyType, startDate, endDate)

	queryString := fmt.Sprintf("SELECT player_id, %s FROM %s WHERE game_code = $1 AND currency_type = $2 AND %s > 0 "+
		" AND updated_at >= $3 AND updated_at <= $4 ORDER BY %s DESC LIMIT %d OFFSET %d ",
		column,
		AchievementDatabaseTableName,
		column,
		column,
		limit,
		offset)
	rows, err := dataCenter.Db().Query(queryString, gameCode, currencyType, startDate, endDate)
	if err != nil {
		log.LogSerious("error fetch leaderboard player %s", err.Error())
		return nil, 0, err
	}
	playerIds := make([]int64, 0)
	values := make([]int64, 0)
	for rows.Next() {
		var playerId int64
		var value int64
		err = rows.Scan(&playerId, &value)
		if err != nil {
			rows.Close()
			log.LogSerious("error fetch leaderboard player %s", err.Error())
			return nil, 0, err
		}
		playerIds = append(playerIds, playerId)
		values = append(values, value)
	}
	rows.Close()

	playersData = make([]map[string]interface{}, 0)
	var lastValue int64
	rank := 0
	for index, playerId := range playerIds {
		player, err := GetPlayer(playerId)
		if err != nil {
			return nil, 0, err
		}
		if player != nil {
			data := player.SerializedDataMinimal()
			value := values[index]
			if lastValue == value {
				data["rank"] = rank
			} else {
				rank++
				data["rank"] = rank
			}
			data["value"] = value
			lastValue = value
			playersData = append(playersData, data)
		}
	}
	return playersData, total, nil
}

func getAchievementOfPlayer(playerId int64, gameCode string, currencyType string) (data map[string]interface{}, err error) {
	player, err := GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	achievement := player.achievementManager.getAchievement(gameCode, currencyType)
	var column string
	data = make(map[string]interface{})
	var rank int64
	// total gain
	column = "total_gain_this_week"
	totalGain := achievement.totalGainThisWeek
	rank, err = getRankFromValue(gameCode, column, totalGain)
	if err != nil {
		return nil, err
	}
	totalGainData := make(map[string]interface{})
	totalGainData["rank"] = rank
	totalGainData["value"] = achievement.totalGainThisWeek

	// daily total gain
	column = "total_gain_this_day"
	dailyTotalGain := achievement.totalGainThisDay
	rank, err = getRankFromValue(gameCode, column, dailyTotalGain)
	if err != nil {
		return nil, err
	}
	dailyTotalGainData := make(map[string]interface{})
	dailyTotalGainData["rank"] = rank
	dailyTotalGainData["value"] = achievement.totalGainThisDay

	// biggest win
	column = "biggest_win_this_week"
	biggestWin := achievement.biggestWinThisWeek
	rank, err = getRankFromValue(gameCode, column, biggestWin)
	if err != nil {
		return nil, err
	}
	biggestWinData := make(map[string]interface{})
	biggestWinData["rank"] = rank
	biggestWinData["value"] = achievement.biggestWinThisWeek

	// daily biggest win
	column = "biggest_win_this_day"
	dailyBiggestWin := achievement.biggestWinThisDay
	rank, err = getRankFromValue(gameCode, column, dailyBiggestWin)
	if err != nil {
		return nil, err
	}
	dailyBiggestWinData := make(map[string]interface{})
	dailyBiggestWinData["rank"] = rank
	dailyBiggestWinData["value"] = achievement.biggestWinThisDay

	data["weekly_biggest_win"] = biggestWinData
	data["daily_biggest_win"] = dailyBiggestWinData
	data["weekly_total_gain"] = totalGainData
	data["daily_total_gain"] = dailyTotalGainData
	data["game_code"] = gameCode
	year, week := utils.CurrentTimeInVN().ISOWeek()
	data["week"] = week
	data["year"] = year

	return data, nil
}

func getRankFromValue(gameCode string, column string, value int64) (rank int64, err error) {
	queryString := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT DISTINCT %s FROM %s WHERE game_code = $1 AND %s > $2) AS temp;",
		column,
		AchievementDatabaseTableName,
		column)
	err = dataCenter.Db().QueryRow(queryString, gameCode, value).Scan(&rank)
	if err != nil {
		return 0, err
	}
	rank = rank + 1
	return rank, nil
}
