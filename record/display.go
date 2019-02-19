package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"math"
	"time"
)

func GetPaymentData(startDate time.Time, endDate time.Time, page int64) map[string]interface{} {
	results := make(map[string]interface{})

	totalList := make([]map[string]interface{}, 0)

	limit := int64(100)
	offset := (page - 1) * limit

	var row *sql.Row
	var rows *sql.Rows
	var totalPayment int64
	var totalClaimedPayment int64
	var totalClaimedPaymentRealMoney int64
	var tax int64
	var err error

	if startDate.IsZero() && endDate.IsZero() {
		queryString := "SELECT COUNT(*) " +
			" FROM payment_record as payment"
		row = dataCenter.Db().QueryRow(queryString)

		queryString = "SELECT payment.id, payment.player_id, player.username,player.player_type, payment.payment,payment.value_before, payment.value_after,payment.status, payment.created_at" +
			" FROM payment_record as payment, player as player" +
			" WHERE payment.player_id = player.id ORDER BY -payment.id LIMIT $1 OFFSET $2"
		rows, err = dataCenter.Db().Query(queryString, limit, offset)
		if err != nil {
			log.LogSerious("Error fetch payment record %v", err)
			return nil
		}
		defer rows.Close()
		queryString = "SELECT SUM(payment) FROM payment_record"
		totalPayment = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(tax) FROM payment_record"
		tax = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(payment) FROM payment_record  WHERE status = 'claimed'"
		totalClaimedPayment = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(payment-tax) FROM payment_record  WHERE status = 'claimed'"
		totalClaimedPaymentRealMoney = getInt64FromQuery(queryString)
	} else {

		queryString := "SELECT COUNT(*) " +
			" FROM payment_record as payment" +
			" WHERE payment.created_at >= $1 AND payment.created_at <= $2"
		row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT payment.id, payment.player_id, player.username,player.player_type, payment.payment,payment.value_before, payment.value_after,payment.status, payment.created_at" +
			" FROM payment_record as payment, player as player" +
			" WHERE payment.created_at >= $1 AND payment.created_at <= $2 AND payment.player_id = player.id ORDER BY -payment.id LIMIT $3 OFFSET $4"
		rows, err = dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC(), limit, offset)
		if err != nil {
			log.LogSerious("Error fetch payment record %v", err)
			return nil
		}
		defer rows.Close()

		queryString = "SELECT SUM(payment) FROM payment_record WHERE created_at >= $1 AND created_at <= $2 "
		totalPayment = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(tax) FROM payment_record  WHERE created_at >= $1 AND created_at <= $2 "
		tax = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(payment) FROM payment_record  WHERE status = 'claimed' AND created_at >= $1 AND created_at <= $2 "
		totalClaimedPayment = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(payment-tax) FROM payment_record  WHERE status = 'claimed' AND created_at >= $1 AND created_at <= $2 "
		totalClaimedPaymentRealMoney = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())
	}
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error count payment record %v", err)
		return nil
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))
	for rows.Next() {
		var id int64
		var playerId int64
		var username string
		var playerType string
		var status string
		var payment, moneyBefore, moneyAfter int64
		var createdAt time.Time
		err = rows.Scan(&id, &playerId, &username, &playerType, &payment, &moneyBefore, &moneyAfter, &status, &createdAt)
		if err != nil {
			log.LogSerious("Error fetch payment record %v", err)
		}
		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerId
		data["username"] = username
		data["payment"] = utils.FormatWithComma(payment)
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		data["status"] = status
		dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
		data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)

		totalList = append(totalList, data)
	}

	results["page"] = page
	results["num_pages"] = numPages
	results["total_list"] = totalList

	results["total_payment"] = utils.FormatWithComma(totalPayment)
	results["total_claimed_payment"] = utils.FormatWithComma(totalClaimedPayment)
	results["total_claimed_payment_real"] = fmt.Sprintf("%s VND", utils.FormatWithComma(totalClaimedPaymentRealMoney))
	results["tax"] = utils.FormatWithComma(tax)
	results["real_payment"] = utils.FormatWithComma(totalPayment - tax)

	return results
}

func GetPaymentDetailData(paymentId int64) map[string]interface{} {
	queryString := "SELECT payment.id, payment.player_id, player.username,player.player_type," +
		" payment.payment, payment.card_code, payment.card_id, payment.status, payment.value_before, payment.value_after, payment.created_at, " +
		" payment.payment_type, payment.data" +
		" FROM payment_record as payment, player as player" +
		" WHERE payment.id = $1 AND payment.player_id = player.id"
	row := dataCenter.Db().QueryRow(queryString, paymentId)
	var id int64
	var playerId int64
	var username string
	var playerType, paymentType, status string
	var cardCode, dataByte []byte
	var cardId sql.NullInt64
	var payment, moneyBefore, moneyAfter int64
	var createdAt time.Time
	err := row.Scan(&id, &playerId, &username, &playerType, &payment, &cardCode, &cardId, &status, &moneyBefore, &moneyAfter, &createdAt,
		&paymentType, &dataByte)
	if err != nil {
		log.LogSerious("Error fetch payment record %v", err)
		return nil
	}

	var paymentData map[string]interface{}
	json.Unmarshal(dataByte, &paymentData)

	data := make(map[string]interface{})
	data["id"] = id
	data["player_id"] = playerId
	data["username"] = username
	data["payment"] = utils.FormatWithComma(payment)
	data["value_before"] = utils.FormatWithComma(moneyBefore)
	data["value_after"] = utils.FormatWithComma(moneyAfter)
	data["card_code"] = string(cardCode)
	data["payment_type"] = paymentType
	data["status"] = status
	data["data"] = paymentData

	if status == "claimed" && paymentType == "card" {

		queryString = "SELECT card_number, serial_code FROM card where id = $1"
		row = dataCenter.Db().QueryRow(queryString, cardId.Int64)
		var serialCode, cardNumber sql.NullString
		err = row.Scan(&cardNumber, &serialCode)
		if err != nil {
			log.LogSerious("Error fetch payment record %v", err)
			return nil
		}
		data["serial_code"] = serialCode.String
		data["card_number"] = cardNumber.String
	}

	dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
	data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)
	return data

}

func GetPurchaseDetailData(purchaseId int64) map[string]interface{} {
	queryString := "SELECT purchase.id, purchase.transaction_id, purchase.card_code, purchase.player_id," +
		" player.username,player.player_type, purchase.purchase, purchase.value_before, purchase.value_after, purchase.created_at" +
		" FROM purchase_record as purchase, player as player" +
		" WHERE purchase.id = $1 AND purchase.player_id = player.id"
	row := dataCenter.Db().QueryRow(queryString, purchaseId)

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
	err := row.Scan(&id, &transactionId, &cardCode, &playerId, &username, &playerType, &purchase, &moneyBefore, &moneyAfter, &createdAt)
	if err != nil {
		log.LogSerious("Error fetch purchase detail record %v", err)
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
	dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
	data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)

	return data
}

func GetMoneyFlowInGameData(gameCode string, currencyType string, startDate time.Time, endDate time.Time) map[string]interface{} {
	fmt.Println("get", gameCode, currencyType)
	queryString := "SELECT COUNT(id), SUM(win), SUM(lose),SUM(bot_win), SUM(bot_lose), SUM(tax), SUM(bet)" +
		" FROM match_record where created_at >= $1 AND created_at <= $2 AND game_code = $3 AND currency_type = $4"
	row := dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), gameCode, currencyType)
	var total, win, lose, botWin, botLose, tax, bet sql.NullInt64
	err := row.Scan(&total, &win, &lose, &botWin, &botLose, &tax, &bet)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	data := make(map[string]interface{})
	data["total_match"] = utils.FormatWithComma(total.Int64)
	data["win"] = utils.FormatWithComma(win.Int64)
	data["lose"] = utils.FormatWithComma(lose.Int64)
	data["bot_win"] = utils.FormatWithComma(botWin.Int64)
	data["bot_lose"] = utils.FormatWithComma(botLose.Int64)
	data["tax"] = utils.FormatWithComma(tax.Int64)
	data["bet"] = utils.FormatWithComma(bet.Int64)

	data["win_bet"] = fmt.Sprintf("%.2f%%", float64(win.Int64+botWin.Int64)/float64(bet.Int64)*100)
	data["lose_bet"] = fmt.Sprintf("%.2f%%", float64(lose.Int64+botLose.Int64)/float64(bet.Int64)*100)
	data["win_lose"] = fmt.Sprintf("%.2f%%", float64(win.Int64+botWin.Int64)/float64(lose.Int64+botLose.Int64)*100)

	data["system_gain_by_game"] = utils.FormatWithComma(lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64)

	return data
}

func GetBotData(currencyType string, startDate time.Time, endDate time.Time) map[string]interface{} {
	queryString := "SELECT SUM(bot_win), SUM(bot_lose) FROM match_record " +
		" where created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	row := dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var botWin, botLose sql.NullInt64
	err := row.Scan(&botWin, &botLose)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	queryString = "SELECT SUM(payment.payment)" +
		" FROM payment_record as payment, player as player" +
		" WHERE payment.created_at >= $1 AND payment.created_at <= $2" +
		" AND player.id = payment.player_id AND player.player_type = 'bot' AND payment.currency_type = $3"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var payment sql.NullInt64
	err = row.Scan(&payment)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	queryString = "SELECT SUM(purchase.purchase)" +
		" FROM purchase_record as purchase, player as player" +
		" WHERE purchase.created_at >= $1 AND purchase.created_at <= $2" +
		" AND player.id = purchase.player_id AND player.player_type = 'bot' AND purchase.currency_type = $3"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var purchase sql.NullInt64
	err = row.Scan(&purchase)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	data := make(map[string]interface{})
	data["bot_win"] = utils.FormatWithComma(botWin.Int64)
	data["bot_lose"] = utils.FormatWithComma(botLose.Int64)
	data["payment"] = utils.FormatWithComma(payment.Int64)
	data["purchase"] = utils.FormatWithComma(purchase.Int64)
	return data
}

func GetBotDataInGame(gameCode string, currencyType string, startDate time.Time, endDate time.Time) map[string]interface{} {
	queryString := "SELECT SUM(bot_win), SUM(bot_lose) FROM match_record where created_at >= $1 AND created_at <= $2 AND game_code = $3 AND currency_type = $4"
	row := dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), gameCode, currencyType)
	var botWin, botLose sql.NullInt64
	err := row.Scan(&botWin, &botLose)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	data := make(map[string]interface{})
	data["bot_win"] = utils.FormatWithComma(botWin.Int64)
	data["bot_lose"] = utils.FormatWithComma(botLose.Int64)
	return data
}

func GetUserData(startDate time.Time, endDate time.Time) map[string]interface{} {
	queryString := "SELECT COUNT(id) FROM player WHERE player_type != 'bot'"
	row := dataCenter.Db().QueryRow(queryString)
	var totalUsers sql.NullInt64
	err := row.Scan(&totalUsers)
	if err != nil {
		log.LogSerious("Error fetch user data %v", err)
	}

	queryString = "SELECT COUNT(id) FROM player WHERE player_type != 'bot' AND created_at >= $1 AND created_at <= $2"
	totalUsersInRange := utils.GetInt64FromQuery(dataCenter.Db(), queryString, startDate.UTC(), endDate.UTC())

	queryString = "SELECT COUNT(id) FROM player WHERE player_type = 'bot'"
	row = dataCenter.Db().QueryRow(queryString)
	var totalBots sql.NullInt64
	err = row.Scan(&totalBots)
	if err != nil {
		log.LogSerious("Error fetch user data %v", err)
	}

	queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT record.player_id " +
		"FROM active_record as record " +
		"WHERE record.start_date >= $1 AND record.start_date <= $2 " +
		"AND record.player_id IN (SELECT player.id FROM player as player WHERE player.player_type != 'bot')) as temp"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC())
	var activeCountSql sql.NullInt64
	err = row.Scan(&activeCountSql)
	if err != nil {
		log.LogSerious("Error get cohort data %v", err)
		return nil
	}

	queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT record.player_id " +
		"FROM active_record as record " +
		"WHERE record.start_date >= $1 AND record.start_date <= $2 AND device_type = 'ios'" +
		"AND record.player_id IN (SELECT player.id FROM player as player WHERE player.player_type != 'bot')) as temp"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC())
	var activeIOSCountSql sql.NullInt64
	err = row.Scan(&activeIOSCountSql)
	if err != nil {
		log.LogSerious("Error get cohort data %v", err)
		return nil
	}

	queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT record.player_id " +
		"FROM active_record as record " +
		"WHERE record.start_date >= $1 AND record.start_date <= $2 AND device_type = 'android'" +
		"AND record.player_id IN (SELECT player.id FROM player as player WHERE player.player_type != 'bot')) as temp"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC())
	var activeAndroidCountSql sql.NullInt64
	err = row.Scan(&activeAndroidCountSql)
	if err != nil {
		log.LogSerious("Error get cohort data %v", err)
		return nil
	}

	queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT purchase.player_id FROM purchase_record as purchase " +
		" WHERE purchase.purchase_type IN ('paybnb','appvn','iap') " +
		"AND purchase.player_id IN (SELECT player.id FROM player as player where player.player_type != 'bot' AND " +
		" player.created_at >= $1 AND player.created_at <= $2)) AS temp;"
	purchasePlayerCounInRange := utils.GetInt64FromQuery(dataCenter.Db(), queryString, startDate.UTC(), endDate.UTC())

	queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT purchase.player_id FROM purchase_record as purchase " +
		" WHERE purchase.purchase_type IN ('paybnb','appvn','iap') " +
		"AND purchase.player_id IN (SELECT player.id FROM player as player where player.player_type != 'bot')) AS temp;"
	row = dataCenter.Db().QueryRow(queryString)
	var purchasePlayerCountSql sql.NullInt64
	err = row.Scan(&purchasePlayerCountSql)
	if err != nil {
		log.LogSerious("Error get purchase count data %v", err)
		return nil
	}

	numberOfOtpUser := utils.GetInt64FromQuery(dataCenter.Db(), "select count(*) from player"+
		" where player.already_receive_otp_reward = true")

	numberOfOtpUserInRange := utils.GetInt64FromQuery(dataCenter.Db(), "select count(*) from player"+
		" where player.already_receive_otp_reward = true AND player.created_at >= $1 AND player.created_at <= $2", startDate.UTC(), endDate.UTC())

	data := make(map[string]interface{})
	data["total_users_in_range"] = utils.FormatWithComma(totalUsersInRange)
	data["total_users"] = utils.FormatWithComma(totalUsers.Int64)
	data["total_bots"] = utils.FormatWithComma(totalBots.Int64)
	data["total_all_users"] = utils.FormatWithComma(totalBots.Int64 + totalUsers.Int64)
	data["purchase_users"] = utils.FormatWithComma(purchasePlayerCountSql.Int64)
	data["purchase_users_percent"] = fmt.Sprintf("%.5f%%", float64(purchasePlayerCountSql.Int64)/float64(totalUsers.Int64)*100)
	data["purchase_users_in_range"] = utils.FormatWithComma(purchasePlayerCounInRange)
	data["purchase_users_in_range_percent"] = fmt.Sprintf("%.5f%%", float64(purchasePlayerCounInRange)/float64(totalUsersInRange)*100)
	data["otp_users"] = utils.FormatWithComma(numberOfOtpUser)
	data["otp_users_in_range"] = utils.FormatWithComma(numberOfOtpUserInRange)
	data["otp_users_percent"] = fmt.Sprintf("%.5f%%", float64(numberOfOtpUser)/float64(totalUsers.Int64)*100)
	data["otp_users_in_range_percent"] = fmt.Sprintf("%.5f%%", float64(numberOfOtpUserInRange)/float64(totalUsersInRange)*100)
	data["bot_user"] = fmt.Sprintf("%.2f%%", float64(totalBots.Int64)/float64(totalUsers.Int64)*100)
	data["active_users"] = utils.FormatWithComma(activeCountSql.Int64)
	data["active_ios_users"] = utils.FormatWithComma(activeIOSCountSql.Int64)
	data["active_android_users"] = utils.FormatWithComma(activeAndroidCountSql.Int64)
	data["active_ios_users_active_users"] = fmt.Sprintf("%.2f%%", float64(activeIOSCountSql.Int64)/float64(activeCountSql.Int64)*100)
	data["active_android_users_active_users"] = fmt.Sprintf("%.2f%%", float64(activeAndroidCountSql.Int64)/float64(activeCountSql.Int64)*100)
	return data
}

func GetDailyReportPage(currencyType string, startDate time.Time, endDate time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	queryString := "SELECT game_code, COUNT(id), SUM(win), SUM(lose),SUM(bot_win), SUM(bot_lose), SUM(tax), SUM(bet)" +
		" FROM match_record where created_at >= $1 AND created_at <= $2 AND currency_type = $3 GROUP BY game_code ORDER BY game_code"
	rows, err := dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	if err != nil {
		return nil, err
	}
	moneyGameData := make([]map[string]interface{}, 0)
	for rows.Next() {
		var total, win, lose, botWin, botLose, tax, bet sql.NullInt64
		var gameCode string

		err = rows.Scan(&gameCode, &total, &win, &lose, &botWin, &botLose, &tax, &bet)
		if err != nil {
			rows.Close()
			return nil, err
		}
		gameData := make(map[string]interface{})
		gameData["game_code"] = gameCode
		gameData["total_match"] = utils.FormatWithComma(total.Int64)
		gameData["win"] = utils.FormatWithComma(win.Int64)
		gameData["lose"] = utils.FormatWithComma(lose.Int64)
		userNetGain := win.Int64 - lose.Int64
		gameData["user_net_gain"] = utils.FormatWithComma(userNetGain)
		if userNetGain > 0 {
			gameData["user_net_gain_color"] = "danger"
		} else {
			gameData["user_net_gain_color"] = "active"
		}
		gameData["bot_win"] = utils.FormatWithComma(botWin.Int64)
		gameData["bot_lose"] = utils.FormatWithComma(botLose.Int64)
		gameData["tax"] = utils.FormatWithComma(tax.Int64)
		gameData["bet"] = utils.FormatWithComma(bet.Int64)

		moneyGameData = append(moneyGameData, gameData)
	}
	rows.Close()

	// cohort
	data["cohort"] = GetCohortData(startDate, endDate)

	// nru
	nruData, err := GetNRU(startDate, endDate)
	if err != nil {
		return nil, err
	}
	data["nru"] = nruData

	// payment
	paymentData, err := GetPaymentGraphData(startDate, endDate)
	if err != nil {
		return nil, err
	}
	data["payment_graph"] = paymentData

	// purchase
	purchaseData, err := GetPurchaseGraphData(startDate, endDate)
	if err != nil {
		return nil, err
	}
	data["purchase_graph"] = purchaseData

	// payment request count
	requestCount := utils.GetInt64FromQuery(dataCenter.Db(), "SELECT COUNT(*) FROM payment_record as payment"+
		" WHERE payment.status = 'requested' AND payment.payment_type IN ('card','gift')")
	data["request_count"] = requestCount

	data["money_game"] = moneyGameData
	return data, nil
}
