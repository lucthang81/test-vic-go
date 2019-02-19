package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	//	"github.com/vic/vic_go/models/gamemini/slotacp/slotacpconfig"
	"github.com/vic/vic_go/models/otp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/zconfig"
)

const (
	// boolean redis var, "true" means already got bonus, reset once per day
	KEY_AGENCY_WEEKLY_BONUS  = "KEY_AGENCY_WEEKLY_BONUS"
	KEY_AGENCY_MONTHLY_BONUS = "KEY_AGENCY_MONTHLY_BONUS"
	TRUE_S                   = "true"
)

var Promotions []PromotionData

func init() {
	// Promotions
	Promotions = make([]PromotionData, 0)
	//	for _, wd := range []time.Weekday{
	//		time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday} {
	//		Promotions = append(Promotions, PromotionData{
	//			Weekday:         wd,
	//			PromotedRate:    0.2,
	//			ClockHour:       11,
	//			DurationInHours: 2,
	//		})
	//		Promotions = append(Promotions, PromotionData{
	//			Weekday:         wd,
	//			PromotedRate:    0.2,
	//			ClockHour:       21,
	//			DurationInHours: 1,
	//		})
	//	}
	//	for _, wd := range []time.Weekday{time.Saturday, time.Sunday} {
	//		Promotions = append(Promotions, PromotionData{
	//			Weekday:         wd,
	//			PromotedRate:    0.3,
	//			ClockHour:       21,
	//			DurationInHours: 2,
	//		})
	//		if wd == time.Sunday {
	//			Promotions = append(Promotions, PromotionData{
	//				Weekday:         wd,
	//				PromotedRate:    0.3,
	//				ClockHour:       8,
	//				DurationInHours: 2,
	//			})
	//		}
	//	}
	//	for i, _ := range Promotions {
	//		Promotions[i].PushMessage = fmt.Sprintf(
	//			"%v Đừng quên từ %vh đến %vh vào %v và nạp tiền "+
	//				"để nhận thêm %v%% Kim %v! Chúc may mắn!",
	//			zconfig.Icon1,
	//			Promotions[i].ClockHour,
	//			Promotions[i].ClockHour+Promotions[i].DurationInHours,
	//			zmisc.CLIENT_NAME_PH,
	//			int(Promotions[i].PromotedRate*100),
	//			zconfig.Icon1,
	//		)
	//	}

	// for  imported and not used
	_, _ = json.Marshal([]int{})
	_ = sql.DB{}
	fmt.Print("")
	_ = time.Now()
	_ = rand.Int63n(100)
	_ = zconfig.SV_01
	_ = otp.SmsChargingMoneyRate
	_ = zmisc.PushAll
}

type PromotionData struct {
	Weekday         time.Weekday
	PromotedRate    float64
	ClockHour       int
	ClockMinute     int
	ClockSecond     int
	DurationInHours int
	PushMessage     string
}

//
func GetPushOfflineData(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	r := map[string]interface{}{
		"ClientNameString": zmisc.CLIENT_NAME_PH,
		"List":             Promotions,
	}
	return r, nil
}

// delete rows from table currency_record, keep nRow = 100 * 10^6
// exe time is 03:00:00 local
func PeriodicallyDeleteCurrencyRecord() {
	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	var alarm <-chan time.Time
	durTo0am := 24*time.Hour - oddDuration
	alarm = time.After(durTo0am + 3*time.Hour)
	<-alarm

	for {
		nRow := int64(1000000000)

		query := "SELECT id FROM currency_record  ORDER BY id DESC LIMIT 1 "
		row := dataCenter.Db().QueryRow(query)
		var lastId int64
		err := row.Scan(&lastId)
		if err != nil {
			fmt.Println("Error PeriodicallyDeleteCurrencyRecord 1 ", err)
		} else {
			query1 := "DELETE FROM currency_record WHERE id<$1 "
			sqlResult, err1 := dataCenter.Db().Exec(query1, lastId-nRow)
			if err1 != nil {
				fmt.Println("Error PeriodicallyDeleteCurrencyRecord 2 ", err1)
			}
			fmt.Println(time.Now(), "Delete rows from currency_record successfully.", sqlResult)
		}
		// sleep to the next day
		time.Sleep(24 * time.Hour)
	}
}

// user login 3 ngày liên tiếp được 5000 test_money,
// user login 7 ngày liên tiếp được 15000 test_money,
// lưu ở bảng player_logins_track
// mỗi lần user login gọi hàm UpdateLoginsTrack
// ở hàm này duyệt qua cả bảng:
//     nếu is_logged_in_today == false thì xóa row,
//     nếu (n_continuous_logins >= 3) && (now - last_gift3_time >= 3)
//	       thì gửi tin báo người dùng nhận quà
//     nếu (n_continuous_logins >= 7) && (now - last_gift7_time >= 7)
//	       thì gửi tin báo người dùng nhận quà
//     cuối cùng đặt lại cả bảng is_logged_in_today = false
// nhận quà bằng funcs ReceiveLoginsGift3, ReceiveLoginsGift7
func PeriodicallyGiveFreeMoney() {
	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	var alarm <-chan time.Time
	durTo0am := 24*time.Hour - oddDuration
	alarm = time.After(durTo0am)
	<-alarm
	//
	for {
		query := "SELECT player_id, n_continuous_logins, is_logged_in_today, " +
			"    last_gift3_time, last_gift7_time " +
			"FROM player_logins_track "
		rows, err := dataCenter.Db().Query(query)
		if err != nil {
			s := fmt.Sprintf("ERROR PeriodicallyGiveFreeMoney 1:  %v", err)
			fmt.Println(s)
		} else {
			for rows.Next() {
				var player_id, n_continuous_logins int64
				var is_logged_in_today bool
				var last_gift3_time, last_gift7_time time.Time
				err = rows.Scan(&player_id, &n_continuous_logins,
					&is_logged_in_today, &last_gift3_time, &last_gift7_time,
				)
				if err != nil {
					s := fmt.Sprintf("ERROR PeriodicallyGiveFreeMoney 2:  %v", err)
					fmt.Println(s)
				} else {
					if is_logged_in_today == false {
						query1 := "DELETE FROM player_logins_track " +
							"WHERE player_id=$1 "
						dataCenter.Db().Exec(query1, player_id)
					} else {
						playerObj, _ := player.GetPlayer(player_id)
						if playerObj != nil {
							durFromLastGift3 := time.Now().Sub(last_gift3_time)
							if n_continuous_logins >= 3 &&
								durFromLastGift3 >= 3*24*time.Hour {
								playerObj.CreateReactingMessage(
									"Bạn hãy nhận quà đăng nhập 3 ngày liền", "",
									map[string]interface{}{
										"ReceiveLoginsGift3": map[string]interface{}{}},
								)
							}
							durFromLastGift7 := time.Now().Sub(last_gift7_time)
							if n_continuous_logins >= 7 &&
								durFromLastGift7 >= 7*24*time.Hour {
								playerObj.CreateReactingMessage(
									"Bạn hãy nhận quà đăng nhập 7 ngày liền", "",
									map[string]interface{}{
										"ReceiveLoginsGift7": map[string]interface{}{}},
								)
							}
						}
						query1 := "UPDATE player_logins_track " +
							"SET is_logged_in_today = FALSE " +
							"WHERE player_id=$1 "
						dataCenter.Db().Exec(query1, player_id)
					}
				}
			}
			rows.Close()
		}
		// sleep to the next day
		time.Sleep(24 * time.Hour)
	}
}

// phục vụ tặng quà đăng nhập liên tục,
// mỗi lần user login gọi hàm UpdateLoginsTrack:
//     nếu chưa có thì tạo row,
//     else if is_logged_in_today == false:
//         n_continuous_logins += 1
//         is_logged_in_today = true
//     else do nothing
func UpdateLoginsTrack(pid int64) {
	q1 := "INSERT INTO player_logins_track " +
		"(player_id, n_continuous_logins, is_logged_in_today) " +
		"VALUES ($1, $2, $3) "
	_, _ = dataCenter.Db().Exec(q1, pid, 1, true)
	q2 := "UPDATE player_logins_track " +
		"SET n_continuous_logins = n_continuous_logins + 1, " +
		"    is_logged_in_today = TRUE " +
		"WHERE player_id = $1 AND is_logged_in_today = FALSE"
	_, err := dataCenter.Db().Exec(q2, pid)
	if err != nil {
		fmt.Println("ERROR UpdateLoginsTrack", err)
	}
}

// giftType in [3, 7]
// đủ đk thì cộng tiền và update player_logins_track.last_gift3_time
func ReceiveLoginsGift(giftType int64, player_id int64) error {
	playerObj, err := player.GetPlayer(player_id)
	if playerObj != nil {
		query := "SELECT player_id, n_continuous_logins, is_logged_in_today, " +
			"    last_gift3_time, last_gift7_time " +
			"FROM player_logins_track WHERE player_id=$1"
		row := dataCenter.Db().QueryRow(query, player_id)
		var player_id, n_continuous_logins int64
		var is_logged_in_today bool
		var last_gift3_time, last_gift7_time time.Time
		err = row.Scan(&player_id, &n_continuous_logins,
			&is_logged_in_today, &last_gift3_time, &last_gift7_time,
		)
		if err != nil {
			return err
		}
		var last_gift_time time.Time
		var colName string
		if giftType == 3 {
			last_gift_time = last_gift3_time
			colName = "last_gift3_time"
		} else {
			last_gift_time = last_gift7_time
			colName = "last_gift7_time"
		}
		durFromLastGift := time.Now().Sub(last_gift_time)
		if n_continuous_logins >= giftType &&
			durFromLastGift >= time.Duration(giftType)*24*time.Hour {
			amount := int64(5000)
			playerObj.CreateRawMessage(fmt.Sprintf(
				"Bạn đã nhận quà đăng nhập %v ngày liền %v Xu", giftType, amount),
				"")
			q2 := fmt.Sprintf("UPDATE player_logins_track "+
				"SET %v = $1 "+
				"WHERE player_id = $2", colName)
			_, err := dataCenter.Db().Exec(q2, time.Now().UTC(), player_id)
			if err != nil {
				return err
			}
			playerObj.ChangeMoneyAndLog(
				amount, currency.Money, false, "",
				record.ACTION_GIFT_LOGINS, "", "")
			return nil
		} else {
			return errors.New("Không đủ điều kiện nhận quà")
		}
	} else {
		return err
	}
}

//  hour, minute, second as same as now
func GetNextMonday() time.Time {
	now := time.Now()
	var nextMonday time.Time
	for i := 1; i < 8; i++ {
		temp := now.Add(time.Duration(i) * 24 * time.Hour)
		if temp.Weekday() == time.Monday {
			nextMonday = temp
			break
		}
	}
	return nextMonday
}

//  hour, minute, second as same as now
func GetNextMM01() time.Time {
	now := time.Now()
	var result time.Time
	for i := 1; i < 32; i++ {
		temp := now.Add(time.Duration(i) * 24 * time.Hour)
		if temp.Day() == 1 {
			result = temp
			break
		}
	}
	return result
}

//
// trả thường theo thứ hạng, phát số weeklyEvents: EVENT_EARNING_TEST_MONEY
// 1 interval per 5*time.Minute
func PeriodicallyPayEventPrize() {
	for {
		query := "SELECT id, event_name, starting_time, finishing_time, " +
			"    map_position_to_prize, full_order " +
			"FROM event_top_result WHERE is_paid=FALSE "
		var rowId int64
		var event_name string
		var starting_time, finishing_time time.Time
		var map_position_to_prizeS, full_orderS string
		var full_order []top.TopRow
		rows, err := dataCenter.Db().Query(query)
		if err != nil {
			fmt.Println("ERROR PeriodicallyPayEventPrize 0 ", err)
		} else {
			for rows.Next() {
				err := rows.Scan(&rowId, &event_name,
					&starting_time, &finishing_time,
					&map_position_to_prizeS, &full_orderS)
				if err != nil {
					fmt.Println("ERROR PeriodicallyPayEventPrize 1 ", err)
				} else {
					map_position_to_prize := map[int]int64{}
					err1 := json.Unmarshal([]byte(map_position_to_prizeS),
						&map_position_to_prize)
					err2 := json.Unmarshal([]byte(full_orderS),
						&full_order)
					if err1 != nil && err2 != nil {
						fmt.Println("ERROR PeriodicallyPayEventPrize 2 ", err1, " | ", err2)
					} else {
						for i, r := range full_order {
							vnName := event_name
							if event_name == top.EVENT_EARNING_TEST_MONEY {
								vnName = "trùm xu"
							}
							if prize, isIn := map_position_to_prize[i]; isIn {
								pObj, _ := player.GetPlayer(r.PlayerId)
								if pObj != nil {
									pObj.ChangeMoneyAndLog(
										prize, currency.Money, false, "",
										record.ACTION_EVENT_PAY_TOPS, "", "")
									pObj.CreateRawMessage(
										"Thưởng event đua top",
										fmt.Sprintf("Bạn được thưởng %v Kim vì đứng hạng %v tại sự kiện %v.",
											prize, i+1, vnName),
									)
								}
							}
							if event_name == top.EVENT_EARNING_TEST_MONEY {
								if i < 10 {
									CreateLuckyNumber(r.PlayerId, rand.Int63n(100),
										time.Now(), 1000000)
								}
							}
						}
					}
					q1 := "UPDATE event_top_result SET is_paid=TRUE WHERE id=$1"
					dataCenter.Db().Exec(q1, rowId)
				}
			}
		}
		//
		time.Sleep(5 * time.Minute)
	}
}

//
func CreateLuckyNumber(
	playerId int64, luckyNumber int64, validDate time.Time, prize int64) {
	validDateStr := GetDateStr(validDate)
	query := "INSERT INTO player_lucky_number " +
		"(player_id, number, valid_date, prize) " +
		"VALUES ($1,$2,$3,$4) "
	_, e := dataCenter.Db().Exec(query, playerId, luckyNumber, validDateStr, prize)
	if e == nil {
		playerObj, _ := player.GetPlayer(playerId)
		if playerObj != nil {
			playerObj.CreateRawMessage(
				fmt.Sprintf("Bạn đã được số may mắn %v", luckyNumber),
				fmt.Sprintf("Số này nếu trùng với hai số cuối kết quả xổ số miền Bắc "+
					"chiều thứ hai bạn sẽ được thưởng %v Kim", prize))
		}
	}
}

// nhập kết quả xổ số và cộng tiền cho người chơi trúng
func SetResultLuckyNumber(luckyNumber int64, validDateStr string) {
	query := "SELECT id, player_id, prize " +
		"FROM player_lucky_number " +
		"WHERE valid_date=$1 AND number=$2 "
	rows, err := dataCenter.Db().Query(query, validDateStr, luckyNumber)
	if err != nil {
		fmt.Println("ERROR SetResultLuckyNumber", err)
		return
	}
	for rows.Next() {
		var rowId, playerId, prize sql.NullInt64
		err := rows.Scan(&rowId, &playerId, &prize)
		if err != nil {
			fmt.Println("ERROR SetResultLuckyNumber 1", err)
			continue
		}
		playerObj, _ := player.GetPlayer(playerId.Int64)
		if playerObj != nil {
			playerObj.ChangeMoneyAndLog(
				prize.Int64, currency.Money, false, "",
				record.ACTION_EVENT_PAY_LUCKY_NUMBER, "", "")
			playerObj.CreateRawMessage(
				fmt.Sprintf("Bạn đã trúng thưởng từ số may mắn %v", luckyNumber),
				fmt.Sprintf("Bạn đã được thưởng %v Kim", prize))
			q1 := "UPDATE player_lucky_number SET is_hit=TRUE WHERE id=$1"
			dataCenter.Db().Exec(q1, rowId.Int64)
		}
	}
	rows.Close()
}

// EVENT_HOURLY_SLOTACP
func PeriodicallyDoHourlyTasks() {
	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second))
	nextHour0000 := now.Add(-oddDuration).Add(1 * time.Hour)
	//
	durToNextHour0000 := nextHour0000.Sub(now)
	alarm := time.After(durToNextHour0000)
	<-alarm
	// loop do hourly tasks
	for {
		if true {
			// 4 tiếng reset ghép tranh slotacp 1 lần
			for _, eventName := range []string{
				event_player.EVENT_HOURLY_SLOTACP_1,
				event_player.EVENT_HOURLY_SLOTACP_50,
				event_player.EVENT_HOURLY_SLOTACP_100,
				event_player.EVENT_HOURLY_SLOTACP_250,
				event_player.EVENT_HOURLY_SLOTACP_500,
				event_player.EVENT_HOURLY_SLOTACP_1000,
				event_player.EVENT_HOURLY_SLOTACP_2500,
				event_player.EVENT_HOURLY_SLOTACP_5000,
				event_player.EVENT_HOURLY_SLOTACP_10000,
			} {
				event := event_player.NewEventCollectingPieces(
					eventName,
					nextHour0000,
					nextHour0000.Add(1*time.Hour-10*time.Second),
					8, 1000000000,
					1, 0.125)
				_ = event
			}
		}

		// sleep to the next hour
		nextHour0000 = nextHour0000.Add(1 * time.Hour)
		time.Sleep(1 * time.Hour)
	}
}

// EVENT_CHARGING_MONEY, EVENT_EARNING_TEST_MONEY, EVENT_EARNING_MONEY,
// EVENT_COLLECTING_PIECES
func PeriodicallyDoWeeklyTasks() {
	now := time.Now()
	nextMonday := GetNextMonday()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	nextMonday0am := nextMonday.Add(-oddDuration)
	//
	durTo0amNextMonday := nextMonday0am.Sub(now)
	alarm := time.After(durTo0amNextMonday)
	<-alarm
	// loop do weekly tasks
	for {
		for _, eventName := range []string{
			top.EVENT_EARNING_TEST_MONEY,
			top.EVENT_CHARGING_MONEY,
			top.EVENT_EARNING_MONEY,
		} {
			mapPositionToPrize := map[int]int64{}
			if eventName == top.EVENT_EARNING_TEST_MONEY {
				mapPositionToPrize = map[int]int64{
					0: 500000, 1: 200000, 2: 100000}
			}
			top.NewEventTop(
				eventName,
				nextMonday0am,
				nextMonday0am.Add(7*24*time.Hour-1*time.Minute),
				mapPositionToPrize,
			)
		}
		//
		for _, eventName := range []string{
			event_player.EVENT_COLLECTING_PIECES,
		} {
			event := event_player.NewEventCollectingPieces(
				eventName,
				nextMonday0am,
				nextMonday0am.Add(7*24*time.Hour-1*time.Minute),
				9, 3,
				3000000, 0.003333)
			_ = event
		}
		// sleep to the next week
		nextMonday0am = nextMonday0am.Add(7 * 24 * time.Hour)
		time.Sleep(7 * 24 * time.Hour)
	}
}

//
func PeriodicallyDoMonthlyTasks() {
	now := time.Now()
	nextMM01 := GetNextMM01()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	nextMM010am := nextMM01.Add(-oddDuration)
	lastDayOfMonth := nextMM010am.AddDate(0, 1, -1)
	//
	durToNextMM010am := nextMM010am.Sub(now)
	alarm := time.After(durToNextMM010am)
	<-alarm
	// loop do monthly tasks
	for {
		for _, eventName := range []string{
			event_player.EVENT_COLLECTING_PIECES_MONTHLY,
		} {
			event := event_player.NewEventCollectingPieces(
				eventName,
				nextMM010am,
				lastDayOfMonth.Add(23*time.Hour+59*time.Minute),
				9, 1,
				50000000, 0.0002)
			_ = event
		}
		// sleep to the next month
		nextMM01 = GetNextMM01()
		nextMM010am = time.Date(nextMM01.Year(), nextMM01.Month(), nextMM01.Day(),
			0, 0, 0, 0, nextMM01.Location())
		temp := lastDayOfMonth
		lastDayOfMonth = nextMM010am.AddDate(0, 1, -1)
		time.Sleep(temp.Sub(time.Now()) + 24*time.Hour)
	}
}

// charge card promotion for verified users
func PeriodicallyDoDailyTasks() {
	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	nextDay0am := now.Add(-oddDuration).Add(24 * time.Hour)
	//
	durToNextDay0am := nextDay0am.Sub(now)
	alarm := time.After(durToNextDay0am)
	<-alarm
	// loop do daily tasks
	for {
		go func() {
			time.Sleep(5 * time.Hour)
			server.SendRequestsToAll("TurnOffCardCharge", nil)
			server.SendRequestsToAll("TurnOffSmsCharge", nil)
		}()
		go func() {
			time.Sleep(17 * time.Hour)
			server.SendRequestsToAll("TurnOnCardCharge", nil)
			server.SendRequestsToAll("TurnOnSmsCharge", nil)
		}()

		// sleep to the next day
		time.Sleep(24 * time.Hour)
	}
}

func PeriodicallyPromote() {
	for {
		for i, _ := range Promotions {
			go func(i int) {
				now := time.Now()
				if Promotions[i].Weekday == now.Weekday() {
					pTime := time.Date(now.Year(), now.Month(), now.Day(),
						Promotions[i].ClockHour,
						Promotions[i].ClockMinute,
						Promotions[i].ClockSecond,
						0, now.Location())
					durToPTime := pTime.Sub(now)
					if durToPTime >= 0 {
						alarm := time.After(durToPTime)
						<-alarm
						zmisc.PushAll(Promotions[i].PushMessage)
						dailyPromotedRateForVerifiedUser = Promotions[i].PromotedRate
						temp := otp.SmsChargingMoneyRate
						otp.SmsChargingMoneyRate = temp * (1 + Promotions[i].PromotedRate)
						time.Sleep(time.Duration(Promotions[i].DurationInHours) * time.Hour)
						dailyPromotedRateForVerifiedUser = 0
						otp.SmsChargingMoneyRate = temp
					}
				}
			}(i)
		}
		time.Sleep(24 * time.Hour)
	}
}

func PeriodicallyUpdateAgencies() {
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			record.RedisSaveString(KEY_AGENCY_WEEKLY_BONUS, "")
			record.RedisSaveString(KEY_AGENCY_MONTHLY_BONUS, "")
		}
	}()
	go func() {
		AGENCY_WEEKLY_LOWER_LIMIT := int64(100000000)
		AGENCY_MONTHLY_LOWER_LIMIT := int64(1000000000)
		for {
			time.Sleep(30 * time.Minute)
			mapPidToAgency, _ := getAgencies(true, 0)
			now := time.Now()
			if now.Weekday() == time.Monday &&
				record.RedisLoadString(KEY_AGENCY_WEEKLY_BONUS) != TRUE_S {
				record.RedisSaveString(KEY_AGENCY_WEEKLY_BONUS, TRUE_S)
				for pidS, _ := range mapPidToAgency {
					pid, _ := strconv.ParseInt(pidS, 10, 64)
					ubTime := time.Date(now.Year(), now.Month(), now.Day(),
						0, 0, 0, 0, now.Location())
					lbTime := ubTime.Add(-7 * 24 * time.Hour)
					sumTransfer := calcSumTransfer(pid, lbTime, ubTime)
					pObj, _ := player.GetPlayer(pid)
					if pObj != nil {
						pObj.CreateType2Message("Tổng kết CTV Tuần",
							fmt.Sprintf("Doanh số của bạn trong tuần (từ %v đến %v) "+
								"là %v Kim",
								lbTime.Format(time.RFC3339),
								ubTime.Format(time.RFC3339),
								sumTransfer))
						if sumTransfer > AGENCY_WEEKLY_LOWER_LIMIT {
							bonus := (sumTransfer - AGENCY_WEEKLY_LOWER_LIMIT) / 10
							pObj.CreateType2Message("Tổng kết CTV Tuần",
								fmt.Sprintf(
									"Bạn được thưởng %v Kim do giao dịch trên %v Kim",
									bonus, AGENCY_WEEKLY_LOWER_LIMIT))
							pObj.ChangeMoneyAndLog(
								bonus, currency.Money, false, "",
								record.ACTION_PROMOTE_AGENCY_WEEKLY, "", "")
						}
					}
				}
			}
			if now.Day() == 1 &&
				record.RedisLoadString(KEY_AGENCY_MONTHLY_BONUS) != TRUE_S {
				record.RedisSaveString(KEY_AGENCY_MONTHLY_BONUS, TRUE_S)
				for pidS, _ := range mapPidToAgency {
					pid, _ := strconv.ParseInt(pidS, 10, 64)
					ubTime := time.Date(now.Year(), now.Month(), now.Day(),
						0, 0, 0, 0, now.Location())
					lbTime := time.Date(now.Year(), now.Month()-1, now.Day(),
						0, 0, 0, 0, now.Location())
					sumTransfer := calcSumTransfer(pid, lbTime, ubTime)
					pObj, _ := player.GetPlayer(pid)
					if pObj != nil {
						pObj.CreateType2Message("Tổng kết CTV Tháng",
							fmt.Sprintf("Doanh số của bạn trong tháng (từ %v đến %v) "+
								"là %v Kim",
								lbTime.Format(time.RFC3339),
								ubTime.Format(time.RFC3339),
								sumTransfer))
						if sumTransfer > AGENCY_MONTHLY_LOWER_LIMIT {
							bonus := (sumTransfer - AGENCY_MONTHLY_LOWER_LIMIT) / 20
							pObj.CreateType2Message("Tổng kết CTV Tháng",
								fmt.Sprintf(
									"Bạn được thưởng %v Kim do giao dịch trên %v Kim",
									bonus, AGENCY_MONTHLY_LOWER_LIMIT))
							pObj.ChangeMoneyAndLog(
								bonus, currency.Money, false, "",
								record.ACTION_PROMOTE_AGENCY_MONTHLY, "", "")
						}
					}
				}
			}
		}
	}()
}
