package wheel

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/language"
)

const (
	WHEEL_GAME_CODE = "wheel"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
}

type WheelGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter int64
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*WheelMatch
	mapPlayerIdToHistory map[int64]*cardgame.SizedList
	// map người chơi đã nhận vòng quay trong khung giờ hiện tại chưa,
	// reset theo khung thời gian
	mapPlayerIdToIsReceivedSpin map[int64]bool
	isReadyToGiveFreeSpin       bool

	ChanActionReceiver chan *Action

	mutex sync.RWMutex
}

func NewWheelGame(currencyType string) *WheelGame {
	wheelG := &WheelGame{
		gameCode:     WHEEL_GAME_CODE,
		currencyType: currencyType,

		matchCounter:                0,
		mapPlayerIdToMatch:          map[int64]*WheelMatch{},
		mapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		mapPlayerIdToIsReceivedSpin: map[int64]bool{},

		ChanActionReceiver: make(chan *Action),
	}

	if wheelG.currencyType == currency.Money {
		wheelG.tax = 0
	} else {
		wheelG.tax = 0
	}

	go LoopResetFreeSpin(wheelG)
	go LoopReceiveActions(wheelG)

	return wheelG
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
						isPlayingXocdia := false
						isPlayingOtherGamemini := false
						if isPlayingXocdia || isPlayingOtherGamemini {
							action.chanResponse <- &ActionResponse{err: errors.New("Cant play wheel when playing xocdia2 or other gamemini")}
						} else {
							wheelG.mutex.RLock()
							_, isPlayingAWheelMatch := wheelG.mapPlayerIdToMatch[action.playerId]
							wheelG.mutex.RUnlock()
							if isPlayingAWheelMatch {
								action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0007))}
							} else {
								if playerObj.GetMoney(currency.WheelSpin) < 1 {
									action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0008))}
								} else {
									playerObj.DecreaseMoney(1, currency.WheelSpin, true)
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
						}
					} else if action.actionName == ACTION_GET_HISTORY {
						var listResultJson []string
						wheelG.mutex.RLock()
						if _, isIn := wheelG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = wheelG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						isReceivedFreeSpin := !wheelG.isReadyToGiveFreeSpin || wheelG.mapPlayerIdToIsReceivedSpin[action.playerId]
						wheelG.mutex.RUnlock()
						data := map[string]interface{}{
							"listResultJson":     listResultJson,
							"isReceivedFreeSpin": isReceivedFreeSpin,
						}
						wheelG.SendDataToPlayerId(
							"WheelHistory",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else if action.actionName == ACTION_RECEIVE_FREE_SPIN {
						wheelG.mutex.RLock()
						isReceived := wheelG.mapPlayerIdToIsReceivedSpin[action.playerId]
						wheelG.mutex.RUnlock()
						if wheelG.isReadyToGiveFreeSpin && !isReceived {
							wheelG.mutex.Lock()
							wheelG.mapPlayerIdToIsReceivedSpin[action.playerId] = true
							wheelG.mutex.Unlock()
							playerObj.IncreaseMoney(1, currency.WheelSpin, true)
							action.chanResponse <- &ActionResponse{err: nil}
						} else {
							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0009))}
						}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New("wrong action")}
					}
				}
			}(wheelG, action)
		}
	}
}

// reset nhận vòng quay theo khung thời gian
func LoopResetFreeSpin(wheelG *WheelGame) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()
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
		"tax":          wheelG.tax,

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
