package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/money"
	"github.com/vic/vic_go/models/otp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

func updateUsername(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	username := utils.GetStringAtPath(data, "username")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateUsername player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = player.UpdateUsername(username)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["username"] = username
	responseData["id"] = player.Id()
	return responseData, nil
}

func updateDisplayName(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	username := utils.GetStringAtPath(data, "display_name")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateUsername player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = player.UpdateDisplayName(username)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["display_name"] = username
	responseData["id"] = player.Id()
	return responseData, nil
}

func updateAvatar(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	avatar := utils.GetStringAtPath(data, "avatar_url")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateAvatar player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = player.UpdateAvatar(avatar)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["avatar_url"] = avatar
	responseData["id"] = player.Id()
	return responseData, nil
}

func updateEmail(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	email := utils.GetStringAtPath(data, "email")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateEmail player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	err = player.UpdateEmail(email)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["email"] = email
	responseData["id"] = player.Id()
	return responseData, nil
}

func updatePhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	return nil, errors.New(l.Get(l.M0062))
	phoneNumber := utils.GetStringAtPath(data, "phone_number")
	player, err := models.GetPlayer(playerId)

	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateAvatar player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if player.IsVerify() {
		// need to register to change phone number through otp code
		if !player.PhoneNumberChangeAvailable() {
			return nil, errors.New(l.Get(l.M0062))
		}

		err = player.UpdatePhoneNumber(phoneNumber)
		if err != nil {
			return nil, err
		}
		responseData = make(map[string]interface{})
		responseData["phone_number"] = phoneNumber
		responseData["id"] = player.Id()
		return responseData, nil

	} else {
		// can just update
		err = player.UpdatePhoneNumber(phoneNumber)
		if err != nil {
			return nil, err
		}
		responseData = make(map[string]interface{})
		responseData["phone_number"] = phoneNumber
		responseData["id"] = player.Id()
		return responseData, nil
	}
}

func setNewPassword(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	password := utils.GetStringAtPath(data, "password")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	if player == nil {
		log.LogSerious("updatePassword player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}
	if !player.IsVerify() {
		return nil, errors.New("Bạn cần xác nhận tài khoản trước khi sử dụng chức năng này")
	}
	err = player.UpdatePassword(password)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["id"] = player.Id()
	return responseData, nil
}

func updatePassword(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	password := utils.GetStringAtPath(data, "password")
	oldPassword := utils.GetStringAtPath(data, "old_password")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	if player == nil {
		log.LogSerious("updatePassword player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if !player.CheckPassword(oldPassword) {
		return nil, errors.New(l.Get(l.M0063))
	}

	err = player.UpdatePassword(password)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["id"] = player.Id()
	return responseData, nil
}

func verifyPhoneNumber(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("updateAvatar player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if player.IsVerify() {
		return nil, errors.New(l.Get(l.M0064))
	}

	err = otp.RegisterVerifyPhoneNumber(player.Id())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getPlayerData(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	player := currentPlayer

	var getCurrentPlayerData bool
	playerIdParams := utils.GetInt64AtPath(data, "player_id")
	fields := utils.GetStringSliceAtPath(data, "fields")
	if playerIdParams != 0 {
		getCurrentPlayerData = false
		playerId = playerIdParams

		player, err = models.GetPlayer(playerId)
		if err != nil {
			return nil, err
		}
	} else {
		getCurrentPlayerData = true
	}

	if player == nil {
		log.LogSerious("get player data player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	responseData = player.SerializedDataWithFields(fields)

	if getCurrentPlayerData {
		if utils.ContainsByString(fields, "room") {
			if player.Room() != nil {
				responseData["room"] = player.Room().SerializedDataFull(player)
			}
		}

		if feature.IsTimeBonusAvailable() {
			if utils.ContainsByString(fields, "time_bonus") {
				responseData["time_bonus"] = player.GetTimeBonusData()
			}
		}

		if utils.ContainsByString(fields, "email") {
			responseData["email"] = player.Email()
		}

		if utils.ContainsByString(fields, "payment_requirement_text") {
			responseData["payment_requirement_text"] = money.GetPaymentRequirement().GetPaymentRequirementTextForPlayer(player)
		}
	}

	if utils.ContainsByString(fields, "relationship") {
		responseData["relationship"] = currentPlayer.GetRelationshipDataWithPlayer(playerId)
	}

	return responseData, nil
}

func searchPlayer(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	keywords := utils.GetStringAtPath(data, "keywords")
	return player.SearchPlayer(keywords)
}

func claimTimeBonus(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return currentPlayer.ClaimTimeBonus()
}

func sendFeedback(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	star := utils.GetIntAtPath(data, "star")
	feedback := utils.GetStringAtPath(data, "feedback")
	version := utils.GetStringAtPath(data, "version")
	err = currentPlayer.SendFeedback(version, star, feedback)
	return nil, err
}

func registerPNDevice(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	apnsDeviceToken := utils.GetStringAtPath(data, "apns_device_token")
	gcmDeviceToken := utils.GetStringAtPath(data, "gcm_device_token")
	err = currentPlayer.RegisterPNDevice(apnsDeviceToken, gcmDeviceToken)
	return nil, err
}

func getInboxMessage(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	limit := utils.GetInt64AtPath(data, "limit")
	offset := utils.GetInt64AtPath(data, "offset")
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	results, total, err := currentPlayer.GetInboxMessages(limit, offset)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = results
	responseData["total"] = total
	return responseData, nil
}

func getInboxMessageByType(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	limit := utils.GetInt64AtPath(data, "limit")
	offset := utils.GetInt64AtPath(data, "offset")
	msgType := utils.GetStringAtPath(data, "msgType")
	currentPlayer, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	results, total, err := currentPlayer.GetInboxMessagesByType(
		limit, offset, msgType)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = results
	responseData["total"] = total
	return responseData, nil
}

func markReadInboxMessages(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)

	if err != nil {
		return nil, err
	}

	err = currentPlayer.MarkReadAllMessages()
	return nil, err
}

func markReadInbox1Message(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)

	msgId := utils.GetInt64AtPath(data, "msgId")

	if err != nil {
		return nil, err
	}

	err = currentPlayer.MarkRead1Message(msgId)
	return nil, err
}

func deleteInbox1Message(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currentPlayer, err := models.GetPlayer(playerId)

	msgId := utils.GetInt64AtPath(data, "msgId")

	if err != nil {
		return nil, err
	}

	err = currentPlayer.Delete1Message(msgId)
	return nil, err
}

func SendMsgToPlayerId(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	targetId := utils.GetInt64AtPath(data, "targetId")
	title := utils.GetStringAtPath(data, "msgTitle")
	content := utils.GetStringAtPath(data, "msgContent")
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	playerObj, _ := models.GetPlayer(playerId)
	if playerObj == nil {
		return nil, errors.New("playerObj == nil")
	}
	targetObj, _ := models.GetPlayer(targetId)
	if targetObj == nil {
		return nil, errors.New(l.Get(l.M0067))
	}
	targetObj.CreateRawMessage2(title, content, playerObj)
	return nil, err
}

func UserSendMsgToAdmin(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	message := utils.GetStringAtPath(data, "message")
	jsonString := userSendMsgToAdmin(playerId, message)
	e := json.Unmarshal([]byte(jsonString), &responseData)
	if e != nil {
		return nil, errors.New(jsonString)
	} else {
		return
	}
}

func GetMsgsForUser(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	limit := utils.GetIntAtPath(data, "limit")
	offset := utils.GetIntAtPath(data, "offset")

	jsonString := getMsgsForUser(playerId, limit, offset)
	e := json.Unmarshal([]byte(jsonString), &responseData)
	if e != nil {
		return nil, errors.New(jsonString)
	} else {
		return
	}
}

func MarkMsgAsRead(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	msgId := utils.GetInt64AtPath(data, "msgId")
	jsonString := markMsgAsRead(msgId)
	e := json.Unmarshal([]byte(jsonString), &responseData)
	if e != nil {
		return nil, errors.New(jsonString)
	} else {
		return
	}
}

func BuyShopItem(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	item_id := utils.GetInt64AtPath(data, "item_id")
	buyer_name := utils.GetStringAtPath(data, "buyer_name")
	buyer_phone := utils.GetStringAtPath(data, "buyer_phone")
	buyer_address := utils.GetStringAtPath(data, "buyer_address")
	playerObj, _ := models.GetPlayer(playerId)
	if playerObj == nil {
		return nil, errors.New("playerObj == nil")
	}
	shopItems := getShopItems()
	item, isIn := shopItems[item_id]
	if !isIn {
		return nil, errors.New("Sai mã vật phẩm")
	}
	itemPriceI := item["price"]
	discount_rateI := item["discount_rate"]
	itemPriceO, isOk := itemPriceI.(int64)
	discount_rate, isOk1 := discount_rateI.(float64)
	itemPrice := int64(float64(itemPriceO) * (1 - discount_rate))
	if !isOk || !isOk1 {
		return nil, errors.New("Lỗi hệ thống: Không có giá vật phẩm")
	}
	if playerObj.GetAvailableMoney(currency.Money) < itemPrice {
		return nil, errors.New(l.Get(l.M0016))
	}
	playerObj.ChangeMoneyAndLog(
		-itemPrice, currency.Money, false, "",
		record.ACTION_BUY_SHOP_ITEM, "", "")
	dataCenter.Db().Exec("INSERT INTO shop_item_buyer "+
		"(item_id, buyer_id, buyer_name, buyer_phone, buyer_address) "+
		"VALUES ($1, $2, $3, $4, $5)",
		item_id, playerId, buyer_name, buyer_phone, buyer_address)
	return nil, err
}

func AgencyRegister(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	pObj, _ := player.GetPlayer(playerId)
	if pObj == nil {
		return nil, errors.New("pObj == nil")
	}
	if pObj.PhoneNumber() == "" {
		return nil, errors.New(l.Get(l.M0079))
	}
	bank_name := utils.GetStringAtPath(data, "bank_name")
	bank_account_number := utils.GetStringAtPath(data, "bank_account_number")
	bank_account__name := utils.GetStringAtPath(data, "bank_account__name")
	address := utils.GetStringAtPath(data, "address")
	_, e := dataCenter.Db().Exec(
		"INSERT INTO player_agency"+
			"(player_id, bank_name, bank_account_number, "+
			"    bank_account__name, address, rate_kim_to_vnd)"+
			"VALUES ($1, $2, $3, $4, $5, $6) ",
		playerId, bank_name, bank_account_number,
		bank_account__name, address, 1/RATE_BANK_CHARGING)
	if e != nil {
		return nil, e
	} else {
		return nil, nil
	}
}

func AgencyEdit(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	bank_name := utils.GetStringAtPath(data, "bank_name")
	bank_account_number := utils.GetStringAtPath(data, "bank_account_number")
	bank_account__name := utils.GetStringAtPath(data, "bank_account__name")
	address := utils.GetStringAtPath(data, "address")
	_, e := dataCenter.Db().Exec(
		"UPDATE player_agency "+
			"SET bank_name=$2, bank_account_number=$3, "+
			"    bank_account__name=$4, address=$5 "+
			"WHERE player_id=$1 ",
		playerId, bank_name, bank_account_number, bank_account__name, address)
	if e != nil {
		return nil, e
	} else {
		return nil, nil
	}
}

func editAgency(pid int64, field string, newValue interface{}) error {
	if _, isIn := map[string]bool{
		"bank_name":           true,
		"bank_account_number": true,
		"bank_account__name":  true,
		"address":             true,
		"rate_kim_to_vnd":     true,
		"email":               true,
		"skype":               true}[field]; !isIn {
		return errors.New("Wrong field")
	} else {
		_, e := dataCenter.Db().Exec(
			"UPDATE player_agency SET $1 = $2 WHERE player_id = $3",
			field, newValue, pid,
		)
		return e
	}
}

func agencyHide(pid int64) error {
	_, e := dataCenter.Db().Exec(
		"UPDATE player_agency SET is_hidding = TRUE "+
			"WHERE player_id = $1 ",
		pid)
	if e != nil {
		return e
	}
	_, e = dataCenter.Db().Exec(
		"UPDATE player_privileges "+
			"SET can_transfer_money=FALSE "+
			"WHERE player_id = $1 ",
		pid)
	return e
}

// input time is local
func calcSumTransfer(sender_id int64, lbTime time.Time, ubTime time.Time) int64 {
	row := dataCenter.Db().QueryRow(
		"SELECT sender_id, sum(amount_kim) FROM player_transfer_record "+
			"WHERE sender_id = $1 AND $1 <= created_time AND created_time < $2 "+
			"GROUP BY sender_id ",
		sender_id, lbTime.UTC(), ubTime.UTC(),
	)
	var sid, sum int64
	e := row.Scan(&sid, &sum)
	if e != nil {
		fmt.Println("ERROR ", e)
	}
	return sum
}

func AgencyGetList(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	isSeller := utils.GetBoolAtPath(data, "isSeller")
	r, e := getAgencies(isSeller, 0)
	// để giữ lại test acc id=7
	// delete(r, "7")
	//
	return r, e
}

func AgencyGetMyInfo(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	r, e := getAgencies(true, playerId)
	return r, e
}

func AgencyCheck(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	pid := utils.GetInt64AtPath(data, "idOrName")
	uname := utils.GetStringAtPath(data, "idOrName")
	return map[string]interface{}{
		"isAgency": agencyCheck(pid, uname),
	}, nil
}

// check a pid or username is a agency
func agencyCheck(pid int64, uname string) bool {
	r1, _ := getAgencies(true, pid)
	if len(r1) > 0 {
		return true
	}
	pid2, e := player.FindPlayerId(uname)
	if e != nil {
		return false
	}
	r2, _ := getAgencies(true, pid2)
	if len(r2) > 0 {
		return true
	} else {
		return false
	}
}

// arg filterPid = 0: dont filter.
// arg isSeller is a little bit misleading, a user !isSeller means he is buyer
// return map[pid]infoMap
func getAgencies(isSeller bool, filterPid int64) (map[string]interface{}, error) {
	temp := "can_transfer_money"
	if !isSeller {
		temp = "can_receive_money"
	}
	rows, e := dataCenter.Db().Query(
		"    SELECT player_agency.player_id, bank_name, bank_account_number, "+
			"    bank_account__name, address, rate_kim_to_vnd, email, skype, "+
			"    currency.value, phone2 "+
			"FROM player_agency "+
			"JOIN player_privileges "+
			"    ON player_agency.player_id = player_privileges.player_id "+
			"JOIN currency ON player_agency.player_id = currency.player_id "+
			fmt.Sprintf("WHERE is_accepted = TRUE AND %v = TRUE ", temp)+
			"    AND currency_type = $1 AND is_hidding = FALSE ",
		currency.Money)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	result := make(map[string]interface{})
	for rows.Next() {
		var player_id, value int64
		var rate_kim_to_vnd float64
		var bank_name, bank_account_number, bank_account__name, address, email,
			skype, phone2 string
		e1 := rows.Scan(&player_id, &bank_name, &bank_account_number,
			&bank_account__name, &address, &rate_kim_to_vnd, &email, &skype,
			&value, &phone2)
		if e1 == nil {
			pObj, _ := player.GetPlayer(player_id)
			if pObj != nil {
				result[fmt.Sprintf("%v", player_id)] = map[string]interface{}{
					"player_id":           player_id,
					"bank_name":           bank_name,
					"bank_account_number": bank_account_number,
					"bank_account__name":  bank_account__name,
					"address":             address,
					"rate_kim_to_vnd":     rate_kim_to_vnd,
					"email":               email,
					"skype":               skype,
					"money":               value,
					"phone2":              phone2,
					"phone_number":        pObj.PhoneNumber(),
					"username":            pObj.Username(),
					"display_name":        pObj.DisplayName(),
				}
			}
		} else {
			fmt.Println("ERROR: ", e1)
		}
	}
	if filterPid != 0 {
		for k, _ := range result {
			if k != fmt.Sprintf("%v", filterPid) {
				delete(result, k)
			}
		}
	}
	return result, nil
}

func AgencyTransferList(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	isSelling := utils.GetBoolAtPath(data, "isSelling")
	limit := utils.GetIntAtPath(data, "limit")
	showCheckedTrans := utils.GetBoolAtPath(data, "showCheckedTrans")
	if limit == 0 {
		limit = 20
	}
	offset := utils.GetIntAtPath(data, "offset")
	var query string
	if isSelling {
		var temp string
		if !showCheckedTrans {
			temp = ", AND has_sender_checked=FALSE "
		}
		query = "SELECT id, sender_id, target_id, amount_kim, created_time, " +
			"    has_sender_checked " +
			"FROM player_transfer_record " +
			"WHERE sender_id=$1 " + temp +
			"ORDER BY created_time DESC LIMIT $2 OFFSET $3 "
	} else {
		var temp string
		if !showCheckedTrans {
			temp = ", AND has_target_checked=FALSE "
		}
		query = "SELECT id, sender_id, target_id, amount_kim, created_time, " +
			"    has_target_checked " +
			"FROM player_transfer_record " +
			"WHERE target_id=$1 " + temp +
			"ORDER BY created_time DESC LIMIT $2 OFFSET $3 "
	}
	rows, e := dataCenter.Db().Query(query, playerId, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	r := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sender_id, target_id, amount_kim int64
		var created_time time.Time
		var has_checked bool
		e1 := rows.Scan(&id, &sender_id, &target_id,
			&amount_kim, &created_time, &has_checked)
		if e1 == nil {
			r = append(r, map[string]interface{}{
				"id":           id,
				"sender_id":    sender_id,
				"target_id":    target_id,
				"amount_kim":   amount_kim,
				"created_time": created_time.Local().Format(time.RFC3339),
				"has_checked":  has_checked,
			})
		}
	}
	return map[string]interface{}{"list": r}, nil
}

func AgencyTransferListMarkComplete(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	transferId := utils.GetInt64AtPath(data, "transferId")
	row := dataCenter.Db().QueryRow(
		"SELECT sender_id, target_id FROM player_transfer_record WHERE id = $1 ",
		transferId)
	var sender_id, target_id int64
	e := row.Scan(&sender_id, &target_id)
	if e != nil {
		return nil, errors.New("Sai mã giao dịch")
	}
	if playerId == sender_id {
		dataCenter.Db().Exec(
			"UPDATE player_transfer_record SET has_sender_checked = TRUE "+
				"WHERE id = $1 ",
			transferId)
		return nil, nil
	} else if playerId == target_id {
		dataCenter.Db().Exec(
			"UPDATE player_transfer_record SET has_target_checked = TRUE "+
				"WHERE id = $1 ",
			transferId)
		return nil, nil
	} else {
		return nil, errors.New("Giao dịch không liên quan đến bạn")
	}
}
