package slotxxx

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/record"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/utils"
)

const (
	SLOTXXX_GAME_CODE = "slotxxx"

	SLOTXXX_JACKPOT_CODE_100   = "SLOTXXX_JACKPOT_CODE_100"
	SLOTXXX_JACKPOT_CODE_1000  = "SLOTXXX_JACKPOT_CODE_1000"
	SLOTXXX_JACKPOT_CODE_10000 = "SLOTXXX_JACKPOT_CODE_10000"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

type SlotxxxGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter int64
	// money only in [0, 100, 1000, 10000]
	mapPlayerIdToMoneyPerLine map[int64]int64
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*SlotxxxMatch
	mapPlayerIdToHistory map[int64]*cardgame.SizedList
	// danh sách những người trúng lớn gần nhất
	bigWinList *cardgame.SizedList
	// danh sách người chơi vừa có thao tác trong 60s gần đây,
	// dùng để public jackpot price 3s một lần
	mapPlayerIdToIsActive       map[int64]bool
	mapPlayerIdToLastActiveTime map[int64]time.Time

	jackpot100   *jackpot.Jackpot
	jackpot1000  *jackpot.Jackpot
	jackpot10000 *jackpot.Jackpot

	ChanActionReceiver chan *Action

	mutex sync.RWMutex
}

func NewSlotxxxGame(currencyType string) *SlotxxxGame {
	slotxxxG := &SlotxxxGame{
		gameCode:     SLOTXXX_GAME_CODE,
		currencyType: currencyType,

		matchCounter:                0,
		mapPlayerIdToMatch:          map[int64]*SlotxxxMatch{},
		mapPlayerIdToMoneyPerLine:   map[int64]int64{},
		mapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		mapPlayerIdToIsActive:       map[int64]bool{},
		mapPlayerIdToLastActiveTime: map[int64]time.Time{},

		jackpot100:   jackpot.GetJackpot(SLOTXXX_JACKPOT_CODE_100, currencyType),
		jackpot1000:  jackpot.GetJackpot(SLOTXXX_JACKPOT_CODE_1000, currencyType),
		jackpot10000: jackpot.GetJackpot(SLOTXXX_JACKPOT_CODE_10000, currencyType),

		ChanActionReceiver: make(chan *Action),
	}

	temp := cardgame.NewSizedList(10)
	slotxxxG.bigWinList = &temp

	if slotxxxG.currencyType == currency.Money {
		slotxxxG.tax = 0
	} else {
		slotxxxG.tax = 0
	}

	go LoopReceiveActions(slotxxxG)
	// go LoopPublicJackpotInfo(slotxxxG)

	return slotxxxG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopReceiveActions(slotxxxG *SlotxxxGame) {
	for {
		action := <-slotxxxG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(slotxxxG *SlotxxxGame, action *Action) {
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
					slotxxxG.SetPlayerActiveStatus(action.playerId)
					if action.actionName == ACTION_SPIN {
						slotxxxG.mutex.RLock()
						_, isPlayingASlotxxxMatch := slotxxxG.mapPlayerIdToMatch[action.playerId]
						slotxxxG.mutex.RUnlock()
						if isPlayingASlotxxxMatch {
							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0007))}
						}

						slotxxxG.mutex.RLock()
						moneyPerLine := slotxxxG.mapPlayerIdToMoneyPerLine[action.playerId]
						payLineIndexs := []int{0}
						slotxxxG.mutex.RUnlock()
						slotxxxG.mutex.Lock()
						neededMoney := moneyPerLine * int64(len(payLineIndexs))
						if playerObj.GetAvailableMoney(slotxxxG.currencyType) < neededMoney {
							slotxxxG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0008))}
						} else {
							playerObj.ChangeMoneyAndLog(
								-neededMoney, slotxxxG.CurrencyType(), false, "",
								ACTION_SPIN, slotxxxG.GameCode(), "")
							//
							slotxxxG.matchCounter += 1
							newMatch := NewSlotxxxMatch(
								slotxxxG,
								playerObj,
								slotxxxG.matchCounter,
								moneyPerLine,
								payLineIndexs,
							)
							slotxxxG.mapPlayerIdToMatch[action.playerId] = newMatch
							slotxxxG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: nil}
						}

					} else if action.actionName == ACTION_CHOOSE_MONEY_PER_LINE {
						moneyPerLine := utils.GetInt64AtPath(action.data, "moneyPerLine")
						if cardgame.FindInt64InSlice(moneyPerLine, MONEYS_PER_LINE) == -1 {
							action.chanResponse <- &ActionResponse{err: errors.New("wrong moneyPerLine")}
						} else {
							slotxxxG.mutex.Lock()
							slotxxxG.mapPlayerIdToMoneyPerLine[action.playerId] = moneyPerLine
							slotxxxG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: nil}
						}
					} else if action.actionName == ACTION_GET_HISTORY {
						var listResultJson []string
						slotxxxG.mutex.RLock()
						if _, isIn := slotxxxG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = slotxxxG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						data := map[string]interface{}{
							"myLast10":     listResultJson,
							"last10BigWin": slotxxxG.bigWinList.Elements,
						}
						slotxxxG.mutex.RUnlock()
						slotxxxG.SendDataToPlayerId(
							"SlotxxxHistory",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else { // các hành động khi đã bắt đầu chơi
						slotxxxG.mutex.RLock()
						hisMatch, isPlayingASlotxxxMatch := slotxxxG.mapPlayerIdToMatch[action.playerId]
						slotxxxG.mutex.RUnlock()
						if isPlayingASlotxxxMatch && hisMatch != nil {
							hisMatch.ChanActionReceiver <- action
						} else {
							action.chanResponse <- &ActionResponse{err: errors.New("need to ACTION_SPIN first")}
						}
					}
				}
			}(slotxxxG, action)
		}
	}
}

// public jackpot, use lock game inside
func LoopPublicJackpotInfo(slotxxxG *SlotxxxGame) {
	for {
		time.Sleep(5 * time.Second)
		slotxxxG.mutex.RLock()
		activePids := make([]int64, len(slotxxxG.mapPlayerIdToIsActive))
		for pid, _ := range slotxxxG.mapPlayerIdToIsActive {
			activePids = append(activePids, pid)
		}
		slotxxxG.mutex.RUnlock()
		for _, pid := range activePids {
			slotxxxG.SendDataToPlayerId(
				"SlotxxxJackpots",
				map[string]interface{}{
					SLOTXXX_JACKPOT_CODE_100:   slotxxxG.jackpot100.Value(),
					SLOTXXX_JACKPOT_CODE_1000:  slotxxxG.jackpot1000.Value(),
					SLOTXXX_JACKPOT_CODE_10000: slotxxxG.jackpot10000.Value(),
				},
				pid)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (slotxxxG *SlotxxxGame) GameCode() string {
	return slotxxxG.gameCode
}
func (slotxxxG *SlotxxxGame) GetGameCode() string {
	return slotxxxG.gameCode
}

func (slotxxxG *SlotxxxGame) CurrencyType() string {
	return slotxxxG.currencyType
}
func (slotxxxG *SlotxxxGame) GetCurrencyType() string {
	return slotxxxG.currencyType
}

func (slotxxxG *SlotxxxGame) SerializeData() map[string]interface{} {
	for _, jackpotObj := range jackpot.Jackpots() {
		_ = jackpotObj
		// fmt.Println("jackpot.Jackpots", jackpotObj.SerializedData())
	}
	// fmt.Println("slotxxxG.currencyType", slotxxxG.currencyType)
	// fmt.Println("slotxxxG.jackpot100", slotxxxG.jackpot100)
	result := map[string]interface{}{
		"gameCode":     slotxxxG.gameCode,
		"currencyType": slotxxxG.currencyType,
		"tax":          slotxxxG.tax,

		"SYMBOLS":                     SYMBOLS,
		"PAYLINES":                    PAYLINES,
		"MONEYS_PER_LINE":             MONEYS_PER_LINE,
		"MAP_LINE_TYPE_TO_PRIZE_RATE": MAP_LINE_TYPE_TO_PRIZE_RATE,

		"jackpot100":   slotxxxG.jackpot100.Value(),
		"jackpot1000":  slotxxxG.jackpot1000.Value(),
		"jackpot10000": slotxxxG.jackpot10000.Value(),
	}
	return result
}

func (slotxxxG *SlotxxxGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

// mark a player active in slotxxx, for public jackpots price
// after 60s mark inactive
func (slotxxxG *SlotxxxGame) SetPlayerActiveStatus(playerId int64) {
	slotxxxG.mutex.Lock()
	slotxxxG.mapPlayerIdToIsActive[playerId] = true
	slotxxxG.mapPlayerIdToLastActiveTime[playerId] = time.Now()
	slotxxxG.mutex.Unlock()
	go func(playerId int64) {
		timeout := time.After(60 * time.Second)
		<-timeout
		slotxxxG.mutex.Lock()
		if time.Now().Sub(slotxxxG.mapPlayerIdToLastActiveTime[playerId]) >= 60*time.Second {
			delete(slotxxxG.mapPlayerIdToIsActive, playerId)
		}
		slotxxxG.mutex.Unlock()
	}(playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(slotxxxG *SlotxxxGame, action *Action) error {
	slotxxxG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (slotxxxG *SlotxxxGame) ChooseMoneyPerLine(player *player.Player, moneyPerLine int64) error {
	action := &Action{
		actionName: ACTION_CHOOSE_MONEY_PER_LINE,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"moneyPerLine": moneyPerLine,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) Spin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) StopPlaying(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_STOP_PLAYING,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) SelectSmall(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SELECT_SMALL,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) SelectBig(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SELECT_BIG,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}

// aaa,
func (slotxxxG *SlotxxxGame) GetMatchInfo(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_MATCH_INFO,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotxxxG, action)
}
