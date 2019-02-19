package otp

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"math"
	"strconv"
	"strings"
	"time"
)

func GetOtpRewardData(keyword string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})
	keywordId, _ := strconv.ParseInt(keyword, 10, 64)

	limit := int64(100)
	offset := (page - 1) * limit

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM otp_reward WHERE player_id IN " +
		"(SELECT id from player where username LIKE $1 OR id = $2 OR phone_number LIKE $1)")
	row := dataCenter.Db().QueryRow(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))
	queryString = "select reward.id, reward.phone_number, player.username, player.id, reward.created_at FROM otp_reward as reward" +
		" LEFT JOIN (SELECT player.username, player.id FROM player as player) player ON player.id = reward.player_id" +
		" WHERE player.username LIKE $1 OR player.id = $2 OR reward.phone_number LIKE $1" +
		" ORDER BY -reward.id LIMIT $3 OFFSET $4"
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, playerId sql.NullInt64
		var phoneNumber, username sql.NullString
		var createdAt time.Time
		err = rows.Scan(&id, &phoneNumber, &username, &playerId, &createdAt)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["player_id"] = playerId.Int64
		data["phone_number"] = phoneNumber.String
		data["username"] = username.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		list = append(list, data)
	}
	results["num_pages"] = numPages
	results["results"] = list
	return
}

func GetOtpCodeData(keyword string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})
	keywordId, _ := strconv.ParseInt(keyword, 10, 64)

	limit := int64(100)
	offset := (page - 1) * limit

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM otp_code WHERE player_id IN " +
		"(SELECT id from player where username LIKE $1 OR id = $2 OR phone_number LIKE $1)")
	row := dataCenter.Db().QueryRow(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))
	queryString = "select code.id, code.phone_number, code.reason, code.otp_code," +
		" code.status, code.retry_count, code.expired_at, code.created_at," +
		" player.username, player.id FROM otp_code as code" +
		" LEFT JOIN (SELECT player.username, player.id FROM player as player) player ON player.id = code.player_id" +
		" WHERE player.username LIKE $1 OR player.id = $2 OR code.phone_number LIKE $1" +
		" ORDER BY -code.id LIMIT $3 OFFSET $4"
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, playerId, retryCount sql.NullInt64
		var phoneNumber, username, reason, otpCode, status sql.NullString
		var createdAt, expiredAt time.Time
		err = rows.Scan(&id, &phoneNumber, &reason, &otpCode, &status, &retryCount, &expiredAt,
			&createdAt, &username, &playerId)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["player_id"] = playerId.Int64
		data["phone_number"] = phoneNumber.String
		data["username"] = username.String
		data["reason"] = reason.String
		data["otp_code"] = otpCode.String
		data["status"] = status.String
		data["retry_count"] = retryCount.Int64
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["expired_at"] = utils.FormatTimeToVietnamDateTimeString(expiredAt)
		list = append(list, data)
	}
	results["num_pages"] = numPages
	results["results"] = list
	return
}

func GetEditOtpCodeForm(id int64) (data map[string]interface{}, editObject *htmlutils.EditObject) {
	queryString := "select code.phone_number, code.reason, code.otp_code," +
		" code.status, code.retry_count, code.expired_at, code.created_at," +
		" player.username, player.id FROM otp_code as code" +
		" LEFT JOIN (SELECT player.username, player.id FROM player as player) player ON player.id = code.player_id" +
		" WHERE code.id = $1"
	row := dataCenter.Db().QueryRow(queryString, id)
	var playerId, retryCount sql.NullInt64
	var phoneNumber, username, reason, otpCode, status sql.NullString
	var createdAt, expiredAt time.Time
	err := row.Scan(&phoneNumber, &reason, &otpCode, &status, &retryCount, &expiredAt,
		&createdAt, &username, &playerId)
	if err != nil {
		return nil, nil
	}
	data = make(map[string]interface{})
	data["id"] = id
	data["player_id"] = playerId.Int64
	data["phone_number"] = phoneNumber.String
	data["username"] = username.String
	data["reason"] = reason.String
	data["otp_code"] = otpCode.String
	data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)

	row1 := htmlutils.NewRadioField("Status", "status",
		status.String,
		[]string{kStatusWaitingForUser, kStatusAlreadySentSms, kStatusInvalid, kStatusValid})
	row2 := htmlutils.NewInt64Field("Retry count", "retry_count", "Retry count", retryCount.Int64)
	row3 := htmlutils.NewStringField("Expired at", "expired_at", "Expired at", utils.FormatTimeToVietnamDateTimeString(expiredAt))
	row5 := htmlutils.NewInt64HiddenField("id", id)
	editObject = htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row5},
		"/admin/otp/code/edit")

	return data, editObject
}

func EditOtpCode(data map[string]interface{}) (err error) {
	id := utils.GetInt64AtPath(data, "id")
	status := utils.GetStringAtPath(data, "status")
	retryCount := utils.GetInt64AtPath(data, "retry_count")
	expiredAtString := utils.GetStringAtPath(data, "expired_at")

	expiredTokens := strings.Split(expiredAtString, " ")
	expiredAt := utils.TimeFromVietnameseTimeString(expiredTokens[0], expiredTokens[1])

	query := "UPDATE otp_code set status = $1, retry_count = $2, expired_at = $3" +
		" WHERE id = $4"
	_, err = dataCenter.Db().Exec(query, status, retryCount, expiredAt.UTC(), id)
	return err
}
