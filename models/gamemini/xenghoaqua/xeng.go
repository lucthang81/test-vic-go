package xeng

//
//import (
//	"errors"
//	"fmt"
//	"runtime/debug"
//	"sync"
//	"time"
//
//	"github.com/vic/vic_go/models/cardgame"
//	"github.com/vic/vic_go/models/currency"
//	"github.com/vic/vic_go/models/game/jackpot"
//	"github.com/vic/vic_go/models/gamemini"
//	"github.com/vic/vic_go/models/player"
//)
//
//const (
//	GAME_CODE = "xenghoaqua"
//)
//
//func init() {
//	fmt.Print("")
//	_ = currency.Money
//}
//
//type Game struct {
//	GameCode     string
//	CurrencyType string
//	MatchCounter int64
//
//	MapPidToMatch   map[int64]*Match
//	MapPidToMapBet  map[int64]map[string]int64
//	MapPidToHistory cardgame.SizedList
//	BigWinList      *cardgame.SizedList
//	Jackpot         *jackpot.Jackpot
//
//	//	ChanActionReceiver chan *Action
//
//	mutex sync.RWMutex
//}
//
//func NewGame(currencyType string) *Game {
//	G := &Game{
//		gameCode:     _GAME_CODE,
//		currencyType: currencyType,
//
//		matchCounter:       0,
//		mapPlayerIdToMatch: map[int64]*Match{},
//
//		SharedMatch: nil,
//		History:     cardgame.NewSizedList(30),
//
//		balance:      0,
//		sumUserBets:  0,
//		stealingRate: 0.07,
//
//		ChanActionReceiver: make(chan *Action),
//		ChanMatchEnded:     make(chan bool),
//	}
//
//	if G.currencyType == currency.Money {
//		G.tax = 0.03
//	} else {
//		G.tax = 0.08
//	}
//
//	go LoopCreateNewMatch(G)
//	go LoopReceiveActions(G)
//	go LoopResetBalance(G)
//
//	return G
//}
//
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//func LoopCreateNewMatch(G *Game) {
//	for {
//		G.mutex.Lock()
//		G.matchCounter += 1
//		G.SharedMatch = NewMatch(G)
//		// gửi thông tin bắt đầu ván về cho những người vừa chơi ván trước
//		var lastPids []int64
//		if G.lastMatchPlayerIds != nil {
//			lastPids = make([]int64, len(G.lastMatchPlayerIds))
//			copy(lastPids, G.lastMatchPlayerIds)
//		}
//		G.mutex.Unlock()
//		if lastPids != nil {
//			for _, pid := range lastPids {
//				G.SendDataToPlayerId(
//					"NotifyStartGame",
//					map[string]interface{}{}, pid)
//			}
//		}
//		// wait for end match,
//		// data to this chan send from G.SharedMatch end match phase
//		<-G.ChanMatchEnded
//		//
//		G.mutex.Lock()
//		G.lastMatchPlayerIds = G.SharedMatch.GetAllPlayerIds()
//		G.mapPlayerIdToMatch = map[int64]*Match{}
//		G.SharedMatch = nil
//		G.mutex.Unlock()
//	}
//}
//
//func LoopResetBalance(G *Game) {
//	for {
//		time.Sleep(24 * time.Hour)
//		G.balance = 0
//	}
//}
//
//func LoopReceiveActions(G *Game) {
//	for {
//		action := <-G.ChanActionReceiver
//		if action.actionName == ACTION_STOP_GAME {
//			action.chanResponse <- &ActionResponse{err: nil}
//			break
//		} else {
//			go func(G *Game, action *Action) {
//				defer func() {
//					if r := recover(); r != nil {
//						bytes := debug.Stack()
//						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
//					}
//				}()
//
//				playerObj, err := player.GetPlayer(action.playerId)
//				if err != nil {
//					action.chanResponse <- &ActionResponse{err: errors.New("Cant find player for this id")}
//				} else {
//					G.mutex.RLock()
//					//fmt.Println("checkPoint 1 pid ", action.playerId, action.actionName)
//					match, isPlayingAMatch := G.mapPlayerIdToMatch[action.playerId]
//					G.mutex.RUnlock()
//					if !isPlayingAMatch {
//						if G.SharedMatch == nil {
//							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
//						} else {
//							G.mutex.Lock()
//							G.mapPlayerIdToMatch[action.playerId] = G.SharedMatch
//							G.mutex.Unlock()
//
//							G.SharedMatch.mutex.Lock()
//							G.SharedMatch.players[action.playerId] = playerObj
//							G.SharedMatch.mutex.Unlock()
//
//							match = G.SharedMatch
//						}
//					}
//					if match != nil {
//						// in case: SharedMatch change to nil immediately
//						// after run "G.mapPlayerIdToMatch[action.playerId] = G.SharedMatch"
//						match.ChanActionReceiver <- action
//					} else {
//						// seldom happen
//						G.mutex.Lock()
//						delete(G.mapPlayerIdToMatch, action.playerId)
//						G.mutex.Unlock()
//						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
//					}
//
//				}
//			}(G, action)
//		}
//	}
//}
//
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//// interface
//func (G *Game) GetGameCode() string {
//	return G.gameCode
//}
//func (G *Game) GameCode() string {
//	return G.gameCode
//}
//
//func (G *Game) GetCurrencyType() string {
//	return G.currencyType
//}
//func (G *Game) CurrencyType() string {
//	return G.currencyType
//}
//
//func (G *Game) SerializeData() map[string]interface{} {
//	result := map[string]interface{}{
//		"gameCode":     G.GameCode(),
//		"currencyType": G.CurrencyType(),
//		"tax":          G.tax,
//		"DURATION_PHASE_1_BET":    DURATION_PHASE_1_BET.Seconds(),
//		"DURATION_PHASE_3_RESULT": DURATION_PHASE_3_RESULT.Seconds(),
//	}
//	return result
//}
//
//func (G *Game) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
//	gamemini.ServerObj.SendRequest(method, data, playerId)
//}
//
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////
//// gameplay funcs
//////////////////////////////////////////////////////////////////////////////////
//
//// aaa,
//func DoPlayerAction(G *Game, action *Action) error {
//	G.ChanActionReceiver <- action
//	timeout := time.After(5 * time.Second)
//	select {
//	case res := <-action.chanResponse:
//		return res.err
//	case <-timeout:
//		return errors.New(l.Get(l.M0006))
//	}
//}
//
//// aaa,
//func (G *Game) GetInfo(player *player.Player) error {
//	action := &Action{
//		actionName:   ACTION_GET_MATCH_INFO,
//		playerId:     player.Id(),
//		data:         map[string]interface{}{},
//		chanResponse: make(chan *ActionResponse),
//	}
//	return DoPlayerAction(G, action)
//}
//
//// aaa,
//func (G *Game) AddBet(player *player.Player, selection string, moneyValue int64) error {
//	action := &Action{
//		actionName: ACTION_ADD_BET,
//		playerId:   player.Id(),
//		data: map[string]interface{}{
//			"selection":  selection,
//			"moneyValue": moneyValue,
//		},
//		chanResponse: make(chan *ActionResponse),
//	}
//	return DoPlayerAction(G, action)
//}
//
//func (G *Game) Chat(player *player.Player, message string) error {
//	action := &Action{
//		actionName: ACTION_CHAT,
//		playerId:   player.Id(),
//		data: map[string]interface{}{
//			"senderName": player.DisplayName(),
//			"message":    message,
//		},
//		chanResponse: make(chan *ActionResponse),
//	}
//	return DoPlayerAction(G, action)
//}
