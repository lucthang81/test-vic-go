package maubinh

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/utils"
	"net/http"
	"time"
)

// minah never call this func
func (gameSession *MauBinhSession) recordTypeAfterGame(matchId int64) {
	for playerId, cardsData := range gameSession.organizedCardsData {
		player := gameSession.GetPlayer(playerId)
		if player != nil {
			playerType := player.PlayerType()
			currencyType := gameSession.currencyType
			playerCards := gameSession.cards[playerId]
			whiteWinType := gameSession.game.getTypeOfWhiteWin(playerCards)
			playerCardsString := fmt.Sprintf("%v", playerCards)
			if whiteWinType == "" {
				var isValid bool
				if gameSession.game.isCardsDataValid(cardsData) {
					isValid = true
				}
				for positionString, cards := range cardsData {
					var typeString string
					if isValid {
						typeString = gameSession.game.getTypeOfCards(cards)
					} else {
						typeString = "invalid"
					}
					cardsString := fmt.Sprintf("%v", cards)
					_, err := dataCenter.Db().Exec("INSERT INTO maubinh_type_record (match_id, player_id, player_type, currency_type, type, position, cards)"+
						" VALUES ($1, $2, $3, $4, $5, $6, $7)",
						matchId,
						playerId,
						playerType,
						currencyType,
						typeString,
						positionString,
						cardsString,
					)
					if err != nil {
						log.LogSerious("err mb record type %v", err)
					}
				}
			}
			_, err := dataCenter.Db().Exec("INSERT INTO maubinh_white_win_record (match_id, player_id, player_type, currency_type, white_win_type, cards)"+
				" VALUES ($1, $2, $3, $4, $5, $6)",
				matchId,
				playerId,
				playerType,
				currencyType,
				whiteWinType,
				playerCardsString,
			)
			if err != nil {
				log.LogSerious("err mb record whitewin type %v", err)
			}
		}

	}
}

func GetTypeAdvanceRecordData(request *http.Request) map[string]interface{} {
	data := make(map[string]interface{})
	recordType := request.URL.Query().Get("record_type")
	data["record_type"] = recordType
	{
		row1 := htmlutils.NewDateField("Start", "start_date", "Start date", time.Now().Add(-24*7*time.Hour))
		row2 := htmlutils.NewDateField("End", "end_date", "End date", time.Now())
		row3 := htmlutils.NewRadioField("Query By Above Date", "use_date", "false", []string{"true", "false"})
		row4 := htmlutils.NewRadioField("Player Type", "player_type", "all", []string{"normal", "bot", "all"})
		row5 := htmlutils.NewRadioField("CurrencyType", "currency_type", "all", []string{currency.Money, currency.TestMoney, "all"})
		row6 := htmlutils.NewRadioField("Position", "position", "all", []string{TopPart, MiddlePart, BottomPart, "all"})
		row7 := htmlutils.NewStringHiddenField("record_type", "type")

		editObject := htmlutils.NewEditObjectGet([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6, row7},
			fmt.Sprintf("/admin/game/maubinh/advance_record/"))

		if recordType == "type" {
			responseData := editObject.ConvertGetRequestToData(request)
			//fmt.Println(responseData)
			editObject.UpdateEntryFromRequestData(responseData)

			query := "SELECT type, COUNT(*) from maubinh_type_record WHERE true = true"
			useDate := utils.GetStringAtPath(responseData, "use_date")
			params := make([]interface{}, 0)
			paramCount := 1
			if useDate == "true" {
				startDate := responseData["start_date"].(time.Time)
				endDate := responseData["end_date"].(time.Time)
				query = fmt.Sprintf("%v created_at >= $%d AND created_at <= $%d", query, paramCount, paramCount+1)
				paramCount += 2
				params = append(params, startDate.UTC())
				params = append(params, endDate.UTC())
			}
			playerType := utils.GetStringAtPath(responseData, "player_type")
			if playerType == "normal" || playerType == "bot" {
				query = fmt.Sprintf("%v AND player_type = $%d", query, paramCount)
				paramCount += 1
				params = append(params, playerType)
			}

			currencyType := utils.GetStringAtPath(responseData, "currency_type")
			if currencyType == currency.Money || currencyType == currency.TestMoney {
				query = fmt.Sprintf("%v AND currency_type = $%d", query, paramCount)
				paramCount += 1
				params = append(params, currencyType)
			}

			position := utils.GetStringAtPath(responseData, "position")
			if position == TopPart || position == MiddlePart || position == BottomPart {
				query = fmt.Sprintf("%v AND position = $%d", query, paramCount)
				paramCount += 1
				params = append(params, position)
			}
			query = fmt.Sprintf("%v GROUP BY type", query)
			rows, err := dataCenter.Db().Query(query, params...)
			if err != nil {
				log.LogSerious("err query type record mb %v", err)
				return data
			}
			results := make([]map[string]interface{}, 0)
			var total int64
			for rows.Next() {
				var typeString string
				var count int64
				err := rows.Scan(&typeString, &count)
				if err != nil {
					log.LogSerious("err scane query type record mb %v", err)
					continue
				}
				result := make(map[string]interface{})
				result["type"] = typeString
				result["count"] = count
				total += count
				results = append(results, result)
			}
			rows.Close()
			data["total"] = total

			headers := []string{"Type", "Count", "Percent"}
			columns := make([][]*htmlutils.TableColumn, 0)
			for _, result := range results {
				count := utils.GetInt64AtPath(result, "count")
				c1 := htmlutils.NewStringTableColumn(utils.GetStringAtPath(result, "type"))
				c2 := htmlutils.NewStringTableColumn(fmt.Sprintf("%v", count))
				percent := float64(count) / float64(total) * 100
				c3 := htmlutils.NewStringTableColumn(fmt.Sprintf("%.2f", percent))
				row := []*htmlutils.TableColumn{c1, c2, c3}
				columns = append(columns, row)
			}
			table := htmlutils.NewTableObject(headers, columns)
			data["result_table"] = table.SerializedData()
		}
		data["script_type"] = editObject.GetScriptHTML()
		data["form_type"] = editObject.GetFormHTML()
	}

	{
		row1 := htmlutils.NewDateField("Start", "start_date", "Start date", time.Now().Add(-24*7*time.Hour))
		row2 := htmlutils.NewDateField("End", "end_date", "End date", time.Now())
		row3 := htmlutils.NewRadioField("Query By Above Date", "use_date", "false", []string{"true", "false"})
		row4 := htmlutils.NewRadioField("Player Type", "player_type", "all", []string{"normal", "bot", "all"})
		row5 := htmlutils.NewRadioField("CurrencyType", "currency_type", "all", []string{currency.Money, currency.TestMoney, "all"})
		row6 := htmlutils.NewStringHiddenField("record_type", "white_win")

		editObject := htmlutils.NewEditObjectGet([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6},
			fmt.Sprintf("/admin/game/maubinh/advance_record/"))

		if recordType == "white_win" {
			responseData := editObject.ConvertGetRequestToData(request)
			editObject.UpdateEntryFromRequestData(responseData)
			query := "SELECT white_win_type, COUNT(*) from maubinh_white_win_record WHERE true = true"
			useDate := utils.GetStringAtPath(responseData, "use_date")
			params := make([]interface{}, 0)
			paramCount := 1
			if useDate == "true" {
				startDate := responseData["start_date"].(time.Time)
				endDate := responseData["end_date"].(time.Time)
				query = fmt.Sprintf("%v AND created_at >= $%d AND created_at <= $%d", query, paramCount, paramCount+1)
				paramCount += 2
				params = append(params, startDate.UTC())
				params = append(params, endDate.UTC())
			}
			playerType := utils.GetStringAtPath(responseData, "player_type")
			if playerType == "normal" || playerType == "bot" {
				query = fmt.Sprintf("%v AND player_type = $%d", query, paramCount)
				paramCount += 1
				params = append(params, playerType)
			}

			currencyType := utils.GetStringAtPath(responseData, "currency_type")
			if currencyType == currency.Money || currencyType == currency.TestMoney {
				query = fmt.Sprintf("%v AND currency_type = $%d", query, paramCount)
				paramCount += 1
				params = append(params, currencyType)
			}

			query = fmt.Sprintf("%v GROUP BY white_win_type", query)
			rows, err := dataCenter.Db().Query(query, params...)
			if err != nil {
				log.LogSerious("err query whitewintype record mb %v", err)
				return data
			}
			results := make([]map[string]interface{}, 0)
			var total int64
			for rows.Next() {
				var typeString string
				var count int64
				err := rows.Scan(&typeString, &count)
				if err != nil {
					log.LogSerious("err scane query whitewintype record mb %v", err)
					continue
				}
				result := make(map[string]interface{})
				if typeString == "" {
					typeString = "Không thắng trắng"
				}
				result["type"] = typeString
				result["count"] = count
				total += count
				results = append(results, result)
			}
			rows.Close()
			data["total"] = total

			headers := []string{"Type", "Count", "Percent"}
			columns := make([][]*htmlutils.TableColumn, 0)
			for _, result := range results {
				count := utils.GetInt64AtPath(result, "count")
				c1 := htmlutils.NewStringTableColumn(utils.GetStringAtPath(result, "type"))
				c2 := htmlutils.NewStringTableColumn(fmt.Sprintf("%v", count))
				percent := float64(count) / float64(total) * 100
				c3 := htmlutils.NewStringTableColumn(fmt.Sprintf("%.2f", percent))
				row := []*htmlutils.TableColumn{c1, c2, c3}
				columns = append(columns, row)
			}
			table := htmlutils.NewTableObject(headers, columns)
			data["result_table"] = table.SerializedData()
		}
		data["script_white_win"] = editObject.GetScriptHTML()
		data["form_white_win"] = editObject.GetFormHTML()
	}
	return data
}
