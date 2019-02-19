package models

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

func init() {
	_, _ = json.Marshal([]int{})
	_ = time.Now()
	fmt.Print("")
	_, _ = strconv.ParseInt("20081992", 10, 64)
	_ = record.ACTION_ADMIN_CHANGE
	_ = utils.GetBoolAtPath
}

func (models *Models) HandleIPN(r *martini.ClassicMartini) {
	r.Post("/appotapay", models.appotapayIpnHandle)

	r.Get("/paytrust88/success", func() string { return "success" })
	r.Get("/paytrust88/fail", func() string { return "fail" })
	r.Post("/paytrust88/status", models.paytrust88IpnHandle)
	r.Post("/paytrust88/cashout_status", models.paytrust88CashoutHandle)

	r.Post("/loangngoang/charge", models.loangngoangChargeHandle)

	r.Get("/luckico", models.luckicoCheck)
}

func (models *Models) appotapayIpnHandle(request *http.Request, params martini.Params) string {
	fmt.Println("appotapayIpnHandle2")
	body, _ := ioutil.ReadAll(request.Body)
	fmt.Println("appotapayIpnHandle body", string(body))
	//body status=1&sandbox=1&transaction_id=AP17111120916B&
	//developer_trans_id=23&transaction_type=BANK&type=ATM&state=&target=&
	//amount=10000&currency=VND&country_code=VN&hash=38e34e462847512b56c4edf2ee41aca1
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "ERROR: url.ParseQuery"
	}
	var transaction_type, status, transaction_id, amount, developer_trans_id string
	if v["transaction_type"] != nil {
		transaction_type = v["transaction_type"][0]
	}
	if v["status"] != nil {
		status = v["status"][0]
	}
	if v["transaction_id"] != nil {
		transaction_id = v["transaction_id"][0]
	}
	if v["amount"] != nil {
		amount = v["amount"][0]
	}
	if v["developer_trans_id"] != nil {
		developer_trans_id = v["developer_trans_id"][0]
	}
	// for auth hash
	var country_code, currency_, sandbox, state, target, type_, hash, vendor string
	if v["country_code"] != nil {
		country_code = v["country_code"][0]
	}
	if v["currency"] != nil {
		currency_ = v["currency"][0]
	}
	if v["sandbox"] != nil {
		sandbox = v["sandbox"][0]
	}
	if v["state"] != nil {
		state = v["state"][0]
	}
	if v["target"] != nil {
		target = v["target"][0]
	}
	if v["type"] != nil {
		type_ = v["type"][0]
	}
	if v["hash"] != nil {
		hash = v["hash"][0]
	}
	if v["vendor"] != nil {
		vendor = v["vendor"][0]
	}
	//
	if transaction_type == "BANK" {
		temp := amount + country_code + currency_ + developer_trans_id +
			sandbox + state + status + target + transaction_id + transaction_type +
			type_ + API_SECRET_APPOTAPAY
		myHash := fmt.Sprintf("%x", md5.Sum([]byte(temp)))
		if myHash != hash {
			s := "Request is not from appota"
			fmt.Println(s)
			return s
		}
	} else {
		_ = vendor
	}
	//
	if status == "1" {
		// giao dịch thành công
		if transaction_type == "CARD" {
			return "OK"
		} else if transaction_type == "BANK" {
			vicTranId, _ := strconv.ParseInt(developer_trans_id, 10, 64)
			record.LogTransactionIdRefererId(vicTranId, transaction_id)
			query := "SELECT player_id FROM purchase_referer WHERE id=$1"
			row := dataCenter.Db().QueryRow(query, vicTranId)
			var pid int64
			err := row.Scan(&pid)
			if err != nil {
				return "error1"
			}
			playerInstance, _ := models.GetPlayer(pid)
			if playerInstance == nil {
				return "error2"
			}
			moneyBefore := playerInstance.GetMoney(currency.Money)
			cardValue, _ := strconv.ParseInt(amount, 10, 64)
			moneyAfter, _ := playerInstance.IncreaseMoney(cardValue, currency.Money, true)
			_ = playerInstance.IncreaseVipScore(cardValue / 100)
			playerInstance.CreateRawMessage("Nạp tiền qua ngân hàng thành công",
				fmt.Sprintf("Bạn đã nạp thành công %v Kim qua ngân hàng.", amount),
			)
			record.LogTransactionIdRefererId(vicTranId, transaction_id)
			record.LogPurchaseRecord(playerInstance.Id(),
				transaction_id, "appotaPayBank", fmt.Sprintf("%v", cardValue),
				currency.Money, cardValue, moneyBefore, moneyAfter)
			// promotion
			if promotedRate > 0 {
				promotedMoney := int64(promotedRate * float64(cardValue))
				playerInstance.ChangeMoneyAndLog(
					promotedMoney, currency.Money, false, "",
					record.ACTION_PROMOTE, "", "")
				playerInstance.CreateRawMessage(
					fmt.Sprintf("Khuyến Mãi %.0f%% Nạp", 100*promotedRate),
					fmt.Sprintf("Chúc mừng bạn đã nhận được %v Kim "+
						"tương đương %.0f%% giá trị nạp.", promotedMoney, 100*promotedRate),
				)
			}
			//
			server.SendRequest("purchase_money",
				map[string]interface{}{"amount": cardValue},
				pid)
			return "OK"
		} else {
			return "error-2"
		}
	} else {
		return "error-1"
	}
}

func (models *Models) paytrust88IpnHandle(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	fmt.Println("paytrust88IpnHandle header", request.Header)
	fmt.Println("paytrust88IpnHandle body", string(body))
	// appotapayIpnHandle header map[Content-Type:[application/x-www-form-urlencoded] Connection:[close] User-Agent:[Python-urllib/2.7] Accept-Encoding:[identity] Content-Length:[478]]
	exBody := map[string]interface{}{
		"status":           "1",
		"bank_name":        "Test Bank",
		"account":          "977",
		"apikey":           "357",
		"name":             "Jon Doe",
		"created_at":       "2018-03-18 19:53:42 Asia/Kuala_Lumpur",
		"telephone":        "",
		"contract":         "323",
		"currency":         "VND",
		"amount":           "17.16",
		"transaction":      "244728",
		"bank_account":     "00000000000",
		"signature":        "e5a25eddda57c97d129a9bec5623468c9afd8f5b5169fbfe6fa6cb4bd4b4312b",
		"item_id":          "item_id",
		"status_message":   "Accepted",
		"email":            "user@tmt.com",
		"item_description": "item_description"}
	_ = exBody
	var data map[string]string
	e := json.Unmarshal(body, &data)
	if e != nil {
		return e.Error()
	}
	var player_id int64
	var amount_vnd, amount_myr, amount_kim, kim_before, kim_after float64
	var paytrust88_data string
	paytrust88_data = string(body)
	t1 := strings.Index(data["email"], "@")
	if t1 != -1 {
		player_idS := data["email"][0:t1]
		player_id, _ = strconv.ParseInt(player_idS, 10, 64)
	}
	// amount_myr, _ =
	amount_vnd, _ = strconv.ParseFloat(data["amount"], 64)
	amount_kim = amount_vnd * RATE_BANK_CHARGING

	pObj, _ := models.GetPlayer(player_id)
	if pObj != nil {
		kim_before = float64(pObj.GetMoney(currency.Money))
		if data["status"] == "1" && data["currency"] == "VND" { // Accepted
			pObj.ChangeMoneyAndLog(
				int64(amount_kim), currency.Money, false, "",
				record.ACTION_PAYTRUST88_CHARGING, "", "")
			pObj.CreateRawMessage("Chuyển khoản thành công",
				fmt.Sprintf("Bạn được cộng %v Kim", amount_kim))
			server.SendRequest("ChargingBankSuccess",
				map[string]interface{}{
					"amount_vnd": amount_vnd, "amount_kim": amount_kim},
				pObj.Id())
		}
		kim_after = float64(pObj.GetMoney(currency.Money))
		// agency bonus
		m, _ := getAgencies(true, pObj.Id())
		if len(m) >= 1 {
			bonusValue := int64(0.05 * amount_kim)
			pObj.ChangeMoneyAndLog(
				bonusValue, currency.Money, false, "",
				record.ACTION_PROMOTE_AGENCY, "", "")
			pObj.CreateRawMessage("Khuyến mại",
				fmt.Sprintf("Bạn được Khuyến mại %v Kim", bonusValue))
		}
	}
	record.LogBankCharging(player_id, amount_vnd, amount_myr, amount_kim,
		kim_before, kim_after, paytrust88_data)
	return "{}"
}

func (models *Models) paytrust88CashoutHandle(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	fmt.Println("paytrust88CashoutHandle header", request.Header)
	fmt.Println("paytrust88CashoutHandle body", string(body))

	//
	exBody := map[string]interface{}{
		"status":           "1",
		"bank_name":        "Test Bank",
		"account":          "977",
		"apikey":           "357",
		"name":             "Jon Doe",
		"created_at":       "2018-03-18 19:53:42 Asia/Kuala_Lumpur",
		"telephone":        "",
		"contract":         "323",
		"currency":         "MYR",
		"amount":           "17.16",
		"transaction":      "244728",
		"bank_account":     "00000000000",
		"signature":        "e5a25eddda57c97d129a9bec5623468c9afd8f5b5169fbfe6fa6cb4bd4b4312b",
		"item_id":          "item_id",
		"status_message":   "Accepted",
		"email":            "user@tmt.com",
		"item_description": "item_description"}
	_ = exBody

	var data map[string]string
	e := json.Unmarshal(body, &data)
	if e != nil {
		return e.Error()
	}

	cashoutIdS := data["item_id"]
	cashoutId, _ := strconv.ParseInt(cashoutIdS, 10, 64)
	_, _ = dataCenter.Db().Exec(
		"UPDATE cash_out_paytrust88_record "+
			"SET is_paid = TRUE, paytrust88_data = $1 WHERE id = $2",
		string(body), cashoutId)

	return "{}"
}

func (models *Models) loangngoangChargeHandle(
	request *http.Request, params martini.Params) string {
	body, _ := ioutil.ReadAll(request.Body)
	var data map[string]interface{}
	e := json.Unmarshal(body, &data)
	if e != nil {
		return e.Error()
	}
	exBody := map[string]interface{}{
		"pid":         123,
		"id_giaodich": "1",
		"real_amount": 10,
		"amount":      12,
	}
	_ = exBody
	playerId := utils.GetInt64AtPath(data, "pid")
	id_giaodich := utils.GetStringAtPath(data, "id_giaodich")
	LangngoangMapLock.Lock()
	i, isIn := LangngoangMapIdgdToInfo[id_giaodich]
	LangngoangMapLock.Unlock()
	if i == nil || !isIn {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": "id_giaodich không tồn tại hoặc đã quá lâu"}, playerId)
		return "id_giaodich không tồn tại hoặc đã quá lâu"
	}
	row := dataCenter.Db().QueryRow(
		"SELECT id FROM purchase_record WHERE transaction_id = $1",
		id_giaodich)
	var idhihi int64
	e1 := row.Scan(&idhihi)
	if e1 == nil {
		return "Trùng id_giaodich"
	}
	step3Data, step3Err := handleChargeResult(
		i.playerInstance, i.refererId, i.purchaseType, i.vendor,
		id_giaodich,
		fmt.Sprintf("%v", utils.GetInt64AtPath(data, "real_amount")),
		fmt.Sprintf("%v", utils.GetInt64AtPath(data, "amount")),
		nil)
	if step3Err != nil {
		return step3Err.Error()
	} else {
		bytes, _ := json.MarshalIndent(step3Data, "", "    ")
		return string(bytes)
	}
}

func (models *Models) luckicoCheck(
	request *http.Request, params martini.Params) string {
	key := request.URL.Query().Get("key")
	userId := record.RedisLoadFloat64(key)
	if userId == 0 {
		return "not_found"
	}
	return fmt.Sprintf("%v", userId)

}
