// general slots
package gamemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

type ServerInterface interface {
	// already run in a goroutine
	SendRequest(requestType string, data map[string]interface{}, toPlayerId int64)
	SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64)
	SendRequestsToAll(requestType string, data map[string]interface{})
}

var ServerObj ServerInterface

func init() {
	fmt.Print("")
	var _ game.GamePlayer
}

func RegisterServer(registeredServer ServerInterface) {
	ServerObj = registeredServer
}

type GameMiniInterface interface {
	GetGameCode() string
	GetCurrencyType() string

	SerializeData() map[string]interface{}
}

// call after slot spin,
// receive action from player,
// can change SumWonMoney, SumLostMoney, jackpot
type AdditionalGameInterface interface {
	Start()
	ReceiveAction(action *Action)
}

// _____________________________________________________________________________
// _____________________________________________________________________________

// represent action from player to server
type Action struct {
	ActionName string
	PlayerId   int64

	Data         map[string]interface{}
	ChanResponse chan *ActionResponse
}

//
type ActionResponse struct {
	Err  error
	Data map[string]interface{}
}

//
type ResultSlot struct {
	// playerId
	Id           int64
	Username     string
	ChangedMoney int64

	MatchId      string
	StartedTime  time.Time
	MoneyPerLine int64

	SlotResult                [][]string
	MapPaylineIndexToWonMoney map[int]int64
	MapPaylineIndexToIsWin    map[int]bool
	//
	SumWonMoney int64
	// SumLostMoney is a negative value
	SumLostMoney int64
	MatchWonType string // MATCH_WON_TYPE_..
}

func (result1p *ResultSlot) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"playerId": result1p.Id,
		"username": result1p.Username,

		"startedTime":  result1p.StartedTime.Format(time.RFC3339),
		"matchId":      result1p.MatchId,
		"moneyPerLine": result1p.MoneyPerLine,

		"slotResult":                result1p.SlotResult,
		"sumWonMoney":               result1p.SumWonMoney,
		"sumLostMoney":              result1p.SumLostMoney,
		"mapPaylineIndexToWonMoney": result1p.MapPaylineIndexToWonMoney,
		"mapPaylineIndexToIsWin":    result1p.MapPaylineIndexToIsWin,
		"matchWonType":              result1p.MatchWonType,
		"changedMoney":              result1p.ChangedMoney,
	}
	return result
}

func (result1p *ResultSlot) String() string {
	bytes, _ := json.Marshal(result1p.Serialize())
	return string(bytes)
}

func (result1p *ResultSlot) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"id":       result1p.Id,
		"username": result1p.Username,

		"change": result1p.ChangedMoney,
	}
	return result
}

type SlotGame struct {
	GameCode     string
	CurrencyType string

	MatchCounter int64
	// money only in [0, 100, 1000, 10000]
	MapPlayerIdToMoneyPerLine  map[int64]int64
	MapPlayerIdToPayLineIndexs map[int64][]int
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	MapPlayerIdToMatch   map[int64]*SlotMatch
	MapPlayerIdToHistory map[int64]*cardgame.SizedList
	// danh sách những người trúng lớn gần nhất
	BigWinList *cardgame.SizedList
	// danh sách người chơi vừa có thao tác trong 60s gần đây,
	// dùng để public jackpot price 3s một lần
	MapPlayerIdToIsActive       map[int64]bool
	MapPlayerIdToLastActiveTime map[int64]time.Time

	JackpotSmall  *jackpot.Jackpot
	JackpotMedium *jackpot.Jackpot
	JackpotBig    *jackpot.Jackpot

	ChanActionReceiver chan *Action

	Mutex sync.RWMutex
}

func (slotG *SlotGame) GetGameCode() string {
	return slotG.GameCode
}

func (slotG *SlotGame) GetCurrencyType() string {
	return slotG.CurrencyType
}

func (slotG *SlotGame) SerializeData() map[string]interface{} {
	for _, jackpotObj := range jackpot.Jackpots() {
		_ = jackpotObj
		// fmt.Println("jackpot.Jackpots", jackpotObj.SerializedData())
	}
	// fmt.Println("slotG.currencyType", slotG.currencyType)
	// fmt.Println("slotG.jackpot100", slotG.jackpot100)
	result := map[string]interface{}{
		"GameCode":     slotG.GameCode,
		"CurrencyType": slotG.CurrencyType,

		"JackpotSmall":  slotG.JackpotSmall.Value(),
		"JackpotMedium": slotG.JackpotMedium.Value(),
		"JackpotBig":    slotG.JackpotBig.Value(),
	}
	return result
}

func (slotG *SlotGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	ServerObj.SendRequest(method, data, playerId)
}

// mark a player active in slot, for public jackpots price
// after 60s mark inactive
func (slotG *SlotGame) SetPlayerActiveStatus(playerId int64) {
	slotG.Mutex.Lock()
	slotG.MapPlayerIdToIsActive[playerId] = true
	slotG.MapPlayerIdToLastActiveTime[playerId] = time.Now()
	slotG.Mutex.Unlock()
	go func(playerId int64) {
		timeout := time.After(60 * time.Second)
		<-timeout
		slotG.Mutex.Lock()
		if time.Now().Sub(slotG.MapPlayerIdToLastActiveTime[playerId]) >= 60*time.Second {
			delete(slotG.MapPlayerIdToIsActive, playerId)
		}
		slotG.Mutex.Unlock()
	}(playerId)
}

// aaa,
func DoPlayerAction(slotG *SlotGame, action *Action) error {
	slotG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.ChanResponse:
		return res.Err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

//
func SlotGameLoopReceiveActions(slotG *SlotGame,
	moneysPerLine []int64, paylines [][]int,
	startMatchFunc TypeStartSlotMatchFunction) {
	for {
		action := <-slotG.ChanActionReceiver
		if action.ActionName == consts.ACTION_STOP_GAME {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func(slotG *SlotGame, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				playerObj, err := player.GetPlayer(action.PlayerId)
				if err != nil {
					action.ChanResponse <- &ActionResponse{Err: errors.New("Cant find player for this id")}
				} else {
					slotG.SetPlayerActiveStatus(action.PlayerId)
					if action.ActionName == consts.ACTION_SPIN {
						slotG.Mutex.RLock()
						_, isPlayingASlotMatch := slotG.MapPlayerIdToMatch[action.PlayerId]
						slotG.Mutex.RUnlock()
						if isPlayingASlotMatch {
							action.ChanResponse <- &ActionResponse{Err: errors.New(l.Get(l.M0007))}
						}

						slotG.Mutex.RLock()
						moneyPerLine := slotG.MapPlayerIdToMoneyPerLine[action.PlayerId]
						payLineIndexs := slotG.MapPlayerIdToPayLineIndexs[action.PlayerId]
						slotG.Mutex.RUnlock()
						slotG.Mutex.Lock()
						neededMoney := moneyPerLine * int64(len(payLineIndexs))
						// fmt.Println("neededMoney playerObj.GetAvailableMoney", neededMoney, playerObj.GetAvailableMoney(slotG.CurrencyType), slotG.CurrencyType)
						if playerObj.GetAvailableMoney(slotG.CurrencyType) < neededMoney {
							slotG.Mutex.Unlock()
							action.ChanResponse <- &ActionResponse{Err: errors.New(l.Get(l.M0008))}
						} else {
							playerObj.ChangeMoneyAndLog(
								-neededMoney, slotG.CurrencyType, false, "",
								consts.ACTION_SPIN, slotG.GameCode, "")
							//
							slotG.MatchCounter += 1
							newMatch := NewSlotMatch(
								slotG,
								playerObj,
								slotG.MatchCounter,
								moneyPerLine,
								payLineIndexs,
								startMatchFunc,
							)
							slotG.MapPlayerIdToMatch[action.PlayerId] = newMatch
							slotG.Mutex.Unlock()
							action.ChanResponse <- &ActionResponse{Err: nil}
						}

					} else if action.ActionName == consts.ACTION_CHOOSE_MONEY_PER_LINE {
						moneyPerLine := utils.GetInt64AtPath(action.Data, "moneyPerLine")
						if cardgame.FindInt64InSlice(moneyPerLine, moneysPerLine) == -1 {
							action.ChanResponse <- &ActionResponse{Err: errors.New("wrong moneyPerLine")}
						} else {
							slotG.Mutex.Lock()
							slotG.MapPlayerIdToMoneyPerLine[action.PlayerId] = moneyPerLine
							slotG.Mutex.Unlock()
							action.ChanResponse <- &ActionResponse{Err: nil}
						}
					} else if action.ActionName == consts.ACTION_CHOOSE_PAYLINES {
						paylineIndexs := action.Data["paylineIndexs"].([]int)
						filtedPaylineIndexs := []int{}
						setIndexs := map[int]bool{}
						for _, index := range paylineIndexs {
							if (0 <= index) && (index < len(paylines)) {
								setIndexs[index] = true
							}
						}
						for index, _ := range setIndexs {
							filtedPaylineIndexs = append(filtedPaylineIndexs, index)
						}
						slotG.Mutex.Lock()
						slotG.MapPlayerIdToPayLineIndexs[action.PlayerId] = filtedPaylineIndexs
						slotG.Mutex.Unlock()
						action.ChanResponse <- &ActionResponse{Err: nil}
					} else if action.ActionName == consts.ACTION_GET_HISTORY {
						var listResultJson []string
						slotG.Mutex.RLock()
						if _, isIn := slotG.MapPlayerIdToHistory[action.PlayerId]; isIn {
							listResultJson = slotG.MapPlayerIdToHistory[action.PlayerId].Elements
						} else {
							listResultJson = []string{}
						}
						data := map[string]interface{}{
							"myLast10":     listResultJson,
							"last10BigWin": slotG.BigWinList.Elements,
						}
						slotG.Mutex.RUnlock()
						slotG.SendDataToPlayerId(
							"SlotHistory",
							data,
							action.PlayerId)
						action.ChanResponse <- &ActionResponse{Err: nil}
					} else { // các hành động khi đã bắt đầu chơi
						slotG.Mutex.RLock()
						hisMatch, isPlayingASlotMatch := slotG.MapPlayerIdToMatch[action.PlayerId]
						slotG.Mutex.RUnlock()
						if isPlayingASlotMatch && hisMatch != nil {
							hisMatch.ChanActionReceiver <- action
						} else {
							action.ChanResponse <- &ActionResponse{Err: errors.New(l.Get(l.M0095))}
						}
					}
				}
			}(slotG, action)
		}
	}
}

///*
// aaa,
func (slotG *SlotGame) ChooseMoneyPerLine(
	player *player.Player, moneyPerLine int64) error {
	action := &Action{
		ActionName: consts.ACTION_CHOOSE_MONEY_PER_LINE,
		PlayerId:   player.Id(),
		Data: map[string]interface{}{
			"moneyPerLine": moneyPerLine,
		},
		ChanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) ChoosePaylines(player *player.Player, paylineIndexs []int) error {
	action := &Action{
		ActionName: consts.ACTION_CHOOSE_PAYLINES,
		PlayerId:   player.Id(),
		Data: map[string]interface{}{
			"paylineIndexs": paylineIndexs,
		},
		ChanResponse: make(chan *ActionResponse),
	}
	// fmt.Printf("hihi %v %+v \n", slotG.GameCode, action.Data)
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) GetHistory(player *player.Player) error {
	action := &Action{
		ActionName:   consts.ACTION_GET_HISTORY,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) Spin(player *player.Player) error {
	action := &Action{
		ActionName:   consts.ACTION_SPIN,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) GetMatchInfo(player *player.Player) error {
	action := &Action{
		ActionName:   consts.ACTION_GET_MATCH_INFO,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

type SlotMatch struct {
	Game          *SlotGame
	Player        *player.Player
	StartedTime   time.Time
	MatchId       string
	MoneyPerLine  int64
	PayLineIndexs []int

	SlotResult     [][]string
	PlayerResult   *ResultSlot
	Phase          string
	AdditionalGame AdditionalGameInterface

	ChanActionReceiver chan *Action

	Mutex sync.RWMutex
}

//
func (match *SlotMatch) GetGameCode() string {
	return match.Game.GetGameCode()
}

func (match *SlotMatch) GetCurrencyType() string {
	return match.Game.GetCurrencyType()
}

// json obj represent general match info
func (match *SlotMatch) SerializedData() map[string]interface{} {
	data := match.PlayerResult.Serialize()
	data["GameCode"] = match.GetGameCode()
	data["CurrencyType"] = match.GetCurrencyType()
	data["Phase"] = match.Phase
	return data
}

func (match *SlotMatch) UpdateMatchStatus() {
	data := match.SerializedData()
	match.Game.SendDataToPlayerId(
		"SlotUpdateMatchStatus",
		data,
		match.Player.Id(),
	)
}

type TypeStartSlotMatchFunction func(match *SlotMatch)

func NewSlotMatch(
	slotG *SlotGame, createdPlayer *player.Player, matchCounter int64,
	moneyPerLine int64, payLineIndexs []int,
	startMatchFunc TypeStartSlotMatchFunction,
) *SlotMatch {
	match := &SlotMatch{
		Game:               slotG,
		Player:             createdPlayer,
		StartedTime:        time.Now(),
		MatchId:            fmt.Sprintf("%v_%v_%v", slotG.GameCode, matchCounter, time.Now().Unix()),
		PlayerResult:       &ResultSlot{},
		MoneyPerLine:       moneyPerLine,
		PayLineIndexs:      payLineIndexs,
		Phase:              "PHASE_0_INITING",
		ChanActionReceiver: make(chan *Action),
	}
	// init vars code in match here
	match.PlayerResult.Id = match.Player.Id()
	match.PlayerResult.Username = match.Player.Name()
	match.PlayerResult.MatchId = match.MatchId
	match.PlayerResult.StartedTime = match.StartedTime
	match.PlayerResult.MoneyPerLine = match.MoneyPerLine
	match.PlayerResult.SumLostMoney = -match.MoneyPerLine * int64(len(payLineIndexs))
	//
	go startMatchFunc(match)
	go InMatchLoopReceiveActions(match)
	return match
}

func InMatchLoopReceiveActions(match *SlotMatch) {
	for {
		action := <-match.ChanActionReceiver
		if action.ActionName == consts.ACTION_FINISH_SESSION {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func(match *SlotMatch, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == consts.ACTION_GET_MATCH_INFO {
					action.ChanResponse <- &ActionResponse{Err: nil}
					match.UpdateMatchStatus()
				} else {
					if match.AdditionalGame != nil {
						match.AdditionalGame.ReceiveAction(action)
					} else {
						action.ChanResponse <- &ActionResponse{
							Err: errors.New("match.AdditionalGame == nil")}
					}
				}
			}(match, action)
		}
	}
}
