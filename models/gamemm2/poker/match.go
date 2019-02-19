package poker

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/gamemm2"
	"github.com/vic/vic_go/models/gamemm2/zhelp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

const (
	ACTION_FINISH_MATCH = "ACTION_FINISH_MATCH"

	DURATION_TURN         = 10 * time.Second
	DURATION_CHANGE_ROUND = 1 * time.Second

	ACTION_M_MAKE_MOVE = "ACTION_M_MAKE_MOVE"
)

func init() {
	_ = json.Marshal
	_ = errors.New
	_ = fmt.Sprintf
	_ = rand.Intn
	_ = debug.Stack
	_ = z.NewDeck
	_ = zmisc.GLOBAL_TEXT_LOWER_BOUND
	_ = record.LogMatchRecord

}

type ExMatch struct {
	Game        *ExGame
	Lobby       *gamemm2.Lobby
	StartedTime time.Time
	MatchId     string

	MapPidToPlayer  map[int64]*player.Player
	MapSeatToPlayer map[int]*player.Player
	MapPidToChip    map[int64]int64

	PokerBoard *PokerBoard

	// receive action from game.LoopReceiveActions
	ChanAction chan *gamemm2.Action
	// receive move info from valid action
	ChanMove chan Move
	// respond for a errMove
	ChanMoveErr chan error

	mutex sync.Mutex
}

func (match *ExMatch) CheckContainingPlayer(pid int64) bool {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result := false
	for inMatchPid, _ := range match.MapPidToPlayer {
		if pid == inMatchPid {
			result = true
			break
		}
	}
	return result
}

// data which all players can read
func (match *ExMatch) ToMap() map[string]interface{} {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result := match.PokerBoard.ToMap()
	return result
}

// shared data and data which only specific player can read
func (match *ExMatch) ToMapForPlayer(pid int64) map[string]interface{} {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result := match.PokerBoard.ToMapForPlayer(pid)
	return result
}

func (match *ExMatch) GetLobby() *gamemm2.Lobby {
	return match.Lobby
}

// forward action to match.LoopReceiveActions
func (match *ExMatch) ReceiveAction(action *gamemm2.Action) {
	t1 := time.After(3 * time.Second)
	select {
	case match.ChanAction <- action:
	case <-t1:
		fmt.Println("ERROR: poker match.ReceiveAction TimeOut")
	}
}

//
func Start(match *ExMatch) {
	match.mutex.Lock()
	listPids := make([]int64, 0)
	seats := make([]int, 0)
	for i := 0; i < len(match.MapSeatToPlayer); i++ {
		seats = append(seats, i)
		if match.MapSeatToPlayer[i] != nil {
			listPids = append(listPids, match.MapSeatToPlayer[i].Id())
		}
	}
	pokerLobbyData := match.Lobby.GameSpecificLobbyData.(*PokerLobbyData)
	dealerSeatI := pokerLobbyData.DealerSeat
	dealerPid := int64(0)
	for {
		p := match.MapSeatToPlayer[dealerSeatI]
		if p != nil {
			dealerPid = p.Id()
			break
		} else {
			dealerSeatI = zhelp.GetNextSeat(dealerSeatI, seats)
		}
	}
	match.PokerBoard = NewPokerBoard(
		listPids,
		match.MapPidToChip,
		dealerPid,
		0, match.Lobby.Rule.BaseMoney, 2*match.Lobby.Rule.BaseMoney)
	match.PokerBoard.StartDealingHoleCards()
	match.mutex.Unlock()
	match.Lobby.UpdateMatchStatus()
	time.Sleep(DURATION_CHANGE_ROUND)

	//
	for _, round := range []string{ROUND_PRE_FLOP, ROUND_FLOP,
		ROUND_TURN, ROUND_RIVER} {
		match.mutex.Lock()
		if round == ROUND_PRE_FLOP {
			match.PokerBoard.StartPreFlop()
		} else {
			match.PokerBoard.StartFlopOrTurnOrRiver(round)
		}
		match.mutex.Unlock()
		match.Lobby.UpdateMatchStatus()
		for {
			if match.PokerBoard.InRoundCurrentTurnPlayer == 0 {
				break
			} else {
				timeout := time.After(DURATION_TURN)
			ReceivingMoveLoop:
				for {
					select {
					case move := <-match.ChanMove:
						match.mutex.Lock()
						err := match.PokerBoard.MakeMove(move)
						match.mutex.Unlock()
						if err == nil {
							break ReceivingMoveLoop
						} else {
							select {
							case match.ChanMoveErr <- err:
							default:
							}
						}
					case <-timeout:
						match.mutex.Lock()
						err1 := match.PokerBoard.MakeMove(Move{
							PlayerId: match.PokerBoard.InRoundCurrentTurnPlayer,
							MoveType: MOVE_CHECK})
						match.mutex.Unlock()
						if err1 != nil {
							match.mutex.Lock()
							match.PokerBoard.MakeMove(Move{
								PlayerId: match.PokerBoard.InRoundCurrentTurnPlayer,
								MoveType: MOVE_FOLD})
							match.mutex.Unlock()
						}
						break ReceivingMoveLoop
					}
				}
			}
		}
		time.Sleep(DURATION_CHANGE_ROUND)
	}
	match.mutex.Lock()
	match.PokerBoard.StartShowdown()
	match.Lobby.MatchMaker.Mutex.Lock()
	mapPidToChange := make(map[int64]int64)
	for pid, _ := range match.PokerBoard.MapPlayerToWonChip {
		mapPidToChange[pid] = match.PokerBoard.MapPlayerToWonChip[pid] +
			match.PokerBoard.MapPlayerToLostChip[pid]
	}
	for pid, change := range mapPidToChange {
		match.Lobby.MapPidToChip[pid] += change
	}
	match.Lobby.MatchMaker.Mutex.Unlock()
	match.mutex.Unlock()
	match.Lobby.UpdateMatchStatus()
	time.Sleep(DURATION_CHANGE_ROUND)
	pokerLobbyData.DealerSeat = zhelp.GetNextSeat(dealerSeatI, seats)
	select {
	case match.ChanAction <- gamemm2.NewAction(ACTION_FINISH_MATCH, 0, nil):
	default:
		fmt.Println("¯\\_(ツ)_/¯")
	}
	match.Lobby.MatchMaker.HandleFinishedMatch(match)
}

func InMatchLoopReceiveActions(match *ExMatch) {
	for {
		action := <-match.ChanAction
		if action.ActionName == ACTION_FINISH_MATCH {
			action.ChanResponse <- nil
			break
		} else {
			go func(match *ExMatch, action *gamemm2.Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == ACTION_M_MAKE_MOVE {
					move := Move{
						PlayerId: action.PlayerId,
						MoveType: utils.GetStringAtPath(action.Data, "MoveType"),
						Value:    utils.GetInt64AtPath(action.Data, "Value"),
						IsAllIn:  utils.GetBoolAtPath(action.Data, "IsAllIn"),
					}
					select {
					case match.ChanMove <- move:
						t := time.After(1 * time.Second)
						select {
						case me := <-match.ChanMoveErr:
							action.ChanResponse <- me
						case <-t:
							action.ChanResponse <- nil
							match.Lobby.UpdateMatchStatus()
						}
					default:
						action.ChanResponse <- errors.New("Cant match.ChanMove <- move")
					}

				} else {
					action.ChanResponse <- errors.New("Wrong ActionName")
				}
			}(match, action)
		}
	}
}
