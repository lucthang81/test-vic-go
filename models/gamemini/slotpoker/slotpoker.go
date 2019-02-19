package slotpoker

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
	//	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/utils"
)

const (
	SLOTPOKER_GAME_CODE = "slotpoker"

	SLOTPOKER_JACKPOT_CODE_100   = "SLOTPOKER_JACKPOT_CODE_100"
	SLOTPOKER_JACKPOT_CODE_1000  = "SLOTPOKER_JACKPOT_CODE_1000"
	SLOTPOKER_JACKPOT_CODE_10000 = "SLOTPOKER_JACKPOT_CODE_10000"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

type SlotpokerGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter int64
	// money only in [0, 100, 1000, 10000]
	mapPlayerIdToMoneyPerLine map[int64]int64
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*SlotpokerMatch
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

func NewSlotpokerGame(currencyType string) *SlotpokerGame {
	slotpokerG := &SlotpokerGame{
		gameCode:     SLOTPOKER_GAME_CODE,
		currencyType: currencyType,

		matchCounter:                0,
		mapPlayerIdToMatch:          map[int64]*SlotpokerMatch{},
		mapPlayerIdToMoneyPerLine:   map[int64]int64{},
		mapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		mapPlayerIdToIsActive:       map[int64]bool{},
		mapPlayerIdToLastActiveTime: map[int64]time.Time{},

		jackpot100:   jackpot.GetJackpot(SLOTPOKER_JACKPOT_CODE_100, currencyType),
		jackpot1000:  jackpot.GetJackpot(SLOTPOKER_JACKPOT_CODE_1000, currencyType),
		jackpot10000: jackpot.GetJackpot(SLOTPOKER_JACKPOT_CODE_10000, currencyType),

		ChanActionReceiver: make(chan *Action),
	}

	temp := cardgame.NewSizedList(10)
	slotpokerG.bigWinList = &temp

	if slotpokerG.currencyType == currency.Money {
		slotpokerG.tax = 0
	} else {
		slotpokerG.tax = 0
	}

	go LoopReceiveActions(slotpokerG)
	//go LoopPublicJackpotInfo(slotpokerG)

	return slotpokerG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopReceiveActions(slotpokerG *SlotpokerGame) {
	for {
		action := <-slotpokerG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(slotpokerG *SlotpokerGame, action *Action) {
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
					slotpokerG.SetPlayerActiveStatus(action.playerId)
					if action.actionName == ACTION_SPIN {
						isPlayingXocdia := false
						isPlayingOtherGamemini := false
						if isPlayingXocdia || isPlayingOtherGamemini {
							action.chanResponse <- &ActionResponse{err: errors.New("Cant play slotpoker when playing xocdia2 or other gamemini")}
						} else {
							slotpokerG.mutex.RLock()
							_, isPlayingASlotpokerMatch := slotpokerG.mapPlayerIdToMatch[action.playerId]
							slotpokerG.mutex.RUnlock()
							if isPlayingASlotpokerMatch {
								action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0007))}
							} else {
								slotpokerG.mutex.RLock()
								moneyPerLine := slotpokerG.mapPlayerIdToMoneyPerLine[action.playerId]
								payLineIndexs := []int{0}
								slotpokerG.mutex.RUnlock()
								slotpokerG.mutex.Lock()
								neededMoney := moneyPerLine * int64(len(payLineIndexs))
								if playerObj.GetAvailableMoney(slotpokerG.currencyType) < neededMoney {
									slotpokerG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0008))}
								} else {
									playerObj.ChangeMoneyAndLog(
										-neededMoney, slotpokerG.CurrencyType(), false, "",
										ACTION_SPIN, slotpokerG.GameCode(), "")
									//
									slotpokerG.matchCounter += 1
									newMatch := NewSlotpokerMatch(
										slotpokerG,
										playerObj,
										slotpokerG.matchCounter,
										moneyPerLine,
										payLineIndexs,
									)
									slotpokerG.mapPlayerIdToMatch[action.playerId] = newMatch
									slotpokerG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: nil}
								}
							}
						}
					} else if action.actionName == ACTION_CHOOSE_MONEY_PER_LINE {
						moneyPerLine := utils.GetInt64AtPath(action.data, "moneyPerLine")
						if cardgame.FindInt64InSlice(moneyPerLine, MONEYS_PER_LINE) == -1 {
							action.chanResponse <- &ActionResponse{err: errors.New("wrong moneyPerLine")}
						} else {
							slotpokerG.mutex.Lock()
							slotpokerG.mapPlayerIdToMoneyPerLine[action.playerId] = moneyPerLine
							slotpokerG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: nil}
						}
					} else if action.actionName == ACTION_GET_HISTORY {
						var listResultJson []string
						slotpokerG.mutex.RLock()
						if _, isIn := slotpokerG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = slotpokerG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						data := map[string]interface{}{
							"myLast10":     listResultJson,
							"last10BigWin": slotpokerG.bigWinList.Elements,
						}
						slotpokerG.mutex.RUnlock()
						slotpokerG.SendDataToPlayerId(
							"SlotpokerHistory",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New("wrong action")}
					}
				}
			}(slotpokerG, action)
		}
	}
}

// public jackpot, use lock game inside
func LoopPublicJackpotInfo(slotpokerG *SlotpokerGame) {
	for {
		time.Sleep(5 * time.Second)
		slotpokerG.mutex.RLock()
		activePids := make([]int64, len(slotpokerG.mapPlayerIdToIsActive))
		for pid, _ := range slotpokerG.mapPlayerIdToIsActive {
			activePids = append(activePids, pid)
		}
		slotpokerG.mutex.RUnlock()
		for _, pid := range activePids {
			slotpokerG.SendDataToPlayerId(
				"SlotpokerJackpots",
				map[string]interface{}{
					SLOTPOKER_JACKPOT_CODE_100:   slotpokerG.jackpot100.Value(),
					SLOTPOKER_JACKPOT_CODE_1000:  slotpokerG.jackpot1000.Value(),
					SLOTPOKER_JACKPOT_CODE_10000: slotpokerG.jackpot10000.Value(),
				},
				pid)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (slotpokerG *SlotpokerGame) GetGameCode() string {
	return slotpokerG.gameCode
}
func (slotpokerG *SlotpokerGame) GameCode() string {
	return slotpokerG.gameCode
}

func (slotpokerG *SlotpokerGame) GetCurrencyType() string {
	return slotpokerG.currencyType
}
func (slotpokerG *SlotpokerGame) CurrencyType() string {
	return slotpokerG.currencyType
}

func (slotpokerG *SlotpokerGame) SerializeData() map[string]interface{} {
	for _, jackpotObj := range jackpot.Jackpots() {
		_ = jackpotObj
		// fmt.Println("jackpot.Jackpots", jackpotObj.SerializedData())
	}
	//fmt.Println("slotpokerG.currencyType", slotpokerG.currencyType)
	//fmt.Println("slotpokerG.jackpot100", slotpokerG.jackpot100)
	result := map[string]interface{}{
		"gameCode":     slotpokerG.gameCode,
		"currencyType": slotpokerG.currencyType,
		"tax":          slotpokerG.tax,

		"SYMBOLS":                     SYMBOLS,
		"PAYLINES":                    PAYLINES,
		"MONEYS_PER_LINE":             MONEYS_PER_LINE,
		"MAP_LINE_TYPE_TO_PRIZE_RATE": MAP_LINE_TYPE_TO_PRIZE_RATE,

		"jackpot100":   slotpokerG.jackpot100.Value(),
		"jackpot1000":  slotpokerG.jackpot1000.Value(),
		"jackpot10000": slotpokerG.jackpot10000.Value(),
	}
	return result
}

func (slotpokerG *SlotpokerGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

// mark a player active in slotpoker, for public jackpots price
// after 60s mark inactive
func (slotpokerG *SlotpokerGame) SetPlayerActiveStatus(playerId int64) {
	slotpokerG.mutex.Lock()
	slotpokerG.mapPlayerIdToIsActive[playerId] = true
	slotpokerG.mapPlayerIdToLastActiveTime[playerId] = time.Now()
	slotpokerG.mutex.Unlock()
	go func(playerId int64) {
		timeout := time.After(60 * time.Second)
		<-timeout
		slotpokerG.mutex.Lock()
		if time.Now().Sub(slotpokerG.mapPlayerIdToLastActiveTime[playerId]) >= 60*time.Second {
			delete(slotpokerG.mapPlayerIdToIsActive, playerId)
		}
		slotpokerG.mutex.Unlock()
	}(playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(slotpokerG *SlotpokerGame, action *Action) error {
	slotpokerG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (slotpokerG *SlotpokerGame) ChooseMoneyPerLine(player *player.Player, moneyPerLine int64) error {
	action := &Action{
		actionName: ACTION_CHOOSE_MONEY_PER_LINE,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"moneyPerLine": moneyPerLine,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotpokerG, action)
}

// aaa,
func (slotpokerG *SlotpokerGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotpokerG, action)
}

// aaa,
func (slotpokerG *SlotpokerGame) Spin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotpokerG, action)
}
