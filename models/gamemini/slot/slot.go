package slot

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	//"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/utils"
)

const (
	SLOT_GAME_CODE = "slot2"

	SLOT_JACKPOT_CODE_100   = "SLOT_JACKPOT_CODE_100"
	SLOT_JACKPOT_CODE_1000  = "SLOT_JACKPOT_CODE_1000"
	SLOT_JACKPOT_CODE_10000 = "SLOT_JACKPOT_CODE_10000"

	IS_FAKE_JACKPOT = true
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

type SlotGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter int64
	// money only in [0, 100, 1000, 10000]
	mapPlayerIdToMoneyPerLine  map[int64]int64
	mapPlayerIdToPayLineIndexs map[int64][]int
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
	mapPlayerIdToMatch   map[int64]*SlotMatch
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

func NewSlotGame(currencyType string) *SlotGame {
	slotG := &SlotGame{
		gameCode:     SLOT_GAME_CODE,
		currencyType: currencyType,

		matchCounter:               0,
		mapPlayerIdToMatch:         map[int64]*SlotMatch{},
		mapPlayerIdToMoneyPerLine:  map[int64]int64{},
		mapPlayerIdToPayLineIndexs: map[int64][]int{},
		mapPlayerIdToHistory:       map[int64]*cardgame.SizedList{},

		mapPlayerIdToIsActive:       map[int64]bool{},
		mapPlayerIdToLastActiveTime: map[int64]time.Time{},

		jackpot100:   jackpot.GetJackpot(SLOT_JACKPOT_CODE_100, currencyType),
		jackpot1000:  jackpot.GetJackpot(SLOT_JACKPOT_CODE_1000, currencyType),
		jackpot10000: jackpot.GetJackpot(SLOT_JACKPOT_CODE_10000, currencyType),

		ChanActionReceiver: make(chan *Action),
	}

	temp := cardgame.NewSizedList(10)
	slotG.bigWinList = &temp

	if slotG.currencyType == currency.Money {
		slotG.tax = 0
	} else {
		slotG.tax = 0
	}

	go LoopReceiveActions(slotG)
	// go LoopPublicJackpotInfo(slotG)
	if IS_FAKE_JACKPOT {
		go LoopFakeJackpot(slotG)
	}

	return slotG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopReceiveActions(slotG *SlotGame) {
	for {
		action := <-slotG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(slotG *SlotGame, action *Action) {
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
					slotG.SetPlayerActiveStatus(action.playerId)
					if action.actionName == ACTION_SPIN {
						isPlayingXocdia := false
						isPlayingOtherGamemini := false
						if isPlayingXocdia || isPlayingOtherGamemini {
							action.chanResponse <- &ActionResponse{err: errors.New("Cant play slot when playing xocdia2 or other gamemini")}
						} else {
							slotG.mutex.RLock()
							_, isPlayingASlotMatch := slotG.mapPlayerIdToMatch[action.playerId]
							slotG.mutex.RUnlock()
							if isPlayingASlotMatch {
								action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0007))}
							} else {
								slotG.mutex.RLock()
								moneyPerLine := slotG.mapPlayerIdToMoneyPerLine[action.playerId]
								payLineIndexs := slotG.mapPlayerIdToPayLineIndexs[action.playerId]
								slotG.mutex.RUnlock()
								slotG.mutex.Lock()
								neededMoney := moneyPerLine * int64(len(payLineIndexs))
								if playerObj.GetAvailableMoney(slotG.currencyType) < neededMoney {
									slotG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0008))}
								} else {
									temp := int64(len(payLineIndexs)) * moneyPerLine
									playerObj.ChangeMoneyAndLog(
										-temp, slotG.GetCurrencyType(), false, "",
										ACTION_SPIN, slotG.GetGameCode(), "")
									//
									slotG.matchCounter += 1
									newMatch := NewSlotMatch(
										slotG,
										playerObj,
										slotG.matchCounter,
										moneyPerLine,
										payLineIndexs,
									)
									slotG.mapPlayerIdToMatch[action.playerId] = newMatch
									slotG.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: nil}
								}
							}
						}
					} else if action.actionName == ACTION_CHOOSE_MONEY_PER_LINE {
						moneyPerLine := utils.GetInt64AtPath(action.data, "moneyPerLine")
						if cardgame.FindInt64InSlice(moneyPerLine, MONEYS_PER_LINE) == -1 {
							action.chanResponse <- &ActionResponse{err: errors.New("wrong moneyPerLine")}
						} else {
							slotG.mutex.Lock()
							slotG.mapPlayerIdToMoneyPerLine[action.playerId] = moneyPerLine
							slotG.mutex.Unlock()
							action.chanResponse <- &ActionResponse{err: nil}
						}
					} else if action.actionName == ACTION_CHOOSE_PAYLINES {
						paylineIndexs := action.data["paylineIndexs"].([]int)
						filtedPaylineIndexs := []int{}
						setIndexs := map[int]bool{}
						for _, index := range paylineIndexs {
							if (0 <= index) && (index < len(PAYLINES)) {
								setIndexs[index] = true
							}
						}
						for index, _ := range setIndexs {
							filtedPaylineIndexs = append(filtedPaylineIndexs, index)
						}
						slotG.mutex.Lock()
						slotG.mapPlayerIdToPayLineIndexs[action.playerId] = filtedPaylineIndexs
						slotG.mutex.Unlock()
						action.chanResponse <- &ActionResponse{err: nil}
					} else if action.actionName == ACTION_GET_HISTORY {
						var listResultJson []string
						slotG.mutex.RLock()
						if _, isIn := slotG.mapPlayerIdToHistory[action.playerId]; isIn {
							listResultJson = slotG.mapPlayerIdToHistory[action.playerId].Elements
						} else {
							listResultJson = []string{}
						}
						data := map[string]interface{}{
							"myLast10":     listResultJson,
							"last10BigWin": slotG.bigWinList.Elements,
						}
						slotG.mutex.RUnlock()
						slotG.SendDataToPlayerId(
							"SlotHistory",
							data,
							action.playerId)
						action.chanResponse <- &ActionResponse{err: nil}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New("wrong action")}
					}
				}
			}(slotG, action)
		}
	}
}

// public jackpot, use lock game inside
func LoopPublicJackpotInfo(slotG *SlotGame) {
	for {
		time.Sleep(5 * time.Second)
		slotG.mutex.RLock()
		activePids := make([]int64, len(slotG.mapPlayerIdToIsActive))
		for pid, _ := range slotG.mapPlayerIdToIsActive {
			activePids = append(activePids, pid)
		}
		slotG.mutex.RUnlock()
		for _, pid := range activePids {
			slotG.SendDataToPlayerId(
				"SlotJackpots",
				map[string]interface{}{
					SLOT_JACKPOT_CODE_100:   slotG.jackpot100.Value(),
					SLOT_JACKPOT_CODE_1000:  slotG.jackpot1000.Value(),
					SLOT_JACKPOT_CODE_10000: slotG.jackpot10000.Value(),
				},
				pid)
		}
	}
}

func LoopFakeJackpot(slotG *SlotGame) {
	time.Sleep(5 * time.Second)
	for i, jackpotObj := range []*jackpot.Jackpot{
		slotG.jackpot100, slotG.jackpot1000, slotG.jackpot10000,
	} {
		go func(i int, jackpotObj *jackpot.Jackpot) {
			for {
				hitPeriodInMinutes := (17*i + 1) * (40 + rand.Intn(20))
				hitValue := (18000 + rand.Int63n(3000)) * int64(math.Pow10(i+2))

				baseJackpotValue := 10000 * int64(math.Pow10(i+2))
				gpm := (hitValue - baseJackpotValue) / int64(hitPeriodInMinutes)
				gps := gpm / 60

				for jackpotObj.Value() <= hitValue {
					jackpotObj.AddMoney(gps)
					time.Sleep(1 * time.Second)
				}
				//
				fakeName := player.BotUsernames[rand.Intn(len(player.BotUsernames))]
				jV := jackpotObj.Value()
				jackpotObj.NotifySomeoneHitJackpot(
					SLOT_GAME_CODE, jV, -1, fakeName)
				jackpotObj.AddMoney(-jV + baseJackpotValue)
				zmisc.InsertNewGlobalText(map[string]interface{}{
					"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
					"username": fakeName,
					"wonMoney": jV,
					"gamecode": SLOT_GAME_CODE,
				})
				slotG.mutex.Lock()
				m1 := map[int]int64{}
				m2 := map[int]bool{}
				nWR := 10 + rand.Intn(10)
				for ti := 0; ti < 20; ti++ {
					if ti <= nWR {
						m1[ti] = 1
						m2[ti] = true
					} else {
						m1[ti] = 0
						m2[ti] = false
					}
				}
				bytes, _ := json.Marshal(map[string]interface{}{
					"id":                        -1,
					"username":                  fakeName,
					"startedTime":               time.Now().Format(time.RFC3339),
					"matchId":                   -1,
					"moneyPerLine":              int64(math.Pow10(i + 2)),
					"slotResult":                [][]string{},
					"sumWonMoney":               jV,
					"mapPaylineIndexToWonMoney": m1,
					"mapPaylineIndexToIsWin":    m2,
					"matchWonType":              MATCH_WON_TYPE_JACKPOT,
					"change":                    jV,
				})
				slotG.bigWinList.Append(string(bytes))
				slotG.mutex.Unlock()
			}
		}(i, jackpotObj)

	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (slotG *SlotGame) GetGameCode() string {
	return slotG.gameCode
}

func (slotG *SlotGame) GetCurrencyType() string {
	return slotG.currencyType
}

func (slotG *SlotGame) SerializeData() map[string]interface{} {
	//	for _, jackpotObj := range jackpot.Jackpots() {
	//		_ = jackpotObj
	// fmt.Println("jackpot.Jackpots", jackpotObj.SerializedData())
	//	}
	//fmt.Println("slotG.currencyType", slotG.currencyType)
	//fmt.Println("slotG.jackpot100", slotG.jackpot100)
	result := map[string]interface{}{
		"gameCode":     slotG.gameCode,
		"currencyType": slotG.currencyType,
		"tax":          slotG.tax,

		"SYMBOLS":                       SYMBOLS,
		"PAYLINES":                      PAYLINES,
		"MONEYS_PER_LINE":               MONEYS_PER_LINE,
		"MAP_SYLBOL_NDUP_TO_PRIZE_RATE": MAP_SYLBOL_NDUP_TO_PRIZE_RATE,

		"jackpot100":   slotG.jackpot100.Value(),
		"jackpot1000":  slotG.jackpot1000.Value(),
		"jackpot10000": slotG.jackpot10000.Value(),
	}
	return result
}

func (slotG *SlotGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

// mark a player active in slot, for public jackpots price
// after 60s mark inactive
func (slotG *SlotGame) SetPlayerActiveStatus(playerId int64) {
	slotG.mutex.Lock()
	slotG.mapPlayerIdToIsActive[playerId] = true
	slotG.mapPlayerIdToLastActiveTime[playerId] = time.Now()
	slotG.mutex.Unlock()
	go func(playerId int64) {
		timeout := time.After(60 * time.Second)
		<-timeout
		slotG.mutex.Lock()
		if time.Now().Sub(slotG.mapPlayerIdToLastActiveTime[playerId]) >= 60*time.Second {
			delete(slotG.mapPlayerIdToIsActive, playerId)
		}
		slotG.mutex.Unlock()
	}(playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(slotG *SlotGame, action *Action) error {
	slotG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		fmt.Println("ERROR: slot timeout, action: ", action.ToMap())
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (slotG *SlotGame) ChooseMoneyPerLine(player *player.Player, moneyPerLine int64) error {
	action := &Action{
		actionName: ACTION_CHOOSE_MONEY_PER_LINE,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"moneyPerLine": moneyPerLine,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) ChoosePaylines(player *player.Player, paylineIndexs []int) error {
	action := &Action{
		actionName: ACTION_CHOOSE_PAYLINES,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"paylineIndexs": paylineIndexs,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) GetHistory(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_HISTORY,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}

// aaa,
func (slotG *SlotGame) Spin(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_SPIN,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(slotG, action)
}
