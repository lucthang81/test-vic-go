package wheel2

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	// "github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/language"
)

const (
	WHEEL2_GAME_CODE          = "wheel2"
	DURATION_TO_GET_FREE_SPIN = 1 * time.Hour
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
}

type WheelGame struct {
	gameCode     string
	currencyType string

	matchCounter int64
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*WheelMatch
	mapPlayerIdToHistory map[int64]*cardgame.SizedList
	bigWinList           *cardgame.SizedList
	// map người chơi đã nhận vòng quay trong khung giờ hiện tại chưa,
	// reset theo khung thời gian
	mapPlayerIdToIsReceivedSpin      map[int64]bool
	mapPlayerIdToLastTimeReceiveSpin map[int64]time.Time
	mapPidToLastMoney1               map[int64]int64
	mapPidToLastMoney2               map[int64]int64
	isReadyToGiveFreeSpin            bool
	mapDeviceIdentifierToFirstPid    map[string]int64

	// tổng tiền hệ thống lãi của game wheel2, thường là số âm, reset theo giờ
	balance int64
	// 10% lợi nhuận của ngày hôm qua trên toàn server, cập nhật ngày 1 lần
	yesterday10PercentAllProfit int64

	ChanActionReceiver chan *Action

	mutex sync.RWMutex
}

func NewWheelGame(currencyType string) *WheelGame {
	wheelG := &WheelGame{
		gameCode:     WHEEL2_GAME_CODE,
		currencyType: currencyType,

		matchCounter:                     0,
		mapPlayerIdToMatch:               map[int64]*WheelMatch{},
		mapPlayerIdToHistory:             map[int64]*cardgame.SizedList{},
		mapPlayerIdToIsReceivedSpin:      map[int64]bool{},
		mapPlayerIdToLastTimeReceiveSpin: map[int64]time.Time{},
		mapPidToLastMoney1:               map[int64]int64{},
		mapPidToLastMoney2:               map[int64]int64{},

		mapDeviceIdentifierToFirstPid: map[string]int64{},

		ChanActionReceiver: make(chan *Action),
	}
	temp := cardgame.NewSizedList(20)
	wheelG.bigWinList = &temp

	go LoopUpdateYesterdayProfit(wheelG)
	go LoopResetFreeSpin(wheelG)
	go LoopReceiveActions(wheelG)
	go LoopResetBalance(wheelG)
	go LoopUpdateHourlyFreeSpin(wheelG)

	return wheelG
}

// reset nhận vòng quay theo khung thời gian
func LoopUpdateYesterdayProfit(wheelG *WheelGame) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()
	// waiting for record.db connect
	time.Sleep(5 * time.Second)
	//
	period := 24 * time.Hour
	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	// alarm := time.After(period - oddDuration)
	// <-alarm
	//
	for {
		queryString := "SELECT game_code, COUNT(id), SUM(win), SUM(lose)" +
			" FROM match_record where created_at >= $1 AND created_at <= $2 " +
			" AND currency_type = $3 " +
			" GROUP BY game_code ORDER BY game_code  "
		rows, err := record.DataCenter.Db().Query(
			queryString, now.Add(-oddDuration).UTC(),
			now.Add(-oddDuration).Add(period).UTC(), currency.Money)
		if err != nil {
			fmt.Printf("ERROR: dbPool.Query %v", err)
			return
		}
		yesterdayAllProfit := int64(0)
		for rows.Next() {
			var nMatch, win, lose sql.NullInt64
			var game_code string
			err = rows.Scan(&game_code, &nMatch, &win, &lose)
			if err != nil {
				fmt.Printf("ERROR: rows.Scan %v", err)
			}
			profit := lose.Int64 - win.Int64
			fmt.Printf("%v %v \n", game_code, profit)
			if game_code != WHEEL2_GAME_CODE {
				yesterdayAllProfit += profit
			}
		}
		fmt.Println("yesterdayAllProfit", yesterdayAllProfit)
		wheelG.yesterday10PercentAllProfit = int64(0.10 * float64(yesterdayAllProfit))

		time.Sleep(period)
	}
}

// reset balance, phục vụ phân phát tiền theo giờ
func LoopResetBalance(wheelG *WheelGame) {
	//
	period := 1 * time.Hour

	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second))
	alarm := time.After(period - oddDuration)
	<-alarm
	//
	for {
		wheelG.balance = 0
		time.Sleep(period)
	}
}

//
func LoopUpdateHourlyFreeSpin(wheelG *WheelGame) {
	//
	period := 1 * time.Minute
	for {
		wheelG.mutex.Lock()
		for pid, isReceivedSpin := range wheelG.mapPlayerIdToIsReceivedSpin {
			if isReceivedSpin {
				d1 := time.Now().Sub(wheelG.mapPlayerIdToLastTimeReceiveSpin[pid])
				playerObj, _ := player.GetPlayer(pid)
				if playerObj != nil {
					d2 := time.Now().Sub(playerObj.LastLoginTime)
					var minD time.Duration
					if d1 < d2 {
						minD = d1
					} else {
						minD = d2
					}
					isMoneyChange :=
						(playerObj.GetMoney(currency.Money) != wheelG.mapPidToLastMoney1[pid]) ||
							(playerObj.GetMoney(currency.TestMoney) != wheelG.mapPidToLastMoney2[pid])
					if DURATION_TO_GET_FREE_SPIN <= minD && isMoneyChange {
						delete(wheelG.mapPlayerIdToIsReceivedSpin, pid)
					}
				}
			}
		}
		wheelG.mutex.Unlock()
		time.Sleep(period)
	}

}

// reset
func LoopResetFreeSpin(wheelG *WheelGame) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	// hihi
	wheelG.isReadyToGiveFreeSpin = true

	//
	period := 24 * time.Hour

	now := time.Now()
	oddDuration := time.Duration(
		int64(now.Minute())*int64(time.Minute) +
			int64(now.Second())*int64(time.Second) +
			int64(now.Hour())*int64(time.Hour))
	var alarm <-chan time.Time
	if now.Hour() >= 6 {
		durTo6am := 30*time.Hour - oddDuration
		alarm = time.After(durTo6am)
	} else {
		durTo6am := 6*time.Hour - oddDuration
		alarm = time.After(durTo6am)
	}
	<-alarm
	//
	for {
		wheelG.isReadyToGiveFreeSpin = true
		wheelG.mutex.Lock()
		wheelG.mapPlayerIdToIsReceivedSpin = map[int64]bool{}
		wheelG.mutex.Unlock()
		time.Sleep(period)
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopReceiveActions(wheelG *WheelGame) {
	for {
		action := <-wheelG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(wheelG *WheelGame, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				playerObj, err := player.GetPlayer(action.playerId)
				if err != nil {
					action.chanResponse <- &ActionResponse{err: errors.New("Cant find player for this id")}
				} else {
					if action.actionName == ACTION_SPIN {
						wheelG.mutex.RLock()
						_, isPlayingAWheelMatch := wheelG.mapPlayerIdToMatch[action.playerId]
						wheelG.mutex.RUnlock()
						if isPlayingAWheelMatch {
							action.chanResponse <- &ActionResponse{err: errors.New("U r playing a wheel match")}
						} else {
							if playerObj.GetMoney(currency.Wheel2Spin) < 1 {
								action.chanResponse <- &ActionResponse{err: errors.New("Dont have free spin")}
							} else {
								playerObj.ChangeMoneyAndLog(
									-1, currency.Wheel2Spin, false, "",
									record.ACTION_SPIN, "", "")
								wheelG.mutex.Lock()
								wheelG.matchCounter += 1
								wheelG.mutex.Unlock()
								newMatch := NewWheelMatch(
									wheelG,
									playerObj,
									wheelG.matchCounter,
								)
								wheelG.mutex.Lock()
								wheelG.mapPlayerIdToMatch[action.playerId] = newMatch
								wheelG.mutex.Unlock()
								action.chanResponse <- &ActionResponse{err: nil}
							}
						}
					} else if action.actionName == ACTION_GET_HISTORY {
						wheelG.mutex.RLock()
						isReceivedFreeSpin := !wheelG.isReadyToGiveFreeSpin || wheelG.mapPlayerIdToIsReceivedSpin[action.playerId]
						var listResultJson []string
						if _, isIn := wheelG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = wheelG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						var remainingDurationToGetFreeSpinInseconds float64
						if wheelG.mapPlayerIdToIsReceivedSpin[action.playerId] == false {
							remainingDurationToGetFreeSpinInseconds = 0
						} else {
							d1 := time.Now().Sub(wheelG.mapPlayerIdToLastTimeReceiveSpin[action.playerId])
							d2 := time.Now().Sub(playerObj.LastLoginTime)
							var minD time.Duration
							if d1 < d2 {
								minD = d1
							} else {
								minD = d2
							}
							remainingDurationToGetFreeSpinInseconds = (DURATION_TO_GET_FREE_SPIN - minD).Seconds()
						}
						wheelG.mutex.RUnlock()
						_ = isReceivedFreeSpin
						data := map[string]interface{}{
							"listResultJson": listResultJson,
							"bigWinList":     wheelG.bigWinList.Elements,
							// "isReceivedFreeSpin":                      isReceivedFreeSpin,
							"isReceivedFreeSpin":                      true,
							"remainingDurationToGetFreeSpinInseconds": remainingDurationToGetFreeSpinInseconds,
						}
						wheelG.SendDataToPlayerId(
							"Wheel2History",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else if action.actionName == ACTION_RECEIVE_FREE_SPIN {
						wheelG.mutex.Lock()
						isDuplicateDi := false
						di := playerObj.DeviceIdentifier()
						if di == "" {

						} else {
							if fpid, isIn := wheelG.mapDeviceIdentifierToFirstPid[di]; isIn {
								if fpid == action.playerId {

								} else {
									isDuplicateDi = true
								}
							} else {
								wheelG.mapDeviceIdentifierToFirstPid[di] = action.playerId
							}
						}
						wheelG.mutex.Unlock()
						if isDuplicateDi {
							action.chanResponse <- &ActionResponse{err: errors.New("bạn đã nhận vòng quay miễn phí rồi")}
						} else {
							wheelG.mutex.RLock()
							isReceived := wheelG.mapPlayerIdToIsReceivedSpin[action.playerId]
							wheelG.mutex.RUnlock()
							if wheelG.isReadyToGiveFreeSpin && !isReceived {
								playerObj.ChangeMoneyAndLog(
									1, currency.Wheel2Spin, false, "",
									record.ACTION_DAILY_FREE_SPIN, "", "")
								wheelG.mutex.Lock()
								wheelG.mapPlayerIdToIsReceivedSpin[action.playerId] = true
								wheelG.mapPlayerIdToLastTimeReceiveSpin[action.playerId] = time.Now()
								wheelG.mapPidToLastMoney1[action.playerId] =
									playerObj.GetMoney(currency.Money)
								wheelG.mapPidToLastMoney2[action.playerId] =
									playerObj.GetMoney(currency.TestMoney)
								wheelG.mutex.Unlock()
								action.chanResponse <- &ActionResponse{err: nil}
							} else {
								action.chanResponse <- &ActionResponse{err: errors.New("bạn đã nhận vòng quay miễn phí rồi, cần chờ lần kế tiếp")}
							}
						}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New("wrong action")}
					}
				}
			}(wheelG, action)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (wheelG *WheelGame) GetGameCode() string {
	return wheelG.gameCode
}
func (wheelG *WheelGame) GameCode() string {
	return wheelG.gameCode
}

func (wheelG *WheelGame) GetCurrencyType() string {
	return wheelG.currencyType
}
func (wheelG *WheelGame) CurrencyType() string {
	return wheelG.currencyType
}

func (wheelG *WheelGame) SerializeData() map[string]interface{} {
	result := map[string]interface{}{
		"gameCode":     wheelG.gameCode,
		"currencyType": wheelG.currencyType,

		"SYMBOLS":  SYMBOLS,
		"SYMBOLS1": SYMBOLS1,
	}
	return result
}

func (wheelG *WheelGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(wheelG *WheelGame, action *Action) error {
	wheelG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (wheelG *WheelGame) ReceiveFreeSpin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_RECEIVE_FREE_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(wheelG, action)
}

// aaa,
func (wheelG *WheelGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(wheelG, action)
}

// aaa,
func (wheelG *WheelGame) Spin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(wheelG, action)
}
