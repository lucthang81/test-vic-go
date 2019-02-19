package models

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
	"github.com/vic/vic_go/zglobal"
)

var KeyPaytrust88Lock sync.Mutex
var Paytrust88MapBankNameToBankCode map[string]string
var LangngoangMapIdgdToInfo map[string]*ChargeInfo
var LangngoangMapLock sync.Mutex

// khong duoc goi ham nap the lien tiep trong vong 2s
var MapSpam map[int64]time.Time
var MapSpamLock sync.Mutex

const (
	KEY_PAYTRUST88_COUNTER = "KEY_PAYTRUST88_COUNTER"
	LANGNGOANG_METHOD_NAME = "loangNgoangPayCardResponse"
)

func init() {
	Paytrust88MapBankNameToBankCode = map[string]string{
		"VietinBank": "5a8d9b3432bc7",
		// "VietComBank": "5a8dbfef271b0",
		"BIDV":        "5a8dc25912217",
		"TechComBank": "5a8ee643945a3",
		"SacomBank":   "5a8eec3fc74e6",
		"DongABank":   "5a904bc3775ba",
	}
	LangngoangMapIdgdToInfo = make(map[string]*ChargeInfo)
	MapSpam = make(map[int64]time.Time)
}

type ChargeInfo struct {
	playerInstance *player.Player
	refererId      int64
	purchaseType   string
	vendor         string
}

func purchaseMoney(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	isSpam := false
	MapSpamLock.Lock()
	if _, isIn := MapSpam[playerId]; isIn {
		if time.Now().Sub(MapSpam[playerId]) < 2*time.Second {
			isSpam = true
		}
	}
	MapSpam[playerId] = time.Now()
	MapSpamLock.Unlock()
	if isSpam {
		return nil, errors.New("¯\\_(ツ)_/¯")
	}

	serialCode := utils.GetStringAtPath(data, "serial_code")
	cardNumber := utils.GetStringAtPath(data, "card_number")
	// paybnb: VIETTEL, MOBIFONE, VINAPHONE, VTC,
	// appotaPayCard: viettel, mobifone, vinaphone
	// c2c: VT, MB, VP
	vendor := utils.GetStringAtPath(data, "vendor")
	purchaseType := utils.GetStringAtPath(data, "purchase_type") // paybnb
	clientIp := utils.GetStringAtPath(data, "clientIp")

	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if playerInstance == nil {
		log.LogSerious("set money player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if purchaseType == "iap" {
		version := utils.GetStringAtPath(data, "version")
		code := utils.GetStringAtPath(data, "code")
		receipt := utils.GetStringAtPath(data, "receipt")
		rate := int64(18000)
		var amount int64
		if utils.ContainsByString([]string{"goi1do"}, code) {
			amount = rate
		} else if utils.ContainsByString([]string{"goi2do"}, code) {
			amount = 2 * rate
		} else if utils.ContainsByString([]string{"goi3do"}, code) {
			amount = 3 * rate
		} else if utils.ContainsByString([]string{"goi4do"}, code) {
			amount = 4 * rate
		} else {
			amount = 0
		}

		if version == "" {
			log.LogSerious("purchase iap error did not verify %d, amount %d", playerId, amount)
			return nil, errors.New(l.Get(l.M0080))
		}

		valid, responseData, transactionId := verifyIAPPurchase(playerId, code, receipt)
		if !valid {
			return nil, errors.New(l.Get(l.M0080))
		}

		moneyBefore := playerInstance.GetMoney(currency.TestMoney)
		_, err = playerInstance.IncreaseMoney(amount, currency.TestMoney, true)
		if err != nil {
			return nil, err
		}
		moneyAfter := playerInstance.GetMoney(currency.TestMoney)
		record.LogPurchaseRecord(playerInstance.Id(),
			transactionId,
			purchaseType,
			fmt.Sprintf("%s_%d", "iap", amount),
			currency.TestMoney,
			amount,
			moneyBefore,
			moneyAfter)
		log.LogSerious("purchase iap success, playerid %d, amount %d, responsedata %v", playerInstance.Id(), amount, responseData)
		return nil, nil
	} else if purchaseType == "iapAndroid" {
		bs, _ := json.Marshal(data)
		record.PsqlSaveString(fmt.Sprintf("%v", time.Now().Unix()), string(bs))
		receipt := utils.GetStringAtPath(data, "receipt")
		pubkey := utils.GetStringAtPath(data, "pubkey")
		sig := utils.GetStringAtPath(data, "sig")
		//  receipt := '{"orderId":"GPA.xxxx-xxxx-xxxx-xxxxx","packageName":"my.package","productId":"myproduct","purchaseTime":1437564796303,"purchaseState":0,"developerPayload":"user001","purchaseToken":"some-token"}'
		// VerifySignature verifies in app billing signature.
		// You need to prepare a public key for your Android app's in app billing
		// at https://play.google.com/apps/publish/
		// pubkey := "JTbngOdvBE0rfdOs3GeuBnPB+YEP1w/peM4VJbnVz+hN9Td25vPjAznX9YKTGQN4iDohZ07wtl+zYygIcpSCc2ozNZUs9pV0s5itayQo22aT5myJrQmkp94ZSGI2npDP4+FE6ZiF+7khl3qoE0rVZq4G2mfk5LIIyTPTSA4UvyQ="
		// sig := "gj0N8LANKXOw4OhWkS1UZmDVUxM1UIP28F6bDzEp7BCqcVAe0DuDxmAY5wXdEgMRx/VM1Nl2crjogeV60OqCsbIaWqS/ZJwdP127aKR0jk8sbX36ssyYZ0DdZdBdCr1tBZ/eSW1GlGuD/CgVaxns0JaWecXakgoV7j+RF2AFbS4="

		var temp interface{}
		err = json.Unmarshal([]byte(receipt), &temp)
		if err != nil {
			fmt.Println("ERROR 1", err)
			return nil, errors.New("err json.Unmarshal")
		}
		receiptObj, isOk := temp.(map[string]interface{})
		if !isOk {
			fmt.Println("receiptObj, !isOk")
			return nil, errors.New("receiptObj, !isOk")
		}
		receiptFieldJsonData, isOk := receiptObj["jsonData"].(string)
		isValid, err := VerifySignature(pubkey, []byte(receiptFieldJsonData), sig)
		// TODO check reuse receipt
		if err != nil {
			fmt.Println("ERROR VerifySignature", err)
			return nil, errors.New(l.Get(l.M0081))
		} else if !isValid {
			fmt.Println("!isValid")
			if !record.GetIsStoreTester(playerId) {
				//				record.SetIapAndCardPay(playerId, false, true)
				//				server.SendRequest("TurnOnCardPay", map[string]interface{}{}, playerId)
			}
			return nil, errors.New("receipt is not valid")
		} else { // receipt isValid
			record.LogChangePartner(playerId, record.PARTNER_IAP_ANDROID)
			var receiptFieldJsonDataObj map[string]interface{}
			json.Unmarshal([]byte(receiptFieldJsonData), &receiptFieldJsonDataObj)
			productId, _ := receiptFieldJsonDataObj["productId"].(string)
			orderId, _ := receiptFieldJsonDataObj["orderId"].(string)
			var moneyAmount int64
			if zglobal.MapIapAndroid != nil {
				if _, isIn := zglobal.MapIapAndroid[productId]; isIn {
					moneyAmount = zglobal.MapIapAndroid[productId]
				} else {
					return nil, errors.New("Undefined productId")
				}
			} else {
				return nil, errors.New("zglobal.MapIapAndroid == nil")
			}
			// fmt.Println("receiptObj", receiptObj)
			// fmt.Printf("receiptObj[jsonData] %T %v", receiptObj["jsonData"], receiptObj["jsonData"])
			isIapOn, _ := record.GetIapAndCardPay(playerId)
			_ = isIapOn
			if true {
				e := record.LogIapAndroidCharging(playerId, orderId, receipt)
				if e != nil {
					return nil, errors.New("Duplicate receipt")
				}
				record.LogSumCharging(playerId, purchaseType, moneyAmount)
				playerInstance.ChangeMoneyAndLog(
					moneyAmount, currency.Money, false, "",
					"IAP", "", "")
				playerInstance.CreateRawMessage("Nạp tiền thành công",
					fmt.Sprintf("Bạn đã nạp thành công %v Kim qua iap.", moneyAmount))
				if record.GetIsStoreTester(playerId) {
					// record.SetIapAndCardPay(playerId, true, false)
				} else {
					record.SetIapAndCardPay(playerId, false, true)
					server.SendRequest("TurnOnCardPay", map[string]interface{}{}, playerId)
				}
			}
		}
	} else if purchaseType == "paybnb" {
		if len(cardNumber) == 0 || len(serialCode) == 0 {
			return nil, errors.New("Mã thẻ không hợp lệ")
		}

		refererId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		if refererId == 0 {
			return nil, errors.New("err:internal_error LogRefererIdForPurchase")
		}

		transactionId, cardValueStr, err := requestPayBnBPurchaseAPIForCardValue(playerId, refererId, cardNumber, serialCode, purchaseType, vendor)
		if err != nil {
			return nil, err
		}
		//fmt.Println("refererId, transactionId, cardValue: ", refererId, transactionId, cardValueStr)

		//		purchaseTypeData := money.GetPurchaseTypeData(purchaseType, cardValue)
		//		money := utils.GetInt64AtPath(purchaseTypeData, "money")
		//		fmt.Println("purchaseTypeData, money: ", purchaseTypeData, money)

		cardValue, err := strconv.ParseInt(cardValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		//
		top.GlobalMutex.Lock()
		event := top.MapEvents[top.EVENT_CHARGING_MONEY]
		top.GlobalMutex.Unlock()
		if event != nil {
			event.ChangeValue(playerId, cardValue)
		}
		event_player.GlobalMutex.Lock()
		event1 := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
		event_player.GlobalMutex.Unlock()
		if event1 != nil {
			event1.GiveAPiece(playerId, true, false, cardValue)
		}
		//

		moneyBefore := playerInstance.GetMoney(currency.Money)
		_, err = playerInstance.IncreaseMoney(cardValue, currency.Money, true)
		if err != nil {
			return nil, err
		}
		moneyAfter := playerInstance.GetMoney(currency.Money)

		err = playerInstance.IncreaseVipScore(cardValue / 100)
		if err != nil {
			return nil, err
		}
		player.CreatePurchaseMessage(playerId, serialCode, cardNumber, cardValue, playerInstance.GetMoney(currency.Money))
		record.LogTransactionIdRefererId(refererId, transactionId)
		record.LogPurchaseRecord(playerInstance.Id(),
			transactionId,
			purchaseType,
			fmt.Sprintf("%v_%v", vendor, cardValue),
			currency.Money,
			cardValue,
			moneyBefore,
			moneyAfter)

		data1 := make(map[string]interface{})
		data1["amount"] = cardValue

		// promotion
		pr := getPromotedRate(playerInstance.IsVerify())
		if pr > 0 {
			promotedMoney := int64(pr * float64(cardValue))
			playerInstance.ChangeMoneyAndLog(
				promotedMoney, currency.Money, false, "",
				record.ACTION_PROMOTE, "", "")
			playerInstance.CreateRawMessage(
				fmt.Sprintf("Khuyến Mãi %.0f%% Thẻ Nạp", 100*promotedRate),
				fmt.Sprintf("Chúc mừng bạn đã nhận được %v Kim "+
					"tương đương %.0f%% giá trị thẻ nạp.", promotedMoney, 100*promotedRate),
			)
		}
		return data1, nil
	} else if purchaseType == "appotaPayBank" {
		amount := utils.GetInt64AtPath(data, "amount")
		vicTranId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		urlToPaymentSite, err := getUrlToPaymentSiteAppotaPayBank(vicTranId, amount, clientIp)
		if err != nil {
			return nil, err
		} else {
			return map[string]interface{}{"UrlToPaymentSite": urlToPaymentSite}, nil
		}
		// continue at appotapayIpnHandle
	} else if purchaseType == "appotaPayWallet" {
		//		amount := utils.GetInt64AtPath(data, "amount")
		//		vicTranId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		//		urlToPaymentSite, err := getUrlToPaymentSiteAppotaPayWallet(vicTranId, amount, clientIp)
		//		if err != nil {
		//			return nil, err
		//		} else {
		//			return map[string]interface{}{"UrlToPaymentSite": urlToPaymentSite}, nil
		//		}
		// continue at appotapayIpnHandle
	} else if purchaseType == "appotaPayCard" ||
		purchaseType == "c2cPayCard" ||
		purchaseType == "hfcPayCard" {
		captchaId := utils.GetStringAtPath(data, "captchaId")
		digits := utils.GetStringAtPath(data, "captchaDigits")
		vr := captcha.VerifyCaptcha(captchaId, digits)
		if vr == false {
			return nil, errors.New(l.Get(l.M0066))
		}

		if len(cardNumber) == 0 || len(serialCode) == 0 {
			return nil, errors.New("Mã thẻ không hợp lệ")
		}
		refererId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		if refererId == 0 {
			return nil, errors.New("err:internal_error LogRefererIdForPurchase")
		}
		var transactionId, cardValueStr string
		var err error
		if purchaseType == "appotaPayCard" {
			transactionId, cardValueStr, err = appotaPayCard(playerId, refererId,
				cardNumber, serialCode, purchaseType, vendor)
		} else if purchaseType == "c2cPayCard" {
			transactionId, cardValueStr, err = c2cPayCard(playerId, refererId,
				cardNumber, serialCode, purchaseType, vendor)
		} else {
			transactionId, cardValueStr, err = hfcPayCard(playerId, refererId,
				cardNumber, purchaseType, vendor)
		}
		_ = data
		return handleChargeResult(
			playerInstance, refererId, purchaseType, vendor,
			transactionId, cardValueStr, cardValueStr, err)
	} else if purchaseType == "langngoangPayCard" {
		clientInputAmount := utils.GetInt64AtPath(data, "clientInputAmount")
		captchaId := utils.GetStringAtPath(data, "captchaId")
		digits := utils.GetStringAtPath(data, "captchaDigits")
		vr := captcha.VerifyCaptcha(captchaId, digits)
		if vr == false {
			return nil, errors.New(l.Get(l.M0066))
		}
		if len(cardNumber) == 0 || len(serialCode) == 0 {
			return nil, errors.New("Mã thẻ không hợp lệ")
		}
		refererId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		if refererId == 0 {
			return nil, errors.New("err:internal_error LogRefererIdForPurchase")
		}
		// step1
		requestUrl := "http://api2.godalex.com:8099/setReq" +
			fmt.Sprintf("?cardtype=%v", vendor) +
			fmt.Sprintf("&mg=%v", clientInputAmount) +
			fmt.Sprintf("&code=%v", cardNumber) +
			fmt.Sprintf("&serial=%v", serialCode) +
			fmt.Sprintf("&whoareyou=slota.win") +
			fmt.Sprintf("&pid=%v", playerId)
		resp, err := http.Get(requestUrl)
		if err != nil {
			fmt.Println("loangNgoangPayCard err", err)
			return nil, errors.New("¯\\_(ツ)_/¯ step1")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		//		fmt.Println("loangNgoangPayCard body", string(body))
		if err != nil {
			return nil, err
		}
		// ex x = {"error":0,"amount": amount,"id_giaodich":id_giaodich},
		// ex x = {"error": 5, "msg": "Yeu cau nhap captra lan nua", "src": base64(img), "id_giaodich":id_giaodich}
		var x map[string]interface{}
		err = json.Unmarshal(body, &x)
		if err != nil {
			return nil, err
		}
		record.LogTransactionIdRefererId(refererId, utils.GetStringAtPath(x, "id_giaodich"))
		resError := utils.GetIntAtPath(x, "error")
		id_giaodich := utils.GetStringAtPath(x, "id_giaodich")
		LangngoangMapLock.Lock()
		LangngoangMapIdgdToInfo[id_giaodich] = &ChargeInfo{
			playerInstance: playerInstance,
			refererId:      refererId,
			purchaseType:   purchaseType,
			vendor:         vendor,
		}
		LangngoangMapLock.Unlock()
		go func() {
			time.Sleep(1 * time.Hour)
			LangngoangMapLock.Lock()
			delete(LangngoangMapIdgdToInfo, id_giaodich)
			LangngoangMapLock.Unlock()
		}()
		switch resError {
		case 0:
			step3Data, step3Err := handleChargeResult(
				playerInstance, refererId, purchaseType, vendor,
				utils.GetStringAtPath(x, "id_giaodich"),
				fmt.Sprintf("%v", utils.GetInt64AtPath(x, "real_amount")),
				fmt.Sprintf("%v", utils.GetInt64AtPath(x, "amount")),
				nil)
			if step3Err != nil {
				step3Data["errorMsg"] = step3Err.Error()
			} else {
				step3Data["errorMsg"] = ""
			}
			server.SendRequest(LANGNGOANG_METHOD_NAME,
				step3Data, playerId)
			return nil, nil
		case 5:
			server.SendRequest(LANGNGOANG_METHOD_NAME,
				map[string]interface{}{
					"error":        resError,
					"id_giaodich":  id_giaodich,
					"captchaImage": utils.GetStringAtPath(x, "src"),
					"errorMsg":     "",
				}, playerId)
			return nil, nil
		default:
			server.SendRequest(LANGNGOANG_METHOD_NAME,
				map[string]interface{}{
					"error":    resError,
					"errorMsg": utils.GetStringAtPath(x, "msg"),
				}, playerId)
			return nil, nil
		}
	} else if purchaseType == "paytrust88Bank" {
		amountVND := utils.GetInt64AtPath(data, "amount")
		bankName := utils.GetStringAtPath(data, "bankName")

		bank_code, isIn := Paytrust88MapBankNameToBankCode[bankName]
		if !isIn {
			return nil, errors.New("Wrong bankName")
		}
		amountKim := int64(float64(amountVND) * RATE_BANK_CHARGING)
		//		amountMYR := int64(float64(amountVND) / RATE_MYR_TO_VND)
		var tnvDomain string
		if zconfig.ServerVersion == zconfig.SV_01 {
			tnvDomain = "tmt1.com"
		} else if zconfig.ServerVersion == zconfig.SV_02 {
			tnvDomain = "tmt2.com"
		} else { // sv test
			tnvDomain = "tmt.com"
		}
		KeyPaytrust88Lock.Lock()
		item_id := record.RedisLoadFloat64(KEY_PAYTRUST88_COUNTER) + 1
		record.RedisSaveFloat64(KEY_PAYTRUST88_COUNTER, item_id)
		KeyPaytrust88Lock.Unlock()
		//
		client := &http.Client{}
		temp := url.Values{}
		temp.Add("return_url", "http://mainsv.choilon.com:8880/paytrust88/success")
		temp.Add("failed_return_url", "http://mainsv.choilon.com:8880/paytrust88/fail")
		temp.Add("http_post_url", "http://smsotp.slota.win/api/payin")
		temp.Add("amount", fmt.Sprintf("%v", amountVND))
		temp.Add("item_id", fmt.Sprintf("%v", item_id))
		temp.Add("item_description", "item_description")
		temp.Add("name", "Jon Doe")
		temp.Add("email", fmt.Sprintf("%v@%v", playerId, tnvDomain))
		temp.Add("bank_code", bank_code)
		if zconfig.ServerVersion == "" { // Test
			temp.Add("currency", "MYR")
		} else {
			temp.Add("currency", "VND")
		}
		requestUrl := "https://paytrust88.com/v1/transaction/start?" + temp.Encode()
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
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("resp body", string(body))
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		//		fmt.Println(utils.PFormat(data))
		return map[string]interface{}{
			"UrlToPaymentSite": data["redirect_to"],
			"amountVND":        amountVND,
			"amountKim":        amountKim,
		}, nil
	} else if purchaseType == "luckicoin" {
		h := md5.New()
		h.Write([]byte(fmt.Sprintf("%v:%v", playerId, time.Now())))
		hashedBs := h.Sum(nil)
		key := hex.EncodeToString(hashedBs)
		record.RedisSaveStringExpire(key, fmt.Sprintf("%v", playerId), 300)
		chargingUrl := fmt.Sprintf("http://lucki.co/deposit/%v", key)
		return map[string]interface{}{
			"ChargingUrl": chargingUrl,
		}, nil
	} else {
		return nil, errors.New("err:purchase_type_invalid")
	}
	return nil, err
}

func purchaseMoneyNewCaptcha(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	id_giaodich := utils.GetStringAtPath(data, "id_giaodich")

	requestUrl := "http://api2.godalex.com:8099/getCaptra" +
		fmt.Sprintf("?id_giaodich=%v", id_giaodich)
	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, errors.New("¯\\_(ツ)_/¯ step1.5")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var x map[string]interface{}
	err = json.Unmarshal(body, &x)
	if err != nil {
		return nil, err
	}
	if utils.GetIntAtPath(x, "error") != 0 {
		return nil, errors.New(utils.GetStringAtPath(x, "msg"))
	} else {
		return map[string]interface{}{
			"id_giaodich":     id_giaodich,
			"newCaptchaImage": utils.GetStringAtPath(x, "src"),
		}, nil
	}
}

func purchaseMoneyStep2(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	id_giaodich := utils.GetStringAtPath(data, "id_giaodich")
	captchaAnswer := utils.GetStringAtPath(data, "captchaAnswer")

	LangngoangMapLock.Lock()
	i, isIn := LangngoangMapIdgdToInfo[id_giaodich]
	LangngoangMapLock.Unlock()
	if i == nil || !isIn {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": "id_giaodich không tồn tại hoặc đã quá lâu"}, playerId)
		return nil, nil
	}
	// step 2
	requestUrl := "http://api2.godalex.com:8099/sendCaptra" +
		fmt.Sprintf("?id_giaodich=%v", id_giaodich) +
		fmt.Sprintf("&captra=%v", captchaAnswer)
	resp, err := http.Get(requestUrl)
	if err != nil {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": "¯\\_(ツ)_/¯ step2"}, playerId)
		// coi nhu nap thanh cong 0 dong, luu lai id_giaodich
		handleChargeResult(
			i.playerInstance, i.refererId, i.purchaseType, i.vendor,
			id_giaodich, "0", "0", nil)
		return nil, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//	fmt.Println("loangNgoangPayCard step2 hihi", string(body), resp.Header)
	if err != nil {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": err.Error()}, playerId)
		return nil, nil
	}
	// ex x = {"error":0,"amount": amount,"id_giaodich":id_giaodich},
	var x map[string]interface{}
	err = json.Unmarshal(body, &x)
	if err != nil {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": err.Error()}, playerId)
		return nil, nil
	}
	if utils.GetIntAtPath(x, "error") != 0 {
		server.SendRequest(LANGNGOANG_METHOD_NAME, map[string]interface{}{
			"errorMsg": utils.GetStringAtPath(x, "msg")}, playerId)
		return nil, nil
	}
	realValueStr := fmt.Sprintf("%v", utils.GetInt64AtPath(x, "real_amount"))
	ingameValueStr := fmt.Sprintf("%v", utils.GetInt64AtPath(x, "amount"))

	step3Data, step3Err := handleChargeResult(
		i.playerInstance, i.refererId, i.purchaseType, i.vendor,
		id_giaodich, realValueStr, ingameValueStr, nil)
	if step3Err != nil {
		step3Data["errorMsg"] = step3Err.Error()
	} else {
		step3Data["errorMsg"] = ""
	}
	server.SendRequest(LANGNGOANG_METHOD_NAME,
		step3Data, playerId)
	return nil, nil
}

// transactionId, realMoneyValueStr, ingameMoneyValue, err are from
// card charging API
func handleChargeResult(
	playerInstance *player.Player, refererId int64,
	purchaseType string, vendor string,
	transactionId string, realMoneyValueStr string, ingameMoneyValueStr string,
	err error) (
	map[string]interface{}, error) {
	playerId := playerInstance.Id()

	if err != nil {
		return nil, err
	}
	realValue, _ := strconv.ParseInt(realMoneyValueStr, 10, 64)
	ingameValue, _ := strconv.ParseInt(ingameMoneyValueStr, 10, 64)
	//
	top.GlobalMutex.Lock()
	event := top.MapEvents[top.EVENT_CHARGING_MONEY]
	top.GlobalMutex.Unlock()
	if event != nil {
		event.ChangeValue(playerId, realValue)
	}
	event_player.GlobalMutex.Lock()
	event1 := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
	event_player.GlobalMutex.Unlock()
	if event1 != nil {
		event1.GiveAPiece(playerId, true, false, ingameValue)
	}
	//
	moneyBefore := playerInstance.GetMoney(currency.Money)
	moneyAfter, _ := playerInstance.IncreaseMoney(ingameValue, currency.Money, true)
	_ = playerInstance.IncreaseVipScore(realValue / 100)
	playerInstance.CreateRawMessage("Nạp tiền qua thẻ thành công",
		fmt.Sprintf("Bạn đã nạp thành công %v Kim qua thẻ.", ingameValue),
	)
	record.LogTransactionIdRefererId(refererId, transactionId)
	// currency.ChargingBonus, need to do before LogPurchaseRecord
	var nCB int64
	if !record.CheckIsInFirstTimePurchase(playerId) {
		nCB = ingameValue / 10
	} else {
		if !record.CheckIsInFirstTimePurchaseDaily(playerId) {
			if time.Now().Hour() == 21 {
				nCB = ingameValue / 20
			} else {
				nCB = ingameValue / 20
			}
		}
	}
	if zconfig.ServerVersion == zconfig.SV_02 {
		if nCB != 0 {
			playerInstance.ChangeMoneyAndLog(
				nCB, currency.ChargingBonus, false, "", "", "", "")
			playerInstance.CreateRawMessage("Thưởng điểm tích lũy",
				fmt.Sprintf("Bạn được thưởng %v điểm tích lũy. "+
					"Điểm này có thể đổi sang Kim khi bạn hết tiền. "+
					"Mỗi lần được đổi tối đa 10000 Kim", nCB),
			)
		}
	}
	//
	record.LogPurchaseRecord2(playerInstance.Id(),
		transactionId, purchaseType, fmt.Sprintf("%v_%v", vendor, realValue),
		currency.Money, ingameValue, moneyBefore, moneyAfter, realValue)
	// promotion
	pr := getPromotedRate(playerInstance.IsVerify())
	if pr > 0 {
		promotedMoney := int64(pr * float64(realValue))
		playerInstance.ChangeMoneyAndLog(
			promotedMoney, currency.Money, false, "",
			record.ACTION_PROMOTE, "", "")
		playerInstance.CreateRawMessage(
			fmt.Sprintf("Khuyến Mãi %.0f%% Nạp", 100*promotedRate),
			fmt.Sprintf("Chúc mừng bạn đã nhận được %v Kim "+
				"tương đương %.0f%% giá trị nạp.", promotedMoney, 100*promotedRate),
		)
	}
	// gift code percentage
	gp := getPercentageGiftCharge(playerInstance.Id())
	if gp != 0 {
		promotedMoney := int64(gp * float64(realValue))
		playerInstance.ChangeMoneyAndLog(
			promotedMoney, currency.Money, false, "",
			record.ACTION_PROMOTE, "", "")
		playerInstance.CreateRawMessage(
			fmt.Sprintf("Khuyến Mãi %.0f%% Nạp", 100*promotedRate),
			fmt.Sprintf("Chúc mừng bạn đã nhận được %v Kim "+
				"tương đương %.0f%% giá trị nạp từ giftcode", promotedMoney, 100*promotedRate),
		)
	}
	//
	data1 := make(map[string]interface{})
	data1["amount"] = ingameValue
	return data1, nil
}

func GetSumCharging(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	result := map[string]interface{}{}
	rows, e := dataCenter.Db().Query("SELECT player_id, purchase_type, sum_value "+
		"FROM purchase_sum WHERE player_id = $1 ",
		playerId)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var player_id, sum_value int64
		var purchase_type string
		e := rows.Scan(&player_id, &purchase_type, &sum_value)
		if e != nil {
			return nil, e
		}
		result[purchase_type] = sum_value
	}
	return result, nil
}
