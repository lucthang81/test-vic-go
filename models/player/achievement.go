package player

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"time"
)

type AchievementManager struct {
	alreadyFetched bool
	achievements   []*Achievement
	playerId       int64
}

func NewAchievementManager() (manager *AchievementManager) {
	return &AchievementManager{
		alreadyFetched: false,
		achievements:   make([]*Achievement, 0),
	}
}

func (manager *AchievementManager) fetchData() (err error) {
	// get achievement
	if manager.alreadyFetched || manager.playerId == 0 {
		return nil
	}
	queryString := fmt.Sprintf("SELECT id, win_count, lose_count, draw_count, quit_count,"+
		" biggest_win, biggest_win_this_week,total_gain_this_week,"+
		" biggest_win_this_day, total_gain_this_day,"+
		" game_code, currency_type FROM %s WHERE player_id = $1", AchievementDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString, manager.playerId)
	if err != nil {
		return err
	}
	manager.achievements = make([]*Achievement, 0)
	for rows.Next() {
		var id int64
		var winCount int
		var loseCount int
		var drawCount int
		var quitCount int
		var biggestWin int64
		var biggestWinThisWeek int64
		var totalGainThisWeek int64
		var biggestWinThisDay int64
		var totalGainThisDay int64
		var gameCode string
		var currencyType string
		err = rows.Scan(&id, &winCount, &loseCount, &drawCount, &quitCount, &biggestWin,
			&biggestWinThisWeek, &totalGainThisWeek, &biggestWinThisDay, &totalGainThisDay, &gameCode, &currencyType)
		if err != nil {
			rows.Close()
			return err
		}
		achievement := &Achievement{
			playerId: manager.playerId,
		}
		achievement.id = id
		achievement.gameCode = gameCode
		achievement.currencyType = currencyType
		achievement.winCount = winCount
		achievement.loseCount = loseCount
		achievement.drawCount = drawCount
		achievement.quitCount = quitCount
		achievement.biggestWin = biggestWin
		achievement.biggestWinThisWeek = biggestWinThisWeek
		achievement.totalGainThisWeek = totalGainThisWeek

		manager.achievements = append(manager.achievements, achievement)
	}
	rows.Close()
	manager.alreadyFetched = true
	return nil
}

func (manager *AchievementManager) createAchievement(gameCode string, currencyType string) (achievement *Achievement, err error) {
	manager.fetchData()
	achievement = &Achievement{
		playerId:     manager.playerId,
		gameCode:     gameCode,
		currencyType: currencyType,
	}

	_, err = dataCenter.InsertObject(achievement,
		[]string{"player_id", "game_code", "currency_type"},
		[]interface{}{achievement.playerId, achievement.gameCode, achievement.currencyType}, true)
	if err != nil {
		return nil, err
	}
	manager.addAchievement(achievement)
	return achievement, nil
}

func (manager *AchievementManager) recordGameResult(gameCode string, result string, change int64, currencyType string) (err error) {
	manager.fetchData()
	achievement := manager.getAchievement(gameCode, currencyType)
	return achievement.recordGameResult(result, change, currencyType)

}

func (manager *AchievementManager) addAchievement(achievement *Achievement) {
	manager.fetchData()
	manager.achievements = append(manager.achievements, achievement)
}

func (manager *AchievementManager) getAchievement(gameCode string, currencyType string) (achievement *Achievement) {
	manager.fetchData()

	for _, achievementInList := range manager.achievements {
		if achievementInList.gameCode == gameCode && achievementInList.currencyType == currencyType {
			achievement = achievementInList
		}
	}

	if achievement == nil {
		achievement, err := manager.createAchievement(gameCode, currencyType)
		if err != nil {
			log.LogSerious("err create achievement %v", err)
		}
		return achievement
	}
	return achievement
}

func (manager *AchievementManager) SerializedData() (data []map[string]interface{}) {
	manager.fetchData()
	data = make([]map[string]interface{}, 0)
	for _, achievement := range manager.achievements {
		data = append(data, achievement.SerializedData())
	}
	return data
}

type Achievement struct {
	id           int64
	playerId     int64
	gameCode     string
	currencyType string

	winCount   int
	loseCount  int
	drawCount  int
	quitCount  int
	biggestWin int64

	biggestWinThisWeek int64
	totalGainThisWeek  int64

	biggestWinThisDay int64
	totalGainThisDay  int64
}

const AchievementCacheKey string = "achievement"
const AchievementDatabaseTableName string = "achievement"
const AchievementClassName string = "Achievement"

func (achievement *Achievement) CacheKey() string {
	return AchievementCacheKey
}

func (achievement *Achievement) DatabaseTableName() string {
	return AchievementDatabaseTableName
}

func (achievement *Achievement) ClassName() string {
	return AchievementClassName
}

func (achievement *Achievement) Id() int64 {
	return achievement.id
}

func (achievement *Achievement) SetId(id int64) {
	achievement.id = id
}

func (achievement *Achievement) recordGameResult(result string, change int64, currencyType string) (err error) {
	fieldsToSave := make([]string, 0)
	dataToSave := make([]interface{}, 0)

	if achievement.biggestWin < change {
		achievement.biggestWin = change
		fieldsToSave = append(fieldsToSave, "biggest_win")
		dataToSave = append(dataToSave, change)
	}

	if achievement.biggestWinThisWeek < change {
		achievement.biggestWinThisWeek = change
		fieldsToSave = append(fieldsToSave, "biggest_win_this_week")
		dataToSave = append(dataToSave, change)
	}

	if achievement.biggestWinThisDay < change {
		achievement.biggestWinThisDay = change
		fieldsToSave = append(fieldsToSave, "biggest_win_this_day")
		dataToSave = append(dataToSave, change)
	}

	if change != 0 {
		achievement.totalGainThisWeek = utils.MaxInt64(0, achievement.totalGainThisWeek+change)
		fieldsToSave = append(fieldsToSave, "total_gain_this_week")
		dataToSave = append(dataToSave, achievement.totalGainThisWeek)

		achievement.totalGainThisDay = utils.MaxInt64(0, achievement.totalGainThisDay+change)
		fieldsToSave = append(fieldsToSave, "total_gain_this_day")
		dataToSave = append(dataToSave, achievement.totalGainThisDay)
	}

	if result == "win" {
		achievement.winCount++
		fieldsToSave = append(fieldsToSave, "win_count")
		dataToSave = append(dataToSave, achievement.winCount)
	} else if result == "lose" {
		achievement.loseCount++
		fieldsToSave = append(fieldsToSave, "lose_count")
		dataToSave = append(dataToSave, achievement.loseCount)
	} else if result == "draw" {
		achievement.drawCount++
		fieldsToSave = append(fieldsToSave, "draw_count")
		dataToSave = append(dataToSave, achievement.drawCount)
	} else if result == "quit" {
		achievement.quitCount++
		fieldsToSave = append(fieldsToSave, "quit_count")
		dataToSave = append(dataToSave, achievement.quitCount)
	}

	fieldsToSave = append(fieldsToSave, "updated_at")
	dataToSave = append(dataToSave, time.Now().UTC())

	err = dataCenter.SaveObject(achievement, fieldsToSave, dataToSave, false)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (achievement *Achievement) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = achievement.Id()
	data["game_code"] = achievement.gameCode
	data["currency_type"] = achievement.currencyType
	data["achievement"] = achievement.currencyType
	data["player_id"] = achievement.playerId
	data["win_count"] = achievement.winCount
	data["lose_count"] = achievement.loseCount
	data["draw_count"] = achievement.drawCount
	data["quit_count"] = achievement.quitCount
	data["biggest_win"] = achievement.biggestWin
	data["biggest_win_this_week"] = achievement.biggestWinThisWeek
	data["total_gain_this_week"] = achievement.totalGainThisWeek
	data["biggest_win_this_day"] = achievement.biggestWinThisDay
	data["total_gain_this_day"] = achievement.totalGainThisDay
	return data
}
