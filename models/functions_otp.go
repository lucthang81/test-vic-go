package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/otp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
	"github.com/vic/vic_go/zglobal"
)

func sendOtpCode(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	otpCode := utils.GetStringAtPath(data, "otp_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	responseData, err = otp.VerifyOtpCode(player.Id(), otpCode)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func registerVerifyPhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = otp.RegisterVerifyPhoneNumber(playerId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func hkResetPassword(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	username := utils.GetStringAtPath(data, "username")
	phone := utils.GetStringAtPath(data, "phone")
	newpasswd := utils.GetStringAtPath(data, "new_password")

	//err = otp.HKRegisterResetPassword(username,phone,newpasswd)
	fmt1, err1 := otp.HKRegisterResetPassword(username, phone, newpasswd)

	responseData = make(map[string]interface{})
	//responseData["phone"] = phone
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	//responseData["id"] = player.Id()
	return responseData, nil
}

func GetLeaderBoard(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")         // ma game = "all" neu muon tat ca
	currencyType := utils.GetStringAtPath(data, "currency_type") // loai tien //money va test_money
	statisticType := utils.GetStringAtPath(data, "type")         // ="week" ="month" ="year" ="date"
	page := utils.GetInt64AtPath(data, "page")                   // =1 la page 1; = 2 la page 2 ....

	responseData, err = otp.GetLeaderBoard(gameCode, currencyType, page, statisticType)
	//responseData= map[string]interface{data}

	return responseData, nil
}

// ham chua viet can viet
func cashOutBank(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	bankName := utils.GetStringAtPath(data, "bank_name")
	branchName := utils.GetStringAtPath(data, "branch_name")
	fullName := utils.GetStringAtPath(data, "full_name")
	bankNumber := utils.GetStringAtPath(data, "bank_number")
	value := utils.GetInt64AtPath(data, "value")
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	//err = otp.HKRegisterVerifyPhoneNumber(playerId,phone)
	fmt1, err1 := otp.HKCashOutBank(playerId, bankName, branchName, fullName, bankNumber, value)

	responseData = make(map[string]interface{})
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	responseData["id"] = player.Id()
	return responseData, nil
}

// ham chua viet can viet

func cashOutCard(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	cardType := utils.GetStringAtPath(data, "card_type")
	value := utils.GetInt64AtPath(data, "value")
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return nil, errors.New(l.Get(l.M0065))
	}

	fmt1, err1 := otp.HKCashOutCard(playerId, cardType, value)

	responseData = make(map[string]interface{})
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	responseData["id"] = player.Id()
	return responseData, nil
}

func CashOutBank88(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	bankName := utils.GetStringAtPath(data, "bankName")
	bankAccountNumber := utils.GetStringAtPath(data, "bankAccountNumber")
	amountVND := utils.GetFloat64AtPath(data, "amount")
	//
	amountKim := amountVND * zglobal.CashOutRate
	_, _, _ = bankName, amountVND, bankAccountNumber
	pObj, _ := models.GetPlayer(playerId)
	if pObj == nil {
		return nil, errors.New("pObj == nil")
	}
	if zconfig.ServerVersion != "" && amountVND < 1200000 {
		return nil, errors.New("Bạn cần rút ít nhất 1200000 VND")
	}
	if _, isIn := Paytrust88MapBankNameToBankCode[bankName]; !isIn {
		return nil, errors.New("Sai tên ngân hàng")
	}
	//
	kimBefore := pObj.GetMoney(currency.Money)
	pObj.LockMoney(currency.Money)
	if int64(amountKim) > pObj.GetMoney(currency.Money) {
		pObj.UnlockMoney(currency.Money)
		return nil, errors.New("Bạn không đủ Kim")
	}
	kimAfter, _ := pObj.DecreaseMoney(int64(amountKim), currency.Money, false)
	pObj.UnlockMoney(currency.Money)
	//
	record.LogCurrencyRecord(
		playerId, "cashout", "",
		map[string]interface{}{},
		currency.Money, kimBefore, kimAfter, -int64(amountKim),
	)
	_, e := dataCenter.Db().Exec(
		"INSERT INTO cash_out_paytrust88_record "+
			"(player_id, amount_vnd, amount_kim, kim_before, kim_after, "+
			"    bank_name, bank_account_number) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) ",
		pObj.Id(), amountVND, amountKim, kimBefore, kimAfter,
		bankName, bankAccountNumber)
	if e != nil {
		fmt.Println("ERROR: ", e)
	}
	msg := fmt.Sprintf("Giao dịch thành công! Bạn đã bị trừ %v KIM. Chúng tôi sẽ chuyển khoản cho bạn trong vòng 24h.", amountKim)
	pObj.CreateRawMessage("Giao dịch Kim", msg)
	return nil, nil
}

//"hk_verify_otp":                   hkVerifyOTP,
//		"hk_send_phone":                   hkSendPhone,
func hkSendPhone(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	phone := utils.GetStringAtPath(data, "phone")
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	fmt1, err1 := otp.HKSendOTP(playerId, phone)

	responseData = make(map[string]interface{})
	responseData["phone"] = phone
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	responseData["id"] = player.Id()
	return responseData, nil
}
func hkVerifyOTP(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	playerInstance, err := models.GetPlayer(playerId)
	otpCode := utils.GetStringAtPath(data, "otp")
	phone := utils.GetStringAtPath(data, "phone")
	if err != nil {
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}
	if playerInstance == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}
	phone = utils.NormalizePhoneNumber(phone)
	responseData, err1 := otp.HKVerifyOtpCode(phone, otpCode)
	if err1 != nil {
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}
	return responseData, nil

}
func hkRegisterVerifyPhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	phone := utils.GetStringAtPath(data, "phone")
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	fmt1, err1 := otp.HKRegisterVerifyPhoneNumber(playerId, phone)

	responseData = make(map[string]interface{})
	responseData["phone"] = phone
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	responseData["id"] = player.Id()
	return responseData, nil
}
func hkRegisterChangePhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}
	new_phone := utils.GetStringAtPath(data, "new_phone")

	fmt1, err1 := otp.HKRegisterChangePhoneNumber(player.Id(), new_phone)

	responseData = make(map[string]interface{})
	//responseData["phone"] = phone
	responseData["msg"] = err1.Error()
	responseData["fmt"] = fmt1
	//responseData["id"] = player.Id()
	return responseData, nil
}

func registerChangePhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = otp.RegisterChangePhoneNumber(playerId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func registerChangePassword(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("sendOtpCode player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = otp.RegisterChangePassword(playerId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (models *Models) registerResetPassword2(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	phoneNumber := utils.NormalizePhoneNumber(
		utils.GetStringAtPath(data, "phone_number"))
	username := strings.Trim(
		strings.ToLower(utils.GetStringAtPath(data, "username")), " ")

	playerInstance := player.FindPlayerWithPhoneNumber(phoneNumber)
	if playerInstance == nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Số điện thoại này chưa được xác nhận",
			"error_code": 0,
		})
		return
	}
	if playerInstance.IsVerify() == false {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Số điện thoai chưa được xác thực",
			"error_code": 0,
		})
		return
	}
	if strings.Trim(strings.ToLower(playerInstance.Username()), " ") != username {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Không tìm thấy người dùng trong hệ thống",
			"error_code": 0,
		})
		return
	}

	cost := int64(1500)
	if playerInstance.GetAvailableMoney(currency.Money) < cost {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Bạn không đủ 1500 kim",
			"error_code": 0,
		})
		return
	}
	passwd := utils.RandSeqLowercase(6)

	//send otpCodeString to phone via messageBird
	//	client := &http.Client{}
	//	requestUrl := fmt.Sprintf(
	//		"http://smsotp.slota.win/sms?whoareyou=slota.win"+
	//			"&mobileno=%v&senderid=pw%v.%v&language=vn&otp=%v",
	//		phoneNumber, phoneNumber, passwd, passwd,
	//	)
	//	// "Content-Type", "application/json"
	//	reqBodyB, err := json.Marshal(map[string]interface{}{})
	//	reqBody := bytes.NewBufferString(string(reqBodyB))
	//	req, _ := http.NewRequest("GET", requestUrl, reqBody)
	//	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//	// send the http request
	//	resp, err := client.Do(req)
	err = SendSms(phoneNumber, passwd)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Bạn không đủ 1500 kim",
			"error_code": 0,
		})
		return
	}
	//	if resp.StatusCode != 200 {
	//		renderer.JSON(200, map[string]interface{}{
	//			"message":    "Hệ thống bận! Vui lòng thử lại sau",
	//			"error_code": 0,
	//		})
	//		return
	//	}
	playerInstance.UpdatePassword(passwd)
	playerInstance.ChangeMoneyAndLog(-cost, currency.Money, false, "", "", "", "")
}
func (models *Models) registerResetPassword(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	phoneNumber := utils.GetStringAtPath(data, "phone_number")
	err = otp.RegisterResetPassword(phoneNumber)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	renderer.JSON(200, map[string]interface{}{})
}

func (models *Models) verifyOtpCode(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	otpCode := utils.GetStringAtPath(data, "otp_code")
	phoneNumber := utils.NormalizePhoneNumber(utils.GetStringAtPath(data, "phone_number"))

	playerInstance := player.FindPlayerWithPhoneNumber(phoneNumber)
	if playerInstance == nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Số điện thoại này chưa được xác nhận",
			"error_code": 0,
		})
		return
	}

	responseData, err := otp.VerifyOtpCode(playerInstance.Id(), otpCode)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	renderer.JSON(200, responseData)
}

func (models *Models) hkVerifyOtpCode(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	otpCode := utils.GetStringAtPath(data, "otp_code")
	phoneNumber := utils.NormalizePhoneNumber(utils.GetStringAtPath(data, "phone_number"))

	_, err = otp.HKVerifyOtpCode(phoneNumber, otpCode)
	renderer.JSON(200, map[string]interface{}{
		"message":    err.Error(),
		"error_code": 0,
	})
}

/*
otp page
*/

func (models *Models) getOtpPage(c martini.Context, adminAccount *AdminAccount,
	request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "OTP", "/admin/otp")
	data["nav_links"] = navLinks
	data["page_title"] = "OTP"

	renderer.HTML(200, "admin/otp", data)
}

func (models *Models) getOtpRewardPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	if page < 1 {
		page = 1
	}

	data, err := otp.GetOtpRewardData(keyword, page)

	if err != nil {
		renderError(renderer, err, "/admin/otp", adminAccount)
		return
	}
	data["keyword"] = keyword
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "OTP", "/admin/otp")
	navLinks = appendCurrentNavLink(navLinks, "OTP reward", "/admin/otp/reward")
	data["nav_links"] = navLinks
	data["page_title"] = "OTP Reward"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/otp_reward_list", data)
}

func (models *Models) getOtpCodePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	if page < 1 {
		page = 1
	}

	data, err := otp.GetOtpCodeData(keyword, page)

	if err != nil {
		renderError(renderer, err, "/admin/otp", adminAccount)
		return
	}
	data["keyword"] = keyword
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "OTP", "/admin/otp")
	navLinks = appendCurrentNavLink(navLinks, "OTP code", "/admin/otp/code")
	data["nav_links"] = navLinks
	data["page_title"] = "OTP code"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/otp_code_list", data)
}

func (models *Models) getOtpCodeEditPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	data, editObject := otp.GetEditOtpCodeForm(id)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "OTP", "/admin/otp")
	navLinks = appendCurrentNavLink(navLinks, "OTP code", "/admin/otp/code")
	data["nav_links"] = navLinks
	data["page_title"] = "OTP code"
	data["admin_username"] = adminAccount.username
	data["form"] = editObject.GetFormHTML()
	renderer.HTML(200, "admin/otp_code_edit", data)
}

func (models *Models) editOtpCode(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)

	_, editObject := otp.GetEditOtpCodeForm(id)
	data := editObject.ConvertRequestToData(request)
	err := otp.EditOtpCode(data)

	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/otp/edit?id=%d", id), adminAccount)
		return
	}

	renderer.Redirect(fmt.Sprintf("/admin/otp/code/edit?id=%d", id))
}

/*
service callback
*/

func (models *Models) getSMSServiceCallback(c martini.Context, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	status, message := otp.HandleServiceRequest(request)
	responseData := make(map[string]interface{})
	responseData["status"] = status
	responseData["sms"] = message
	responseData["type"] = "text"

	renderer.JSON(200, responseData)
}
func ChangePhone(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	phone := utils.GetStringAtPath(data, "phone")
	player, err := models.GetPlayer(playerId)

	phone = utils.NormalizePhoneNumber(phone)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return map[string]interface{}{}, errors.New("player == nil")
	}

	if player.IsVerify() == false {
		return nil, errors.New("Bạn chưa kích hoạt tài khoản rồi")
	}
	cost := int64(1500)
	if player.GetAvailableMoney(currency.Money) < cost {
		return nil, errors.New("Bạn không đủ 1500 kim")
	}
	if phone == "" || len(phone) < 8 {
		return nil, errors.New("Số điện thoại không hợp lệ")
	}
	row1 := dataCenter.Db().QueryRow(
		"SELECT username FROM player WHERE phone_number=$1",
		phone,
	)
	var uname string
	err1 := row1.Scan(&uname)
	if err1 == nil {
		// so dien thoai da dung
		return nil, errors.New("Số điện thoại đã dùng rồi")
	}
	otpCodeString := utils.RandSeqLowercase(4)

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code "+
		"(player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phone,
		"change_phone_number3",
		otpCodeString,
		"wait",
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return nil, errors.New("Số điện thoại đã dùng rồi")
	}

	//send otpCodeString to phone via messageBird
	//	client := &http.Client{}
	//	requestUrl := fmt.Sprintf(
	//		"http://smsotp.slota.win/sms?whoareyou=slota.win"+
	//			"&mobileno=%v&senderid=%v&language=vn&otp=%v",
	//		player.PhoneNumber(), id, otpCodeString,
	//	)
	//	// "Content-Type", "application/json"
	//	reqBodyB, err := json.Marshal(map[string]interface{}{})
	//	reqBody := bytes.NewBufferString(string(reqBodyB))
	//	req, _ := http.NewRequest("GET", requestUrl, reqBody)
	//	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//	// send the http request
	//	resp, err := client.Do(req)
	//	if err != nil {
	//		return nil, err
	//	}
	err = SendSms(player.PhoneNumber(), otpCodeString)
	if err != nil {
		return nil, errors.New("Lỗi hệ thống")
	}
	player.ChangeMoneyAndLog(-cost, currency.Money, false, "", "", "", "")
	return nil, nil
}

//
func RegisterPhone(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	phone := utils.GetStringAtPath(data, "phone")
	player, err := models.GetPlayer(playerId)

	phone = utils.NormalizePhoneNumber(phone)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return map[string]interface{}{}, errors.New("player == nil")
	}

	if player.IsVerify() {
		return nil, errors.New("Bạn đã kích hoạt tài khoản rồi")
	}
	cost := int64(1500)
	if player.GetAvailableMoney(currency.Money) < cost {
		return nil, errors.New("Bạn không đủ 1500 kim")
	}
	if phone == "" || len(phone) < 8 {
		return nil, errors.New("Số điện thoại không hợp lệ")
	}
	row1 := dataCenter.Db().QueryRow(
		"SELECT username FROM player WHERE phone_number=$1",
		phone,
	)
	var uname string
	err1 := row1.Scan(&uname)
	if err1 == nil {
		// so dien thoai da dung
		return nil, errors.New("Số điện thoại đã dùng rồi")
	}
	otpCodeString := utils.RandSeqLowercase(4)

	row := dataCenter.Db().QueryRow("INSERT INTO otp_code "+
		"(player_id, phone_number, reason, otp_code, status, expired_at, created_at) "+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		playerId,
		phone,
		"change_phone_number2",
		otpCodeString,
		"wait",
		time.Now().Add(game_config.OtpExpiredAfter()).UTC(), time.Now().UTC())
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.LogSerious("err insert otp_code %v,playerId %d", err, playerId)
		return nil, errors.New("Số điện thoại đã dùng rồi")
	}

	//send otpCodeString to phone via messageBird
	client := &http.Client{}
	requestUrl := fmt.Sprintf(
		"http://smsotp.slota.win/sms?whoareyou=slota.win"+
			"&mobileno=%v&senderid=%v&language=vn&otp=%v",
		phone, id, otpCodeString,
	)
	// "Content-Type", "application/json"
	reqBodyB, err := json.Marshal(map[string]interface{}{})
	reqBody := bytes.NewBufferString(string(reqBodyB))
	req, _ := http.NewRequest("GET", requestUrl, reqBody)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// send the http request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Lỗi hệ thống")
	}
	player.ChangeMoneyAndLog(-cost, currency.Money, false, "", "", "", "")

	return nil, nil
}

//

func InputOtp(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	otp := utils.GetStringAtPath(data, "otp")
	playerInstance, err := player.GetPlayer(playerId)
	if err != nil {
		return nil, errors.New("playerInstance == nil")
	}
	otp = strings.ToLower(otp)

	query := "SELECT id, otp_code, phone_number, reason, status, retry_count, expired_at, created_at" +
		" FROM otp_code WHERE player_id = $1 and otp_code=$2 ORDER BY -id LIMIT 1"
	row := dataCenter.Db().QueryRow(query, playerId, otp)
	var id, retryCount int64
	var otpCodeString, reason, status, phoneNumber sql.NullString
	var expiredAt, createdAt time.Time
	err = row.Scan(&id, &otpCodeString, &phoneNumber, &reason, &status, &retryCount, &expiredAt, &createdAt)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, errors.New("Không tìm thấy mã otp trong hệ thống")
	}
	if reason.String == "change_phone_number2" { // dang ki so dien thoai
		playerInstance.SetPhoneNumber(phoneNumber.String)
		dataCenter.Db().Exec("UPDATE player SET phone_number=$1, is_verify=$2 "+
			"WHERE id=$3 ",
			phoneNumber.String, true, playerInstance.Id())
		err = playerInstance.SetIsVerify(true)
		if err != nil {
			fmt.Println("ERROR", err)
			return nil, errors.New("Không tìm thấy mã otp trong hệ thống")
		}
		return nil, nil
	} else if reason.String == "change_phone_number3" { // doi so dien thoai
		playerInstance.UpdatePhoneNumber2(phoneNumber.String)
		return nil, nil
	} else {
		return nil, errors.New("Lỗi hệ thống")
	}
}

func SendSms(mobileno string, message string) error {
	if zglobal.SmsSender == "onewaysms" {
		return SendSms1(mobileno, message)
	} else {
		return SendSms2(mobileno, message)
	}
}

func SendSms1(mobileno string, message string) error {
	apiusername := "APIGUYQ88OFLT"
	apipassword := "APIGUYQ88OFLTGUYQ8"
	senderid := "otpVerify" // meaningless
	languagetype := "1"     // 70 chars unicode
	mturl := "http://gateway.onewaysms.vn:10001/api.aspx"

	client := &http.Client{}
	temp := url.Values{}
	temp.Add("apiusername", apiusername)
	temp.Add("apipassword", apipassword)
	temp.Add("senderid", senderid)
	temp.Add("languagetype", languagetype)
	temp.Add("mobileno", mobileno)
	temp.Add("message", message)
	requestUrl := fmt.Sprintf("%v?%v", mturl, temp.Encode())
	reqBody := bytes.NewBufferString(string(""))
	req, e := http.NewRequest("GET", requestUrl, reqBody)
	if e != nil {
		return e
	}
	resp, e := client.Do(req)
	if e != nil {
		return e
	}
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	fmt.Println("resp header", resp.StatusCode, resp.Status, resp.Header)
	fmt.Println("resp body", string(body))
	mtId := string(body)
	mtIdInt64, e := strconv.ParseInt(mtId, 10, 64)
	if e != nil {
		return e
	}
	if mtIdInt64 < 0 {
		return errors.New("OneWaySms return error: " + mtId)
	}

	return nil
}

func SendSms2(mobileno string, message string) error {
	accessKey := "34meU1MSWWYWmItMM0KGjFMek"
	client := &http.Client{}
	requestUrl := fmt.Sprintf("https://rest.messagebird.com/messages")
	// to create body
	//	temp := url.Values{}
	//	temp.Add("originator", "84976670672")
	//	temp.Add("body", message)
	//	temp.Add("recipients", mobileno)
	args := map[string]interface{}{
		"originator": "84976670672",
		"body":       message,
		"recipients": []string{mobileno},
	}
	argsJson, e := json.Marshal(args)
	if e != nil {
		return e
	}
	reqBody := bytes.NewBufferString(string(argsJson))
	req, e := http.NewRequest("POST", requestUrl, reqBody)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "AccessKey "+accessKey)
	req.Header.Set("Accept", "application/json")
	resp, e := client.Do(req)
	if e != nil {
		return e
	}
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	fmt.Println("resp header", resp.StatusCode, resp.Status, resp.Header)
	fmt.Println("resp body", string(body))
	return nil
}
