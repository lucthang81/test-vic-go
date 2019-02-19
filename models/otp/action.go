package otp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	//"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zglobal"
)

const (
	VIETTEL_SMS_RESET_PASSWORD     = "qk 1000 ide kh rs "
	NON_VIETTEL_SMS_RESET_PASSWORD = "ip tnk nap1 ide rs "
)

var SmsChargingMoneyRate float64

func init() {
	SmsChargingMoneyRate = 0.4
}

func RegisterVerifyPhoneNumber(playerId int64) (err error) {
	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return err
	}
	if playerInstance.PhoneNumber() == "" {
		return errors.New("Bạn chưa nhập số điện thoại")
	}

	if playerInstance.IsVerify() {
		return errors.New(l.Get(l.M0091))
	}

	err = canRequestNewOtpCode(playerId)
	if err != nil {
		return err
	}

	phoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())
	otpCodeString := utils.RandSeqLowercase(4)

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"register_phone_number",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return errors.New(l.Get(l.M0003))
	}

	return nil
}

//HKCashOutBank(playerId, bankName, branchName, fullName, bankNumber, value)

func HKCashOutBank(playerId int64, bankName string, branchName string, fullName string,
	bankNumber string, value int64) (fmt1 string, err error) {

	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	if !playerInstance.IsVerify() {
		return "", errors.New("Bạn phải xác thực số điện thoại trước để chúng tôi trả thể cho bạn!")
	}
	currencyType := "money"

	before := playerInstance.GetMoney(currencyType)
	moneyGain := int64(float64(value) * zglobal.CashOutRate)
	if moneyGain > playerInstance.GetAvailableMoney(currencyType) {
		return "", errors.New(l.Get(l.M0016))
	}

	after, err1 := playerInstance.DecreaseMoney(moneyGain, currencyType, true)
	if err1 != nil {
		return "", errors.New(l.Get(l.M0016))
	}

	additionalData := map[string]interface{}{}
	additionalData["bankName"] = bankName
	additionalData["branchName"] = branchName
	additionalData["fullName"] = fullName
	additionalData["bankNumber"] = bankNumber
	additionalData["value"] = value
	additionalData["action"] = fmt.Sprintf("cob-%v-%v", playerId, utils.RandSeqLowercase(4))
	additionalData["cashOutBank"] = "cashOutBank"

	var additionalDataByte []byte
	var err2 error
	if len(additionalData) != 0 {
		additionalDataByte, err2 = json.Marshal(additionalData)
		if err2 != nil {
			log.LogSerious("Err log currency record %v %v", err2.Error(), additionalData)
			return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
		}
	}
	queryString := "INSERT INTO cash_out_record " +
		" (player_id, action, game_code, additional_data,currency_type, " +
		"change, value_before, value_after) " +
		"VALUES ($1, $2, $3, $4, $5, $6,$7,$8) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, playerId, additionalData["action"],
		additionalData["cashOutCard"], string(additionalDataByte),
		currencyType, moneyGain, before, after)
	var id int64
	err3 := row.Scan(&id)
	if err3 != nil {
		log.LogSerious("Error log currency record detail in cashOutBank playerid %v cardtype %v value %v details %v", playerId, fullName, value, err3.Error())
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}
	msg := fmt.Sprintf("Giao dịch thành công! Bạn đã bị trừ %v KIM. Chúng tôi sẽ Chuyển tiền cho bạn trong vòng 24h! Mã giao dịch: %v", moneyGain, id)

	playerInstance.CreateRawMessage("Giao dịch KIM", msg)
	return "",
		errors.New(msg)
}
func HKCashOutCard(playerId int64, cardType string, value int64) (fmt1 string, err error) {

	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	//	if !playerInstance.IsVerify() {
	//		return "", errors.New("Bạn phải xác thực số điện thoại trước để chúng tôi trả thể cho bạn!")
	//	}

	currencyType := "money"

	before := playerInstance.GetMoney(currencyType)
	moneyGain := int64(float64(value) * zglobal.CashOutRate)

	playerInstance.LockMoney(currencyType)
	if moneyGain > playerInstance.GetAvailableMoney(currencyType) {
		playerInstance.UnlockMoney(currencyType)
		return "", errors.New(l.Get(l.M0016))
	}
	after, err1 := playerInstance.DecreaseMoney(moneyGain, currencyType, false)
	playerInstance.UnlockMoney(currencyType)

	if err1 != nil {
		return "", errors.New("Có lỗi khi trừ tiền.")
	}

	record.LogCurrencyRecord(
		playerId, "cashout", "",
		map[string]interface{}{},
		currencyType, before, after, -moneyGain,
	)

	additionalData := map[string]interface{}{}
	additionalData["cardType"] = cardType
	additionalData["value"] = value
	additionalData["action"] = fmt.Sprintf("coc-%v-%v", playerId, utils.RandSeqLowercase(4))
	additionalData["cashOutCard"] = "cashOutCard"

	var additionalDataByte []byte
	var err2 error
	if len(additionalData) != 0 {
		additionalDataByte, err2 = json.Marshal(additionalData)
		if err2 != nil {
			log.LogSerious("Err log currency record %v %v", err2.Error(), additionalData)
			return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
		}
	}
	queryString := "INSERT INTO cash_out_record (player_id, action, game_code, additional_data,currency_type, change, value_before, value_after, real_money_value) " +
		"VALUES ($1, $2, $3, $4, $5, $6,$7,$8, $9) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, playerId, additionalData["action"],
		additionalData["cashOutCard"], string(additionalDataByte),
		currencyType, moneyGain, before, after, value)
	var id int64
	err3 := row.Scan(&id)
	if err3 != nil {
		log.LogSerious("Error log currency record detail in cashOut playerid %v cardtype %v value %v details %v", playerId, cardType, value, err3.Error())
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	msg := fmt.Sprintf("Giao dịch thành công! Bạn đã bị trừ %v KIM. Chúng tôi sẽ trả thẻ cào cho bạn trong vòng 24h! Mã giao dịch: %v", moneyGain, id)
	playerInstance.CreateRawMessage("Giao dịch KIM", msg)
	return "", errors.New(msg)
}
func HKSendOTP(playerId int64, phone string) (fmt1 string, err error) {

	phoneNumber := utils.NormalizePhoneNumber(phone)
	lenPhone := len(phoneNumber)
	if lenPhone != 11 && lenPhone != 12 {
		return "", errors.New("Số điện thoại không hợp lệ!")
	}

	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	//phoneNumberphoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())
	otp := hkCanRequestNewOtpCode(playerId)
	//otpCodeString := ""
	if otp == "" {
		otp = utils.RandSeqLowercase(4)
	}
	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"register_phone_number",
		otp,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}
	var enoughMoney int64
	enoughMoney = 300 // số kim phải trừ để kích hoạt tài khoản
	currencyType := "money"
	playerInstance.LockMoney(currencyType)
	if playerInstance.GetMoney(currencyType) < enoughMoney {
		enoughMoney = 0
	} else {
		playerInstance.DecreaseMoney(enoughMoney, currencyType, false)
	}
	playerInstance.UnlockMoney(currencyType)

	if enoughMoney == 0 {
		return "", errors.New("Tài khoản KIM không đủ. Bạn phải có ít nhất 300 KIM để kích hoạt!")
	}

	// send http get de lay ve otp cho user nhap
	url := fmt.Sprintf("http://sms.xoac.us/sms?phone=%s&body=%s", phoneNumber, otp)
	resp, err1 := http.Get(url) //, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	if err1 != nil {
		resp, err1 = http.Get(url) //, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
		if err1 != nil {
			return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
		}
	}
	defer resp.Body.Close()

	body, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}
	s := string(body)
	if s != "1" {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}
	return "", nil
}

func HKRegisterVerifyPhoneNumber(playerId int64, phone string) (
	fmt1 string, err error) {

	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	phoneNumber := utils.NormalizePhoneNumber(phone)
	if phoneNumber == "" || len(phoneNumber) < 8 {
		return "", errors.New("Bạn chưa nhập số điện thoại")
	}
	row1 := dataCenter.Db().QueryRow(
		"SELECT username FROM player WHERE phone_number=$1",
		phoneNumber,
	)
	var uname string
	err1 := row1.Scan(&uname)
	if err1 != nil {
		// so dien thoai chua dung
	} else {
		return "", errors.New("Số điện thoại đã dùng rồi")
	}
	if playerInstance.IsVerify() {
		return "", errors.New(l.Get(l.M0091))
	}

	// get format tin nhan otp
	formatSMS := "sms:9029?body=IP TNK NAP1 IDE kh tk %s"
	formatTB := "Để xác thực bạn vui lòng gửi tin nhắn<br/> <font color='#FF0000'> IP TNK NAP1 IDE kh tk %s</font><br/>đến 9029"

	dauviettel := ",8496,8497,8498,84162,84163,84165,84166,84167,84168,84169,8486,"
	if strings.Index(dauviettel, phoneNumber[0:5]) >= 0 ||
		strings.Index(dauviettel, phoneNumber[0:4]) >= 0 {
		formatSMS = "sms:9029?body=QK 1000 IDE kh tk %s"
		formatTB = "Để xác thực bạn vui lòng gửi tin nhắn<br/> <font color='#FF0000'>QK 1000 IDE kh tk %s</font><br/>đến 9029"
	}
	otp := strings.ToLower(hkCanRequestNewOtpCode(playerId))
	if otp != "" {
		return fmt.Sprintf(formatSMS, otp),
			errors.New(fmt.Sprintf(formatTB, otp))

	}

	//phoneNumberphoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())
	otpCodeString := strings.ToLower(utils.RandSeqLowercase(4))

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"register_phone_number",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return "", errors.New("Hệ thông đang nâng cấp chức năng này!")
	}

	return fmt.Sprintf(formatSMS, otpCodeString),
		errors.New(fmt.Sprintf(formatTB, otpCodeString))
}

func RegisterChangePhoneNumber(playerId int64) (err error) {
	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return err
	}
	if playerInstance.PhoneNumber() == "" {
		return errors.New("Bạn chưa nhập số điện thoại")
	}

	if !playerInstance.IsVerify() {
		return errors.New("Bạn chưa xác nhận số điện thoại hiện tại")
	}

	err = canRequestNewOtpCode(playerId)
	if err != nil {
		return err
	}

	otpCodeString := utils.RandSeqLowercase(4)
	phoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"change_phone_number",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return errors.New(l.Get(l.M0003))
	}

	return nil
}

//(playerId int64, phone string)
func HKRegisterChangePhoneNumber(playerId int64, phone string) (ftm string, err error) {
	//yêu cầu người dùng phải đăng nhập rồi nhập vào số điện thoại cần đổi và
	//yêu cầu người này nhắn tin từ số điện thoại này
	// yêu cầu tài khoản phải kích hoạt rồi mới đợợc đổi pass

	_, err = player.GetPlayer(playerId)
	if err != nil {
		return "", errors.New("User không tồn tại")
	}
	phoneNumber := utils.NormalizePhoneNumber(phone)
	if phoneNumber == "" {
		return "", errors.New("Bạn chưa nhập số điện thoại")
	}

	formatSMS := "sms:9029?body=IP TNK NAP1 IDE kh %s"
	formatTB := "Để thay số điện thoại bạn cần dùng số  %s nhắn: <br/> <font color='#FF0000'>IP TNK NAP1 IDE kh %s</font><br />đến 9029"

	dauviettel := ",8496,8497,8498,84162,84163,84165,84166,84167,84168,84169,8486,"
	if strings.Index(dauviettel, phoneNumber[0:5]) >= 0 ||
		strings.Index(dauviettel, phoneNumber[0:4]) >= 0 {
		formatSMS = "sms:9029?body=QK 1000 IDE kh tk %s"
		formatTB = "Để thay số điện thoại bạn cần dùng số  %s nhắn<br/> <font color='#FF0000'>QK 1000 IDE kh tk %s</font><br />đến 9029"
	}
	otp := strings.ToLower(hkCanRequestNewOtpCode(playerId))
	if otp != "" {
		return fmt.Sprintf(formatSMS, otp),
			errors.New(fmt.Sprintf(formatTB, phoneNumber, otp))
	}

	otpCodeString := strings.ToLower(utils.RandSeqLowercase(4))
	//phoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"change_phone_number",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return "", errors.New("Hệ thông hiện đang nâng cấp tính năng này!")
	}

	return fmt.Sprintf(formatSMS, otpCodeString),
		errors.New(fmt.Sprintf(formatTB, phoneNumber, otpCodeString))
}
func RegisterChangePassword(playerId int64) (err error) {
	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return err
	}
	if playerInstance.PhoneNumber() == "" {
		return errors.New("Bạn chưa nhập số điện thoại")
	}

	if !playerInstance.IsVerify() {
		return errors.New("Bạn chưa xác nhận số điện thoại hiện tại")
	}

	err = canRequestNewOtpCode(playerId)
	if err != nil {
		return err
	}

	otpCodeString := utils.RandSeqLowercase(4)
	phoneNumber := utils.NormalizePhoneNumber(playerInstance.PhoneNumber())

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phoneNumber,
		"change_password",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return errors.New(l.Get(l.M0003))
	}

	return nil
}

// ( username string, phoneNumber string,newPassWD string) (err error)
func HKRegisterResetPassword(username string, phoneNumber string, newPassWD string) (s1 string, err error) {
	// nhập số phone , và password lưu tạm
	otpPhone, err := dataCenter.GetOtpPhone()
	if err != nil {
		log.LogSerious("err insert otp_code %v,phoneNumber %v", err, phoneNumber)
		return "", errors.New("Hệ thông hiện đang nâng cấp tính năng này!")
	}
	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)
	fmt.Println(phoneNumber)
	playerInstance := player.FindPlayerWithPhoneNumber(phoneNumber)
	if playerInstance == nil {
		return "", errors.New("Số điện thoại này không có trong hệ thống")
	}

	if !playerInstance.IsVerify() {
		return "", errors.New("Bạn chưa xác nhận số điện thoại này")
	}
	if strings.ToLower(playerInstance.Username()) != strings.ToLower(username) {
		return "", errors.New("Tên đăng nhập và số điện thoại không khớp!")
	}
	otp := hkCanRequestNewOtpCode(playerInstance.Id())
	if otp != "" {
		return fmt.Sprintf("sms:%s?body=%s", otpPhone, otp),
			errors.New(fmt.Sprintf("Để đặt lại mật khẩu bạn vui lòng gửi tin nhắn %s đến số điện thoại %s!", otp, otpPhone))
	}
	playerInstance.SetNewPassWD(newPassWD)
	//player.newPasswd = newPassWD
	otpCodeString := utils.RandSeqLowercase(4)

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at,passwd) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		playerInstance.Id(),
		phoneNumber,
		"reset_password",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC(),
		newPassWD)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerInstance.Id())
		return "", errors.New("Hệ thông hiện đang nâng cấp tính năng này!")
	}

	return fmt.Sprintf("sms:%s?body=%s", otpPhone, otpCodeString),
		errors.New(fmt.Sprintf("Để đặt lại mật khẩu bạn vui lòng gửi tin nhắn %s đến số điện thoại %s!", otpCodeString, otpPhone))
}

func RegisterResetPassword(phoneNumber string) (err error) {
	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)
	fmt.Println(phoneNumber)
	playerInstance := player.FindPlayerWithPhoneNumber(phoneNumber)
	if playerInstance == nil {
		return errors.New("Số điện thoại này không có trong hệ thống")
	}

	if !playerInstance.IsVerify() {
		return errors.New("Bạn chưa xác nhận số điện thoại này")
	}

	err = canRequestNewOtpCode(playerInstance.Id())
	if err != nil {
		return err
	}

	otpCodeString := utils.RandSeqLowercase(4)

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code (player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerInstance.Id(),
		phoneNumber,
		"reset_password",
		otpCodeString,
		kStatusWaitingForUser,
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerInstance.Id())
		return errors.New(l.Get(l.M0003))
	}

	return nil
}

// all sms method
func HKVerifyOtpCode(phone string, otpCodeString1 string) (
	data map[string]interface{}, err error) {
	// duoc goi boi he thong sms cua gate sim
	phoneNumber := utils.NormalizePhoneNumber(phone)
	otpCodeString := strings.ToLower(otpCodeString1)
	if strings.Index(otpCodeString, VIETTEL_SMS_RESET_PASSWORD) >= 0 ||
		strings.Index(otpCodeString, NON_VIETTEL_SMS_RESET_PASSWORD) >= 0 {
		pos := strings.LastIndex(otpCodeString, " ")
		username := strings.ToLower(strings.Trim(otpCodeString[pos:], " "))
		//phoneNumber

		row := dataCenter.Db().QueryRow(
			"SELECT phone_number, id FROM player WHERE username=$1", username)
		verifyPhoneNumber := ""
		pId := int64(0)
		e := row.Scan(&verifyPhoneNumber, &pId)
		if e != nil {
			return nil, errors.New("Username khong ton tai! Vui long lien he 0961744715!")
		}
		if verifyPhoneNumber != phoneNumber {
			return nil, errors.New("so dien thoai nhan tin khong hop le! Vui long lien he 0961744715!")
		}
		newPassword := strings.ToLower(utils.RandSeq(6))
		pObj, _ := player.GetPlayer(pId)
		if pObj == nil {
			encryptedNewPassword := utils.HashPassword(newPassword)
			_, err := dataCenter.Db().Exec(
				"UPDATE player SET password=$1 WHERE username=$2",
				encryptedNewPassword, username)
			if err != nil {
				return nil, err
			} else {

			}
		} else {
			pObj.UpdatePassword(newPassword)
		}
		// reset
		return nil, errors.New(fmt.Sprintf("MK moi %v\n Vui long lien he 0961744715!", newPassword))
	} else if strings.Index(otpCodeString, "qk 1000 ide kh ") >= 0 ||
		strings.Index(otpCodeString, "ip tnk nap1 ide kh ") >= 0 {
		pos := strings.LastIndex(otpCodeString, " ")
		otpCodeString = strings.ToLower(strings.Trim(otpCodeString[pos:], " "))
		otpCode := HKGetLatestOtpCodeForPlayer(phoneNumber, otpCodeString)
		if otpCode == nil {
			return nil, errors.New("Khong tim thay ma otp trong he thong! Vui long lien he 0961744715!")
		}
		playerInstance, err := player.GetPlayer(otpCode.playerId)
		if err != nil {
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
		}
		//fmt.Printf("%v - %v \r\n", otpCode.otpCode, phoneNumber)

		if otpCode.reason == "register_phone_number" {
			err = playerInstance.SetIsVerifyWithPhone(true, phoneNumber)
			if err != nil {
				return nil, errors.New("Khong tim thay ma otp trong he thong! Vui long lien he 0961744715!")
			}
			otpCode.updateCurrentStatus(kStatusValid)
			playerInstance.ChangeMoneyAndLog(
				1000, currency.Money, false, "",
				record.ACTION_SMS_CHARGE, "", "")
			req := map[string]interface{}{}
			req["title"] = "Thông báo"
			req["content"] = fmt.Sprintf(
				"Cập nhật số điện thoại thành công, "+
					"bạn được cộng 1000 Kim. "+
					"Số điện thoại của bạn là: %v", phoneNumber)
			req["chang"] = 0
			req["currency"] = "money"
			req["id_msg"] = 01
			req["test_money"] = 0
			req["type"] = "1"
			playerInstance.SendForceRequest("force_popup", req)
			///
			return nil, errors.New(fmt.Sprintf("Thanh cong! Ban duoc cong 1000 kim vao tai khoan %s Vui long lien he 0961744715!", playerInstance.Username()))
		} else if otpCode.reason == "change_phone_number" {
			err = playerInstance.SetPhoneNumberChange(phoneNumber)
			if err != nil {
				return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
			}
			playerInstance.ChangeMoneyAndLog(
				1000, currency.Money, false, "",
				record.ACTION_SMS_CHARGE, "", "")
			req := map[string]interface{}{}
			req["title"] = "Thông báo"
			req["content"] = fmt.Sprintf("Cập nhật số điện thoại thành công! Số điện thoại của bạn là: %v", phoneNumber)
			req["chang"] = 0
			req["currency"] = "money"
			req["id_msg"] = 01
			req["test_money"] = 0
			req["type"] = "1"
			playerInstance.SendForceRequest("force_popup", req)
			otpCode.updateCurrentStatus(kStatusValid)
			return nil, errors.New(fmt.Sprintf("Thanh cong! Ban duoc cong 1000 kim vao tai khoan %s Vui long lien he 0961744715!", playerInstance.Username()))
		} else if otpCode.reason == "change_password" {
			err = playerInstance.SetPasswordChangeAvailable(true)
			if err != nil {
				return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
			}

			otpCode.updateCurrentStatus(kStatusValid)
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
		} else if otpCode.reason == "reset_password" {
			playerInstance.SetPasswordChangeAvailable(true)
			err := playerInstance.HKGenerateResetPasswordCode(otpCode.passwd)
			if err != nil {
				return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
			}

			otpCode.updateCurrentStatus(kStatusValid)

			//data = make(map[string]interface{})
			//data["reset_password_code"] = resetCode
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")
		} else {
			otpCode.updateCurrentStatus(kStatusValid)

			//data = make(map[string]interface{})
			//data["reset_password_code"] = resetCode
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")

		}
	} else {
		//cong tien SMS
		pos := strings.LastIndex(otpCodeString, " ")
		pIdS := strings.Trim(otpCodeString[pos:], " ")
		//		row := dataCenter.Db().QueryRow(
		//			"SELECT id FROM player WHERE username=$1", username)
		pId, e := strconv.ParseInt(pIdS, 10, 64)
		//		e := row.Scan(&pId)
		if e != nil {
			return nil, errors.New("Tin nhan sai cu phap! Vui long lien he 0961744715!")

		}
		pObj, e := player.GetPlayer(pId)
		if pObj == nil {
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")

		}
		tokens := strings.Split(otpCodeString, " ")
		var moneyS string
		if tokens[0] == "qk" {
			moneyS = tokens[1]
		} else {
			temp := tokens[2]
			if len(temp) < 4 {
				return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")

			}
			moneyS = temp[3:]
			moneyS += "000"
		}
		moneyValue, e := strconv.ParseInt(moneyS, 10, 64)
		if e != nil {
			return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")

		}

		inGameMoneyValue := int64(SmsChargingMoneyRate * float64(moneyValue))
		moneyBefore := pObj.GetMoney(currency.Money)
		record.LogPurchaseRecord(
			pObj.Id(), "", record.ACTION_SMS_CHARGE, "",
			currency.Money, moneyValue, moneyBefore, moneyBefore+inGameMoneyValue)
		pObj.ChangeMoneyAndLog(
			inGameMoneyValue, currency.Money, false, "",
			record.ACTION_SMS_CHARGE, "", "")

		gamemini.ServerObj.SendRequest(
			"SmsCharge",
			map[string]interface{}{"moneyValue": moneyValue},
			pObj.Id())

		return nil, errors.New(fmt.Sprintf("Thanh cong! Ban duoc cong %v kim vao tai khoan %s Vui long lien he 0961744715!", inGameMoneyValue, pObj.Username()))

	}

	//	if otpCode.isExpired() {
	//		req := map[string]interface{}{}
	//		req["title"] = "Thông báo"
	//		req["content"] = "Mã otp của bạn đã hết hiệu lực, vui lòng yêu cầu mã mới"
	//		req["chang"] = 0
	//		req["currency"] = "money"
	//		req["id_msg"] = 01
	//		req["test_money"] = 0
	//		req["type"] = "1"
	//		playerInstance.SendForceRequest("force_popup", req)
	//		return nil, errors.New("Mã otp của bạn đã hết hiệu lực, vui lòng yêu cầu mã mới")
	//	}

	//	if otpCode.status == kStatusInvalid {
	//		req := map[string]interface{}{}
	//		req["title"] = "Thông báo"
	//		req["content"] = "Mã otp của bạn đã hết hiệu lực, vui lòng yêu cầu mã mới"
	//		req["chang"] = 0
	//		req["currency"] = "money"
	//		req["id_msg"] = 01
	//		req["test_money"] = 0
	//		req["type"] = "1"
	//		playerInstance.SendForceRequest("force_popup", req)
	//		return nil, errors.New("Mã otp của bạn đã bị vô hiệu hoá, vui lòng yêu cầu mã mới")
	//	}
	/*
		if otpCode.status != kStatusAlreadySentSms {
			return nil, errors.New("Bạn chưa gửi tin nhắn. Vui lòng gửi tin nhắn theo cú pháp để nhận mã OTP")
		}
	*/

	return nil, errors.New("Khong tim thay username trong he thong! Vui long lien he 0961744715!")

}

func VerifyOtpCode(playerId int64, otpCodeString string) (data map[string]interface{}, err error) {
	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	otpCodeString = strings.ToLower(otpCodeString)

	otpCode := getLatestOtpCodeForPlayer(playerId)
	if otpCode == nil {
		return nil, errors.New("Không tìm thấy mã otp trong hệ thống")
	}

	if otpCode.isExpired() {
		return nil, errors.New("Mã otp của bạn đã hết hiệu lực, vui lòng yêu cầu mã mới")
	}

	if otpCode.status == kStatusInvalid {
		return nil, errors.New("Mã otp của bạn đã bị vô hiệu hoá, vui lòng yêu cầu mã mới")
	}

	if otpCode.status != kStatusAlreadySentSms {
		return nil, errors.New("Bạn chưa gửi tin nhắn. Vui lòng gửi tin nhắn theo cú pháp để nhận mã OTP")
	}

	if otpCode.otpCode != otpCodeString {
		otpCode.retryCount++
		retryLeft := int64(game_config.OtpCodeRetryCount()) - otpCode.retryCount
		if retryLeft == 0 {
			otpCode.updateCurrentStatus(kStatusInvalid)
			return nil, errors.New("Mã otp không chính xác và đã bị vô hiệu hoá. Vui lòng yêu cầu mã mới")
		}
		otpCode.updateCurrentRetryCount()
		return nil, errors.New(fmt.Sprintf("Mã otp không chính xác. Bạn còn %d lần nhập nữa", retryLeft))
	}

	if otpCode.reason == "register_phone_number" {
		err = playerInstance.SetIsVerify(true)
		if err != nil {
			return nil, err
		}
		otpCode.updateCurrentStatus(kStatusValid)

		if !AlreadyReceiveVerifyReward(playerInstance.PhoneNumber(), playerInstance.Id()) {
			MarkAlreadyReceiveVerifyReward(playerInstance.PhoneNumber(), playerInstance.Id(), playerInstance.Username())
			money := playerInstance.GetMoney(currency.Money)
			testMoney := playerInstance.GetMoney(currency.TestMoney)
			playerInstance.IncreaseMoney(game_config.OtpRewardMoney(), currency.Money, true)
			playerInstance.IncreaseMoney(game_config.OtpRewardTestMoney(), currency.TestMoney, true)

			record.LogPurchaseRecord(playerInstance.Id(), "otp_reward", "otp_reward", "otp_reward",
				currency.Money, game_config.OtpRewardMoney(), money, playerInstance.GetMoney(currency.Money))
			record.LogPurchaseRecord(playerInstance.Id(), "otp_reward", "otp_reward", "otp_reward",
				currency.TestMoney, game_config.OtpRewardTestMoney(), testMoney, playerInstance.GetMoney(currency.TestMoney))
		}
		return nil, nil
	} else if otpCode.reason == "change_phone_number" {
		err = playerInstance.SetPhoneNumberChangeAvailable(true)
		if err != nil {
			return nil, err
		}

		otpCode.updateCurrentStatus(kStatusValid)
		return nil, nil
	} else if otpCode.reason == "change_password" {
		err = playerInstance.SetPasswordChangeAvailable(true)
		if err != nil {
			return nil, err
		}

		otpCode.updateCurrentStatus(kStatusValid)
		return nil, nil
	} else if otpCode.reason == "reset_password" {
		playerInstance.SetPasswordChangeAvailable(true)
		resetCode, err := playerInstance.GenerateResetPasswordCode()
		if err != nil {
			return nil, err
		}

		otpCode.updateCurrentStatus(kStatusValid)

		data = make(map[string]interface{})
		data["reset_password_code"] = resetCode
		return data, nil
	}

	return nil, nil
}
