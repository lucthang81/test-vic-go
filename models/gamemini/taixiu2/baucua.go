package baucua

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
)

const (
	TAIXIU_GAME_CODE = "baucua"

	ACTION_GET_HISTORY       = "ACTION_GET_HISTORY"
	KEY_BAUCUA_MATCH_COUNTER = "KEY_BAUCUA_MATCH_COUNTER"
)

func init() {
	fmt.Print("")
	_ = currency.Money
}

type TaixiuGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter       int64
	mapPlayerIdToMatch map[int64]*TaixiuMatch

	SharedMatch *TaixiuMatch

	taixiuHistory cardgame.SizedList
	// full mapBet and shakingResult
	history2           cardgame.SizedList
	lastBetInfo        map[int64]map[string]int64
	lastMatchPlayerIds []int64

	// tiền hệ thống lãi tổng để hư cấu kết quả
	balance     int64
	sumUserBets int64
	// balance >= stealingRate * sumUserBets
	stealingRate float64

	ChanActionReceiver chan *Action
	ChanMatchEnded     chan bool

	mutex sync.RWMutex
}

func NewTaixiuGame(currencyType string) *TaixiuGame {
	taixiuG := &TaixiuGame{
		gameCode:     TAIXIU_GAME_CODE,
		currencyType: currencyType,

		matchCounter: int64(
			record.RedisLoadFloat64(KEY_BAUCUA_MATCH_COUNTER)),
		mapPlayerIdToMatch: map[int64]*TaixiuMatch{},

		SharedMatch:   nil,
		taixiuHistory: cardgame.NewSizedList(100),
		history2:      cardgame.NewSizedList(100),

		balance:      0,
		sumUserBets:  0,
		stealingRate: 0.03,

		ChanActionReceiver: make(chan *Action),
		ChanMatchEnded:     make(chan bool),
	}

	if taixiuG.currencyType == currency.Money {
		taixiuG.tax = 0.03
	} else {
		taixiuG.tax = 0.08
	}

	go LoopCreateNewMatch(taixiuG)
	go LoopReceiveActions(taixiuG)
	go LoopResetBalance(taixiuG)

	return taixiuG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopCreateNewMatch(taixiuG *TaixiuGame) {
	for {
		taixiuG.mutex.Lock()
		taixiuG.matchCounter += 1
		record.RedisSaveFloat64(KEY_BAUCUA_MATCH_COUNTER,
			float64(taixiuG.matchCounter))
		taixiuG.SharedMatch = NewTaixiuMatch(taixiuG)
		// gửi thông tin bắt đầu ván về cho những người vừa chơi ván trước
		var lastPids []int64
		if taixiuG.lastMatchPlayerIds != nil {
			lastPids = make([]int64, len(taixiuG.lastMatchPlayerIds))
			copy(lastPids, taixiuG.lastMatchPlayerIds)
		}
		taixiuG.mutex.Unlock()
		if lastPids != nil {
			for _, pid := range lastPids {
				taixiuG.SendDataToPlayerId(
					"BaucuaNotifyStartGame",
					map[string]interface{}{}, pid)
			}
		}
		// wait for end match,
		// data to this chan send from taixiuG.SharedMatch end match phase
		<-taixiuG.ChanMatchEnded
		//
		taixiuG.mutex.Lock()
		taixiuG.lastMatchPlayerIds = taixiuG.SharedMatch.GetAllPlayerIds()
		taixiuG.mapPlayerIdToMatch = map[int64]*TaixiuMatch{}
		taixiuG.SharedMatch = nil
		taixiuG.mutex.Unlock()
	}
}

func LoopResetBalance(taixiuG *TaixiuGame) {
	for {
		time.Sleep(30 * 24 * time.Hour)
		taixiuG.balance = 0
	}
}

func LoopReceiveActions(taixiuG *TaixiuGame) {
	for {
		action := <-taixiuG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(taixiuG *TaixiuGame, action *Action) {
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
					if action.actionName == ACTION_GET_HISTORY {
						action.chanResponse <- &ActionResponse{err: nil}
						taixiuG.mutex.RLock()
						response := make([]map[string]interface{}, len(taixiuG.history2.Elements))
						fmt.Println(taixiuG.history2.Elements)
						for i, matchDetailS := range taixiuG.history2.Elements {
							var matchDetail map[string]interface{}
							err := json.Unmarshal([]byte(matchDetailS), &matchDetail)
							if err == nil {
								fullMapBet, isOk := matchDetail["mapBetInfo"].(map[string]interface{})
								// fmt.Printf("%v %T", matchDetail["mapBetInfo"], matchDetail["mapBetInfo"])
								if isOk {
									response[i] = map[string]interface{}{
										"shakingResult": matchDetail["shakingResult"],
										"yourBet":       fullMapBet[fmt.Sprintf("%v", action.playerId)],
									}
								} else {
									response[i] = map[string]interface{}{
										"shakingResult": matchDetail["shakingResult"],
										"yourBet":       nil,
									}
								}
							}
						}
						taixiuG.mutex.RUnlock()
						taixiuG.SendDataToPlayerId(
							"BaucuaGetHistory",
							map[string]interface{}{"response": response},
							action.playerId)
					} else {
						taixiuG.mutex.RLock()
						//fmt.Println("checkPoint 1 pid ", action.playerId, action.actionName)
						match, isPlayingATaixiuMatch := taixiuG.mapPlayerIdToMatch[action.playerId]
						taixiuG.mutex.RUnlock()
						if !isPlayingATaixiuMatch {
							if taixiuG.SharedMatch == nil {
								action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
							} else {
								taixiuG.mutex.Lock()
								taixiuG.mapPlayerIdToMatch[action.playerId] = taixiuG.SharedMatch
								taixiuG.mutex.Unlock()

								taixiuG.SharedMatch.mutex.Lock()
								taixiuG.SharedMatch.players[action.playerId] = playerObj
								taixiuG.SharedMatch.mutex.Unlock()

								match = taixiuG.SharedMatch
							}
						}
						if match != nil {
							// in case: SharedMatch change to nil immediately
							// after run "taixiuG.mapPlayerIdToMatch[action.playerId] = taixiuG.SharedMatch"
							match.ChanActionReceiver <- action
						} else {
							// seldom happen
							taixiuG.mutex.Lock()
							delete(taixiuG.mapPlayerIdToMatch, action.playerId)
							taixiuG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
						}
					}
				}
			}(taixiuG, action)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (taixiuG *TaixiuGame) GetGameCode() string {
	return taixiuG.gameCode
}
func (taixiuG *TaixiuGame) GameCode() string {
	return taixiuG.gameCode
}

func (taixiuG *TaixiuGame) GetCurrencyType() string {
	return taixiuG.currencyType
}
func (taixiuG *TaixiuGame) CurrencyType() string {
	return taixiuG.currencyType
}

func (taixiuG *TaixiuGame) SerializeData() map[string]interface{} {
	result := map[string]interface{}{
		"gameCode":     taixiuG.GameCode(),
		"currencyType": taixiuG.CurrencyType(),
		"tax":          taixiuG.tax,
		"DURATION_PHASE_1_BET":    DURATION_PHASE_1_BET.Seconds(),
		"DURATION_PHASE_3_RESULT": DURATION_PHASE_3_RESULT.Seconds(),
	}
	return result
}

func (taixiuG *TaixiuGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(taixiuG *TaixiuGame, action *Action) error {
	taixiuG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (taixiuG *TaixiuGame) GetInfo(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_MATCH_INFO,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

// aaa,
func (taixiuG *TaixiuGame) AddBet(player *player.Player, selection string, moneyValue int64) error {
	action := &Action{
		actionName: ACTION_ADD_BET,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"selection":  selection,
			"moneyValue": moneyValue,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func (taixiuG *TaixiuGame) Chat(player *player.Player, message string) error {
	action := &Action{
		actionName: ACTION_CHAT,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"senderName": player.DisplayName(),
			"message":    message,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func (taixiuG *TaixiuGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}
