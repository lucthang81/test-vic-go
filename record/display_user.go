package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	// "github.com/vic/vic_go/utils"
	"time"
)

func GetCohortData(startDate time.Time, endDate time.Time) map[string]interface{} {
	var queryString string
	var row *sql.Row
	var err error
	var data map[string]interface{}
	data = make(map[string]interface{})

	eachDaysData := make([]map[string]interface{}, 0)
	loopRange := make([]int, 0)

	beginFirstDayTime := utils.StartOfDayFromTime(startDate)
	beginLastDayTime := utils.StartOfDayFromTime(endDate)
	timeRange := int(beginLastDayTime.Sub(beginFirstDayTime).Hours() / 24)

	for i := 0; i < timeRange; i++ {
		loopRange = append(loopRange, i+1)
		currentDate := startDate.Add(time.Duration(i) * 86400 * time.Second)
		beginDayTime := utils.StartOfDayFromTime(currentDate)
		endDayTime := utils.EndOfDayFromTime(currentDate)

		if beginDayTime.After(endDate.UTC()) {
			return data
		}
		dateString, _ := utils.FormatTimeToVietnamTime(beginDayTime)
		dayData := make(map[string]interface{})
		dayData["date_string"] = dateString

		queryString = "SELECT COUNT(id) FROM player WHERE player_type != 'bot' AND created_at >= $1 AND created_at <= $2"
		row = dataCenter.Db().QueryRow(queryString, beginDayTime.UTC(), endDayTime.UTC())
		var cohortSizeSql sql.NullInt64
		err = row.Scan(&cohortSizeSql)
		if err != nil {
			log.LogSerious("Error get cohort data %v", err)
			return nil
		}

		cohortSize := cohortSizeSql.Int64
		cohortsData := make([]map[string]interface{}, 0)
		dayData["cohort_size"] = cohortSize
		if cohortSize == 0 {
			for j := 1; j < timeRange-i; j++ {
				cohortData := make(map[string]interface{})
				cohortData["percent"] = "0.00%"
				cohortData["alpha"] = 0
				cohortsData = append(cohortsData, cohortData)
			}
			dayData["cohort"] = cohortsData
		} else {
			for j := 1; j < timeRange-i; j++ {
				afterInstallDate := currentDate.Add(time.Duration(j) * 86400 * time.Second)
				afterInstallBeginDate := utils.StartOfDayFromTime(afterInstallDate)
				afterInstallEndDate := utils.EndOfDayFromTime(afterInstallDate)

				queryString = "SELECT COUNT(*) FROM (SELECT DISTINCT record.player_id " +
					"FROM active_record as record " +
					"WHERE record.start_date >= $1 AND record.start_date <= $2 " +
					"AND record.player_id IN (SELECT player.id FROM player as player WHERE player.player_type != 'bot' AND player.created_at >= $3 AND player.created_at <= $4)) as temp"
				row = dataCenter.Db().QueryRow(queryString, afterInstallBeginDate.UTC(), afterInstallEndDate.UTC(), beginDayTime.UTC(), endDayTime.UTC())
				var activeCountSql sql.NullInt64
				err = row.Scan(&activeCountSql)
				if err != nil {
					log.LogSerious("Error get cohort data %v", err)
					return nil
				}

				percent := float64(activeCountSql.Int64) / float64(cohortSize) * 100
				cohortData := make(map[string]interface{})
				cohortData["percent"] = fmt.Sprintf("%.2f%%", percent)
				cohortData["alpha"] = percent / 100
				cohortsData = append(cohortsData, cohortData)
			}
			dayData["cohort"] = cohortsData
		}

		eachDaysData = append(eachDaysData, dayData)
	}
	data["range"] = timeRange
	data["days"] = eachDaysData
	data["loop_range"] = loopRange

	return data
}

func GetDAU(startDateParams time.Time, endDateParams time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfDayFromTime(startDateParams)
	endDate := utils.EndOfDayFromTime(endDateParams)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := timeRangeStart.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		counter := dataCenter.GetInt64FromQuery("SELECT COUNT(*) FROM "+
			"(SELECT DISTINCT player_id FROM active_record WHERE start_date >= $1 AND start_date <= $2 "+
			"AND player_id IN (SELECT id FROM player where player_type != 'bot')) as temp ", timeRangeStart.UTC(), timeRangeEnd.UTC())

		dayData := make(map[string]interface{})
		dayData["count"] = counter
		dayData["date"], _ = utils.FormatTimeToVietnamTime(timeRangeStart)
		results = append(results, dayData)
		timeRangeStart = timeRangeStart.Add(24 * time.Hour)
	}
	data["results"] = results
	return
}

func GetMAU(startDateParams time.Time, endDateParams time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfMonthFromTime(startDateParams)
	endDate := utils.EndOfMonthFromTime(endDateParams)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := utils.EndOfMonthFromTime(timeRangeStart)
		counter := dataCenter.GetInt64FromQuery("SELECT COUNT(*) FROM "+
			"(SELECT DISTINCT player_id FROM active_record WHERE start_date >= $1 AND start_date <= $2 "+
			"AND player_id IN (SELECT id FROM player where player_type != 'bot')) as temp ", timeRangeStart.UTC(), timeRangeEnd.UTC())

		monthData := make(map[string]interface{})
		monthData["count"] = counter
		monthData["date"] = utils.FormatTimeToVietnamTimeMonthYear(timeRangeStart)
		results = append(results, monthData)
		timeRangeStart = utils.StartOfMonthFromTime(timeRangeEnd.Add(24 * time.Hour))
	}
	data["results"] = results
	return
}

func GetHAU(date time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfDayFromTime(date)
	endDate := utils.EndOfDayFromTime(date)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := timeRangeStart.Add(59*time.Minute + 59*time.Second)
		counter := dataCenter.GetInt64FromQuery("SELECT COUNT(*) FROM "+
			"(SELECT DISTINCT player_id FROM active_record WHERE start_date >= $1 AND start_date <= $2 "+
			"AND player_id IN (SELECT id FROM player where player_type != 'bot')) as temp ", timeRangeStart.UTC(), timeRangeEnd.UTC())

		monthData := make(map[string]interface{})
		monthData["count"] = counter
		_, monthData["date"] = utils.FormatTimeToVietnamTime(timeRangeStart)
		results = append(results, monthData)
		timeRangeStart = timeRangeStart.Add(1 * time.Hour)
	}
	data["results"] = results
	return
}

func GetNRU(startDateParams time.Time, endDateParams time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate := utils.StartOfDayFromTime(startDateParams)
	endDate := utils.EndOfDayFromTime(endDateParams)

	results := make([]map[string]interface{}, 0)
	timeRangeStart := startDate
	for timeRangeStart.Before(endDate) {
		timeRangeEnd := timeRangeStart.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		counter := dataCenter.GetInt64FromQuery("SELECT COUNT(id) FROM player"+
			" WHERE player_type != 'bot' AND created_at >= $1 AND created_at <= $2", timeRangeStart.UTC(), timeRangeEnd.UTC())

		dayData := make(map[string]interface{})
		dayData["count"] = counter
		dayData["date"], _ = utils.FormatTimeToVietnamTime(timeRangeStart)
		results = append(results, dayData)
		timeRangeStart = timeRangeStart.Add(24 * time.Hour)
	}
	data["results"] = results
	return
}

func GetCCU(startDate time.Time, endDate time.Time) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	startDate = utils.StartOfDayFromTime(startDate)
	endDate = utils.EndOfDayFromTime(endDate)
	queryString := "SELECT online_total_count, online_bot_count, online_normal_count, game_online_data, created_at " +
		"FROM ccu_record WHERE created_at >= $1 AND created_at <= $2 ORDER BY -id"
	rows, err := dataCenter.Db().Query(queryString, startDate.UTC(), endDate.UTC())
	if err != nil {
		return
	}
	defer rows.Close()
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var onlineTotalCount, onlineBotCount, onlineNormalCount sql.NullInt64
		var gameOnlineDataRaw []byte
		var createdAt time.Time
		err = rows.Scan(&onlineTotalCount, &onlineBotCount, &onlineNormalCount, &gameOnlineDataRaw, &createdAt)
		if err != nil {
			return
		}
		var gameOnlineData map[string]interface{}
		err = json.Unmarshal(gameOnlineDataRaw, &gameOnlineData)
		if err != nil {
			return
		}
		markData := make(map[string]interface{})
		markData["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		markData["online_total_count"] = onlineTotalCount.Int64
		markData["online_bot_count"] = onlineBotCount.Int64
		markData["online_normal_count"] = onlineNormalCount.Int64
		markData["game_online_data"] = gameOnlineData
		results = append(results, markData)
	}
	data["results"] = results
	return
}
