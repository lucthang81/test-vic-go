package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vic/vic_go/utils"
)

// 16:58:01 28-09-2017
func toHihiFormat(t time.Time) string {
	return fmt.Sprintf(
		"%02d:%02d:%02d    %02d-%02d-%04d",
		t.Hour(), t.Minute(), t.Second(),
		t.Day(), t.Month(), t.Year(),
	)
}

func GetPaymentHistory(playerId int64, limit int64, offset int64) (
	results []map[string]interface{}, total int64, err error) {
	cashoutRows := make([]map[string]interface{}, 0)
	queryString := fmt.Sprintf(
		"SELECT id, cash_out_record.player_id, action, game_code, change, currency_type, "+
			"value_before, value_after, additional_data, created_at, is_verified_by_admin, is_paid "+
			"FROM cash_out_record "+
			"WHERE player_id = %v "+
			"ORDER BY created_at DESC LIMIT %v OFFSET %v",
		playerId, limit, offset,
	)
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, player_id, change, value_before, value_after sql.NullInt64
		var action, game_code, currency_type, additional_data sql.NullString
		var is_verified_by_admin, is_paid bool
		var created_at time.Time
		err = rows.Scan(
			&id, &player_id, &action, &game_code, &change, &currency_type,
			&value_before, &value_after, &additional_data, &created_at,
			&is_verified_by_admin, &is_paid,
		)
		if err != nil {
			return nil, 0, err
		}
		var tempMap map[string]interface{}
		err := json.Unmarshal([]byte(additional_data.String), &tempMap)
		if err != nil {
			tempMap = map[string]interface{}{}
		}
		r1 := map[string]interface{}{
			"id":                   id.Int64,
			"player_id":            player_id.Int64,
			"change":               change.Int64,
			"value_before":         value_before.Int64,
			"value_after":          value_after.Int64,
			"action":               action.String,
			"game_code":            game_code.String,
			"currency_type":        currency_type.String,
			"card_provider":        tempMap["cardType"],
			"card_value":           tempMap["value"],
			"created_at":           toHihiFormat(created_at.Local()),
			"is_verified_by_admin": is_verified_by_admin,
			"is_paid":              is_paid,
		}
		// TODO: calc total
		cashoutRows = append(cashoutRows, r1)
		results = cashoutRows
	}
	return results, total, nil
}

func GetPurchaseHistory(playerId int64, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	queryString := "SELECT COUNT(id) FROM purchase_record WHERE player_id = $1 AND currency_type = 'money'" // currently hard code to show only money
	row := dataCenter.Db().QueryRow(queryString, playerId)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryString = "SELECT id, created_at, purchase_type, card_code FROM purchase_record WHERE player_id = $1 AND currency_type = 'money' ORDER BY -id LIMIT $2 OFFSET $3"
	rows, err := dataCenter.Db().Query(queryString, playerId, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var createdAt time.Time
		var purchaseType, cardCode sql.NullString
		err = rows.Scan(&id, &createdAt, &purchaseType, &cardCode)
		if err != nil {
			return nil, 0, err
		}
		data := make(map[string]interface{})
		data["id"] = id
		data["created_at"] = utils.FormatTime(createdAt)
		data["purchase_type"] = purchaseType.String
		data["card_code"] = cardCode.String
		results = append(results, data)
	}
	return results, total, nil
}

// func GetCurrentMoneyRange() (data map[string]interface{}, err error) {
// 	queryString := "SELECT COUNT(id) FROM purchase_record WHERE player_id = $1"
// 	row := dataCenter.Db().QueryRow(queryString, playerId)
// 	err = row.Scan(&total)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	queryString = "SELECT id, created_at, purchase_type, card_code FROM purchase_record WHERE player_id = $1 ORDER BY -id LIMIT $2 OFFSET $3"
// 	rows, err := dataCenter.Db().Query(queryString, playerId, limit, offset)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	defer rows.Close()

// 	results = make([]map[string]interface{}, 0)
// 	for rows.Next() {
// 		var id int64
// 		var createdAt time.Time
// 		var purchaseType, cardCode sql.NullString
// 		err = rows.Scan(&id, &createdAt, &purchaseType, &cardCode)
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		data := make(map[string]interface{})
// 		data["id"] = id
// 		data["created_at"] = utils.FormatTime(createdAt)
// 		data["purchase_type"] = purchaseType.String
// 		data["card_code"] = cardCode.String
// 		results = append(results, data)
// 	}
// 	return results, total, nil
// }
