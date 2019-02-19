package record

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"time"
)

func GetTotalMoneyReportData(currencyType string) map[string]interface{} {
	queryString := "SELECT SUM(win), SUM(lose),SUM(bot_win), SUM(bot_lose), SUM(tax), SUM(bet) FROM match_record WHERE currency_type = $1"
	row := dataCenter.Db().QueryRow(queryString, currencyType)
	var win, lose, botWin, botLose, tax, bet sql.NullInt64
	err := row.Scan(&win, &lose, &botWin, &botLose, &tax, &bet)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	queryString = "SELECT SUM(value) FROM currency WHERE currency_type = $1"
	totalPlayerMoney := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(value) FROM currency WHERE currency_type = $1 AND player_id IN (select id from player where player_type = $2)"
	totalNormalMoney := getInt64FromQuery(queryString, currencyType, "normal")

	queryString = "SELECT SUM(value) FROM currency WHERE currency_type = $1 AND player_id IN (select id from player where player_type = $2)"
	totalBotMoney := getInt64FromQuery(queryString, currencyType, "bot")

	queryString = "SELECT SUM(payment) FROM payment_record WHERE currency_type = $1"
	row = dataCenter.Db().QueryRow(queryString, currencyType)
	var payment sql.NullInt64
	err = row.Scan(&payment)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	queryString = "SELECT SUM(purchase) FROM purchase_record WHERE currency_type = $1"
	row = dataCenter.Db().QueryRow(queryString, currencyType)
	var purchase sql.NullInt64
	err = row.Scan(&purchase)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	queryString = "SELECT SUM(tax) FROM match_record WHERE currency_type = $1"
	matchTax := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(tax) FROM payment_record WHERE currency_type = $1"
	paymentTax := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT DISTINCT game_code, SUM(tax) FROM match_record  WHERE currency_type = $1 GROUP BY game_code"
	rowsTax, err := dataCenter.Db().Query(queryString, currencyType)
	if err != nil {
		log.LogSerious("Error fetch tax each game data %v", err)
	}
	defer rowsTax.Close()
	taxData := make(map[string]string)
	for rowsTax.Next() {
		var gameCode string
		var tax sql.NullInt64
		err = rowsTax.Scan(&gameCode, &tax)
		if err != nil {
			log.LogSerious("Error fetch tax each game data %v", err)
		} else {
			taxData[gameCode] = utils.FormatWithComma(tax.Int64)
		}
	}

	queryString = "SELECT SUM(value) FROM bank WHERE currency_type = $1"
	totalBankMoney := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT id, game_code, value FROM bank  WHERE currency_type = $1"
	rowsBank, err := dataCenter.Db().Query(queryString, currencyType)
	if err != nil {
		log.LogSerious("Error fetch tax each game data %v", err)
	}
	defer rowsBank.Close()
	bankData := make(map[string]string)
	for rowsBank.Next() {
		var gameCode string
		var id, money int64
		err = rowsBank.Scan(&id, &gameCode, &money)
		if err != nil {
			log.LogSerious("Error fetch tax each game data %v", err)
		} else {
			bankData[gameCode] = utils.FormatWithComma(money)
		}
	}

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add' AND currency_type = $1 AND player_id " +
		"IN (SELECT id from player where player_type ='bot')"
	moneyAddToBot := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add' AND currency_type = $1 AND player_id " +
		"IN (SELECT id from player where player_type != 'bot')"
	moneyAddToUser := getInt64FromQuery(queryString, currencyType)
	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add'  AND currency_type = $1"
	moneyAddToAll := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game' AND currency_type = $1 AND player_id " +
		"IN (SELECT id from player where player_type != 'bot')"
	moneyAddWhenStartGameForPlayer := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game' AND currency_type = $1 AND player_id " +
		"IN (SELECT id from player where player_type = 'bot')"
	moneyAddWhenStartGameForBot := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game' AND currency_type = $1"
	moneyAddWhenStartGameForAll := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(bot_win) FROM match_record  WHERE currency_type = $1"
	totalBotWin := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(bot_lose) FROM match_record  WHERE currency_type = $1"
	totalBotLose := getInt64FromQuery(queryString, currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'appvn' OR purchase_type = 'paybnb' AND currency_type = $1"
	totalPurchaseByCard := getInt64FromQuery(queryString, currencyType)

	data := make(map[string]interface{})
	data["money_add_to_bot"] = utils.FormatWithComma(moneyAddToBot)
	data["money_add_to_user"] = utils.FormatWithComma(moneyAddToUser)
	data["money_add_to_all"] = utils.FormatWithComma(moneyAddToAll)

	data["payment_tax"] = utils.FormatWithComma(paymentTax)
	data["match_tax"] = utils.FormatWithComma(matchTax)
	data["total_tax"] = utils.FormatWithComma(matchTax + paymentTax)

	data["tax_data"] = taxData
	data["bank_data"] = bankData

	data["money_add_when_start_to_bot"] = utils.FormatWithComma(moneyAddWhenStartGameForBot)
	data["money_add_when_start_to_user"] = utils.FormatWithComma(moneyAddWhenStartGameForPlayer)
	data["money_add_when_start_to_all"] = utils.FormatWithComma(moneyAddWhenStartGameForAll)

	data["total_bot_money"] = utils.FormatWithComma(totalBotMoney)
	data["total_bot_win"] = utils.FormatWithComma(totalBotWin)
	data["total_bot_lose"] = utils.FormatWithComma(totalBotLose)

	data["total_purchase_by_card"] = utils.FormatWithComma(totalPurchaseByCard)

	totalMoney := totalBankMoney + totalPlayerMoney
	data["percent_payment_total"] = fmt.Sprintf("%.2f%%", float64(payment.Int64)/float64(totalMoney)*100)
	data["total_money"] = utils.FormatWithComma(totalMoney)
	data["total_bank_money"] = utils.FormatWithComma(totalBankMoney)
	data["total_player_money"] = utils.FormatWithComma(totalPlayerMoney)
	data["total_player_normal_money"] = utils.FormatWithComma(totalNormalMoney)
	data["total_player_bot_money"] = utils.FormatWithComma(totalBotMoney)

	data["system_gain_by_payment_purchase"] = utils.FormatWithComma(purchase.Int64 - payment.Int64)
	data["system_gain_by_game"] = utils.FormatWithComma(lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64)
	data["total_system_gain"] = utils.FormatWithComma(purchase.Int64 - payment.Int64 + lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64)
	data["total_money_in_system"] = utils.FormatWithComma(totalMoney)
	data["total_money_player_in_system"] = utils.FormatWithComma(totalPlayerMoney)
	data["ratio_player_money"] = fmt.Sprintf("%.2f%%", float64(totalNormalMoney)/float64(totalPlayerMoney)*100)
	data["total_money_bot_in_system"] = utils.FormatWithComma(totalBotMoney)
	data["ratio_bot_money"] = fmt.Sprintf("%.2f%%", float64(totalBotMoney)/float64(totalPlayerMoney)*100)
	data["payment"] = utils.FormatWithComma(payment.Int64)
	data["purchase"] = utils.FormatWithComma(purchase.Int64)

	return data
}

func GetGeneralMoneyData(currencyType string, startDate time.Time, endDate time.Time) map[string]interface{} {
	queryString := "SELECT SUM(win), SUM(lose),SUM(bot_win), SUM(bot_lose), SUM(tax), SUM(bet) FROM match_record" +
		" where created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	row := dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var win, lose, botWin, botLose, tax, bet sql.NullInt64
	err := row.Scan(&win, &lose, &botWin, &botLose, &tax, &bet)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	queryString = "SELECT SUM(payment) FROM payment_record where created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var payment sql.NullInt64
	err = row.Scan(&payment)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	queryString = "SELECT SUM(purchase) FROM purchase_record where created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	row = dataCenter.Db().QueryRow(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	var purchase sql.NullInt64
	err = row.Scan(&purchase)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}

	queryString = "SELECT SUM(tax) FROM match_record WHERE created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	matchTax := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(tax) FROM payment_record WHERE created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	paymentTax := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT DISTINCT game_code, SUM(tax) FROM match_record " +
		"WHERE created_at >= $1 AND created_at <= $2 AND currency_type = $3 GROUP BY game_code"
	taxRows, err := dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	if err != nil {
		log.LogSerious("Error fetch tax each game data %v", err)
	}
	defer taxRows.Close()
	taxData := make(map[string]string)
	for taxRows.Next() {
		var gameCode string
		var tax sql.NullInt64
		err = taxRows.Scan(&gameCode, &tax)
		if err != nil {
			log.LogSerious("Error fetch tax each game data %v", err)
		} else {
			taxData[gameCode] = utils.FormatWithComma(tax.Int64)
		}
	}

	queryString = "SELECT id, game_code, value FROM bank where currency_type = $1"
	rowsBank, err := dataCenter.Db().Query(queryString, currencyType)
	if err != nil {
		log.LogSerious("Error fetch tax each game data %v", err)
	}
	defer rowsBank.Close()
	bankData := make(map[string]string)
	for rowsBank.Next() {
		var gameCode string
		var id, money int64
		err = rowsBank.Scan(&id, &gameCode, &money)
		if err != nil {
			log.LogSerious("Error fetch tax each game data %v", err)
		} else {
			bankData[gameCode] = utils.FormatWithComma(money)
		}
	}

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add' AND created_at >= $1 AND created_at <= $2 AND currency_type = $3 AND player_id " +
		"IN (SELECT id from player where player_type ='bot')"
	moneyAddToBot := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add' AND created_at >= $1 AND created_at <= $2 AND currency_type = $3 AND player_id " +
		"IN (SELECT id from player where player_type != 'bot')"
	moneyAddToUser := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)
	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'admin_add'   AND created_at >= $1 AND created_at <= $2  AND currency_type = $3"
	moneyAddToAll := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game' AND created_at >= $1 AND created_at <= $2 AND currency_type = $3 AND player_id " +
		"IN (SELECT id from player where player_type != 'bot')"
	moneyAddWhenStartGameForPlayer := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game' AND created_at >= $1 AND created_at <= $2  AND currency_type = $3 AND player_id " +
		"IN (SELECT id from player where player_type = 'bot')"
	moneyAddWhenStartGameForBot := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record " +
		"WHERE purchase_type = 'start_game'  AND created_at >= $1 AND created_at <= $2 AND currency_type = $3"
	moneyAddWhenStartGameForAll := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(value) FROM currency WHERE player_id IN (select id from player where player_type = 'bot')"
	totalBotMoney := getInt64FromQuery(queryString)

	queryString = "SELECT SUM(bot_win) FROM match_record  WHERE created_at >= $1 AND created_at <= $2  AND currency_type = $3"
	totalBotWin := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(bot_lose) FROM match_record  WHERE created_at >= $1 AND created_at <= $2  AND currency_type = $3"
	totalBotLose := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	queryString = "SELECT SUM(purchase) FROM purchase_record WHERE (purchase_type = 'appvn' OR purchase_type = 'paybnb')  " +
		"AND created_at >= $1 AND created_at <= $2  AND currency_type = $3"
	totalPurchaseByCard := getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC(), currencyType)

	data := make(map[string]interface{})
	data["money_add_to_bot"] = utils.FormatWithComma(moneyAddToBot)
	data["money_add_to_user"] = utils.FormatWithComma(moneyAddToUser)
	data["money_add_to_all"] = utils.FormatWithComma(moneyAddToAll)

	data["total_tax"] = utils.FormatWithComma(paymentTax + matchTax)
	data["match_tax"] = utils.FormatWithComma(matchTax)
	data["payment_tax"] = utils.FormatWithComma(paymentTax)
	data["tax_data"] = taxData
	data["bank_data"] = bankData

	data["money_add_when_start_to_bot"] = utils.FormatWithComma(moneyAddWhenStartGameForBot)
	data["money_add_when_start_to_user"] = utils.FormatWithComma(moneyAddWhenStartGameForPlayer)
	data["money_add_when_start_to_all"] = utils.FormatWithComma(moneyAddWhenStartGameForAll)

	data["total_bot_money"] = utils.FormatWithComma(totalBotMoney)
	data["total_bot_win"] = utils.FormatWithComma(totalBotWin)
	data["total_bot_lose"] = utils.FormatWithComma(totalBotLose)

	data["total_purchase_by_card"] = utils.FormatWithComma(totalPurchaseByCard)

	totalMoney := purchase.Int64 + lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64
	data["percent_payment_total"] = fmt.Sprintf("%.2f%%", float64(payment.Int64)/float64(totalMoney)*100)
	data["total_money"] = utils.FormatWithComma(totalMoney)

	data["win"] = utils.FormatWithComma(win.Int64)
	data["lose"] = utils.FormatWithComma(lose.Int64)
	data["bot_win"] = utils.FormatWithComma(botWin.Int64)
	data["bot_lose"] = utils.FormatWithComma(botLose.Int64)
	data["tax"] = utils.FormatWithComma(tax.Int64)
	data["bet"] = utils.FormatWithComma(bet.Int64)

	data["win_bet"] = fmt.Sprintf("%.2f%%", float64(win.Int64+botWin.Int64)/float64(bet.Int64)*100)
	data["lose_bet"] = fmt.Sprintf("%.2f%%", float64(lose.Int64+botLose.Int64)/float64(bet.Int64)*100)
	data["win_lose"] = fmt.Sprintf("%.2f%%", float64(win.Int64+botWin.Int64)/float64(lose.Int64+botLose.Int64)*100)

	data["payment"] = utils.FormatWithComma(payment.Int64)
	data["purchase"] = utils.FormatWithComma(purchase.Int64)

	data["total_system_gain"] = utils.FormatWithComma(purchase.Int64 - payment.Int64 + lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64)
	data["system_gain_by_payment_purchase"] = utils.FormatWithComma(purchase.Int64 - payment.Int64)
	data["system_gain_by_game"] = utils.FormatWithComma(lose.Int64 + botLose.Int64 + tax.Int64 - win.Int64 - botWin.Int64)
	data["total_win"] = utils.FormatWithComma(win.Int64 + botWin.Int64)
	data["total_lose"] = utils.FormatWithComma(lose.Int64 + botLose.Int64 + tax.Int64)

	return data
}

func GetPaymentGraphData(startDateParams time.Time, endDateParams time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfDayFromTime(startDateParams)
	endDate := utils.EndOfDayFromTime(endDateParams)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := timeRangeStart.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		value := dataCenter.GetInt64FromQuery("SELECT SUM(payment) FROM payment_record WHERE created_at >= $1 AND created_at <= $2", timeRangeStart.UTC(), timeRangeEnd.UTC())

		dayData := make(map[string]interface{})
		dayData["value"] = value
		dayData["value_format"] = utils.FormatWithComma(value)
		dayData["date"], _ = utils.FormatTimeToVietnamTime(timeRangeStart)
		results = append(results, dayData)
		timeRangeStart = timeRangeStart.Add(24 * time.Hour)
	}
	data["results"] = results
	return
}

func GetPurchaseGraphData(startDateParams time.Time, endDateParams time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfDayFromTime(startDateParams)
	endDate := utils.EndOfDayFromTime(endDateParams)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := timeRangeStart.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		value := dataCenter.GetInt64FromQuery("SELECT SUM(purchase) FROM purchase_record "+
			"WHERE created_at >= $1 AND created_at <= $2 AND (purchase_type = 'paybnb' OR purchase_type = 'appvn' OR purchase_type = 'iap')", timeRangeStart.UTC(), timeRangeEnd.UTC())

		dayData := make(map[string]interface{})
		dayData["value"] = value
		dayData["value_format"] = utils.FormatWithComma(value)
		dayData["date"], _ = utils.FormatTimeToVietnamTime(timeRangeStart)
		results = append(results, dayData)
		timeRangeStart = timeRangeStart.Add(24 * time.Hour)
	}
	data["results"] = results
	return
}
