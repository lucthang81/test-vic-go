package slotagoldminer

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
	SLOTAGM_GAME_CODE = "slotagm"

	SLOTAGM_JACKPOT_CODE_SMALL  = "SLOTAGM_JACKPOT_CODE_SMALL"
	SLOTAGM_JACKPOT_CODE_MEDIUM = "SLOTAGM_JACKPOT_CODE_MEDIUM"
	SLOTAGM_JACKPOT_CODE_BIG    = "SLOTAGM_JACKPOT_CODE_BIG"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = cardgame.SizedList{}
	_ = jackpot.GetJackpot("", "")
}

func NewSlotagmGame(currencyType string) *gamemini.SlotGame {
	slotG := &gamemini.SlotGame{
		GameCode:     SLOTAGM_GAME_CODE,
		CurrencyType: currencyType,

		MatchCounter:                0,
		MapPlayerIdToMatch:          map[int64]*gamemini.SlotMatch{},
		MapPlayerIdToMoneyPerLine:   map[int64]int64{},
		MapPlayerIdToPayLineIndexs:  map[int64][]int{},
		MapPlayerIdToHistory:        map[int64]*cardgame.SizedList{},
		MapPlayerIdToIsActive:       map[int64]bool{},
		MapPlayerIdToLastActiveTime: map[int64]time.Time{},

		JackpotSmall:  jackpot.GetJackpot(SLOTAGM_JACKPOT_CODE_SMALL, currencyType),
		JackpotMedium: jackpot.GetJackpot(SLOTAGM_JACKPOT_CODE_MEDIUM, currencyType),
		JackpotBig:    jackpot.GetJackpot(SLOTAGM_JACKPOT_CODE_BIG, currencyType),

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
					SLOTAGM_JACKPOT_CODE_SMALL:  slotG.JackpotSmall.Value(),
					SLOTAGM_JACKPOT_CODE_MEDIUM: slotG.JackpotMedium.Value(),
					SLOTAGM_JACKPOT_CODE_BIG:    slotG.JackpotBig.Value(),
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
func ChooseGoldPotIndex(
	slotG *gamemini.SlotGame, player *player.Player, potIndex int) error {
	action := &gamemini.Action{
		ActionName:   consts.ACTION_CHOOSE_GOLD_POT_INDEX,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{"potIndex": potIndex},
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	return gamemini.DoPlayerAction(slotG, action)
}

func StopPlaying(slotG *gamemini.SlotGame, player *player.Player) error {
	action := &gamemini.Action{
		ActionName:   consts.ACTION_STOP_PLAYING,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	return gamemini.DoPlayerAction(slotG, action)
}
