package money

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

var cardTypes []*CardType

type CardType struct {
	id       int64
	cardName string
	cardCode string
	money    int64
}

func (cardType *CardType) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = cardType.id
	data["card_name"] = cardType.cardName
	data["card_code"] = cardType.cardCode
	data["money"] = cardType.money
	return data
}

type ByCardMoney []map[string]interface{}

func (a ByCardMoney) Len() int      { return len(a) }
func (a ByCardMoney) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCardMoney) Less(i, j int) bool {
	iValue := utils.GetInt64AtPath(a[i], "money")
	jValue := utils.GetInt64AtPath(a[j], "money")

	return iValue < jValue
}

type Card struct {
	id         int64
	cardType   string
	cardCode   string
	serialCode string
	cardNumber string
	cardValue  string
	status     string

	claimedByPlayerId int64
	playerUsername    string

	acceptedByAdminId int64
	adminUsername     string

	createdAt time.Time
	claimedAt time.Time
}

func (card *Card) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = card.id
	data["card_type"] = card.cardType
	data["card_code"] = card.cardCode
	data["serial_code"] = card.serialCode
	data["card_number"] = card.cardNumber
	data["status"] = card.status
	data["claimed_by_player_id"] = card.claimedByPlayerId
	data["claimed_by_player_username"] = card.playerUsername
	data["accepted_by_admin_id"] = card.acceptedByAdminId
	data["accepted_by_admin_username"] = card.adminUsername
	data["created_at"] = utils.FormatTime(card.createdAt)
	data["claimed_at"] = utils.FormatTime(card.claimedAt)
	return data
}

func fetchCardTypes() {
	queryString := "SELECT id, card_name, card_code, money FROM card_type"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("error when fetch card types %v", err)
		return
	}

	defer rows.Close()

	cardTypes = make([]*CardType, 0)
	for rows.Next() {
		var id int64
		var code string
		var name string
		var money int64
		err = rows.Scan(&id, &name, &code, &money)
		if err != nil {
			log.LogSerious("error fetch card type data %v", err)
			return
		}
		cardType := &CardType{
			id:       id,
			cardName: name,
			cardCode: code,
			money:    money,
		}
		cardTypes = append(cardTypes, cardType)
	}
}

func getCard(cardCode string) *Card {
	queryString := "SELECT id, serial_code, card_number FROM card WHERE card_code = $1 AND status = $2 LIMIT 1"
	row := dataCenter.Db().QueryRow(queryString, cardCode, "unclaimed")
	var id int64
	var serialCode, cardNumber string
	err := row.Scan(&id, &serialCode, &cardNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.LogSerious("err get card %v", err)
		return nil
	}
	card := &Card{
		id:         id,
		cardCode:   cardCode,
		serialCode: serialCode,
		cardNumber: cardNumber,
	}
	return card
}

func GetCardsData(cardType string, status string, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	if cardType == "" && status == "" {
		return getAllCardsData(limit, offset)
	}
	if cardType == "" && status != "" {
		return getCardsDataWithStatus(status, limit, offset)
	}

	if cardType != "" && status == "" {
		return getCardsDataWithCardType(cardType, limit, offset)
	}

	return getCardsData(cardType, status, limit, offset)
}

func getAllCardsData(limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	queryString := "SELECT COUNT(id) FROM card"
	row := dataCenter.Db().QueryRow(queryString)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = "SELECT card.id, card.card_type, card.card_code, card.serial_code, card.card_number," +
		" player.id,player.username,admin.username,card.created_at,card.claimed_at" +
		" FROM card as card" +
		" LEFT OUTER JOIN player as player ON player.id = card.claimed_by_player_id" +
		" LEFT OUTER JOIN admin_account as admin ON admin.id = card.accepted_by_admin_id" +
		"  ORDER BY -card.id LIMIT $1 OFFSET $2"

	rows, err := dataCenter.Db().Query(queryString, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	results = make([]map[string]interface{}, 0)

	for rows.Next() {
		var id int64
		var playerId sql.NullInt64
		var playerUsername, adminUsername sql.NullString
		var cardType, cardCode, serialCode, cardNumber string
		var createdAt, claimedAt time.Time

		err = rows.Scan(&id, &cardType, &cardCode, &serialCode, &cardNumber, &playerId, &playerUsername, &adminUsername, &createdAt, &claimedAt)
		if err != nil {
			return nil, 0, err
		}

		var status string
		if playerUsername.String == "" {
			status = "Chưa sử dụng"
		} else {

			claimedAtDateString, claimedAtTimeString := utils.FormatTimeToVietnamTime(claimedAt)
			status = fmt.Sprintf("Đã được nhận bởi %s đồng ý bởi %s vào lúc %s", playerUsername.String, adminUsername.String, fmt.Sprintf("%s %s", claimedAtDateString, claimedAtTimeString))
		}

		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerId.Int64
		data["card_type"] = cardType
		data["card_code"] = cardCode
		data["serial_code"] = serialCode
		data["card_number"] = hideString(cardNumber)
		data["status"] = status
		data["player_username"] = playerUsername.String
		data["admin_username"] = adminUsername.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["claimed_at"] = utils.FormatTimeToVietnamDateTimeString(claimedAt)

		results = append(results, data)
	}

	return results, total, nil
}

func getCardsDataWithStatus(status string, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	queryString := "SELECT COUNT(id) FROM card"
	row := dataCenter.Db().QueryRow(queryString)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = "SELECT card.id, card.card_type, card.card_code, card.serial_code, card.card_number," +
		" player.id,player.username,admin.username,card.created_at,card.claimed_at" +
		" FROM card as card" +
		" LEFT OUTER JOIN player as player ON player.id = card.claimed_by_player_id" +
		" LEFT OUTER JOIN admin_account as admin ON admin.id = card.accepted_by_admin_id" +
		" WHERE status = $1" +
		"  ORDER BY -card.id LIMIT $2 OFFSET $3"

	rows, err := dataCenter.Db().Query(queryString, status, limit, offset)
	if err != nil {

		return nil, 0, err
	}
	defer rows.Close()
	results = make([]map[string]interface{}, 0)

	for rows.Next() {
		var id int64
		var playerId sql.NullInt64
		var playerUsername, adminUsername sql.NullString
		var cardType, cardCode, serialCode, cardNumber string
		var createdAt, claimedAt time.Time

		err = rows.Scan(&id, &cardType, &cardCode, &serialCode, &cardNumber, &playerId, &playerUsername, &adminUsername, &createdAt, &claimedAt)
		if err != nil {
			return nil, 0, err
		}

		var status string
		if playerUsername.String == "" {
			status = "Chưa sử dụng"
		} else {
			claimedAtString := utils.FormatTimeToVietnamDateTimeString(claimedAt)
			status = fmt.Sprintf("Đã được nhận bởi %s đồng ý bởi %s vào lúc %s", playerUsername.String, adminUsername.String, claimedAtString)
		}

		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerId.Int64
		data["card_type"] = cardType
		data["card_code"] = cardCode
		data["serial_code"] = serialCode
		data["card_number"] = hideString(cardNumber)
		data["status"] = status
		data["player_username"] = playerUsername.String
		data["admin_username"] = adminUsername.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["claimed_at"] = utils.FormatTimeToVietnamDateTimeString(claimedAt)

		results = append(results, data)
	}

	return results, total, nil
}

func getCardsDataWithCardType(cardType string, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	queryString := "SELECT COUNT(id) FROM card"
	row := dataCenter.Db().QueryRow(queryString)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = "SELECT card.id, card.card_type, card.card_code, card.serial_code, card.card_number," +
		" player.id,player.username,admin.username,card.created_at,card.claimed_at" +
		" FROM card as card" +
		" LEFT OUTER JOIN player as player ON player.id = card.claimed_by_player_id" +
		" LEFT OUTER JOIN admin_account as admin ON admin.id = card.accepted_by_admin_id" +
		" WHERE card_type = $1" +
		"  ORDER BY -card.id LIMIT $2 OFFSET $3"

	rows, err := dataCenter.Db().Query(queryString, cardType, limit, offset)
	if err != nil {

		return nil, 0, err
	}
	defer rows.Close()
	results = make([]map[string]interface{}, 0)

	for rows.Next() {
		var id int64
		var playerId sql.NullInt64
		var playerUsername, adminUsername sql.NullString
		var cardType, cardCode, serialCode, cardNumber string
		var createdAt, claimedAt time.Time

		err = rows.Scan(&id, &cardType, &cardCode, &serialCode, &cardNumber, &playerId, &playerUsername, &adminUsername, &createdAt, &claimedAt)
		if err != nil {
			return nil, 0, err
		}

		var status string
		if playerUsername.String == "" {
			status = "Chưa sử dụng"
		} else {
			claimedAtString := utils.FormatTimeToVietnamDateTimeString(claimedAt)
			status = fmt.Sprintf("Đã được nhận bởi %s đồng ý bởi %s vào lúc %s", playerUsername.String, adminUsername.String, claimedAtString)
		}

		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerId.Int64
		data["card_type"] = cardType
		data["card_code"] = cardCode
		data["serial_code"] = serialCode
		data["card_number"] = hideString(cardNumber)
		data["status"] = status
		data["player_username"] = playerUsername.String
		data["admin_username"] = adminUsername.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["claimed_at"] = utils.FormatTimeToVietnamDateTimeString(claimedAt)

		results = append(results, data)
	}

	return results, total, nil
}

func getCardsData(cardType string, status string, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	queryString := "SELECT COUNT(id) FROM card"
	row := dataCenter.Db().QueryRow(queryString)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = "SELECT card.id, card.card_type, card.card_code, card.serial_code, card.card_number," +
		" player.id,player.username,admin.username,card.created_at,card.claimed_at" +
		" FROM card as card" +
		" LEFT OUTER JOIN player as player ON player.id = card.claimed_by_player_id" +
		" LEFT OUTER JOIN admin_account as admin ON admin.id = card.accepted_by_admin_id" +
		" WHERE status = $1 AND card_type = $2" +
		"  ORDER BY -card.id LIMIT $3 OFFSET $4"

	rows, err := dataCenter.Db().Query(queryString, status, cardType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	results = make([]map[string]interface{}, 0)

	for rows.Next() {
		var id int64
		var playerId sql.NullInt64
		var playerUsername, adminUsername sql.NullString
		var cardType, cardCode, serialCode, cardNumber string
		var createdAt, claimedAt time.Time

		err = rows.Scan(&id, &cardType, &cardCode, &serialCode, &cardNumber, &playerId, &playerUsername, &adminUsername, &createdAt, &claimedAt)
		if err != nil {
			return nil, 0, err
		}

		var status string
		if playerUsername.String == "" {
			status = "Chưa sử dụng"
		} else {
			claimedAtString := utils.FormatTimeToVietnamDateTimeString(claimedAt)
			status = fmt.Sprintf("Đã được nhận bởi %s đồng ý bởi %s vào lúc %s", playerUsername.String, adminUsername.String, claimedAtString)
		}

		data := make(map[string]interface{})
		data["id"] = id
		data["player_id"] = playerId.Int64
		data["card_type"] = cardType
		data["card_code"] = cardCode
		data["serial_code"] = serialCode
		data["card_number"] = hideString(cardNumber)
		data["status"] = status
		data["player_username"] = playerUsername.String
		data["admin_username"] = adminUsername.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		data["claimed_at"] = utils.FormatTimeToVietnamDateTimeString(claimedAt)

		results = append(results, data)
	}

	return results, total, nil
}

func CreateCard(cardType string, cardCode string, serialCode string, cardNumber string) (err error) {
	var found bool
	for _, cardType := range cardTypes {
		if cardType.cardCode == cardCode {
			found = true
			break
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("err:card_code_not_exist_%s", cardCode))
	}

	queryString := "INSERT INTO card (card_type, card_code, serial_code, card_number) VALUES ($1,$2,$3,$4)"
	_, err = dataCenter.Db().Exec(queryString, cardType, cardCode, serialCode, cardNumber)
	return err
}

func GetCardsDataSummary() (resultsData map[string]interface{}, err error) {
	resultsData = make(map[string]interface{})
	var unclaimedSum, claimedSum, totalSum int64

	results := make([]map[string]interface{}, 0)

	for _, telco := range []string{"mobi", "vina", "viettel"} {
		for _, value := range []string{"50", "100", "200", "500"} {
			cardCode := fmt.Sprintf("%s_%s", telco, value)
			data := make(map[string]interface{})
			data["card_code"] = cardCode
			var telcoString string
			if telco == "mobi" {
				telcoString = "Mobifone"
			} else if telco == "vina" {
				telcoString = "Vinaphone"
			} else if telco == "viettel" {
				telcoString = "viettel"
			}
			data["telco"] = telcoString
			realMoneyValue := fmt.Sprintf("%s.000 VND", value)
			data["real_money"] = realMoneyValue
			results = append(results, data)
		}
	}

	queryString := "SELECT card_code, COUNT(card_code) FROM card WHERE status = $1 GROUP BY card_code"
	rows, err := dataCenter.Db().Query(queryString, "unclaimed")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["unclaimed_count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				unclaimedSum += realMoneyInt * count
				data["unclaimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	queryString = "SELECT card_code, COUNT(card_code) FROM card WHERE status = $1 GROUP BY card_code"
	rows, err = dataCenter.Db().Query(queryString, "claimed")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["claimed_count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				claimedSum += realMoneyInt * count
				data["claimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	queryString = "SELECT card_code, COUNT(card_code) FROM card GROUP BY card_code"
	rows, err = dataCenter.Db().Query(queryString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				totalSum += realMoneyInt * count
				data["sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	resultsData["results"] = results
	resultsData["total_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(totalSum))
	resultsData["claimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(claimedSum))
	resultsData["unclaimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(unclaimedSum))
	return resultsData, nil
}

func GetCardsDataHistory(startDate time.Time, endDate time.Time) (resultsData map[string]interface{}, err error) {
	if startDate.IsZero() || endDate.IsZero() {
		return GetCardsDataSummary()
	}
	resultsData = make(map[string]interface{})
	var unclaimedSum, claimedSum, totalSum int64

	results := make([]map[string]interface{}, 0)
	for _, telco := range []string{"mobi", "vina", "viettel"} {
		for _, value := range []string{"50", "100", "200", "500"} {
			cardCode := fmt.Sprintf("%s_%s", telco, value)
			data := make(map[string]interface{})
			data["card_code"] = cardCode
			var telcoString string
			if telco == "mobi" {
				telcoString = "Mobifone"
			} else if telco == "vina" {
				telcoString = "Vinaphone"
			} else if telco == "viettel" {
				telcoString = "viettel"
			}
			data["telco"] = telcoString
			realMoneyValue := fmt.Sprintf("%s.000 VND", value)
			data["real_money"] = realMoneyValue
			results = append(results, data)
		}
	}

	queryString := "SELECT card_code, COUNT(card_code) FROM card WHERE status = $1 AND created_at >= $2 AND created_at <= $3 GROUP BY card_code"
	rows, err := dataCenter.Db().Query(queryString, "unclaimed", startDate.UTC(), endDate.UTC())
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["unclaimed_count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				unclaimedSum += realMoneyInt * count
				data["unclaimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	queryString = "SELECT card_code, COUNT(card_code) FROM card WHERE status = $1 AND claimed_at >= $2 AND claimed_at <= $3 GROUP BY card_code"
	rows, err = dataCenter.Db().Query(queryString, "claimed", startDate.UTC(), endDate.UTC())
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["claimed_count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				claimedSum += realMoneyInt * count
				data["claimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	queryString = "SELECT card_code, COUNT(card_code) FROM card WHERE created_at >= $1 AND created_at <= $2 GROUP BY card_code"
	rows, err = dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC())
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cardCode string
		var count int64

		err = rows.Scan(&cardCode, &count)
		if err != nil {
			rows.Close()
			return nil, err
		}

		tokens := strings.Split(cardCode, "_")
		for _, data := range results {
			if utils.GetStringAtPath(data, "card_code") == cardCode {
				data["count"] = count
				realMoneyInt, _ := strconv.ParseInt(tokens[1], 10, 64)
				totalSum += realMoneyInt * count
				data["sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(realMoneyInt*count))
				break
			}
		}
	}
	rows.Close()

	resultsData["results"] = results
	resultsData["total_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(totalSum))
	resultsData["claimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(claimedSum))
	resultsData["unclaimed_sum"] = fmt.Sprintf("%s.000 VND", utils.FormatWithComma(unclaimedSum))

	return resultsData, nil
}

func GetCardTypesData() (results []map[string]interface{}) {
	results = make([]map[string]interface{}, 0)
	for _, cardType := range cardTypes {
		data := make(map[string]interface{})
		data["id"] = cardType.id
		data["money_format"] = utils.FormatWithComma(cardType.money)
		data["money"] = cardType.money
		data["card_code"] = cardType.cardCode
		results = append(results, data)
	}

	sort.Sort(ByCardMoney(results))
	return results
}

func UpdateCardType(id int64, code string, money int64) (err error) {
	queryString := "UPDATE card_type SET money = $1, card_code = $2 WHERE id = $3 "
	_, err = dataCenter.Db().Exec(queryString, money, code, id)
	if err != nil {
		return err
	}
	cardType := GetCardType(id)
	if cardType != nil {
		cardType.money = money
		cardType.cardCode = code
	}
	return nil
}

func DeleteCardType(id int64) (err error) {
	queryString := "DELETE FROM card_type WHERE id = $1"
	_, err = dataCenter.Db().Exec(queryString, id)
	if err != nil {
		log.LogSerious("err delete card %v", err)
		return err
	}
	newCardTypes := make([]*CardType, 0)
	for _, cardType := range cardTypes {
		if cardType.id != id {
			newCardTypes = append(newCardTypes, cardType)
		}
	}
	cardTypes = newCardTypes
	return nil
}

func GetCardTypeDataById(id int64) map[string]interface{} {
	cardType := GetCardType(id)
	if cardType == nil {
		return nil
	}

	data := make(map[string]interface{})
	data["id"] = cardType.id
	data["card_code"] = cardType.cardCode
	data["money"] = cardType.money
	return data
}

func GetCardTypeData(code string) map[string]interface{} {
	cardType := GetCardTypeByCode(code)
	if cardType == nil {
		return nil
	}
	data := make(map[string]interface{})
	data["card_code"] = cardType.cardCode
	data["money"] = cardType.money
	return data
}

func GetCardType(id int64) *CardType {
	for _, cardType := range cardTypes {
		if cardType.id == id {
			return cardType
		}
	}
	return nil
}

func GetCardTypeByCode(code string) *CardType {
	for _, cardType := range cardTypes {
		if cardType.cardCode == code {
			return cardType
		}
	}
	return nil
}

func CreateCardType(code string, money int64) (err error) {
	queryString := "INSERT INTO card_type (card_code, money) VALUES ($1,$2) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, code, money)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	cardType := &CardType{
		id:       id,
		cardCode: code,
		money:    money,
	}
	cardTypes = append(cardTypes, cardType)
	return nil
}

func hideString(originString string) string {
	return fmt.Sprintf("%sxxxxx", originString[:len(originString)-5])
}
