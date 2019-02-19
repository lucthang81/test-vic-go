package jackpot

//import (
//	"fmt"
//	"github.com/vic/vic_go/utils"
//
//	"time"
//)

//func GetJackpotRecord(gameCode string, currencyType string, limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
//	// limit the limit to 30
//	limit = utils.MinInt64(30, limit)
//
//	queryString := fmt.Sprintf("SELECT COUNT(id) FROM jackpot_winner_record WHERE currency_type = $1 AND" +
//		" code = $2")
//	row := dataCenter.Db().QueryRow(queryString, currencyType, gameCode)
//	err = row.Scan(&total)
//	if err != nil {
//		return nil, 0, err
//	}
//
//	queryString = fmt.Sprintf("SELECT record.player_id, player.username, record.tiles, record.requirement, record.value, record.created_at" +
//		" FROM jackpot_winner_record as record, player as player" +
//		" WHERE record.player_id = player.id AND record.currency_type = $1 AND record.code = $2" +
//		" ORDER BY -record.id LIMIT $3 OFFSET $4")
//	rows, err := dataCenter.Db().Query(queryString, currencyType, gameCode, limit, offset)
//	if err != nil {
//		return nil, 0, err
//	}
//	defer rows.Close()
//
//	results = make([]map[string]interface{}, 0)
//	for rows.Next() {
//		var playerId int64
//		var username string
//		var tiles string
//		var requirement int64
//		var money int64
//		var createdAt time.Time
//		err = rows.Scan(&playerId, &username, &tiles, &requirement, &money, &createdAt)
//		if err != nil {
//			return nil, 0, err
//		}
//
//		data := make(map[string]interface{})
//		data["player_id"] = playerId
//		data["username"] = username
//		data["requirement"] = requirement
//		data["money"] = money
//		data["tiles"] = tiles
//		data["created_at"] = utils.FormatTime(createdAt)
//
//		results = append(results, data)
//	}
//	return results, total, nil
//}
