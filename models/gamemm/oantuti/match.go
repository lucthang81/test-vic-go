package oantuti

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/gamemm"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
)

const (
	ACTION_FINISH_MATCH = "ACTION_FINISH_MATCH"

	ACTION_HAND_ROCK     = "ACTION_HAND_ROCK"
	ACTION_HAND_PAPER    = "ACTION_HAND_PAPER"
	ACTION_HAND_SCISSORS = "ACTION_HAND_SCISSORS"

	PHASE_1_CHOOSE_HAND = "PHASE_1_CHOOSE_HAND"
	PHASE_2_RESULT      = "PHASE_2_RESULT"
)

type OantutiMatch struct {
	game    *OantutiGame
	lobby   *gamemm.Lobby
	matchId string

	phase string

	tax int64

	p1Id                   int64
	p1Hand                 string
	p1Streak               int64
	p1LosingStreak         int64
	p1EndMatchWinningMoney int64

	p2Id                   int64
	p2Hand                 string
	p2Streak               int64
	p2LosingStreak         int64
	p2EndMatchWinningMoney int64

	ChanAction chan *gamemm.Action

	mutex sync.Mutex
}

func (match *OantutiMatch) GetLobby() *gamemm.Lobby {
	return match.lobby
}

func NewOantutiMatch(game *OantutiGame, lobby *gamemm.Lobby) *OantutiMatch {
	hands := []string{ACTION_HAND_PAPER, ACTION_HAND_ROCK, ACTION_HAND_SCISSORS}

	match := &OantutiMatch{
		game:    game,
		lobby:   lobby,
		matchId: fmt.Sprintf("%v_%v", game.GetGameCode(), time.Now().UnixNano()),
		phase:   "PHASE_0_INIT",
		p1Hand:  hands[rand.Intn(len(hands))],
		p2Hand:  hands[rand.Intn(len(hands))],
	}

	game.matchMaker.Mutex.Lock()
	pids := make([]int64, 0)
	for pid, _ := range lobby.MapPidToPlayers {
		pids = append(pids, pid)
	}
	if len(pids) >= 2 {
		match.p1Id = pids[0]
		match.p2Id = pids[1]
	}
	game.matchMaker.Mutex.Unlock()

	top.GlobalMutex.Lock()
	trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
	trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
	top.GlobalMutex.Unlock()
	if trackerW != nil {
		trackerW.Mutex.Lock()
		match.p1Streak = trackerW.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2Streak = trackerW.MapPlayerIdToCurrentValue[match.p2Id]
		trackerW.Mutex.Unlock()

		trackerL.Mutex.Lock()
		match.p1LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p2Id]
		trackerL.Mutex.Unlock()
	}

	match.ChanAction = make(chan *gamemm.Action)

	go Start(match)
	go InMatchLoopReceiveActions(match)

	return match
}

func (match *OantutiMatch) ReceiveAction(action *gamemm.Action) *gamemm.ActionResponse {
	match.ChanAction <- action
	timeout := time.After(2 * time.Second)
	select {
	case res := <-action.ChanResponse:
		return res
	case <-timeout:
		res := &gamemm.ActionResponse{Err: errors.New("err:matchOnetuti_time_out")}
		return res
	}
}

func (match *OantutiMatch) SerializeData() map[string]interface{} {
	data := map[string]interface{}{
		"matchId": match.matchId,
		"phase":   match.phase,

		"p1Id":                   match.p1Id,
		"p1Hand":                 match.p1Hand,
		"p1Streak":               match.p1Streak,
		"p1LosingStreak":         match.p1LosingStreak,
		"p1EndMatchWinningMoney": match.p1EndMatchWinningMoney,

		"p2Id":                   match.p2Id,
		"p2Hand":                 match.p2Hand,
		"p2Streak":               match.p2Streak,
		"p2LosingStreak":         match.p1LosingStreak,
		"p2EndMatchWinningMoney": match.p2EndMatchWinningMoney,
	}
	return data
}

func (match *OantutiMatch) updateMatchStatus() {
	data := match.SerializeData()
	for _, pid := range []int64{match.p1Id, match.p2Id} {
		gamemm.ServerObj.SendRequest(
			"OantutiUpdateMatchStatus",
			data,
			pid)
	}
}

func Start(match *OantutiMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	match.phase = PHASE_1_CHOOSE_HAND
	match.updateMatchStatus()
	timer := time.After(2 * time.Second)
	<-timer

	match.phase = PHASE_2_RESULT
	baseMoney := match.lobby.Rule.RequirementMoney
	moneyType := match.lobby.Rule.MoneyType
	taxRate := match.game.tax
	if match.p1Hand == match.p2Hand {
		match.tax = int64(taxRate * float64(baseMoney))
		match.p1EndMatchWinningMoney = baseMoney - match.tax/2
		match.p2EndMatchWinningMoney = baseMoney - match.tax/2
	} else if ((match.p1Hand == ACTION_HAND_ROCK) && (match.p2Hand == ACTION_HAND_SCISSORS)) ||
		((match.p1Hand == ACTION_HAND_SCISSORS) && (match.p2Hand == ACTION_HAND_PAPER)) ||
		((match.p1Hand == ACTION_HAND_PAPER) && (match.p2Hand == ACTION_HAND_ROCK)) {
		match.tax = int64(taxRate * float64(baseMoney))
		match.p1EndMatchWinningMoney = 2*baseMoney - match.tax

		top.GlobalMutex.Lock()
		trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
		trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
		maxStreakW := top.MapEvents[top.EVENT_OANTUTI_WINNING_STREAK]
		maxStreakL := top.MapEvents[top.EVENT_OANTUTI_LOSING_STREAK]
		top.GlobalMutex.Unlock()

		trackerW.ChangeValue(match.p1Id, 1)
		trackerW.SetNewValue(match.p2Id, 0)
		trackerL.SetNewValue(match.p1Id, 0)
		trackerL.ChangeValue(match.p2Id, 1)

		if maxStreakW != nil {
			maxStreakW.ChangeValue(match.p1Id, 1)
			maxStreakW.SetNewValue(match.p2Id, 0)
		}
		if maxStreakL != nil {
			maxStreakL.SetNewValue(match.p1Id, 0)
			maxStreakL.ChangeValue(match.p2Id, 1)
		}

		trackerW.Mutex.Lock()
		match.p1Streak = trackerW.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2Streak = trackerW.MapPlayerIdToCurrentValue[match.p2Id]
		trackerW.Mutex.Unlock()
		trackerL.Mutex.Lock()
		match.p1LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p2Id]
		trackerL.Mutex.Unlock()

	} else {
		match.tax = int64(taxRate * float64(baseMoney))
		match.p2EndMatchWinningMoney = 2*baseMoney - match.tax

		top.GlobalMutex.Lock()
		trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
		trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
		maxStreakW := top.MapEvents[top.EVENT_OANTUTI_WINNING_STREAK]
		maxStreakL := top.MapEvents[top.EVENT_OANTUTI_LOSING_STREAK]
		top.GlobalMutex.Unlock()

		trackerW.ChangeValue(match.p2Id, 1)
		trackerW.SetNewValue(match.p1Id, 0)
		trackerL.SetNewValue(match.p2Id, 0)
		trackerL.ChangeValue(match.p1Id, 1)

		if maxStreakW != nil {
			maxStreakW.ChangeValue(match.p1Id, 1)
			maxStreakW.SetNewValue(match.p2Id, 0)
		}
		if maxStreakL != nil {
			maxStreakL.SetNewValue(match.p1Id, 0)
			maxStreakL.ChangeValue(match.p2Id, 1)
		}

		trackerW.Mutex.Lock()
		match.p1Streak = trackerW.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2Streak = trackerW.MapPlayerIdToCurrentValue[match.p2Id]
		trackerW.Mutex.Unlock()
		trackerL.Mutex.Lock()
		match.p1LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p1Id]
		match.p2LosingStreak = trackerL.MapPlayerIdToCurrentValue[match.p2Id]
		trackerL.Mutex.Unlock()

	}
	//
	if match.p1Streak >= 20 || match.p1LosingStreak >= 20 ||
		match.p2Streak >= 20 || match.p2LosingStreak >= 20 {
		var jpWinnerId int64
		if match.p1Streak >= 20 || match.p1LosingStreak >= 20 {
			jpWinnerId = match.p1Id
		} else {
			jpWinnerId = match.p2Id
		}
		var jpRate float64
		if match.lobby.Rule.RequirementMoney == 1000 {
			jpRate = 0.10
		} else if match.lobby.Rule.RequirementMoney == 2000 {
			jpRate = 0.20
		} else if match.lobby.Rule.RequirementMoney == 5000 {
			jpRate = 0.50
		} else if match.lobby.Rule.RequirementMoney == 10000 {
			jpRate = 0.60
		} else if match.lobby.Rule.RequirementMoney == 20000 {
			jpRate = 0.70
		} else if match.lobby.Rule.RequirementMoney == 50000 {
			jpRate = 0.80
		} else {
			jpRate = 0
		}

		var winnerName string
		pObj, _ := player.GetPlayer(jpWinnerId)
		if pObj != nil && pObj.PlayerType() == "normal" {
			winnerName = pObj.Name()
			amount := int64(float64(match.game.jackpot.Value()) * jpRate)
			if jpWinnerId == match.p1Id {
				match.p1EndMatchWinningMoney += amount
			} else {
				match.p2EndMatchWinningMoney += amount
			}
			match.game.jackpot.AddMoney(-amount)
			match.game.jackpot.NotifySomeoneHitJackpot(
				match.game.gameCode,
				amount,
				jpWinnerId,
				winnerName,
			)
		}

		//
		top.GlobalMutex.Lock()
		trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
		trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
		top.GlobalMutex.Unlock()
		if match.p1Streak >= 20 || match.p2Streak >= 20 {
			trackerW.SetNewValue(jpWinnerId, 0)
		} else {
			trackerL.SetNewValue(jpWinnerId, 0)
		}
	}
	//
	p1Obj, _ := player.GetPlayer(match.p1Id)
	p2Obj, _ := player.GetPlayer(match.p2Id)
	if match.p1EndMatchWinningMoney > 0 {
		if p1Obj != nil {
			p1Obj.ChangeMoneyAndLog(
				match.p1EndMatchWinningMoney, moneyType, false, "",
				ACTION_FINISH_MATCH, match.game.GetGameCode(), match.matchId)
		}
	}
	if match.p2EndMatchWinningMoney > 0 {
		if p2Obj != nil {
			p2Obj.ChangeMoneyAndLog(
				match.p2EndMatchWinningMoney, moneyType, false, "",
				ACTION_FINISH_MATCH, match.game.GetGameCode(), match.matchId)
		}
	}
	// LogMatchRecord2
	isHumanInMatch := false
	var humanWon, humanLost, botWon, botLost int64
	if p1Obj != nil {
		if p1Obj.PlayerType() == "bot" {
			botWon += match.p1EndMatchWinningMoney
			botLost += baseMoney
		} else {
			humanWon += match.p1EndMatchWinningMoney
			humanLost += baseMoney
			isHumanInMatch = true
		}
	}
	if p2Obj != nil {
		if p2Obj.PlayerType() == "bot" {
			botWon += match.p2EndMatchWinningMoney
			botLost += baseMoney
		} else {
			humanWon += match.p2EndMatchWinningMoney
			humanLost += baseMoney
			isHumanInMatch = true
		}
	}
	if isHumanInMatch {
		temp := match.tax / 3
		match.game.jackpot.AddMoney(temp)
		record.LogMatchRecord2(
			match.game.GetGameCode(), moneyType, baseMoney, match.tax,
			humanWon, humanLost, botWon, botLost,
			match.matchId, map[int64]string{match.p1Id: "", match.p2Id: ""},
			[]map[string]interface{}{})
	}
	//
	match.updateMatchStatus()
	//
	match.game.matchMaker.ClearLobby(match.lobby)
	//
	action := gamemm.Action{
		ActionName:   ACTION_FINISH_MATCH,
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	match.ChanAction <- &action
	<-action.ChanResponse
	//fmt.Println("finished match", match)
}

func InMatchLoopReceiveActions(match *OantutiMatch) {
	for {
		action := <-match.ChanAction
		if action.ActionName == ACTION_FINISH_MATCH {
			action.ChanResponse <- &gamemm.ActionResponse{Err: nil}
			break
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == ACTION_HAND_ROCK ||
					action.ActionName == ACTION_HAND_PAPER ||
					action.ActionName == ACTION_HAND_SCISSORS {
					if match.phase == PHASE_1_CHOOSE_HAND {
						if action.PlayerId == match.p1Id {
							match.p1Hand = action.ActionName
						} else {
							match.p2Hand = action.ActionName
						}
						action.ChanResponse <- &gamemm.ActionResponse{Err: nil}
					} else {
						action.ChanResponse <- &gamemm.ActionResponse{
							Err: errors.New("Wrong phase")}
					}
				} else {
					action.ChanResponse <- &gamemm.ActionResponse{
						Err: errors.New("Wrong action")}
				}

			}()
		}
	}
	//fmt.Println("finished actionLoop match", match)
}
