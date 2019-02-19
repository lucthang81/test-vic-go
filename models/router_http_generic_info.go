package models

import (
	"database/sql"
	"encoding/json"
	//	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-martini/martini"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slot"
	"github.com/vic/vic_go/models/gamemini/slotacp"
	"github.com/vic/vic_go/models/gamemini/slotagm"
	"github.com/vic/vic_go/models/gamemini/slotax1to5"
	"github.com/vic/vic_go/models/gamemini/slotbacay"
	"github.com/vic/vic_go/models/gamemini/slotpoker"
	"github.com/vic/vic_go/models/gamemini/taixiu"
	"github.com/vic/vic_go/models/gamemini/taixiu2"
	"github.com/vic/vic_go/models/otp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/quarantine"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
)

const (
	MAX_N_MESSAGES_TO_ADMIN            = int64(5) // without reply
	DURATION_RESET_N_MESSAGES_TO_ADMIN = 5 * 24 * time.Hour
)

func init() {
	fmt.Printf("")
	_, _ = strconv.ParseInt("5", 10, 64)
	_ = utils.GetInt64AtPath
}

func (models *Models) GetUserInfoRouter(r *martini.ClassicMartini) {
	r.Get("/test", func() string { return "hihi ngon" })
	r.Group("/test1", func(r martini.Router) {
		r.Group("/:uid", func(r martini.Router) {
			r.Post("/change_money")
		})
	})
	r.Get("/crossdomain.xml", models.CrossdomainTextHandle)
	r.Get("/CreateCaptcha", models.CreateCaptchaHandle)

	r.Post("/login", models.loginUserHandle)
	r.Post("/register", models.registerUserHandle)
	r.Post("/register2", models.registerUser2Handle)
	r.Get("/users/:uid", models.getUserDetail)

	r.Get("/get_jackpot_data", models.GetJackpotsInfoHandle)
	r.Get("/get_game_list", models.GetGamesInfoHandle)
	r.Get("/ClientGetLast5GlobalTexts", models.GetGlobalTextsHandle)
	r.Get("/TaixiuGetInfo", models.TaixiuGetInfoHandle)
	r.Get("/BaucuaGetInfo", models.BaucuaGetInfoHandle)
	r.Get("/EventGetList", models.EventGetListHandle)
	r.Post("/EventTopGetLeaderBoard", models.EventTopGetLeaderBoardHandle)
	r.Get("/GetFailPasswordBlockingDurationInSeconds",
		models.GetFailPasswordBlockingDurationInSecondsHandle)
	r.Get("/GetResetPasswordSms", models.GetResetPasswordSmsHandle)
	r.Get("/GetSmsChargingMoneyRate", models.GetSmsChargingMoneyRateHandle)

	// chat voi admin
	r.Post("/UserSendMsgToAdmin", models.UserSendMsgToAdmin)
	r.Post("/AdminSendMsgToUser", models.AdminSendMsgToUser)
	r.Post("/GetMsgsForUser", models.GetMsgsForUser)
	r.Post("/GetNewMsgUserIds", models.GetNewMsgUserIds)
	r.Post("/MarkMsgAsRead", models.MarkMsgAsRead)

	r.Get("/GetPushOfflineData", models.GetPushOfflineData)
	r.Get("/GetCcu", models.GetCcu)
	r.Get("/GetNHuman", models.GetNHuman)
	r.Post("/SaveString", models.SaveString)
	r.Post("/LoadString", models.LoadString)
	r.Get("/GetShopItems", models.GetShopItems)
	r.Get("/GetExchangeChargingBonusRate", models.GetExchangeChargingBonusRate)
	r.Get("/GetSlotsMoneyPerLine", models.GetSlotsMoneyPerLine)
	r.Get("/GetBanks", models.GetBanks)
	r.Get("/GetSellingAgencies", models.GetSellingAgencies)
	r.Get("/GetBuyingAgencies", models.GetBuyingAgencies)
	r.Post("/ChangePhone", models.ChangePhone)
}

func (models *Models) CrossdomainTextHandle(request *http.Request, params martini.Params) string {
	return `<?xml version="1.0" ?>
<cross-domain-policy> 
  <site-control permitted-cross-domain-policies="master-only"/>
  <allow-access-from domain="*"/>
  <allow-http-request-headers-from domain="*" headers="*"/>
</cross-domain-policy>`
}

func (models *Models) CreateCaptchaHandle(
	request *http.Request, params martini.Params) string {
	aCaptcha := captcha.CreateCaptcha()
	data := map[string]interface{}{
		"CaptchaId": aCaptcha.CaptchaId,
		"PngImage":  aCaptcha.PngImage,
	}
	bytes, _ := json.MarshalIndent(data, "", "    ")
	return string(bytes)
}

func (models *Models) loginUserHandle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	var username, password string
	var facebookAppId, facebookUserId, accessToken string
	if v["username"] != nil {
		username = v["username"][0]
	}
	if v["password"] != nil {
		password = v["password"][0]
	}
	if v["facebookAppId"] != nil {
		facebookAppId = v["facebookAppId"][0]
	}
	if v["accessToken"] != nil {
		accessToken = v["accessToken"][0]
	}
	if v["facebookUserId"] != nil {
		facebookUserId = v["facebookUserId"][0]
	}
	//
	result := map[string]interface{}{}
	if quarantine.IsQuarantine(username, "account") {
		result = map[string]interface{}{
			"error":   "Bạn đăng nhập sai nhiều quá",
			"message": "Tài khoản của bạn đã bị khoá trong 1 tiếng",
		}
	} else {
		var playerInstance *player.Player
		var err error
		if facebookAppId == "" {
			playerInstance, err = player.AuthenticateOldPlayerByPassword(username, password, "", "")
		} else {
			var hasCreatedNewPlayer bool
			hasCreatedNewPlayer, playerInstance, err =
				player.AuthenticatePlayerByFacebook(
					accessToken, facebookUserId, username, "", "", "", facebookAppId)
			_ = hasCreatedNewPlayer
		}
		if err != nil {
			result = map[string]interface{}{
				"error":   err.Error(),
				"message": "Sai tên tài khoản hoặc password",
			}
		} else {
			quarantine.ResetFailAttempt(username, "account")
			result = map[string]interface{}{
				"player_id":    playerInstance.Id(),
				"token":        playerInstance.Token(),
				"identifier":   playerInstance.Identifier(),
				"username":     playerInstance.Username(),
				"display_name": playerInstance.DisplayName(),
				"avatar_url":   playerInstance.AvatarUrl(),
				"money":        playerInstance.GetMoney("money"),
				"test_money":   playerInstance.GetMoney("test_money"),
				"is_verify":    playerInstance.IsVerify(),
			}
		}
	}

	//
	bytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func register(username string, password string) string {
	result := map[string]interface{}{}
	playerInstance, err := player.GenerateNewPlayer2(username, password, "", "", username)
	if err != nil {
		result = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		result = map[string]interface{}{
			"player_id":    playerInstance.Id(),
			"token":        playerInstance.Token(),
			"username":     playerInstance.Username(),
			"display_name": playerInstance.DisplayName(),
			"avatar_url":   playerInstance.AvatarUrl(),
			"money":        playerInstance.GetMoney("money"),
			"test_money":   playerInstance.GetMoney("test_money"),
			"identifier":   playerInstance.Identifier(),
		}
	}

	bytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) registerUserHandle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	var username, password string
	if v["username"] != nil {
		username = v["username"][0]
	}
	if v["password"] != nil {
		password = v["password"][0]
	}
	return register(username, password)
}

func (models *Models) registerUser2Handle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	var username, password, captchaId, digits string
	if v["username"] != nil {
		username = v["username"][0]
	}
	if v["password"] != nil {
		password = v["password"][0]
	}
	if v["captchaId"] != nil {
		captchaId = v["captchaId"][0]
	}
	if v["digits"] != nil {
		digits = v["digits"][0]
	}
	vr := captcha.VerifyCaptcha(captchaId, digits)
	// fmt.Printf("vr, captchaId, digits %v | %v | %v\n", vr, captchaId, digits)
	if vr == false {
		result := map[string]interface{}{
			"error": l.Get(l.M0066),
		}
		bytes, _ := json.MarshalIndent(result, "", "    ")
		return string(bytes)
	}
	return register(username, password)
}

func (models *Models) GetJackpotsInfoHandle(request *http.Request, params martini.Params) string {
	responseData, _ := getJackpotData(models, map[string]interface{}{}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	// fmt.Println("GetJackpotsInfoHandle body:")
	// fmt.Println(string(bytes))
	return string(bytes)
}

func (models *Models) GetGamesInfoHandle(request *http.Request, params martini.Params) string {
	responseData, _ := getGameList(models, map[string]interface{}{}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetGlobalTextsHandle(request *http.Request, params martini.Params) string {
	responseData, _ := ClientGetLast5GlobalTexts(models, map[string]interface{}{}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) EventGetListHandle(request *http.Request, params martini.Params) string {
	responseData, _ := EventGetList(models, map[string]interface{}{}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetPushOfflineData(request *http.Request, params martini.Params) string {
	responseData, _ := GetPushOfflineData(models, map[string]interface{}{}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) EventTopGetLeaderBoardHandle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	// fmt.Printf("header %+v\n", request.Header)
	// fmt.Println("body", string(body))
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	eventName := utils.GetStringAtPath(data, "eventName")
	//
	responseData, _ := EventTopGetLeaderBoard(models, map[string]interface{}{
		"eventName": eventName}, 0)
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) TaixiuGetInfoHandle(request *http.Request) string {
	gameCode := "taixiu"
	currencyType := currency.Money
	gameMiniInterface := models.GetGameMini(gameCode, currencyType)
	if gameMiniInterface == nil {
		return "ERROR: TaixiuGetInfoHandle 1"
	}
	taixiuG, isOk := gameMiniInterface.(*taixiu.TaixiuGame)
	if !isOk {
		return "ERROR: TaixiuGetInfoHandle 2"
	}
	if taixiuG.SharedMatch == nil {
		return "ERROR: need_to_wait_new_match_start"
	} else {
		temp := taixiuG.SharedMatch.SerializedData()
		//
		bytes, err := json.MarshalIndent(temp, "", "    ")
		if err != nil {
			s := fmt.Sprintf("ERROR: TaixiuGetInfoHandle 4 %v", err)
			return s
		}
		return string(bytes)
	}
}

func (models *Models) BaucuaGetInfoHandle(request *http.Request) string {
	gameCode := "baucua"
	currencyType := currency.Money
	gameMiniInterface := models.GetGameMini(gameCode, currencyType)
	if gameMiniInterface == nil {
		return "ERROR: BaucuaGetInfoHandle 1"
	}
	taixiuG, isOk := gameMiniInterface.(*baucua.TaixiuGame)
	if !isOk {
		return "ERROR: BaucuaGetInfoHandle 2"
	}
	if taixiuG.SharedMatch == nil {
		return "ERROR: Baucua need_to_wait_new_match_start"
	} else {
		temp := taixiuG.SharedMatch.SerializedData()
		//
		bytes, err := json.MarshalIndent(temp, "", "    ")
		if err != nil {
			s := fmt.Sprintf("ERROR: BaucuaGetInfoHandle 4 %v", err)
			return s
		}
		return string(bytes)
	}
}

func (models *Models) UserSendMsgToAdmin(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	player_id := utils.GetInt64AtPath(data, "player_id")
	message := utils.GetStringAtPath(data, "message")
	return userSendMsgToAdmin(player_id, message)
}

func (models *Models) AdminSendMsgToUser(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	player_id := utils.GetInt64AtPath(data, "player_id")
	is_from_user := false
	has_read := false
	message := utils.GetStringAtPath(data, "message")
	// for check user can send msg
	dataCenter.Db().Exec(
		"INSERT INTO message_with_admin_counter "+
			"(player_id, n_remaining_message, last_time) "+
			"VALUES ($1, $2, $3)",
		player_id, MAX_N_MESSAGES_TO_ADMIN, time.Now().UTC())
	dataCenter.Db().Exec(
		"UPDATE message_with_admin_counter "+
			"SET n_remaining_message=$1, last_time= $2"+
			"WHERE player_id=$3",
		MAX_N_MESSAGES_TO_ADMIN, time.Now().UTC(), player_id)
	//
	query := "INSERT INTO message_with_admin " +
		"(player_id, is_from_user, has_read, message) " +
		"VALUES ($1, $2, $3, $4) " +
		"RETURNING id "
	row := dataCenter.Db().QueryRow(
		query, player_id, is_from_user, has_read, message)
	var msgId int64
	err = row.Scan(&msgId)
	if err != nil {
		return "ERROR QueryRow: " + err.Error()
	}
	responseData := map[string]interface{}{"msgId": msgId}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}
func (models *Models) GetMsgsForUser(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	player_id := utils.GetInt64AtPath(data, "player_id")
	limit := utils.GetIntAtPath(data, "limit")
	offset := utils.GetIntAtPath(data, "offset")
	return getMsgsForUser(player_id, limit, offset)
}
func (models *Models) GetNewMsgUserIds(request *http.Request, params martini.Params) string {
	//	body, _ := ioutil.ReadAll(request.Body)
	//	var data map[string]interface{}
	//	err := json.Unmarshal(body, &data)
	//	if err != nil {
	//		return err.Error()
	//	}

	query := "SELECT DISTINCT player_id" +
		"FROM message_with_admin " +
		"WHERE has_read=false AND is_from_user=true " +
		"ORDER BY player_id "

	rows, err := dataCenter.Db().Query(query)
	if err != nil {
		return "ERROR: Query " + err.Error()
	}
	defer rows.Close()

	pids := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid int64
		err = rows.Scan(&pid)
		if err != nil {
			return "ERROR: Scan " + err.Error()
		}
		r := map[string]interface{}{
			"player_id": pid,
		}
		pids = append(pids, r)
	}
	responseData := map[string]interface{}{"pids": pids}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}
func (models *Models) MarkMsgAsRead(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	msgId := utils.GetInt64AtPath(data, "msgId")
	return markMsgAsRead(msgId)
}

// return json or error msg
func userSendMsgToAdmin(player_id int64, message string) string {
	is_from_user := true
	has_read := false
	var err error
	// check can user send msg
	r := dataCenter.Db().QueryRow(
		"SELECT n_remaining_message, last_time "+
			"FROM message_with_admin_counter WHERE player_id=$1 ",
		player_id)
	var n_remaining_message int64
	var last_time time.Time
	err = r.Scan(&n_remaining_message, &last_time)
	if err != nil {
		// no rows in result set
		dataCenter.Db().Exec(
			"INSERT INTO message_with_admin_counter "+
				"(player_id, n_remaining_message, last_time) "+
				"VALUES ($1, $2, $3)",
			player_id, MAX_N_MESSAGES_TO_ADMIN-1, time.Now().UTC())
	} else {
		if time.Now().UTC().Sub(last_time) >= DURATION_RESET_N_MESSAGES_TO_ADMIN {
			dataCenter.Db().Exec(
				"UPDATE message_with_admin_counter "+
					"SET n_remaining_message=$1, last_time= $2"+
					"WHERE player_id=$3",
				MAX_N_MESSAGES_TO_ADMIN-1, time.Now().UTC(), player_id)
		} else if n_remaining_message > 0 {
			dataCenter.Db().Exec(
				"UPDATE message_with_admin_counter "+
					"SET n_remaining_message=$1, last_time= $2"+
					"WHERE player_id=$3",
				n_remaining_message-1, time.Now().UTC(), player_id)
		} else {
			return "Bạn cần chờ admin trả lời tin nhắn"
		}
	}
	// send msg to admin
	query := "INSERT INTO message_with_admin " +
		"(player_id, is_from_user, has_read, message) " +
		"VALUES ($1, $2, $3, $4) " +
		"RETURNING id "
	row := dataCenter.Db().QueryRow(
		query, player_id, is_from_user, has_read, message)
	var msgId int64
	err = row.Scan(&msgId)
	if err != nil {
		return "ERROR QueryRow: " + err.Error()
	}
	responseData := map[string]interface{}{"msgId": msgId}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func getMsgsForUser(player_id int64, limit int, offset int) string {
	if limit == 0 {
		limit = 50
	}
	//
	query := "SELECT id, is_from_user, has_read, message, created_at " +
		"FROM message_with_admin " +
		"WHERE player_id=$1 ORDER BY created_at DESC " +
		"LIMIT $2 OFFSET $3"
	rows, err := dataCenter.Db().Query(query, player_id, limit, offset)
	if err != nil {
		return "ERROR: Query " + err.Error()
	}
	defer rows.Close()

	msgs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var is_from_user, has_read bool
		var message string
		var created_at time.Time
		err = rows.Scan(&id, &is_from_user, &has_read, &message, &created_at)
		if err != nil {
			return "ERROR: Scan " + err.Error()
		}
		msg := map[string]interface{}{
			"id":           id,
			"is_from_user": is_from_user,
			"has_read":     has_read,
			"message":      message,
			"created_at":   created_at.Local(),
		}
		msgs = append(msgs, msg)
	}
	responseData := map[string]interface{}{"msgs": msgs}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func markMsgAsRead(msgId int64) string {
	var err error
	query := "UPDATE message_with_admin " +
		"SET has_read = true " +
		"WHERE id=$1 "
	_, err = dataCenter.Db().Exec(query, msgId)
	if err != nil {
		return err.Error()
	}
	return "{}"
}

func (models *Models) GetFailPasswordBlockingDurationInSecondsHandle(
	request *http.Request, params martini.Params) string {
	username := request.URL.Query().Get("username")
	account := quarantine.GetQuarantineAdminAccount(username, "account")
	var blockingDurInSecs float64
	if account == nil {
		blockingDurInSecs = 0
	} else {
		blockingDurInSecs = account.EndDate().Sub(time.Now()).Seconds()
		if blockingDurInSecs < 0 {
			blockingDurInSecs = 0
		}
	}
	responseData := map[string]interface{}{
		"blockingDurInSecs": blockingDurInSecs,
	}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetResetPasswordSmsHandle(
	request *http.Request, params martini.Params) string {
	username := request.URL.Query().Get("username")
	username = strings.Trim(strings.ToLower(username), " ")

	var responseData map[string]interface{}
	var phoneNS sql.NullString
	row := dataCenter.Db().QueryRow(
		"SELECT phone_number FROM player WHERE username=$1",
		username,
	)
	e := row.Scan(&phoneNS)
	phone := phoneNS.String
	if e != nil {
		responseData = map[string]interface{}{
			"info":     "Sai username",
			"debugErr": e.Error(),
		}
	} else {
		if len(phone) < 5 {
			responseData = map[string]interface{}{
				"info": "Tài khoản chưa được kích hoạt bởi số điện thoại, " +
					fmt.Sprintf("liên hệ %v để được hỗ trợ.", zconfig.CustomerServicePhone),
			}
		} else {
			var sms string
			dauviettel := ",8496,8497,8498,84162,84163,84165,84166,84167,84168,84169,8486,"
			if strings.Index(dauviettel, phone[0:5]) >= 0 ||
				strings.Index(dauviettel, phone[0:4]) >= 0 {
				sms = otp.VIETTEL_SMS_RESET_PASSWORD + username
			} else {

				sms = otp.NON_VIETTEL_SMS_RESET_PASSWORD + username
			}
			smsNumber := 9029
			responseData = map[string]interface{}{
				"info": fmt.Sprintf(
					"Bạn hãy nhắn tin %v đến %v từ số điện thoại %vxxx để đặt lại mật khẩu",
					sms, smsNumber, phone[0:len(phone)-3]),
				"sms":       sms,
				"smsNumber": smsNumber,
			}
		}
	}

	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetSmsChargingMoneyRateHandle(
	request *http.Request, params martini.Params) string {
	responseData := map[string]interface{}{
		"SmsChargingMoneyRate": otp.SmsChargingMoneyRate,
	}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

// return json string
func (models *Models) GetCcu(
	request *http.Request, params martini.Params) string {
	responseData := map[string]interface{}{
		"ccu": models.getNumOnlinePlayers(),
	}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

// return json string
func (models *Models) GetNHuman(
	request *http.Request, params martini.Params) string {
	responseData := map[string]interface{}{
		"ccu": models.GetNHumanOnline(),
	}
	//
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) SaveString(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	key := utils.GetStringAtPath(data, "key")
	value := utils.GetStringAtPath(data, "value")
	record.RedisSaveString(key, value)
	//
	responseData := map[string]interface{}{}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) LoadString(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	key := utils.GetStringAtPath(data, "key")
	//
	value := record.RedisLoadString(key)
	responseData := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetShopItems(
	request *http.Request, params martini.Params) string {
	//

	//
	responseData := getShopItems()
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func getShopItems() map[int64]map[string]interface{} {
	shopItems := make(map[int64]map[string]interface{})
	rows, err := dataCenter.Db().Query(
		"SELECT id, name, price , discount_rate, url, addditional_data, created_time " +
			"FROM shop_item",
	)
	if err != nil {
		return shopItems
	}
	defer rows.Close()
	for rows.Next() {
		var id, price int64
		var name, url, addditional_data string
		var discount_rate float64
		var created_time time.Time
		rows.Scan(&id, &name, &price, &discount_rate, &url,
			&addditional_data, &created_time)
		shopItems[id] = map[string]interface{}{
			"id":               id,
			"name":             name,
			"price":            price,
			"discount_rate":    discount_rate,
			"url":              url,
			"addditional_data": addditional_data,
			"created_time":     created_time,
		}
	}
	return shopItems
}

func (models *Models) GetExchangeChargingBonusRate(
	request *http.Request, params martini.Params) string {
	//
	responseData := map[string]interface{}{
		"exchangeChargingBonusRate": exchangeChargingBonusRate,
	}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetSlotsMoneyPerLine(
	request *http.Request, params martini.Params) string {
	//
	responseData := map[string]interface{}{
		slot.SLOT_GAME_CODE:              slot.MONEYS_PER_LINE,
		slotacp.SLOTACP_GAME_CODE:        slotacp.MONEYS_PER_LINE,
		slotagoldminer.SLOTAGM_GAME_CODE: slotagoldminer.MONEYS_PER_LINE,
		slotax1to5.SLOTAX1TO5_GAME_CODE:  slotax1to5.MONEYS_PER_LINE,
		slotbacay.SLOTBACAY_GAME_CODE:    slotbacay.MONEYS_PER_LINE,
		slotpoker.SLOTPOKER_GAME_CODE:    slotpoker.MONEYS_PER_LINE,
	}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetBanks() string {
	bytes, err := json.MarshalIndent(Paytrust88MapBankNameToBankCode, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) GetSellingAgencies() string {
	r, e := getAgencies(true, 0)
	if e != nil {
		return e.Error()
	}
	bytes, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}
func (models *Models) GetBuyingAgencies() string {
	r, e := getAgencies(false, 0)
	if e != nil {
		return e.Error()
	}
	bytes, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}
func (models *Models) getUserDetail(request *http.Request, params martini.Params) string {
	userId, _ := strconv.ParseInt(params["uid"], 10, 64)
	var result map[string]interface{}
	playerInstance, err := player.GetPlayer(userId)
	if playerInstance == nil {
		if err != nil {
			result = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			result = map[string]interface{}{
				"error": "playerInstance == nil",
			}
		}
	}
	result = map[string]interface{}{
		"player_id":    playerInstance.Id(),
		"token":        playerInstance.Token(),
		"identifier":   playerInstance.Identifier(),
		"username":     playerInstance.Username(),
		"display_name": playerInstance.DisplayName(),
		"avatar_url":   playerInstance.AvatarUrl(),
		"money":        playerInstance.GetMoney("money"),
		"test_money":   playerInstance.GetMoney("test_money"),
		"is_verify":    playerInstance.IsVerify(),
	}
	bytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) ChangePhone(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	username := utils.GetStringAtPath(data, "Username")
	phone := utils.GetStringAtPath(data, "NewPhone")
	pObj := player.FindPlayerWithUsername(username)
	if pObj == nil {
		return "pObj == nil"
	}
	err = pObj.UpdatePhoneNumber2(phone)
	if err != nil {
		return err.Error()
	}
	responseData := map[string]interface{}{"IsSuccessful": true}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}
