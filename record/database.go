package record

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/vic/vic_go/utils"

	"github.com/vic/vic_go/log"
)

func logStartActiveRecord(playerId int64, deviceCode string, deviceType string, ipAddress string) {
	querySelect := "SELECT id FROM active_record WHERE player_id = $1 AND start_date = end_date ORDER BY -id LIMIT 1"
	row := dataCenter.Db().QueryRow(querySelect, playerId)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		// end the last active record first
		logEndActiveRecord(playerId)
	}

	queryString := "INSERT INTO active_record (player_id, device_code, device_type, ip_address) " +
		"VALUES ($1, $2, $3, $4)"
	_, err = dataCenter.Db().Exec(queryString, playerId, deviceCode, deviceType, ipAddress)
	if err != nil {
		log.LogSerious("Error log start active record %v", err)
	}
}

func logEndActiveRecord(playerId int64) {
	queryString := "UPDATE active_record SET end_date = $1 WHERE id IN " +
		"(SELECT id FROM active_record WHERE player_id = $2 ORDER BY -id LIMIT 1)"
	_, err := dataCenter.Db().Exec(queryString, time.Now().UTC(), playerId)
	if err != nil {
		log.LogSerious("Error log end active record %v", err)
	}
}

func logMatchRecord(gameCode string, currencyType string, requirement int64, bet int64, tax int64, win int64, lose int64, botWin int64, botLose int64, matchData map[string]interface{}) int64 {
	matchData = utils.ConvertData(matchData)
	rawBytes, err := json.Marshal(matchData)
	if err != nil {
		log.LogSerious("error json match record %v %v", err, matchData)
		return 0
	}

	// var rawZip bytes.Buffer
	// zipper := gzip.NewWriter(&rawZip)
	// _, err = zipper.Write(rawBytes)
	// if err != nil {
	// 	fmt.Printf("zipper.Write ERROR: %+v", err)
	// }
	// err = zipper.Close()
	playerIds := utils.GetInt64SliceAtPath(matchData, "players_id_when_start")

	queryString := "INSERT INTO match_record (game_code, currency_type, requirement, bet, tax, win, lose, bot_win, bot_lose, match_data) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, gameCode, currencyType, requirement, bet, tax, win, lose, botWin, botLose, string(rawBytes))
	var matchRecordId int64
	err = row.Scan(&matchRecordId)
	if err != nil {
		log.LogSerious("Error log match record %v", err)
		return 0
	}

	normalCount := utils.GetIntAtPath(matchData, "normal_count")
	botCount := utils.GetIntAtPath(matchData, "bot_count")
	ipAddress := utils.GetMapAtPath(matchData, "players_ip_when_start")
	additionalData := map[string]interface{}{
		"match_record_id": matchRecordId,
		"game_code":       gameCode,
		"normal_count":    normalCount,
		"bot_count":       botCount,
		"ip_address":      ipAddress,
	}

	for _, playerId := range playerIds {
		queryString := "INSERT INTO player_match_record (player_id, match_record_id) VALUES ($1,$2)"
		_, err = dataCenter.Db().Exec(queryString, playerId, matchRecordId)
		if err != nil {
			log.LogSerious("err insert to match record %v, playerId %d, matchId %d", err, playerId, matchRecordId)
		}

		moneyBefore := utils.GetInt64AtPath(matchData, fmt.Sprintf("players_money_when_start/%d", playerId))
		var moneyAfter int64
		var change int64
		matchResults := utils.GetMapSliceAtPath(matchData, "results")
		for _, matchResult := range matchResults {
			playerIdInResult := utils.GetInt64AtPath(matchResult, "id")
			change = utils.GetInt64AtPath(matchResult, "change")
			if playerIdInResult == playerId {
				// moneyAfter = utils.GetInt64AtPath(matchResult, "money") + utils.GetInt64AtPath(matchResult, "money_on_table")
				moneyAfter = utils.GetInt64AtPath(matchResult, "money")
				break
			}
		}

		LogCurrencyRecord(playerId,
			"match",
			gameCode,
			additionalData,
			currencyType,
			moneyBefore,
			moneyAfter,
			change)

	}
	return matchRecordId
}

// log for minah new games,
// log to tables: match_record, player_match_record
// playerResult main field:
//	- id: playerId
//  - username: playerUsername
//	- change: tiền kiếm từ trận đấu, có thể âm
func LogMatchRecord2(
	gameCode string, currencyType string, requirement int64, tax int64,
	humanWon int64, humanLost int64, botWon int64, botLost int64,
	minahMatchId string, playerIpAdds map[int64]string,
	playerResults []map[string]interface{},
) int64 {
	//
	playerIds := []int64{}
	for pid, _ := range playerIpAdds {
		playerIds = append(playerIds, pid)
	}
	//
	matchData := map[string]interface{}{
		"players_id_when_start": playerIds,
		"players_ip_when_start": playerIpAdds,
		"results":               playerResults,
	}
	rawBytes, _ := json.Marshal(matchData)
	//
	queryString := "INSERT INTO match_record (game_code, currency_type, requirement, bet, tax, win, lose, bot_win, bot_lose, match_data, minah_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, gameCode, currencyType, requirement, 0, tax, humanWon, humanLost, botWon, botLost, string(rawBytes), minahMatchId)
	var matchRecordId int64
	var err error
	err = row.Scan(&matchRecordId)
	if err != nil {
		fmt.Println("row.Scan(&matchRecordId)", err)
	}
	//
	for _, playerId := range playerIds {
		queryString := "INSERT INTO player_match_record (player_id, match_record_id) VALUES ($1,$2)"
		_, err = dataCenter.Db().Exec(queryString, playerId, matchRecordId)
		if err != nil {
			log.LogSerious("err insert to match record %v, playerId %d, matchId %d", err, playerId, matchRecordId)
		}
	}
	//
	return matchRecordId
}

// view comment LogMatchRecord2,
// moreMatchData use for save startingHands, .., more specific info
func LogMatchRecord3(
	gameCode string, currencyType string, requirement int64, tax int64,
	humanWon int64, humanLost int64, botWon int64, botLost int64,
	minahMatchId string, playerIpAdds map[int64]string,
	playerResults []map[string]interface{}, moreMatchData map[string]interface{},
) int64 {
	//
	playerIds := []int64{}
	for pid, _ := range playerIpAdds {
		playerIds = append(playerIds, pid)
	}
	//
	matchData := map[string]interface{}{
		"players_id_when_start": playerIds,
		"players_ip_when_start": playerIpAdds,
		"results":               playerResults,
	}
	for k, v := range moreMatchData {
		matchData[k] = v
	}
	rawBytes, _ := json.Marshal(matchData)
	//
	queryString := "INSERT INTO match_record (game_code, currency_type, requirement, bet, tax, win, lose, bot_win, bot_lose, match_data, minah_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, gameCode, currencyType, requirement, 0, tax, humanWon, humanLost, botWon, botLost, string(rawBytes), minahMatchId)
	var matchRecordId int64
	var err error
	err = row.Scan(&matchRecordId)
	if err != nil {
		fmt.Println("row.Scan(&matchRecordId)", err)
	}
	//
	for _, playerId := range playerIds {
		queryString := "INSERT INTO player_match_record (player_id, match_record_id) VALUES ($1,$2)"
		_, err = dataCenter.Db().Exec(queryString, playerId, matchRecordId)
		if err != nil {
			log.LogSerious("err insert to match record %v, playerId %d, matchId %d", err, playerId, matchRecordId)
		}
	}
	//
	return matchRecordId
}

func logRefererIdForPurchase(playerId int64, purchaseType string, cardCode string, cardSerial string) int64 {
	queryString := "INSERT INTO purchase_referer (player_id, purchase_type," +
		" card_code, card_serial) " +
		" VALUES ($1, $2, $3, $4) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, playerId, purchaseType, cardCode, cardSerial)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error log purchase referer detail %s %s %s %s %v, playerid %d, cardCode %s, cardSerial %s",
				val.Detail, val.Hint, val.InternalQuery, val.Message, err, playerId, cardCode, cardSerial)
		}
	}
	return id
}

func logTransactionIdRefererId(refererId int64, transactionId string) {
	queryString := "UPDATE purchase_referer SET transaction_id = $1 WHERE id = $2"
	_, err := dataCenter.Db().Exec(queryString, transactionId, refererId)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error update purchase referer detail %s %s %s %s %v, referer %d, transaction %s", val.Detail, val.Hint, val.InternalQuery, val.Message, err,
				refererId, transactionId)
		}
	}
}

// include log currency record
func LogPurchaseRecord(playerId int64, transactionId string, purchaseType string,
	cardCode string, currencyType string,
	purchase int64, moneyBefore int64, moneyAfter int64) {
	queryString := "INSERT INTO purchase_record (player_id, transaction_id, purchase_type," +
		" card_code, currency_type, purchase, value_before, value_after) " +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, playerId, transactionId, purchaseType, cardCode, currencyType, purchase, moneyBefore, moneyAfter)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error log purchase record detail %s %s %s %s %v", val.Detail, val.Hint, val.InternalQuery, val.Message, err)
		}
	}

	if purchaseType != "start_game" &&
		purchaseType != "admin_add" &&
		purchaseType != ACTION_ADMIN_CHANGE {
		logFirstTimePurchase(playerId)
		logFirstTimePurchaseDaily(playerId)
	}

	LogCurrencyRecord(playerId,
		"purchase",
		"",
		map[string]interface{}{
			"purchase_record_id": id,
			"transaction_id":     transactionId,
			"purchase_type":      purchaseType,
		},
		currencyType,
		moneyBefore,
		moneyAfter,
		moneyAfter-moneyBefore)
}

// include log currency record
func LogPurchaseRecord2(playerId int64, transactionId string, purchaseType string,
	cardCode string, currencyType string,
	purchase int64, moneyBefore int64, moneyAfter int64, realMoneyValue int64) {
	queryString := "INSERT INTO purchase_record (player_id, transaction_id, purchase_type," +
		" card_code, currency_type, purchase, value_before, value_after, real_money_value) " +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString,
		playerId, transactionId, purchaseType,
		cardCode, currencyType, purchase,
		moneyBefore, moneyAfter, realMoneyValue)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error log purchase record detail %s %s %s %s %v", val.Detail, val.Hint, val.InternalQuery, val.Message, err)
		}
	}

	if purchaseType != "start_game" &&
		purchaseType != "admin_add" &&
		purchaseType != ACTION_ADMIN_CHANGE {
		logFirstTimePurchase(playerId)
		logFirstTimePurchaseDaily(playerId)
	}

	LogCurrencyRecord(playerId,
		"purchase",
		"",
		map[string]interface{}{
			"purchase_record_id": id,
			"transaction_id":     transactionId,
			"purchase_type":      purchaseType,
		},
		currencyType,
		moneyBefore,
		moneyAfter,
		moneyAfter-moneyBefore)
}

func logCurrencyRecord(playerId int64, action string, gameCode string, additionalData map[string]interface{}, currencyType string, valueBefore int64, valueAfter int64, change int64) {
	var additionalDataByte []byte
	var err error
	if len(additionalData) != 0 {
		additionalDataByte, err = json.Marshal(additionalData)
		if err != nil {
			log.LogSerious("Err log currency record %v %v", err, additionalData)
			return
		}
	}
	queryString := "INSERT INTO currency_record (player_id, action, game_code, additional_data,currency_type, change, value_before, value_after) " +
		"VALUES ($1, $2, $3, $4, $5, $6,$7,$8)"
	_, err = dataCenter.Db().Exec(queryString, playerId, action, gameCode, string(additionalDataByte), currencyType, change, valueBefore, valueAfter)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error log currency record detail %s %s %s %s %v", val.Detail, val.Hint, val.InternalQuery, val.Message, err)
		}
	}
}

func logVipPointRecord(playerId int64, action string, gameCode string, additionalData map[string]interface{}, vipPointBefore int64, vipPointAfter int64, change int64) {
	var additionalDataByte []byte
	var err error
	if len(additionalData) != 0 {
		additionalDataByte, err = json.Marshal(additionalData)
		if err != nil {
			log.LogSerious("Err log vipPoint record %v %v", err, additionalData)
			return
		}
	}
	queryString := "INSERT INTO vip_point_record (player_id, action, game_code, additional_data, change, vip_point_before, vip_point_after) " +
		"VALUES ($1, $2, $3, $4, $5, $6,$7)"
	_, err = dataCenter.Db().Exec(queryString, playerId, action, gameCode, string(additionalDataByte), change, vipPointBefore, vipPointAfter)
	if err != nil {
		if val, ok := err.(*pq.Error); ok {
			log.LogSerious("Error log vipPoint record detail %s %s %s %s %v", val.Detail, val.Hint, val.InternalQuery, val.Message, err)
		}
	}
}

func logCCU(totalOnline int, totalNormalOnline int, totalBotOnline int, onlineGameData map[string]interface{}) {
	var onlineGameByte []byte
	var err error
	onlineGameByte, err = json.Marshal(onlineGameData)
	if err != nil {
		log.LogSerious("Err log ccu record %v %v", err, onlineGameData)
		return
	}

	queryString := "INSERT INTO ccu_record (online_total_count, online_bot_count, online_normal_count, game_online_data) " +
		"VALUES ($1,$2,$3,$4)"
	_, err = dataCenter.Db().Exec(queryString, totalOnline, totalBotOnline, totalNormalOnline, string(onlineGameByte))
	if err != nil {
		log.LogSerious("err log ccu %v", err)
	}
}

func logBankRecord(matchId int64, gameCode string, currencyType string, moneyBefore int64, moneyAfter int64) {
	if matchId != 0 {
		queryString := "INSERT INTO bank_record (match_id, game_code, currency_type, value_before, value_after) VALUES ($1, $2, $3, $4, $5)"
		_, err := dataCenter.Db().Exec(queryString, matchId, gameCode, currencyType, moneyBefore, moneyAfter)
		if err != nil {
			log.LogSerious("err log bank record %v", err)
		}
	}
}

func logBankRecordByBot(playerId int64, gameCode string, currencyType string, moneyBefore int64, moneyAfter int64) {
	queryString := "INSERT INTO bank_record (player_id, game_code, currency_type, value_before, value_after) VALUES ($1, $2, $3, $4, $5)"
	_, err := dataCenter.Db().Exec(queryString, playerId, gameCode, currencyType, moneyBefore, moneyAfter)
	if err != nil {
		log.LogSerious("err log bot bank record %v", err)
	}
}

func logAdminActivity(id int64, possibleIps string) {
	fmt.Println("log acti")
	queryString := "INSERT INTO admin_login_activity (admin_id, possible_ips) VALUES ($1,$2)"
	_, err := dataCenter.Db().Exec(queryString, id, possibleIps)
	if err != nil {
		log.LogSerious("err log admin activity %v", err)
	}
}

// player_id is PK of table purchase_first_time
func logFirstTimePurchase(playerId int64) {
	query := "INSERT INTO purchase_first_time " +
		"(player_id, datetime) " +
		"VALUES ($1, $2)"
	_, err := dataCenter.Db().Exec(query, playerId, time.Now().UTC())
	if err != nil {
		// s := fmt.Sprintf("ERROR: %v", err)
		// duplicate
		return
	}
	return
}

// player_id, date_s is PK of table purchase_first_time
func logFirstTimePurchaseDaily(playerId int64) {
	query := "INSERT INTO purchase_first_time_daily " +
		"(player_id, date_s) " +
		"VALUES ($1, $2)"
	_, err := dataCenter.Db().Exec(query, playerId, getDateStr(time.Now()))
	if err != nil {
		// s := fmt.Sprintf("ERROR: %v", err)
		// duplicate
		return
	}
	return
}

func CheckIsInFirstTimePurchase(playerId int64) bool {
	query := "SELECT player_id FROM purchase_first_time " +
		"WHERE player_id = $1 "
	row := dataCenter.Db().QueryRow(query, playerId)
	var pid int64
	err := row.Scan(&pid)
	if err != nil {
		return false
	} else {
		return true
	}
}

func CheckIsInFirstTimePurchaseDaily(playerId int64) bool {
	query := "SELECT player_id, date_s FROM purchase_first_time_daily " +
		"WHERE player_id = $1 AND date_s = $2 "
	row := dataCenter.Db().QueryRow(query, playerId, getDateStr(time.Now()))
	var pid int64
	var ds string
	err := row.Scan(&pid, &ds)
	if err != nil {
		return false
	} else {
		return true
	}
}

func LogPlatformPartner(playerId int64, isRegister bool, platform string, partner string) error {
	query := "INSERT INTO player_source (player_id) VALUES ($1)"
	_, err := dataCenter.Db().Exec(query, playerId)
	if err != nil {
		// dont do anything if duplicate
	}
	//
	var query1 string
	if isRegister {
		query1 = "UPDATE player_source " +
			"SET register_platform = $1, register_partner = $2, register_time = $3 " +
			"WHERE player_id = $4; "
	} else {
		query1 = "UPDATE player_source " +
			"SET last_login_platform = $1, last_login_partner = $2, last_login_time = $3 " +
			"WHERE player_id = $4; "
	}
	_, err = dataCenter.Db().Exec(query1, platform, partner, time.Now().UTC(), playerId)
	if err != nil {
		fmt.Println("ERROR ERROR ERROR", err)
		return err
	}
	return nil
}

// hihi
func LogChangePartner(playerId int64, partner string) error {
	var query1 string
	query1 = "UPDATE player_source " +
		"SET register_partner = $1 " +
		"WHERE player_id = $2; "
	_, _ = dataCenter.Db().Exec(query1, partner, playerId)
	return nil
}

//
func GetPartner(playerId int64) string {
	row := dataCenter.Db().QueryRow(
		"SELECT register_partner FROM player_source WHERE player_id = $1 ",
		playerId)
	var r string
	row.Scan(&r)
	return r
}

// for turn on and off payment methods in client
func SetIapAndCardPay(playerId int64, isIapOn bool, isCardPayOn bool) {
	query := "UPDATE player_source " +
		"SET is_iap_on = $1, is_card_pay_on = $2 " +
		"WHERE player_id = $3;"
	_, err := dataCenter.Db().Exec(query, isIapOn, isCardPayOn, playerId)
	if err != nil {
		fmt.Println("ERROR SetIapAndCardPay", err)
	}
}

// for turn on and off payment methods in client
func GetIapAndCardPay(playerId int64) (isIapOn bool, isCardPayOn bool) {
	query := "SELECT is_iap_on, is_card_pay_on FROM player_source WHERE player_id = $1"
	row := dataCenter.Db().QueryRow(query, playerId)
	err := row.Scan(&isIapOn, &isCardPayOn)
	if err != nil {
		fmt.Println("ERROR GetIapAndCardPay", err)
		isIapOn = true
		isCardPayOn = false
	}
	return
}

//
func GetRegisterPlatform(playerId int64) string {
	var rp sql.NullString
	query := "SELECT register_platform FROM player_source WHERE player_id = $1"
	row := dataCenter.Db().QueryRow(query, playerId)
	err := row.Scan(&rp)
	if err != nil {
		fmt.Println("ERROR GetRegisterPlatform", err)
		return ""
	}
	return rp.String
}

// for turn on and off payment methods in client
func SetIsStoreTester(playerId int64, isTester bool) {
	query := "UPDATE player_source " +
		"SET is_store_tester = $1 " +
		"WHERE player_id = $2;"
	_, err := dataCenter.Db().Exec(query, isTester, playerId)
	if err != nil {
		fmt.Println("ERROR SetIapAndCardPay", err)
	}
}

// for turn on and off payment methods in client
func GetIsStoreTester(playerId int64) (isTester bool) {
	query := "SELECT is_store_tester FROM player_source WHERE player_id = $1"
	row := dataCenter.Db().QueryRow(query, playerId)
	err := row.Scan(&isTester)
	if err != nil {
		fmt.Println("ERROR GetIapAndCardPay", err)
		isTester = true
	}
	return
}

//
func LogEventTopResult(
	event_name string, starting_time time.Time, finishing_time time.Time,
	map_position_to_prize string, full_order string, is_paid bool) {
	query := "INSERT INTO event_top_result " +
		"(event_name, starting_time, finishing_time, map_position_to_prize, " +
		"    full_order, is_paid) " +
		"VALUES ($1, $2, $3, $4, $5, $6) "
	dataCenter.Db().Exec(query, event_name, starting_time, finishing_time,
		map_position_to_prize, full_order, is_paid)
}

//
func LogEventCollectingPiecesResult(
	event_name string, starting_time time.Time, finishing_time time.Time,
	n_pieces_to_complete int, n_limit_prizes int, n_rare_pieces int,
	map_pid_to_map_pieces string, is_paid bool) {
	query := "INSERT INTO event_collecting_pieces_result " +
		"(event_name, starting_time, finishing_time, n_pieces_to_complete, " +
		"    n_limit_prizes, n_rare_pieces, map_pid_to_map_pieces, is_paid) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) "
	dataCenter.Db().Exec(query, event_name, starting_time, finishing_time,
		n_pieces_to_complete, n_limit_prizes, n_rare_pieces,
		map_pid_to_map_pieces, is_paid)
}

// return format: 1992-08-20
func getDateStr(t time.Time) string {
	t = t.Add(7 * time.Hour)
	s := t.Format(time.RFC3339)
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] == 'T' {
			break
		}
		if i == len(s)-1 { // cant find 'T' in s
			i = -1
		}
	}
	if i == -1 {
		// cant happen
		return ""
	} else {
		return s[:i]
	}
}

func LogBankCharging(player_id int64, amount_vnd float64, amount_myr float64,
	amount_kim float64, kim_before float64, kim_after float64, paytrust88_data string) {
	query := "INSERT INTO purchase_record_bank " +
		"(player_id, amount_vnd, amount_myr, amount_kim, " +
		"    kim_before, kim_after, paytrust88_data) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) "
	dataCenter.Db().Exec(query, player_id, amount_vnd, amount_myr, amount_kim,
		kim_before, kim_after, paytrust88_data)
}

func LogIapAndroidCharging(
	player_id int64, order_id string, receipt string) error {
	query := "INSERT INTO purchase_record_iap_android " +
		"(order_id, player_id, receipt) " +
		"VALUES ($1, $2, $3) "
	_, err := dataCenter.Db().Exec(query, order_id, player_id, receipt)
	return err
}

func LogSumCharging(
	player_id int64, purchase_type string, change int64) error {
	timeout := time.Now().Add(2 * time.Second)
	var e error
	for {
		e = logSumCharging(player_id, purchase_type, change)
		if e == nil {
			break
		} else if time.Now().After(timeout) {
			e = errors.New("LogSumCharging err1 timeout")
			break
		} else {
			// continue
		}
	}
	return e
}

func logSumCharging(
	player_id int64, purchase_type string, change int64) error {
	tx, err := dataCenter.Db().Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`SET TRANSACTION ISOLATION LEVEL Serializable`)
	if err != nil {
		return err
	}
	var sum_value int64
	{
		stmt, err := tx.Prepare(
			"SELECT sum_value FROM purchase_sum " +
				"WHERE player_id = $1 AND purchase_type = $2 ")
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		row := stmt.QueryRow(player_id, purchase_type)
		err = row.Scan(&sum_value)
		if err != nil {
			stmtL1, errL1 := tx.Prepare("INSERT INTO purchase_sum " +
				"    (player_id, purchase_type, sum_value) " +
				"VALUES ($1, $2, $3) ")
			if errL1 != nil {
				tx.Rollback()
				return errL1
			}
			defer stmtL1.Close()
			_, errL1 = stmtL1.Exec(player_id, purchase_type, 0)
			if errL1 != nil {
				tx.Rollback()
				return errL1
			}
		}
	}
	{
		stmt, err := tx.Prepare(
			"UPDATE purchase_sum SET sum_value = $1 " +
				"WHERE player_id = $2 AND purchase_type = $3 ")
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(sum_value+change, player_id, purchase_type)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// return inserted row id
func LogTransfer(sender_id int64, target_id int64, amount_kim int64) int64 {
	row := dataCenter.Db().QueryRow(
		"INSERT INTO player_transfer_record "+
			"(sender_id, target_id, amount_kim) "+
			"VALUES ($1, $2, $3) RETURNING id",
		sender_id, target_id, amount_kim)
	var rowid int64
	e := row.Scan(&rowid)
	if e != nil {
		fmt.Println("ERROR ", e)
	}
	return rowid
}
