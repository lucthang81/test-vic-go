package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/go-martini/martini"

	"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/gamemini/taixiu"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
)

var CashoutMutex sync.Mutex
var TaixiuFakeName string

func init() {
	chosenBotnameIndex := rand.Intn(len(player.BotUsernames))
	TaixiuFakeName = player.BotUsernames[chosenBotnameIndex]
}

// ¯\\_(ツ)_/¯
// "ERROR: no error occurred"
func (models *Models) HandleVicManager(r *martini.ClassicMartini) {
	r.Get("/test", func() string { return "hihi ngon" })
	r.Group("/users", func(r martini.Router) {
		r.Group("/:uid", func(r martini.Router) {
			r.Post("/change_money", models.changeUserMoneyHandle)
			r.Post("/ban", models.banUserHandle)
			r.Post("/unban", models.unbanUserHandle)
			r.Post("/change_money2", models.changeUserMoney2Handle)
			r.Post("/set_new_password", models.setUserPasswordHandle)
		})
	})
	r.Post("/cashout", models.cashoutHandle)
	r.Post("/decline_cashout", models.declineCashoutHandle)
	r.Post("/cashout88", models.cashout88Handle)
	r.Post("/decline_cashout88", models.declineCashout88Handle)

	r.Post("/change_promoted_rate", models.changePromotedRateHandle)
	r.Post("/create_event_top", models.createEventTop)
	r.Post("/create_event_sc", models.createEventSC)
	r.Post("/set_result_lucky_number", models.setResultLuckyNumberHandle)
	r.Post("/create_event_cp", models.createEventCollectingPiecesHandle)
	r.Post("/event_cp_change_n_limit_prizes", models.eventCpChangeNLimitPrizes)
	r.Post("/taixiu_fake_chat", models.TaixiuFakeChat)
	r.Post("/taixiu_fake_chat_change_name", models.TaixiuFakeChatChangeName)

	r.Post("/pop-up", models.CreatePopUp)
	r.Post("/pop-up-all", models.CreatePopUpAll)
}

func (models *Models) changeUserMoneyHandle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	playerIdS := params["uid"]
	var money_amountS, currency_type string
	if v["money_amount"] != nil {
		money_amountS = v["money_amount"][0]
	}
	if v["currency_type"] != nil {
		currency_type = v["currency_type"][0]
	}
	var admin_name string
	if v["admin_name"] != nil {
		admin_name = v["admin_name"][0]
	}

	playerId, _ := strconv.ParseInt(playerIdS, 10, 64)
	money_amount, _ := strconv.ParseInt(money_amountS, 10, 64)

	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return "ERROR: wrong pid"
	}
	var err1 error
	err1 = playerInstance.ChangeMoneyAndLog(
		money_amount, currency_type, false, "",
		record.ACTION_ADMIN_CHANGE, admin_name, "")

	if err1 == nil {
		return "{}"
	} else {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		fmt.Println("currency_type money_amount ", currency_type, money_amount)
		return s
	}
}

func (models *Models) changeUserMoney2Handle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	playerIdS := params["uid"]
	var money_amountS, currency_type string
	if v["money_amount"] != nil {
		money_amountS = v["money_amount"][0]
	}
	if v["currency_type"] != nil {
		currency_type = v["currency_type"][0]
	}
	var admin_name string
	if v["admin_name"] != nil {
		admin_name = v["admin_name"][0]
	}
	var mno, thirdPartyName, thirdPartyTransactionId string
	if v["mno"] != nil {
		mno = v["mno"][0]
	}
	if v["thirdPartyName"] != nil {
		thirdPartyName = v["thirdPartyName"][0]
	}
	if v["thirdPartyTransactionId"] != nil {
		thirdPartyTransactionId = v["thirdPartyTransactionId"][0]
	}

	playerId, _ := strconv.ParseInt(playerIdS, 10, 64)
	money_amount, _ := strconv.ParseInt(money_amountS, 10, 64)

	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return "ERROR: wrong pid"
	}
	var err1 error
	moneyBefore := playerInstance.GetMoney(currency_type)
	err1 = playerInstance.ChangeMoneyAndLog(
		money_amount, currency_type, false, "",
		record.ACTION_ADMIN_CHANGE, admin_name, "")
	moneyAfter := playerInstance.GetMoney(currency.Money)
	if money_amount > 0 {
		record.LogPurchaseRecord2(playerId, thirdPartyTransactionId, thirdPartyName,
			fmt.Sprintf("%v_%v", mno, money_amount), currency_type,
			money_amount, moneyBefore, moneyAfter, money_amount)
	} else {
		queryString := `
    		INSERT INTO cash_out_record
        		(player_id ,currency_type, change, value_before, value_after,
        		real_money_value, is_paid, verified_time)
			VALUES ($1, $2, $3, $4, $5, $6,$7, $8)`
		dataCenter.Db().Exec(queryString,
			playerId, currency_type, money_amount, moneyBefore, moneyAfter,
			money_amount, true, time.Now().UTC())
	}
	if err1 == nil {
		return "{}"
	} else {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		fmt.Println("currency_type money_amount ", currency_type, money_amount)
		return s
	}
}

func (models *Models) setUserPasswordHandle(request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	playerIdS := params["uid"]
	var newPassword string
	if v["newPassword"] != nil {
		newPassword = v["newPassword"][0]
	}
	playerId, _ := strconv.ParseInt(playerIdS, 10, 64)
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return "ERROR: wrong pid"
	}
	playerInstance.UpdatePassword(newPassword)
	return "Ngon"

}

func (models *Models) banUserHandle(request *http.Request, params martini.Params) string {
	playerIdS := params["uid"]

	playerId, _ := strconv.ParseInt(playerIdS, 10, 64)

	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return "ERROR: wrong pid"
	}
	playerInstance.SetIsBanned(true)
	return "{}"
}

func (models *Models) unbanUserHandle(request *http.Request, params martini.Params) string {
	playerIdS := params["uid"]

	playerId, _ := strconv.ParseInt(playerIdS, 10, 64)

	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return "ERROR: wrong pid"
	}
	playerInstance.SetIsBanned(false)
	return "{}"
}

func (models *Models) cashoutHandle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}

	var cashoutIdS string
	if v["cashoutId"] != nil {
		cashoutIdS = v["cashoutId"][0]
	}

	cashoutId, err1 := strconv.ParseInt(cashoutIdS, 10, 64)
	if err1 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		return s
	}

	CashoutMutex.Lock()
	defer CashoutMutex.Unlock()
	query := "UPDATE cash_out_record SET is_verified_by_admin = $1, verified_time = $2 WHERE id = $3"
	_, err = dataCenter.Db().Exec(query, true, time.Now().UTC(), cashoutId)
	if err != nil {
		fmt.Println("ERROR ERROR ERROR ERROR ERROR ", err)
	}

	err2 := acceptCashout(cashoutId)
	if err2 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err2)
		return s
	} else {
		return "{}"
	}
}

func (models *Models) declineCashoutHandle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}

	var cashoutIdS string
	var reason string
	if v["cashoutId"] != nil {
		cashoutIdS = v["cashoutId"][0]
		reason = v["reason"][0]
	}

	cashoutId, err1 := strconv.ParseInt(cashoutIdS, 10, 64)
	fmt.Println("cashoutId, reason", cashoutId, reason)
	if err1 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		return s
	}

	err2 := declineCashout(cashoutId, reason)
	if err2 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err2)
		return s
	} else {
		return "{}"
	}
}

func (models *Models) cashout88Handle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}

	var cashoutIdS string
	if v["cashoutId"] != nil {
		cashoutIdS = v["cashoutId"][0]
	}

	cashoutId, err1 := strconv.ParseInt(cashoutIdS, 10, 64)
	if err1 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		return s
	}

	CashoutMutex.Lock()
	defer CashoutMutex.Unlock()
	query := "UPDATE cash_out_paytrust88_record SET is_verified_by_admin = $1, verified_time = $2 WHERE id = $3"
	_, err = dataCenter.Db().Exec(query, true, time.Now().UTC(), cashoutId)
	if err != nil {
		fmt.Println("ERROR ERROR ERROR ERROR ERROR ", err)
	}
	///////////////////////////////////////////////////////
	row := dataCenter.Db().QueryRow(
		"SELECT player_id, bank_name, bank_account_number, amount_vnd "+
			"FROM cash_out_paytrust88_record "+
			"WHERE id = $1 AND is_paid = FALSE ", cashoutId)
	var bank_name, bank_account_number string
	var player_id int64
	var amount_vnd float64
	e := row.Scan(&player_id, &bank_name, &bank_account_number, &amount_vnd)
	fmt.Println("player_id, bank_name, bank_account_number, amount_vnd",
		player_id, bank_name, bank_account_number, amount_vnd)
	if e != nil {
		fmt.Println("ERROR cashout88Handle 1:", err)
		return err.Error()
	}

	//
	client := &http.Client{}
	temp := url.Values{}
	var http_post_url string
	if zconfig.ServerVersion == zconfig.SV_02 {
		http_post_url = "http://smsotp.slota.win/api_7749_payout"
	} else {
		http_post_url = "http://smsotp.slota.win/api_test_payout"
	}
	temp.Add("http_post_url", http_post_url)
	temp.Add("amount", fmt.Sprintf("%v", amount_vnd))
	if zconfig.ServerVersion == "" { // Test
		temp.Add("currency", "MYR")
	} else {
		temp.Add("currency", "VND")
	}
	temp.Add("item_id", fmt.Sprintf("%v", cashoutId))
	temp.Add("item_description", "item_description")
	temp.Add("name", "Jon Doe")
	temp.Add("bank_code", Paytrust88MapBankNameToBankCode[bank_name])
	temp.Add("iban", bank_account_number)
	requestUrl := "https://paytrust88.com/v1/payout/start?" + temp.Encode()
	// "Content-Type", "application/json"
	reqBodyB, err := json.Marshal(map[string]interface{}{})
	reqBody := bytes.NewBufferString(string(reqBodyB))
	req, _ := http.NewRequest("POST", requestUrl, reqBody)
	if zconfig.ServerVersion == "" { // Test
		req.Header.Set("Authorization", zconfig.AP88Test)
	} else {
		req.Header.Set("Authorization", zconfig.AP88)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// send the http request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	fmt.Println("resp body", string(body))
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	fmt.Println(utils.PFormat(data))

	return string(body)
}

func (models *Models) declineCashout88Handle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: declineCashout88Handle0 url.ParseQuery"
	}

	var cashoutIdS string
	var reason string
	if v["cashoutId"] != nil {
		cashoutIdS = v["cashoutId"][0]
		if v["reason"] != nil {
			reason = v["reason"][0]
		} else {
			reason = "Mình không thích thì mình không trả thôi"
		}
	}

	cashoutId, err1 := strconv.ParseInt(cashoutIdS, 10, 64)
	fmt.Println("cashoutId, reason", cashoutId, reason)
	if err1 != nil {
		fmt.Println("ERROR: declineCashout88Handle1:", err1)
		return err1.Error()
	}

	//
	query := "SELECT player_id, amount_kim FROM cash_out_paytrust88_record " +
		"WHERE id = $1 AND is_paid = FALSE"
	row := dataCenter.Db().QueryRow(query, cashoutId)
	var player_id sql.NullInt64
	var amount_kim sql.NullFloat64
	err = row.Scan(&player_id, &amount_kim)
	if err != nil {
		fmt.Println("ERROR: declineCashout88Handle2", err)
		return err.Error()
	}
	playerObj, err := player.GetPlayer(player_id.Int64)
	if err != nil {
		fmt.Println("ERROR: declineCashout88Handle3", err)
		return err.Error()
	}
	playerObj.CreateType2Message(
		"Đổi thưởng thất bại",
		fmt.Sprintf("Id giao dịch: %v. Số Kim hoàn lại: %v. Nguyên nhân: %v",
			cashoutId, amount_kim.Float64, reason),
	)

	query = "DELETE FROM cash_out_paytrust88_record WHERE id = $1"
	_, err = dataCenter.Db().Exec(query, cashoutId)
	if err != nil {
		fmt.Println("ERROR: declineCashout88Handle4 ", err)
		return err.Error()
	}

	playerObj.ChangeMoneyAndLog(
		int64(amount_kim.Float64), currency.Money, false, "",
		"DECLINE_CASH_OUT", "", "")

	return "{}"
}

func (models *Models) changePromotedRateHandle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}

	var promoted_rateS string
	if v["promoted_rate"] != nil {
		promoted_rateS = v["promoted_rate"][0]
	}

	promoted_rate, err1 := strconv.ParseFloat(promoted_rateS, 64)
	if err1 != nil {
		s := fmt.Sprintf("FAIL. err: %v", err1)
		return s
	}

	promotedRate = promoted_rate

	return "{}"
}

func (models *Models) createEventTop(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	//
	var event_name string
	if v["event_name"] != nil {
		event_name = v["event_name"][0]
	}
	var starting_timeS string
	if v["starting_time"] != nil {
		starting_timeS = v["starting_time"][0]
	}
	starting_time, err := time.Parse(time.RFC3339, starting_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var finishing_timeS string
	if v["finishing_time"] != nil {
		finishing_timeS = v["finishing_time"][0]
	}
	finishing_time, err := time.Parse(time.RFC3339, finishing_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var map_position_to_prizeS string
	if v["map_position_to_prize"] != nil {
		map_position_to_prizeS = v["map_position_to_prize"][0]
	}
	var map_position_to_prize map[int]int64
	err = json.Unmarshal([]byte(map_position_to_prizeS), &map_position_to_prize)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}

	//
	top.NewEventTop(event_name, starting_time, finishing_time, map_position_to_prize)
	return "{}"
}

func (models *Models) createEventSC(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	//
	var event_name string
	if v["event_name"] != nil {
		event_name = v["event_name"][0]
	}
	var starting_timeS string
	if v["starting_time"] != nil {
		starting_timeS = v["starting_time"][0]
	}
	starting_time, err := time.Parse(time.RFC3339, starting_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var finishing_timeS string
	if v["finishing_time"] != nil {
		finishing_timeS = v["finishing_time"][0]
	}
	finishing_time, err := time.Parse(time.RFC3339, finishing_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var limit_number_of_bonusS string
	if v["limit_number_of_bonus"] != nil {
		limit_number_of_bonusS = v["limit_number_of_bonus"][0]
	}
	limit_number_of_bonus, err := strconv.ParseInt(limit_number_of_bonusS, 10, 64)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var time_unitS string
	if v["time_unit"] != nil { // duration measure by time.Minute
		time_unitS = v["time_unit"][0]
	}
	time_unitI, err := strconv.ParseInt(time_unitS, 10, 64)
	var time_unit time.Duration
	if err != nil {
		time_unit = 24 * time.Hour
	} else {
		time_unit = time.Duration(time_unitI) * time.Minute
	}
	//
	top.NewEventSC(event_name, starting_time, finishing_time, limit_number_of_bonus, time_unit)
	return "{}"
}

func (models *Models) setResultLuckyNumberHandle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: setResultLuckyNumberHandle 0 " + err.Error()
	}
	//
	var lucky_numberS, validDateStr string
	if v["lucky_number"] != nil {
		lucky_numberS = v["lucky_number"][0]
	}
	if v["valid_date"] != nil {
		validDateStr = v["valid_date"][0]
	}
	if validDateStr == "" {
		validDateStr = GetDateStr(time.Now())
	}
	lucky_number, err := strconv.ParseInt(lucky_numberS, 10, 64)
	if err != nil {
		return "ERROR: setResultLuckyNumberHandle 1 " + err.Error()
	}
	SetResultLuckyNumber(lucky_number, validDateStr)
	return "{}"
}

func (models *Models) createEventCollectingPiecesHandle(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	//
	var event_name string
	if v["event_name"] != nil {
		event_name = v["event_name"][0]
	}
	var starting_timeS string
	if v["starting_time"] != nil {
		starting_timeS = v["starting_time"][0]
	}
	starting_time, err := time.Parse(time.RFC3339, starting_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var finishing_timeS string
	if v["finishing_time"] != nil {
		finishing_timeS = v["finishing_time"][0]
	}
	finishing_time, err := time.Parse(time.RFC3339, finishing_timeS)
	if err != nil {
		s := fmt.Sprintf("FAIL. err: %v", err)
		return s
	}
	var n_pieces_to_completeS, n_limit_prizeS string
	var total_prizeS, chance_to_drop_rare_pieceS string
	if v["n_pieces_to_complete"] != nil {
		n_pieces_to_completeS = v["n_pieces_to_complete"][0]
	}
	if v["n_limit_prize"] != nil {
		n_limit_prizeS = v["n_limit_prize"][0]
	}
	if v["total_prize"] != nil {
		total_prizeS = v["total_prize"][0]
	}
	if v["chance_to_drop_rare_piece"] != nil {
		chance_to_drop_rare_pieceS = v["chance_to_drop_rare_piece"][0]
	}
	n_pieces_to_complete, _ := strconv.Atoi(n_pieces_to_completeS)
	n_limit_prize, _ := strconv.Atoi(n_limit_prizeS)
	total_prize, _ := strconv.ParseInt(total_prizeS, 10, 64)
	chance_to_drop_rare_piece, _ := strconv.ParseFloat(chance_to_drop_rare_pieceS, 64)
	//
	event_player.NewEventCollectingPieces(
		event_name, starting_time, finishing_time,
		n_pieces_to_complete, n_limit_prize,
		total_prize, chance_to_drop_rare_piece)
	return "{}"
}

func (models *Models) eventCpChangeNLimitPrizes(request *http.Request) string {
	body, _ := ioutil.ReadAll(request.Body)
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	//
	var event_name string
	if v["event_name"] != nil {
		event_name = v["event_name"][0]
	}

	var n_limit_prizeS string
	if v["n_limit_prize"] != nil {
		n_limit_prizeS = v["n_limit_prize"][0]
	}
	n_limit_prize, _ := strconv.Atoi(n_limit_prizeS)
	//
	event_player.GlobalMutex.Lock()
	event := event_player.MapEvents[event_name]
	event_player.GlobalMutex.Unlock()
	if event != nil {
		event.ChangeNLimitPrizes(n_limit_prize)
	}
	return "{}"
}

func (models *Models) TaixiuFakeChat(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err.Error()
	}
	senderName := TaixiuFakeName
	message := utils.GetStringAtPath(data, "message")
	gameInstance := models.GetGameMini(taixiu.TAIXIU_GAME_CODE, currency.Money)
	if gameInstance == nil {
		return "wtf"
	}
	taixiuGame, isOk := gameInstance.(*taixiu.TaixiuGame)
	if !isOk {
		return "wtf2"
	}
	err = taixiuGame.Chat(nil, message, senderName)
	if err != nil {
		return err.Error()
	}
	//
	responseData := map[string]interface{}{}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) TaixiuFakeChatChangeName(
	request *http.Request, params martini.Params) string {
	chosenBotnameIndex := rand.Intn(len(player.BotUsernames))
	TaixiuFakeName = player.BotUsernames[chosenBotnameIndex]
	//
	responseData := map[string]interface{}{}
	bytes, err := json.MarshalIndent(responseData, "", "    ")
	if err != nil {
		s := fmt.Sprintf("ERROR: json.MarshalIndent %v", err)
		return s
	}
	return string(bytes)
}

func (models *Models) CreatePopUp(request *http.Request, w http.ResponseWriter) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return ""
	}
	targetUserId := utils.GetInt64AtPath(data, "targetUserId")
	msg := utils.GetStringAtPath(data, "msg")
	pObj, err := models.GetPlayer(targetUserId)
	if pObj == nil {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		return ""
	}
	pObj.CreatePopUp(msg)
	return "{}"
}

func (models *Models) CreatePopUpAll(request *http.Request, w http.ResponseWriter) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return ""
	}
	msg := utils.GetStringAtPath(data, "msg")
	server.SendRequestsToAll("PopUp", map[string]interface{}{"msg": msg})
	return "{}"
}
