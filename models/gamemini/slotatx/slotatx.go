package slotatx

import (
	//	"errors"
	"fmt"
	//	"runtime/debug"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/record"
	//	"github.com/vic/vic_go/utils"
)

const (
	SLOTATX_GAME_CODE = "slotatx"

	SLOTATX_JACKPOT_CODE_SMALL  = "SLOTATX_JACKPOT_CODE_SMALL"
	SLOTATX_JACKPOT_CODE_MEDIUM = "SLOTATX_JACKPOT_CODE_MEDIUM"
	SLOTATX_JACKPOT_CODE_BIG    = "SLOTATX_JACKPOT_CODE_BIG"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

func NewSlotatxGame(currencyType string) *gamemini.SlotGame {
	slotG := &gamemini.SlotGame{
		GameCode:     SLOTATX_GAME_CODE,
		CurrencyType: currencyType,

		MatchCounter:                0,
		MapPlayerIdToMatch:          map[int64]*gamemini.SlotMatch{},
		MapPlayerIdToMoneyPerLine:   map[int64]int64{},
		MapPlayerIdToPayLineIndexs:  map[int64][]int{},
		MapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		MapPlayerIdToIsActive:       map[int64]bool{},
		MapPlayerIdToLastActiveTime: map[int64]time.Time{},

		JackpotSmall:  jackpot.GetJackpot(SLOTATX_JACKPOT_CODE_SMALL, currencyType),
		JackpotMedium: jackpot.GetJackpot(SLOTATX_JACKPOT_CODE_MEDIUM, currencyType),
		JackpotBig:    jackpot.GetJackpot(SLOTATX_JACKPOT_CODE_BIG, currencyType),

		ChanActionReceiver: make(chan *gamemini.Action),
	}

	temp := cardgame.NewSizedList(10)
	slotG.BigWinList = &temp

	go gamemini.SlotGameLoopReceiveActions(slotG,
		MONEYS_PER_LINE, PAYLINES,
		StartMatchFunc)
	// go LoopPublicJackpotInfo(slotG)

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
					SLOTATX_JACKPOT_CODE_SMALL:  slotG.JackpotSmall.Value(),
					SLOTATX_JACKPOT_CODE_MEDIUM: slotG.JackpotMedium.Value(),
					SLOTATX_JACKPOT_CODE_BIG:    slotG.JackpotBig.Value(),
				},
				pid)
		}
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
func SelectSmall(slotG *gamemini.SlotGame, player *player.Player) error {
	action := &gamemini.Action{
		ActionName:   consts.ACTION_SELECT_SMALL,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	return gamemini.DoPlayerAction(slotG, action)
}

func SelectBig(slotG *gamemini.SlotGame, player *player.Player) error {
	action := &gamemini.Action{
		ActionName:   consts.ACTION_SELECT_BIG,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	return gamemini.DoPlayerAction(slotG, action)
}
