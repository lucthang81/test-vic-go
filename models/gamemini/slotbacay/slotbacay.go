package slotbacay

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
	SLOTBACAY_GAME_CODE = "slotbacay"

	SLOTBACAY_JACKPOT_CODE_100   = "SLOTBACAY_JACKPOT_CODE_100"
	SLOTBACAY_JACKPOT_CODE_1000  = "SLOTBACAY_JACKPOT_CODE_1000"
	SLOTBACAY_JACKPOT_CODE_10000 = "SLOTBACAY_JACKPOT_CODE_10000"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

type SlotbacayGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter int64
	// money only in [0, 100, 1000, 10000]
	mapPlayerIdToMoneyPerLine map[int64]int64
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*SlotbacayMatch
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

func NewSlotbacayGame(currencyType string) *SlotbacayGame {
	slotbacayG := &SlotbacayGame{
		gameCode:     SLOTBACAY_GAME_CODE,
		currencyType: currencyType,

		matchCounter:                0,
		mapPlayerIdToMatch:          map[int64]*SlotbacayMatch{},
		mapPlayerIdToMoneyPerLine:   map[int64]int64{},
		mapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		mapPlayerIdToIsActive:       map[int64]bool{},
		mapPlayerIdToLastActiveTime: map[int64]time.Time{},

		jackpot100:   jackpot.GetJackpot(SLOTBACAY_JACKPOT_CODE_100, currencyType),
		jackpot1000:  jackpot.GetJackpot(SLOTBACAY_JACKPOT_CODE_1000, currencyType),
		jackpot10000: jackpot.GetJackpot(SLOTBACAY_JACKPOT_CODE_10000, currencyType),

		ChanActionReceiver: make(chan *Action),
	}

	temp := cardgame.NewSizedList(10)
	slotbacayG.bigWinList = &temp

	if slotbacayG.currencyType == currency.Money {
		slotbacayG.tax = 0
	} else {
		slotbacayG.tax = 0
	}

	go LoopReceiveActions(slotbacayG)
	// go LoopPublicJackpotInfo(slotbacayG)

	return slotbacayG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopReceiveActions(slotbacayG *SlotbacayGame) {
	for {
		action := <-slotbacayG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(slotbacayG *SlotbacayGame, action *Action) {
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
					slotbacayG.SetPlayerActiveStatus(action.playerId)
					if action.actionName == ACTION_SPIN {
						isPlayingXocdia := false
						isPlayingOtherGamemini := false
						if isPlayingXocdia || isPlayingOtherGamemini {
							action.chanResponse <- &ActionResponse{err: errors.New("Cant play slotbacay when playing xocdia2 or other gamemini")}
						} else {
							slotbacayG.mutex.RLock()
							_, isPlayingASlotbacayMatch := slotbacayG.mapPlayerIdToMatch[action.playerId]
							slotbacayG.mutex.RUnlock()
							if isPlayingASlotbacayMatch {
								action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0007))}
							} else {
								slotbacayG.mutex.RLock()
								moneyPerLine := slotbacayG.mapPlayerIdToMoneyPerLine[action.playerId]
								payLineIndexs := []int{0}
								slotbacayG.mutex.RUnlock()
								slotbacayG.mutex.Lock()
								neededMoney := moneyPerLine * int64(len(payLineIndexs))
								if playerObj.GetAvailableMoney(slotbacayG.currencyType) < neededMoney {
									slotbacayG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0008))}
								} else {
									playerObj.ChangeMoneyAndLog(
										-neededMoney, slotbacayG.CurrencyType(), false, "",
										ACTION_SPIN, slotbacayG.GameCode(), "")
									//
									slotbacayG.matchCounter += 1
									newMatch := NewSlotbacayMatch(
										slotbacayG,
										playerObj,
										slotbacayG.matchCounter,
										moneyPerLine,
										payLineIndexs,
									)
									slotbacayG.mapPlayerIdToMatch[action.playerId] = newMatch
									slotbacayG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: nil}
								}
							}
						}
					} else if action.actionName == ACTION_CHOOSE_MONEY_PER_LINE {
						moneyPerLine := utils.GetInt64AtPath(action.data, "moneyPerLine")
						if cardgame.FindInt64InSlice(moneyPerLine, MONEYS_PER_LINE) == -1 {
							action.chanResponse <- &ActionResponse{err: errors.New("wrong moneyPerLine")}
						} else {
							slotbacayG.mutex.Lock()
							slotbacayG.mapPlayerIdToMoneyPerLine[action.playerId] = moneyPerLine
							slotbacayG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: nil}
						}
					} else if action.actionName == ACTION_GET_HISTORY {
						var listResultJson []string
						slotbacayG.mutex.RLock()
						if _, isIn := slotbacayG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = slotbacayG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						data := map[string]interface{}{
							"myLast10":     listResultJson,
							"last10BigWin": slotbacayG.bigWinList.Elements,
						}
						slotbacayG.mutex.RUnlock()
						slotbacayG.SendDataToPlayerId(
							"SlotbacayHistory",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New("wrong action")}
					}
				}
			}(slotbacayG, action)
		}
	}
}

// public jackpot, use lock game inside
func LoopPublicJackpotInfo(slotbacayG *SlotbacayGame) {
	for {
		time.Sleep(5 * time.Second)
		slotbacayG.mutex.RLock()
		activePids := make([]int64, len(slotbacayG.mapPlayerIdToIsActive))
		for pid, _ := range slotbacayG.mapPlayerIdToIsActive {
			activePids = append(activePids, pid)
		}
		slotbacayG.mutex.RUnlock()
		for _, pid := range activePids {
			slotbacayG.SendDataToPlayerId(
				"SlotbacayJackpots",
				map[string]interface{}{
					SLOTBACAY_JACKPOT_CODE_100:   slotbacayG.jackpot100.Value(),
					SLOTBACAY_JACKPOT_CODE_1000:  slotbacayG.jackpot1000.Value(),
					SLOTBACAY_JACKPOT_CODE_10000: slotbacayG.jackpot10000.Value(),
				},
				pid)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (slotbacayG *SlotbacayGame) GetGameCode() string {
	return slotbacayG.gameCode
}

func (slotbacayG *SlotbacayGame) GameCode() string {
	return slotbacayG.gameCode
}

func (slotbacayG *SlotbacayGame) GetCurrencyType() string {
	return slotbacayG.currencyType
}

func (slotbacayG *SlotbacayGame) CurrencyType() string {
	return slotbacayG.currencyType
}

func (slotbacayG *SlotbacayGame) SerializeData() map[string]interface{} {
	for _, jackpotObj := range jackpot.Jackpots() {
		_ = jackpotObj
		// fmt.Println("jackpot.Jackpots", jackpotObj.SerializedData())
	}
	// fmt.Println("slotbacayG.currencyType", slotbacayG.currencyType)
	// fmt.Println("slotbacayG.jackpot100", slotbacayG.jackpot100)
	result := map[string]interface{}{
		"gameCode":     slotbacayG.gameCode,
		"currencyType": slotbacayG.currencyType,
		"tax":          slotbacayG.tax,

		"SYMBOLS":                     SYMBOLS,
		"PAYLINES":                    PAYLINES,
		"MONEYS_PER_LINE":             MONEYS_PER_LINE,
		"MAP_LINE_TYPE_TO_PRIZE_RATE": MAP_LINE_TYPE_TO_PRIZE_RATE,

		"jackpot100":   slotbacayG.jackpot100.Value(),
		"jackpot1000":  slotbacayG.jackpot1000.Value(),
		"jackpot10000": slotbacayG.jackpot10000.Value(),
	}
	return result
}

func (slotbacayG *SlotbacayGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

// mark a player active in slotbacay, for public jackpots price
// after 60s mark inactive
func (slotbacayG *SlotbacayGame) SetPlayerActiveStatus(playerId int64) {
	slotbacayG.mutex.Lock()
	slotbacayG.mapPlayerIdToIsActive[playerId] = true
	slotbacayG.mapPlayerIdToLastActiveTime[playerId] = time.Now()
	slotbacayG.mutex.Unlock()
	go func(playerId int64) {
		timeout := time.After(60 * time.Second)
		<-timeout
		slotbacayG.mutex.Lock()
		if time.Now().Sub(slotbacayG.mapPlayerIdToLastActiveTime[playerId]) >= 60*time.Second {
			delete(slotbacayG.mapPlayerIdToIsActive, playerId)
		}
		slotbacayG.mutex.Unlock()
	}(playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(slotbacayG *SlotbacayGame, action *Action) error {
	slotbacayG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (slotbacayG *SlotbacayGame) ChooseMoneyPerLine(player *player.Player, moneyPerLine int64) error {
	action := &Action{
		actionName: ACTION_CHOOSE_MONEY_PER_LINE,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"moneyPerLine": moneyPerLine,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotbacayG, action)
}

// aaa,
func (slotbacayG *SlotbacayGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotbacayG, action)
}

// aaa,
func (slotbacayG *SlotbacayGame) Spin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotbacayG, action)
}
