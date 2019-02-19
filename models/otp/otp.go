package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
	"net/http"
	"strings"
	"time"
)

//8095

const (
	kAccessKey = "jsqvd53030vband1z5go"
	kSecretKey = "nkbfi6hlfozlexbdyew7b1frwohfv1m7"

	kStatusWaitingForUser = "wait"
	kStatusAlreadySentSms = "sms"
	kStatusInvalid        = "invalid"
	kStatusValid          = "valid"
)

var dataCenter *datacenter.DataCenter

type OTPCode struct {
	id          int64
	playerId    int64
	phoneNumber string
	otpCode     string
	reason      string
	passwd      string
	data        map[string]interface{}
	status      string
	retryCount  int64
	expiredAt   time.Time
	createdAt   time.Time
}

func (otpCode *OTPCode) isExpired() bool {
	if otpCode.expiredAt.Before(time.Now()) {
		return true
	}
	return false
}

func (otpCode *OTPCode) updateCurrentStatus(status string) {
	otpCode.status = status
	_, err := dataCenter.Db().Exec("UPDATE otp_code SET status = $1 WHERE id = $2", otpCode.status, otpCode.id)
	if err != nil {
		log.LogSerious("err update otp status %v", err)
	}
}

func (otpCode *OTPCode) updateCurrentRetryCount() {
	_, err := dataCenter.Db().Exec("UPDATE otp_code SET retry_count = $1 WHERE id = $2", otpCode.retryCount, otpCode.id)
	if err != nil {
		log.LogSerious("err update otp retrycount %v", err)
	}
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

func HandleServiceRequest(request *http.Request) (status int, sms string) {
	status = 0
	sms = "Da co loi trong qua trinh xu ly, mong quy khach vui long thu lai sau"

	accessKey := request.FormValue("access_key")
	command := request.FormValue("command")
	content := request.FormValue("mo_message")
	phoneNumberFull := request.FormValue("msisdn")
	requestId := request.FormValue("request_id")
	requestTime := request.FormValue("request_time")
	shortCode := request.FormValue("short_code")
	signature := request.FormValue("signature")

	if accessKey != kAccessKey {
		log.LogSerious("err access key not match handle service request,%s %s", accessKey, kAccessKey)
		return
	}

	willBeHased := fmt.Sprintf("access_key=%s&command=%s&mo_message=%s&msisdn=%s&request_id=%s&request_time=%s&short_code=%s",
		kAccessKey, command, content, phoneNumberFull, requestId, requestTime, shortCode)
	// fmt.Println("willbehash", willBeHased)

	hashedSignature := ComputeHmac256(willBeHased, kSecretKey)
	// fmt.Println("signare compare", hashedSignature, signature)
	if hashedSignature != signature {
		log.LogSerious("err signature not match")
		return
	}

	if !isRequestIdUnique(requestId) {
		sms = ""
		return
	}
	recordRequestId(requestId)

	phoneNumber := utils.NormalizePhoneNumber(phoneNumberFull)

	otpCode := getOtpCodeForPhoneNumber(phoneNumber)
	if otpCode == nil {
		sms = "Ban chua dang ky nhan ma OTP, tin nhan khong co hieu luc"
		return
	}
	sms = fmt.Sprintf("Ma xac thuc OTP cua ban la %s, ma xac thuc chi co hieu luc trong 5 phut ke tu khi ban gui tin nhan. CSKH: %s",
		otpCode.otpCode,
		game_config.SupportPhoneNumber())
	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	status = 1
	return
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func canRequestNewOtpCode(playerId int64) (err error) {
	otpCode := getLatestOtpCodeForPlayer(playerId)
	if otpCode != nil && otpCode.status != kStatusValid {
		if !otpCode.isExpired() {
			return details_error.NewError("err:wait_before_request_otp", map[string]interface{}{
				"first_message": fmt.Sprintf("Xin hãy đợi %s trước khi yêu cầu mã OTP mới", game_config.OtpRequestAgainDuration().String()),
			})
		}
	}
	return nil
}

func hkCanRequestNewOtpCode(playerId int64) (otp string) {
	otpCode := getLatestOtpCodeForPlayer(playerId)
	if otpCode != nil && otpCode.status != kStatusValid {
		if !otpCode.isExpired() {
			return otpCode.otpCode
		}
	}
	return ""
}

func getLatestOtpCodeForPlayer(playerId int64) *OTPCode {
	query := "SELECT id, otp_code, phone_number, reason, status, retry_count, expired_at, created_at" +
		" FROM otp_code WHERE player_id = $1 ORDER BY -id LIMIT 1"
	row := dataCenter.Db().QueryRow(query, playerId)
	var id, retryCount int64
	var otpCodeString, reason, status, phoneNumber sql.NullString
	var expiredAt, createdAt time.Time
	err := row.Scan(&id, &otpCodeString, &phoneNumber, &reason, &status, &retryCount, &expiredAt, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.LogSerious("err fetch latest otp code %v,playerId %d", err, playerId)
		return nil
	}

	otpCode := &OTPCode{
		id:          id,
		playerId:    playerId,
		otpCode:     otpCodeString.String,
		phoneNumber: phoneNumber.String,
		reason:      reason.String,
		status:      status.String,
		retryCount:  retryCount,
		expiredAt:   expiredAt,
		createdAt:   createdAt,
	}
	return otpCode
}

func HKGetLatestOtpCodeForPlayer(phone string, otp string) *OTPCode {
	query := fmt.Sprintf("SELECT id, otp_code, phone_number, reason, status, retry_count, expired_at, created_at,player_id, passwd"+
		" FROM otp_code WHERE phone_number = '%v' AND otp_code = '%v' ORDER BY id DESC LIMIT 1", phone, otp)
	row := dataCenter.Db().QueryRow(query)
	var id, retryCount, playerId int64
	var passwd, otpCodeString, reason, status, phoneNumber sql.NullString
	var expiredAt, createdAt time.Time
	err := row.Scan(&id, &otpCodeString, &phoneNumber, &reason, &status, &retryCount, &expiredAt, &createdAt, &playerId, &passwd)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.LogSerious("err fetch latest otp code %v,phone %d", err, phone)
		return nil
	}

	otpCode := &OTPCode{
		id:          id,
		playerId:    playerId,
		otpCode:     otpCodeString.String,
		phoneNumber: phoneNumber.String,
		reason:      reason.String,
		status:      status.String,
		retryCount:  retryCount,
		expiredAt:   expiredAt,
		createdAt:   createdAt,
		passwd:      passwd.String,
	}
	return otpCode
}

func getOtpCodeForPhoneNumber(phoneNumber string) *OTPCode {
	query := "SELECT id, otp_code, player_id, reason, data, status,retry_count, expired_at, created_at" +
		" FROM otp_code WHERE phone_number = $1 ORDER BY -id LIMIT 1"
	row := dataCenter.Db().QueryRow(query, phoneNumber)
	var id, playerId, retryCount int64
	var otpCodeString, reason, data, status sql.NullString
	var expiredAt, createdAt time.Time
	err := row.Scan(&id, &otpCodeString, &playerId, &reason, &data, &status, &retryCount, &expiredAt, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.LogSerious("err fetch latest otp code %v,playerId %d", err, playerId)
		return nil
	}

	dataMap := make(map[string]interface{})
	json.Unmarshal([]byte(data.String), &dataMap)

	otpCode := &OTPCode{
		id:          id,
		playerId:    playerId,
		otpCode:     otpCodeString.String,
		phoneNumber: phoneNumber,
		reason:      reason.String,
		data:        dataMap,
		status:      status.String,
		retryCount:  retryCount,
		expiredAt:   expiredAt,
		createdAt:   createdAt,
	}
	return otpCode
}

//GetLeaderBoard(gameCode string, currencyType string,page int64,statisticType string )
func GetLeaderBoard(gameCode string, currencyType string, page int64, statisticType string) (
	map[string]interface{}, error) {
	data := map[string]interface{}{
		"gameCode":     gameCode,
		"currencyType": currencyType,
		"lines":        []map[string]interface{}{},
	}
	query := fmt.Sprintf("SELECT id, player_id, win, lose, draw, "+
		" won_money, lost_money, username"+
		" FROM public.match_statistic WHERE currency_type = '%s' and game_code='%s' LIMIT 10 OFFSET %v ;",
		strings.ToUpper(currencyType), strings.ToUpper(gameCode), (page-1)*10)
	rows, err := dataCenter.Db().Query(query)

	if err != nil {
		fmt.Println(err.Error())
		return data, nil
	}
	defer rows.Close()

	lines := []map[string]interface{}{}
	for rows.Next() {
		var id, player_id, win, lose, draw, won_money, lost_money int64
		var username sql.NullString
		if err := rows.Scan(&id, &player_id, &win, &lose, &draw, &won_money, &lost_money, &username); err != nil {
			break
		}
		line := map[string]interface{}{}
		line["id"] = id
		line["username"] = username.String
		line["player_id"] = player_id
		line["won_money"] = won_money
		line["lost_money"] = lost_money
		line["win"] = win
		line["lose"] = lose
		line["draw"] = draw
		lines = append(lines, line)
	}

	if err := rows.Err(); err != nil {
		return data, nil
	}

	data["lines"] = lines
	return data, nil
}
