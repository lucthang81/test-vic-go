package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/bot"
	"github.com/vic/vic_go/utils"
	"math"
	"strconv"
	"time"
)

func GetMatchData(startDate time.Time, endDate time.Time, gameCode string, currencyType string, playersNum int64, page int64) map[string]interface{} {
	results := make(map[string]interface{})

	limit := int64(100)
	offset := (page - 1) * limit

	var queryForCount, queryForRows string
	var row *sql.Row
	var rows *sql.Rows
	var err error

	if startDate.IsZero() && endDate.IsZero() {
		if gameCode == "" {
			queryForCount = "SELECT COUNT(*) " +
				" FROM match_record as match WHERE currency_type = $1"

			queryForRows = "SELECT match.id, match.win, match.lose, match.bot_win, match.bot_lose, match.tax, match.bet, match.match_data, match.game_code, match.requirement," +
				" match.created_at" +
				" FROM match_record as match WHERE currency_type = $1" +
				" ORDER BY -match.id LIMIT $2 OFFSET $3"
			row = dataCenter.Db().QueryRow(queryForCount, currencyType)
			rows, err = dataCenter.Db().Query(queryForRows, currencyType, limit, offset)
			if err != nil {
				log.LogSerious("Error fetch match record %v", err)
				return nil
			}
			defer rows.Close()
		} else {
			queryForCount = "SELECT COUNT(*) " +
				" FROM match_record as match" +
				" WHERE game_code = $1 AND currency_type = $2"
			row = dataCenter.Db().QueryRow(queryForCount, gameCode, currencyType)
			queryForRows = "SELECT match.id, match.win, match.lose, match.bot_win, match.bot_lose, match.tax, match.bet, match.match_data, match.game_code, match.requirement," +
				" match.created_at" +
				" FROM match_record as match" +
				" WHERE game_code = $1 AND currency_type = $2 ORDER BY -match.id LIMIT $3 OFFSET $4"
			rows, err = dataCenter.Db().Query(queryForRows, gameCode, currencyType, limit, offset)
			if err != nil {
				log.LogSerious("Error fetch match record %v", err)
				return nil
			}
			defer rows.Close()
		}
	} else {
		if gameCode == "" {
			queryForCount = "SELECT COUNT(*) " +
				" FROM match_record as match" +
				" WHERE match.created_at >= $1 AND match.created_at <= $2 AND currency_type = $3"

			queryForRows = "SELECT match.id, match.win, match.lose, match.bot_win, match.bot_lose, match.tax, match.bet, match.match_data, match.game_code, match.requirement," +
				" match.created_at" +
				" FROM match_record as match" +
				" WHERE match.created_at >= $1 AND match.created_at <= $2 AND currency_type = $3 ORDER BY -match.id LIMIT $4 OFFSET $5"
			row = dataCenter.Db().QueryRow(queryForCount, startDate.UTC(), endDate.UTC(), currencyType)
			rows, err = dataCenter.Db().Query(queryForRows, startDate.UTC(), endDate.UTC(), currencyType, limit, offset)
			if err != nil {
				log.LogSerious("Error fetch match record %v", err)
				return nil
			}
			defer rows.Close()
		} else {
			queryForCount = "SELECT COUNT(*) " +
				" FROM match_record as match" +
				" WHERE match.created_at >= $1 AND match.created_at <= $2 AND game_code = $3 AND currency_type = $4"
			row = dataCenter.Db().QueryRow(queryForCount, startDate.UTC(), endDate.UTC(), gameCode, currencyType)
			queryForRows = "SELECT match.id, match.win, match.lose, match.bot_win, match.bot_lose, match.tax, match.bet, match.match_data, match.game_code, match.requirement," +
				" match.created_at" +
				" FROM match_record as match" +
				" WHERE match.created_at >= $1 AND match.created_at <= $2 AND game_code = $3 AND currency_type = $4 ORDER BY -match.id LIMIT $5 OFFSET $6"
			rows, err = dataCenter.Db().Query(queryForRows, startDate.UTC(), endDate.UTC(), gameCode, currencyType, limit, offset)
			if err != nil {
				log.LogSerious("Error fetch match record %v", err)
				return nil
			}
			defer rows.Close()
		}
	}

	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error count match record %v", err)
		return nil
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))

	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var win, lose, botWin, botLose, tax, bet, requirement int64
		var matchData, gameCode string
		var createdAt time.Time

		err = rows.Scan(&id, &win, &lose, &botWin, &botLose, &tax, &bet, &matchData, &gameCode, &requirement, &createdAt)
		if err != nil {
			log.LogSerious("Error fetch purchase record %v", err)
		}
		data := make(map[string]interface{})
		data["id"] = id
		data["win"] = utils.FormatWithComma(win)
		data["lose"] = utils.FormatWithComma(lose)
		data["bot_win"] = utils.FormatWithComma(botWin)
		data["bot_lose"] = utils.FormatWithComma(botLose)
		data["tax"] = utils.FormatWithComma(tax)
		data["bet"] = utils.FormatWithComma(bet)
		data["requirement"] = utils.FormatWithComma(requirement)
		data["game_code"] = gameCode

		var matchDataJson map[string]interface{}
		err = json.Unmarshal([]byte(matchData), &matchDataJson)
		data["match_data"] = matchDataJson

		if gameCode == "roulette" || gameCode == "xocdia" {
			if utils.GetInt64AtPath(matchDataJson, "owner_id") == 0 {
				data["more_info"] = "không có nhà cái"
			} else {
				data["more_info"] = fmt.Sprintf("Nhà cái: %d", utils.GetInt64AtPath(matchDataJson, "owner_id"))
			}
		}

		if gameCode == "sicbo" {
			betData := utils.GetMapAtPath(matchDataJson, "bet_data")
			moreInfo := ""
			for _, betCode := range []string{"tri_1", "tri_2", "tri_3", "tri_4", "tri_5", "tri_6"} {
				chipValue := utils.GetInt64AtPath(betData, betCode)
				if chipValue > 0 {
					if len(moreInfo) == 0 {
						moreInfo = fmt.Sprintf("%s:%s", betCode, utils.FormatWithComma(chipValue))
					} else {
						moreInfo = fmt.Sprintf("%s %s:%s", moreInfo, betCode, utils.FormatWithComma(chipValue))
					}
				}
			}
			data["more_info"] = moreInfo
		}

		playerIds := utils.GetInt64SliceAtPath(matchDataJson, "players_id_when_start")
		if playersNum != 0 {
			if len(playerIds) != int(playersNum) {
				continue
			}
		}
		playerIdsData := make([]map[string]interface{}, 0)
		for _, playerId := range playerIds {
			playerIdData := make(map[string]interface{})
			playerIdData["id"] = playerId
			if bot.IsBot(playerId) {
				playerIdData["player_type"] = "bot"
			} else {
				playerIdData["player_type"] = "normal"
			}
			playerIdsData = append(playerIdsData, playerIdData)
		}
		data["player_ids"] = playerIdsData

		playerIpsRaw := utils.GetMapAtPath(matchDataJson, "players_ip_when_start")
		playerIpsData := make([]map[string]interface{}, 0)
		for playerIdString, ipAddress := range playerIpsRaw {
			playerIpData := make(map[string]interface{})
			playerId, _ := strconv.ParseInt(playerIdString, 10, 64)
			playerIpData["id"] = playerId
			if bot.IsBot(playerId) {
				playerIpData["player_type"] = "bot"
			} else {
				playerIpData["player_type"] = "normal"
			}
			playerIpData["ip_address"] = ipAddress
			playerIpsData = append(playerIpsData, playerIpData)
		}
		data["player_ips"] = playerIpsData

		data["normal_count"] = utils.GetIntAtPath(matchDataJson, "normal_count")
		data["bot_count"] = utils.GetIntAtPath(matchDataJson, "bot_count")

		dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
		data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)

		list = append(list, data)
	}

	results["page"] = page
	results["num_pages"] = numPages
	results["results"] = list
	return results
}

func GetMatchDetailData(matchId int64) map[string]interface{} {
	queryString := "SELECT match.id, match.win, match.lose, match.bot_win, match.bot_lose, match.tax, match.bet, match.match_data, match.game_code, match.requirement," +
		" match.created_at" +
		" FROM match_record as match" +
		" WHERE match.id = $1"
	row := dataCenter.Db().QueryRow(queryString, matchId)
	var id int64
	var win, lose, botWin, botLose, tax, bet, requirement int64
	var matchData, gameCode string
	var createdAt time.Time

	err := row.Scan(&id, &win, &lose, &botWin, &botLose, &tax, &bet, &matchData, &gameCode, &requirement, &createdAt)
	if err != nil {
		log.LogSerious("Error fetch purchase record %v", err)
	}
	data := make(map[string]interface{})
	data["id"] = id
	data["game_code"] = gameCode
	data["win"] = utils.FormatWithComma(win)
	data["lose"] = utils.FormatWithComma(lose)
	data["bot_win"] = utils.FormatWithComma(botWin)
	data["bot_lose"] = utils.FormatWithComma(botLose)
	data["tax"] = utils.FormatWithComma(tax)
	data["bet"] = utils.FormatWithComma(bet)
	data["requirement"] = utils.FormatWithComma(requirement)

	var matchDataJson map[string]interface{}
	err = json.Unmarshal([]byte(matchData), &matchDataJson)
	data["match_data"] = matchDataJson

	playerIds := utils.GetInt64SliceAtPath(matchDataJson, "players_id_when_start")
	playerIdsData := make([]map[string]interface{}, 0)
	for _, playerId := range playerIds {
		playerIdData := make(map[string]interface{})
		playerIdData["id"] = playerId
		if bot.IsBot(playerId) {
			playerIdData["player_type"] = "bot"
		} else {
			playerIdData["player_type"] = "normal"
		}
		playerIdsData = append(playerIdsData, playerIdData)
	}
	data["player_ids"] = playerIdsData

	playerIpsRaw := utils.GetMapAtPath(matchDataJson, "players_ip_when_start")
	playerIpsData := make([]map[string]interface{}, 0)
	for playerIdString, ipAddress := range playerIpsRaw {
		playerIpData := make(map[string]interface{})
		playerId, _ := strconv.ParseInt(playerIdString, 10, 64)
		playerIpData["id"] = playerId
		if bot.IsBot(playerId) {
			playerIpData["player_type"] = "bot"
		} else {
			playerIpData["player_type"] = "normal"
		}
		playerIpData["ip_address"] = ipAddress
		playerIpsData = append(playerIpsData, playerIpData)
	}
	data["player_ips"] = playerIpsData

	data["normal_count"] = utils.GetIntAtPath(matchDataJson, "normal_count")
	data["bot_count"] = utils.GetIntAtPath(matchDataJson, "bot_count")
	dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
	data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)

	return data
}
