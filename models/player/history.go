package player

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/bot"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
	"math"
	"strconv"
	"time"
)

func (player *Player) IsInvalidForAutoPayment() bool {
	queryString := "SELECT action, additional_data" +
		" FROM currency_record" +
		" WHERE action = $1 AND player_id = $2 AND currency_type = $3 " +
		" ORDER BY created_at DESC LIMIT 30"
	rows, err := dataCenter.Db().Query(queryString, "match", player.Id(), currency.Money)
	if err != nil {
		log.LogSerious("err scan invalid auto payment %v", err)
		return false
	}

	defer rows.Close()
	var count int
	var totalCount int
	for rows.Next() {
		var addtionalDataRaw []byte
		var action string
		err = rows.Scan(&action, &addtionalDataRaw)
		if err != nil {
			log.LogSerious("err scan invalid auto payment %v", err)
			return false
		}

		var addtionalData map[string]interface{}
		json.Unmarshal(addtionalDataRaw, &addtionalData)

		if action == "match" {
			botCount := utils.GetIntAtPath(addtionalData, "bot_count")
			if botCount == 1 {
				count++
			}
			totalCount++
		}
	}
	if totalCount == 0 {
		return false
	}
	ratio := float64(count) / float64(totalCount)
	if ratio >= game_config.AutoAcceptPaymentBotFarmRatio() {
		return true
	}
	return false
}

func (player *Player) GetHistoryData(startDate time.Time, endDate time.Time, currencyType string, page int64) (data map[string]interface{}, total int64, err error) {
	var row *sql.Row
	var rows *sql.Rows
	var queryString string
	limit := int64(100)
	offset := (page - 1) * limit

	if startDate.IsZero() && endDate.IsZero() {
		queryString = "SELECT COUNT(*)" +
			" FROM currency_record" +
			" WHERE player_id = $1 AND currency_type = $2"
		row = dataCenter.Db().QueryRow(queryString, player.Id(), currencyType)

		queryString = "SELECT id, action, additional_data," +
			" value_before, value_after, change, created_at" +
			" FROM currency_record" +
			" WHERE player_id = $1 AND currency_type = $2 " +
			" ORDER BY created_at DESC LIMIT $3 OFFSET $4"
		rows, err = dataCenter.Db().Query(queryString, player.Id(), currencyType, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

	} else {
		queryString = "SELECT COUNT(*)" +
			" FROM currency_record" +
			" WHERE player_id = $1 AND created_at >= $2 AND created_at <= $3 AND currency_type = $4 "
		row = dataCenter.Db().QueryRow(queryString, player.Id(), startDate.UTC(), endDate.UTC(), currencyType)

		queryString = "SELECT id, action, additional_data," +
			" value_before, value_after, change, created_at" +
			" FROM currency_record" +
			" WHERE player_id = $1 AND created_at >= $2 AND created_at <= $3 AND currency_type = $4 " +
			" ORDER BY created_at DESC LIMIT $5 OFFSET $6"
		rows, err = dataCenter.Db().Query(queryString, player.Id(), startDate.UTC(), endDate.UTC(), currencyType, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
	}

	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, valueBefore, valueAfter, change int64
		var action string
		var addtionalDataRaw []byte
		var createdAt time.Time
		err = rows.Scan(&id, &action, &addtionalDataRaw, &valueBefore, &valueAfter, &change, &createdAt)
		if err != nil {
			return nil, 0, err
		}

		var addtionalData map[string]interface{}
		json.Unmarshal(addtionalDataRaw, &addtionalData)

		data := make(map[string]interface{})
		data["id"] = id
		data["value_after_raw"] = valueAfter
		data["value_before"] = utils.FormatWithComma(valueBefore)
		data["value_after"] = utils.FormatWithComma(valueAfter)
		data["change"] = utils.FormatWithComma(change)
		data["action"] = action
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)

		if action == "match" {
			data["match_record_id"] = utils.GetInt64AtPath(addtionalData, "match_record_id")
			data["game_code"] = utils.GetStringAtPath(addtionalData, "game_code")
			data["normal_count"] = utils.GetIntAtPath(addtionalData, "normal_count")
			data["bot_count"] = utils.GetIntAtPath(addtionalData, "bot_count")

			playerIpsRaw := utils.GetMapAtPath(addtionalData, "ip_address")
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
		} else if action == "payment" {
			data["payment_record_id"] = utils.GetInt64AtPath(addtionalData, "payment_record_id")
		} else if action == "purchase" {
			data["purchase_record_id"] = utils.GetInt64AtPath(addtionalData, "purchase_record_id")
			data["transaction_id"] = utils.GetInt64AtPath(addtionalData, "transaction_id")
			data["purchase_type"] = utils.GetInt64AtPath(addtionalData, "purchase_type")
		}

		results = append(results, data)
	}

	numPages := int64(math.Ceil(float64(total) / float64(limit)))
	data = make(map[string]interface{})

	playerData := player.SerializedData()
	playerData["money"] = utils.FormatWithComma(player.GetMoney(currency.Money))
	playerData["test_money"] = utils.FormatWithComma(player.GetMoney(currency.TestMoney))
	playerData["email"] = player.email
	playerData["phone_number"] = player.phoneNumber

	// device
	queryString = "SELECT device_code, device_type, ip_address FROM active_record WHERE player_id = $1 ORDER BY -id LIMIT 1"
	row = dataCenter.Db().QueryRow(queryString, player.Id())
	var deviceCode, deviceType, ipAddress sql.NullString
	err = row.Scan(&deviceCode, &deviceType, &ipAddress)
	if err == nil {
		playerData["device_code"] = deviceCode.String
		playerData["device_type"] = deviceType.String
		playerData["ip_address"] = ipAddress.String
	} else {
		fmt.Println("%v", err)
	}

	data["player"] = playerData
	data["results"] = results
	data["num_pages"] = numPages
	data["total"] = total

	// total purchase
	queryString = "SELECT SUM(purchase) FROM purchase_record WHERE player_id = $1 AND currency_type = $2"
	totalPurchase := dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType)
	data["total_purchase"] = utils.FormatWithComma(totalPurchase)

	queryString = "SELECT SUM(payment) FROM payment_record WHERE player_id = $1"
	payment := dataCenter.GetInt64FromQuery(queryString, player.Id())
	data["payment"] = utils.FormatWithComma(payment)

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'admin_add' AND currency_type = $2"
	data["admin_add"] = utils.FormatWithComma(dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType))

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'start_game' AND currency_type = $2"
	data["start_game"] = utils.FormatWithComma(dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType))

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'iap' AND currency_type = $2"
	data["iap_purchase"] = utils.FormatWithComma(dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType))

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'paybnb' AND currency_type = $2"
	data["paybnb_purchase"] = utils.FormatWithComma(dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType))

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'appvn' AND currency_type = $2"
	data["appvn_purchase"] = utils.FormatWithComma(dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType))

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND (purchase_type = 'paybnb' OR purchase_type = 'appvn' OR purchase_type = 'iap') AND currency_type = $2"
	purchase := dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType)
	data["purchase"] = utils.FormatWithComma(purchase)

	queryString = "SELECT SUM(purchase) FROM purchase_record where player_id = $1 AND purchase_type = 'otp_reward' AND currency_type = $2"
	otpReward := dataCenter.GetInt64FromQuery(queryString, player.Id(), currencyType)
	data["otp_reward"] = utils.FormatWithComma(otpReward)

	data["payment"] = utils.FormatWithComma(payment)
	if purchase == 0 {
		data["payment_purchase"] = "N/A"
	} else {
		data["payment_purchase"] = fmt.Sprintf("%.2f%%", float64(payment)/float64(purchase)*100)
	}

	data["money"] = utils.FormatWithComma(player.currencyGroup.GetValue(currencyType))

	return data, total, nil

}

func (player *Player) GetPaymentHistory(page int64) (fullData map[string]interface{}, err error) {
	playerId := player.Id()
	limit := int64(100)
	offset := limit * (page - 1)

	queryString := "SELECT COUNT(id) FROM payment_record WHERE player_id = $1"
	total := dataCenter.GetInt64FromQuery(queryString, playerId)
	numPages := int64(math.Ceil(float64(total) / float64(limit)))

	queryString = "SELECT payment.id, player.username, payment.card_code, payment.payment, payment.value_before," +
		" payment.value_after, payment.created_at, payment.status" +
		" FROM payment_record as payment, player as player" +
		" WHERE player.id = payment.player_id AND player.id = $1" +
		" ORDER BY -payment.id LIMIT $2 OFFSET $3"
	rows, err := dataCenter.Db().Query(queryString, playerId, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, moneyBefore, moneyAfter, payment int64
		var username, cardCode, status string
		var createdAt time.Time
		err = rows.Scan(&id, &username, &cardCode, &payment, &moneyBefore, &moneyAfter, &createdAt, &status)
		if err != nil {
			return nil, err
		}

		data := make(map[string]interface{})
		data["id"] = id
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		data["payment"] = utils.FormatWithComma(payment)
		data["username"] = username
		data["card_code"] = cardCode
		data["status"] = status
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)

		results = append(results, data)
	}

	queryString = "SELECT SUM(payment) FROM payment_record WHERE player_id = $1"
	payment := dataCenter.GetInt64FromQuery(queryString, player.Id())

	fullData = make(map[string]interface{})
	fullData["num_pages"] = numPages
	fullData["page"] = page
	fullData["results"] = results
	fullData["payment"] = utils.FormatWithComma(payment)

	return fullData, nil
}

func (player *Player) GetPurchaseHistory(page int64) (fullData map[string]interface{}, err error) {
	playerId := player.Id()
	results := make([]map[string]interface{}, 0)

	limit := int64(100)
	offset := (page - 1) * limit

	queryString := "SELECT COUNT(id) FROM purchase_record WHERE player_id = $1"
	total := dataCenter.GetInt64FromQuery(queryString, playerId)
	numPages := int64(math.Ceil(float64(total) / float64(limit)))

	queryString = "SELECT purchase.id, purchase.transaction_id, purchase.card_code, purchase.player_id, player.username, player.player_type, purchase.purchase," +
		" purchase.value_before, purchase.value_after, purchase.created_at" +
		" FROM purchase_record as purchase, player as player" +
		" WHERE purchase.player_id = player.id AND player.id = $1 ORDER BY -purchase.id LIMIT $2 OFFSET $3"
	rows, err := dataCenter.Db().Query(queryString, playerId, limit, offset)
	if err != nil {
		log.LogSerious("Error fetch purchase record %v", err)
		return nil, nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var playerId int64
		var transactionId string
		var cardCode string
		var username string
		var playerType string
		var purchase int64
		var moneyBefore int64
		var moneyAfter int64
		var createdAt time.Time
		err = rows.Scan(&id, &transactionId, &cardCode, &playerId, &username, &playerType, &purchase, &moneyBefore, &moneyAfter, &createdAt)
		if err != nil {
			log.LogSerious("Error fetch purchase record %v", err)
			return nil, err
		}
		data := make(map[string]interface{})
		data["id"] = id
		data["transaction_id"] = transactionId
		data["card_code"] = cardCode
		data["player_id"] = playerId
		data["username"] = username
		data["purchase"] = utils.FormatWithComma(purchase)
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)

		results = append(results, data)
	}
	queryString = "SELECT SUM(purchase) FROM purchase_record WHERE player_id = $1"
	purchase := dataCenter.GetInt64FromQuery(queryString, player.Id())

	fullData = make(map[string]interface{})
	fullData["num_pages"] = numPages
	fullData["page"] = page
	fullData["results"] = results
	fullData["purchase"] = utils.FormatWithComma(purchase)
	return fullData, nil
}

func getMatchHistory(playerId int64, offsetUnit int64) (results []map[string]interface{}, err error) {
	limit := int64(100)
	offset := offsetUnit * 100

	queryString := "SELECT id, game_code, created_at, match_data FROM match_record WHERE id IN" +
		" (SELECT match_record_id FROM player_match_record WHERE player_id = $1)" +
		" ORDER BY -id LIMIT $2 OFFSET $3"
	rows, err := dataCenter.Db().Query(queryString, playerId, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var gameCode string
		var matchDataRaw []byte
		var createdAt time.Time

		err = rows.Scan(&id, &gameCode, &matchDataRaw, &createdAt)
		if err != nil {
			return nil, err
		}

		data := make(map[string]interface{})
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["game_code"] = gameCode
		data["id"] = id

		var matchData map[string]interface{}
		err = json.Unmarshal(matchDataRaw, &matchData)
		if err != nil {
			return nil, err
		}

		playersMoneyWhenStart := utils.GetMapAtPath(matchData, "players_money_when_start")
		moneyBefore := utils.GetInt64AtPath(playersMoneyWhenStart, fmt.Sprintf("%d", playerId))
		var moneyAfter int64
		matchResults := utils.GetMapSliceAtPath(matchData, "results")
		for _, matchResult := range matchResults {
			playerIdInResult := utils.GetInt64AtPath(matchResult, "id")
			if playerIdInResult == playerId {
				moneyAfter = utils.GetInt64AtPath(matchResult, "money")
				break
			}
		}
		data["value_after_raw"] = moneyAfter
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		results = append(results, data)
	}
	return results, nil
}
