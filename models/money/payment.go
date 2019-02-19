package money

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"strings"
	"time"
)

type Payment struct {
	id          int64
	playerId    int64
	payment     int64
	moneyBefore int64
	moneyAfter  int64
	status      string
	adminId     int64

	paymentType string

	cardCode string
	cardId   int64

	data map[string]interface{}

	createdAt time.Time
	repliedAt time.Time
}

func (payment *Payment) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = payment.id
	data["player_id"] = payment.playerId
	data["card_code"] = payment.cardCode
	data["payment"] = payment.payment
	data["value_before"] = payment.moneyBefore
	data["value_after"] = payment.moneyAfter
	data["status"] = payment.status
	data["admin_id"] = payment.adminId
	data["card_id"] = payment.cardId
	data["created_at"] = utils.FormatTime(payment.createdAt)
	data["replied_at"] = utils.FormatTime(payment.repliedAt)
	return data
}

func getPayment(paymentId int64) *Payment {
	queryString := "SELECT id,player_id, card_code, data, payment_type FROM payment_record WHERE id = $1 AND status = $2"
	var id, playerId int64
	var cardCode []byte
	var dataByte []byte
	var paymentType string
	row := dataCenter.Db().QueryRow(queryString, paymentId, "requested")
	err := row.Scan(&id, &playerId, &cardCode, &dataByte, &paymentType)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.LogSerious("err check payment already accepted %v", err)
		return nil
	}

	var data map[string]interface{}
	json.Unmarshal(dataByte, &data)
	payment := &Payment{
		id:          id,
		playerId:    playerId,
		cardCode:    string(cardCode),
		data:        data,
		paymentType: paymentType,
	}
	return payment
}

func GetRequestedPaymentData(keyword string, paymentType string,
	limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {

	paymentTypes := make([]string, 0)
	if paymentType == "" {
		paymentTypes = []string{"'card'", "'gift'"}
	} else {
		paymentTypes = []string{fmt.Sprintf("'%s'", paymentType)}
	}
	paymentTypeString := fmt.Sprintf("(%s)", strings.Join(paymentTypes, ","))

	queryString := fmt.Sprintf("SELECT COUNT(*)"+
		" FROM payment_record as payment, player as player"+
		" WHERE payment.status = $1 "+
		" AND player.id = payment.player_id AND player.username LIKE $2 AND payment.payment_type IN %s", paymentTypeString)
	row := dataCenter.Db().QueryRow(queryString, "requested", fmt.Sprintf("%%%s%%", keyword))
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = fmt.Sprintf("SELECT payment.id, player.id, player.username, payment.card_code, payment.value_before,"+
		" payment.value_after, payment.created_at, payment.status, payment.payment_type, payment.data"+
		" FROM payment_record as payment, player as player"+
		" WHERE payment.status = $1 "+
		" AND player.id = payment.player_id AND player.username LIKE $2 AND payment.payment_type IN %s"+
		" ORDER BY -payment.id LIMIT $3 OFFSET $4", paymentTypeString)
	rows, err := dataCenter.Db().Query(queryString, "requested", fmt.Sprintf("%%%s%%", keyword), limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, playerId, moneyBefore, moneyAfter int64
		var username, status string
		var cardCode, dataByte []byte
		var paymentType string
		var createdAt time.Time
		err = rows.Scan(&id, &playerId, &username, &cardCode, &moneyBefore, &moneyAfter, &createdAt, &status, &paymentType, &dataByte)
		if err != nil {
			return nil, 0, err
		}

		var paymentData map[string]interface{}
		json.Unmarshal(dataByte, &paymentData)

		data := make(map[string]interface{})
		data["id"] = id
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		data["player_id"] = playerId
		data["username"] = username
		data["card_code"] = string(cardCode)
		data["status"] = status
		data["payment_type"] = paymentType
		data["data"] = paymentData
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)

		results = append(results, data)
	}
	return results, total, nil
}

func GetRepliedPaymentData(keyword string, paymentType string,
	startDate time.Time, endDate time.Time,
	limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {

	paymentTypes := make([]string, 0)
	if paymentType == "" {
		paymentTypes = []string{"'card'", "'gift'"}
	} else {
		paymentTypes = []string{fmt.Sprintf("'%s'", paymentType)}
	}
	paymentTypeString := fmt.Sprintf("(%s)", strings.Join(paymentTypes, ","))

	queryString := fmt.Sprintf("SELECT COUNT(*)"+
		" FROM payment_record as payment, player as player"+
		" WHERE payment.status != $1 AND payment.created_at >= $2 AND payment.created_at <= $3"+
		" AND player.id = payment.player_id AND player.username LIKE $4 AND payment.payment_type IN %s", paymentTypeString)
	row := dataCenter.Db().QueryRow(queryString, "requested", startDate.UTC(), endDate.UTC(), fmt.Sprintf("%%%s%%", keyword))
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = fmt.Sprintf("SELECT payment.id, player.id, player.username, payment.card_code, payment.value_before,"+
		" payment.value_after, payment.created_at,admin.username,payment.replied_at, payment.status, payment.payment_type, payment.data"+
		" FROM payment_record as payment, player as player, admin_account as admin"+
		" WHERE payment.status != $1 AND payment.created_at >= $2 AND payment.created_at <= $3 AND admin.id = payment.replied_by_admin_id"+
		" AND player.id = payment.player_id AND player.username LIKE $4 AND payment.payment_type IN %s"+
		" ORDER BY -payment.id LIMIT $5 OFFSET $6", paymentTypeString)
	rows, err := dataCenter.Db().Query(queryString, "requested", startDate.UTC(), endDate.UTC(), fmt.Sprintf("%%%s%%", keyword), limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, playerid, moneyBefore, moneyAfter int64
		var username, paymentType, adminUsername, status string
		var cardCode, dataByte []byte
		var createdAt, repliedAt time.Time
		err = rows.Scan(&id, &playerid, &username, &cardCode, &moneyBefore, &moneyAfter, &createdAt, &adminUsername, &repliedAt, &status, &paymentType, &dataByte)
		if err != nil {
			return nil, 0, err
		}
		var paymentData map[string]interface{}
		json.Unmarshal(dataByte, &paymentData)

		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerid
		data["value_before"] = utils.FormatWithComma(moneyBefore)
		data["value_after"] = utils.FormatWithComma(moneyAfter)
		data["username"] = username
		data["card_code"] = string(cardCode)
		data["admin_username"] = adminUsername
		data["status"] = status
		data["data"] = paymentData
		data["payment_type"] = paymentType
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["replied_at"] = utils.FormatTimeToVietnamDateTimeString(repliedAt)

		results = append(results, data)
	}
	return results, total, nil
}
