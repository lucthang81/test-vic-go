package record

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"math"
	"strings"
	"time"
)

func GetPurchaseData(startDate time.Time, endDate time.Time, reportType string, page int64) map[string]interface{} {
	purchaseTypes := make([]string, 0)
	if reportType == "" {
		purchaseTypes = []string{"'paybnb'", "'appvn'", "'iap'", "'admin_add'"}
	} else {
		purchaseTypes = []string{fmt.Sprintf("'%s'", reportType)}
	}
	purchaseTypesString := fmt.Sprintf("(%s)", strings.Join(purchaseTypes, ","))

	results := make(map[string]interface{})
	var totalPayment, botPayment, normalPayment int64

	totalList := make([]map[string]interface{}, 0)
	botList := make([]map[string]interface{}, 0)
	normalPlayerList := make([]map[string]interface{}, 0)

	limit := int64(100)
	offset := (page - 1) * limit

	var totalQueryString, queryString string
	if startDate.IsZero() && endDate.IsZero() {
		totalQueryString = fmt.Sprintf("SELECT COUNT(*) "+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.purchase_type IN %s AND purchase.player_id = player.id AND player.player_type != 'bot'", purchaseTypesString)
		queryString = fmt.Sprintf("SELECT purchase.id, purchase.transaction_id, purchase.purchase_type, purchase.card_code, purchase.player_id, player.username,"+
			" player.player_type, purchase.purchase, purchase.value_before, purchase.value_after, purchase.created_at"+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.purchase_type IN %s "+
			" AND purchase.player_id = player.id AND player.player_type != 'bot' ORDER BY -purchase.id LIMIT $1 OFFSET $2", purchaseTypesString)
	} else {
		totalQueryString = fmt.Sprintf("SELECT COUNT(*) "+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.created_at >= $1 AND purchase.created_at <= $2 "+
			" AND purchase.purchase_type IN %s AND purchase.player_id = player.id AND player.player_type != 'bot'", purchaseTypesString)
		queryString = fmt.Sprintf("SELECT purchase.id, purchase.transaction_id, purchase.purchase_type, purchase.card_code, purchase.player_id, player.username,"+
			" player.player_type, purchase.purchase, purchase.value_before, purchase.value_after, purchase.created_at"+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.created_at >= $1 AND purchase.created_at <= $2 AND purchase.purchase_type IN %s"+
			" AND purchase.player_id = player.id AND player.player_type != 'bot' ORDER BY -purchase.id LIMIT $3 OFFSET $4", purchaseTypesString)
	}
	var row *sql.Row
	var rows *sql.Rows

	if startDate.IsZero() && endDate.IsZero() {
		row = dataCenter.Db().QueryRow(totalQueryString)
	} else {
		row = dataCenter.Db().QueryRow(totalQueryString, startDate.UTC(), endDate.UTC())
	}

	var count int64
	err := row.Scan(&count)
	if err != nil {
		count = 0
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))
	if startDate.IsZero() && endDate.IsZero() {
		rows, err = dataCenter.Db().Query(queryString, limit, offset)
	} else {
		rows, err = dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC(), limit, offset)
	}
	if err != nil {
		log.LogSerious("Error fetch purchase record %v", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var playerId int64
		var transactionId string
		var purchaseType string
		var cardCode string
		var username string
		var playerType string
		var purchase int64
		var moneyAfter int64
		var moneyBefore int64
		var createdAt time.Time
		err = rows.Scan(&id, &transactionId, &purchaseType, &cardCode, &playerId, &username, &playerType, &purchase, &moneyBefore, &moneyAfter, &createdAt)
		if err != nil {
			log.LogSerious("Error fetch purchase record %v", err)
		}
		data := make(map[string]interface{})
		data["id"] = id
		data["transaction_id"] = transactionId
		data["card_code"] = cardCode
		data["player_id"] = playerId
		data["player_type"] = playerType
		data["purchase_type"] = purchaseType
		data["username"] = username
		data["purchase"] = utils.FormatWithComma(purchase)
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		dateString, timeString := utils.FormatTimeToVietnamTime(createdAt)
		data["created_at"] = fmt.Sprintf("%s %s", dateString, timeString)

		totalList = append(totalList, data)
		totalPayment += purchase
		if playerType == "bot" {
			botList = append(botList, data)
			botPayment += purchase
		} else {
			normalPlayerList = append(normalPlayerList, data)
			normalPayment += purchase
		}
	}

	results["page"] = page
	results["num_pages"] = numPages
	results["total_list"] = totalList
	results["normal_list"] = normalPlayerList
	results["bot_list"] = botList

	var totalPurchaseByCard int64
	var totalPurchaseByIAP int64
	var totalPurchasePayBnb int64
	var totalPurchaseAppvn int64
	var totalPurchaseIOS int64
	var totalPurchaseAndroid int64
	var groupCount []map[string]interface{}
	groupCount = make([]map[string]interface{}, 0)

	if startDate.IsZero() && endDate.IsZero() {
		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE (purchase_type = 'appvn' OR purchase_type = 'paybnb')"
		totalPurchaseByCard = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'iap'"
		totalPurchaseByIAP = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'paybnb'"
		totalPurchasePayBnb = getInt64FromQuery(queryString)

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'appvn'"
		totalPurchaseAppvn = getInt64FromQuery(queryString)

		queryString = "select sum(purchase) from purchase_record where" +
			" purchase_type IN ('appvn','paybnb','iap') AND player_id in" +
			" (select distinct player_id from active_record where device_type = 'ios')"
		totalPurchaseIOS = getInt64FromQuery(queryString)

		queryString = "select sum(purchase) from purchase_record where" +
			" purchase_type IN ('appvn','paybnb','iap') AND player_id in" +
			" (select distinct player_id from active_record where device_type = 'android')"
		totalPurchaseAndroid = getInt64FromQuery(queryString)

		queryString = fmt.Sprintf("select purchase, count(id) from purchase_record where purchase_type IN %s "+
			" AND player_id IN (select id from player where player_type != 'bot')"+
			" group by purchase order by purchase", purchaseTypesString)
		rows, err := dataCenter.Db().Query(queryString)
		if err != nil {
			log.LogSerious("err fetch group purchase %v", err)
			return nil
		}
		for rows.Next() {
			var purchaseAmountType int64
			var purchaseAmountCount int64
			err := rows.Scan(&purchaseAmountType, &purchaseAmountCount)
			if err != nil {
				log.LogSerious("err fetch group purchase %v", err)
				rows.Close()
				return nil
			}
			groupData := make(map[string]interface{})
			groupData["amount"] = utils.FormatWithComma(purchaseAmountType)
			groupData["count"] = purchaseAmountCount
			groupCount = append(groupCount, groupData)
		}
		rows.Close()
	} else {
		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE (purchase_type = 'appvn' OR purchase_type = 'paybnb') AND created_at >= $1 AND created_at <= $2"
		totalPurchaseByCard = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'iap' AND created_at >= $1 AND created_at <= $2"
		totalPurchaseByIAP = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'paybnb' AND created_at >= $1 AND created_at <= $2"
		totalPurchasePayBnb = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "SELECT SUM(purchase) FROM purchase_record WHERE purchase_type = 'appvn' AND created_at >= $1 AND created_at <= $2"
		totalPurchaseAppvn = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "select sum(purchase) from purchase_record " +
			"where purchase_type IN ('appvn','paybnb','iap') AND created_at >= $1 AND " +
			" created_at <= $2 AND player_id in" +
			" (select distinct player_id from active_record where device_type = 'ios')"
		totalPurchaseIOS = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = "select sum(purchase) from purchase_record" +
			" where purchase_type IN ('appvn','paybnb','iap')" +
			" AND created_at >= $1 AND " +
			" created_at <= $2 AND player_id in" +
			" (select distinct player_id from active_record where device_type = 'android')"
		totalPurchaseAndroid = getInt64FromQuery(queryString, startDate.UTC(), endDate.UTC())

		queryString = fmt.Sprintf("select purchase, count(id) from purchase_record where "+
			"purchase_type IN %s AND created_at >= $1 AND created_at <= $2"+
			" AND player_id IN (select id from player where player_type != 'bot')"+
			" group by purchase order by purchase", purchaseTypesString)
		rows, err := dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC())
		if err != nil {
			log.LogSerious("err fetch group purchase %v", err)
			return nil
		}
		for rows.Next() {
			var purchaseAmountType int64
			var purchaseAmountCount int64
			err := rows.Scan(&purchaseAmountType, &purchaseAmountCount)
			if err != nil {
				log.LogSerious("err fetch group purchase %v", err)
				rows.Close()
				return nil
			}
			groupData := make(map[string]interface{})
			groupData["amount"] = utils.FormatWithComma(purchaseAmountType)
			groupData["count"] = purchaseAmountCount
			groupCount = append(groupCount, groupData)
		}
		rows.Close()
	}

	results["group_count"] = groupCount
	results["total_purchase"] = utils.FormatWithComma(totalPurchaseByCard + totalPurchaseByIAP)
	results["total_purchase_by_card"] = utils.FormatWithComma(totalPurchaseByCard)
	results["total_purchase_by_iap"] = utils.FormatWithComma(totalPurchaseByIAP)
	results["total_purchase_by_paybnb"] = utils.FormatWithComma(totalPurchasePayBnb)
	results["total_purchase_by_appvn"] = utils.FormatWithComma(totalPurchaseAppvn)
	results["total_purchase_by_ios"] = utils.FormatWithComma(totalPurchaseIOS)
	results["total_purchase_by_android"] = utils.FormatWithComma(totalPurchaseAndroid)
	return results
}

func GetTopPurchaseData(startDate time.Time, endDate time.Time, reportType string, page int64) map[string]interface{} {
	purchaseTypes := make([]string, 0)
	if reportType == "" {
		purchaseTypes = []string{"'paybnb'", "'appvn'", "'iap'", "'admin_add'"}
	} else {
		purchaseTypes = []string{fmt.Sprintf("'%s'", reportType)}
	}
	purchaseTypesString := fmt.Sprintf("(%s)", strings.Join(purchaseTypes, ","))

	results := make(map[string]interface{})

	totalList := make([]map[string]interface{}, 0)

	limit := int64(100)
	offset := (page - 1) * limit

	var totalQueryString, queryString string
	if startDate.IsZero() && endDate.IsZero() {
		totalQueryString = fmt.Sprintf("SELECT COUNT(*) "+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.purchase_type IN %s AND purchase.player_id = player.id AND player.player_type != 'bot' GROUP BY player_id", purchaseTypesString)
		queryString = fmt.Sprintf("SELECT player.id, player.username, SUM(purchase.purchase) as sum_purchase"+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.purchase_type IN %s "+
			" AND purchase.player_id = player.id AND player.player_type != 'bot' GROUP BY (player.id) ORDER BY sum_purchase DESC LIMIT $1 OFFSET $2", purchaseTypesString)
	} else {
		totalQueryString = fmt.Sprintf("SELECT COUNT(*) "+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.created_at >= $1 AND purchase.created_at <= $2 "+
			" AND purchase.purchase_type IN %s AND purchase.player_id = player.id AND player.player_type != 'bot' GROUP BY player_id", purchaseTypesString)
		queryString = fmt.Sprintf("SELECT player.id, player.username, SUM(purchase.purchase) as sum_purchase"+
			" FROM purchase_record as purchase, player as player"+
			" WHERE purchase.created_at >= $1 AND purchase.created_at <= $2 AND purchase.purchase_type IN %s "+
			" AND purchase.player_id = player.id AND player.player_type != 'bot' GROUP BY (player.id) ORDER BY sum_purchase DESC LIMIT $3 OFFSET $4", purchaseTypesString)
	}
	var row *sql.Row
	var rows *sql.Rows

	if startDate.IsZero() && endDate.IsZero() {
		row = dataCenter.Db().QueryRow(totalQueryString)
	} else {
		row = dataCenter.Db().QueryRow(totalQueryString, startDate.UTC(), endDate.UTC())
	}

	var count int64
	err := row.Scan(&count)
	if err != nil {
		count = 0
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))
	if startDate.IsZero() && endDate.IsZero() {
		rows, err = dataCenter.Db().Query(queryString, limit, offset)
	} else {
		rows, err = dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC(), limit, offset)
	}
	if err != nil {
		log.LogSerious("Error fetch purchase record %v", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var playerId int64
		var username string
		var purchase int64
		err = rows.Scan(&playerId, &username, &purchase)
		if err != nil {
			log.LogSerious("Error fetch purchase record %v", err)
		}
		data := make(map[string]interface{})
		data["player_id"] = playerId
		data["username"] = username
		data["purchase"] = utils.FormatWithComma(purchase)

		totalList = append(totalList, data)
	}

	results["page"] = page
	results["num_pages"] = numPages
	results["total_list"] = totalList
	return results
}
