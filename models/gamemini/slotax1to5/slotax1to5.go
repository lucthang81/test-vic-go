package slotax1to5

import (
	//	"errors"
	"fmt"
	//	"runtime/debug"
	"encoding/json"
	"math"
	"math/rand"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/record"
	//	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/zconfig"
)

const (
	SLOTAX1TO5_GAME_CODE = "slotax1to5"

	SLOTAX1TO5_JACKPOT_CODE_SMALL  = "SLOTAX1TO5_JACKPOT_CODE_SMALL"
	SLOTAX1TO5_JACKPOT_CODE_MEDIUM = "SLOTAX1TO5_JACKPOT_CODE_MEDIUM"
	SLOTAX1TO5_JACKPOT_CODE_BIG    = "SLOTAX1TO5_JACKPOT_CODE_BIG"

	IS_FAKE_JACKPOT = true
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

func NewSlotax1to5Game(currencyType string) *gamemini.SlotGame {
	slotG := &gamemini.SlotGame{
		GameCode:     SLOTAX1TO5_GAME_CODE,
		CurrencyType: currencyType,

		MatchCounter:                0,
		MapPlayerIdToMatch:          map[int64]*gamemini.SlotMatch{},
		MapPlayerIdToMoneyPerLine:   map[int64]int64{},
		MapPlayerIdToPayLineIndexs:  map[int64][]int{},
		MapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		MapPlayerIdToIsActive:       map[int64]bool{},
		MapPlayerIdToLastActiveTime: map[int64]time.Time{},

		JackpotSmall:  jackpot.GetJackpot(SLOTAX1TO5_JACKPOT_CODE_SMALL, currencyType),
		JackpotMedium: jackpot.GetJackpot(SLOTAX1TO5_JACKPOT_CODE_MEDIUM, currencyType),
		JackpotBig:    jackpot.GetJackpot(SLOTAX1TO5_JACKPOT_CODE_BIG, currencyType),

		ChanActionReceiver: make(chan *gamemini.Action),
	}

	temp := cardgame.NewSizedList(10)
	slotG.BigWinList = &temp

	go gamemini.SlotGameLoopReceiveActions(slotG,
		MONEYS_PER_LINE, PAYLINES,
		StartMatchFunc)
	// go LoopPublicJackpotInfo(slotG)

	if IS_FAKE_JACKPOT {
		go LoopFakeJackpot(slotG)
	}

	return slotG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// public jackpot, use lock game inside
func LoopPublicJackpotInfo(slotG *gamemini.SlotGame) {
	for {
		time.Sleep(5 * time.Second)
		slotG.Mutex.RLock()
		activePids := make([]int64, len(slotG.MapPlayerIdToIsActive))
		for pid, _ := range slotG.MapPlayerIdToIsActive {
			activePids = append(activePids, pid)
		}
		slotG.Mutex.RUnlock()
		for _, pid := range activePids {
			slotG.SendDataToPlayerId(
				"SlotJackpots",
				map[string]interface{}{
					SLOTAX1TO5_JACKPOT_CODE_SMALL:  slotG.JackpotSmall.Value(),
					SLOTAX1TO5_JACKPOT_CODE_MEDIUM: slotG.JackpotMedium.Value(),
					SLOTAX1TO5_JACKPOT_CODE_BIG:    slotG.JackpotBig.Value(),
				},
				pid)
		}
	}
}

func LoopFakeJackpot(slotG *gamemini.SlotGame) {
	time.Sleep(6 * time.Second)
	for i, jackpotObj := range []*jackpot.Jackpot{
		slotG.JackpotSmall, slotG.JackpotMedium, slotG.JackpotBig,
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
				jackpotObj.AddMoney(-jV + baseJackpotValue)
				if zconfig.ServerVersion == zconfig.SV_02 {
					jackpotObj.NotifySomeoneHitJackpot(
						SLOTAX1TO5_GAME_CODE, jV, -1, fakeName)
					zmisc.InsertNewGlobalText(map[string]interface{}{
						"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
						"username": fakeName,
						"wonMoney": jV,
						"gamecode": SLOTAX1TO5_GAME_CODE,
					})
				}
				slotG.Mutex.Lock()
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
					"matchWonType":              consts.MATCH_WON_TYPE_JACKPOT,
					"change":                    jV,
				})
				slotG.BigWinList.Append(string(bytes))
				slotG.Mutex.Unlock()
			}
		}(i, jackpotObj)

	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////
func Choose(slotG *gamemini.SlotGame, player *player.Player) error {
	action := &gamemini.Action{
		ActionName:   consts.ACTION_CHOOSE,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	return gamemini.DoPlayerAction(slotG, action)
}
