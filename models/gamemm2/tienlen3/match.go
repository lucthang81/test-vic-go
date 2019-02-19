package tienlen3

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
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
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

	TienlenBoard     *TienlenBoard
	TurnStartingTime time.Time

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

// full data, just for debug
func (match *ExMatch) ToMap() map[string]interface{} {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result := match.TienlenBoard.ToMap()
	result["TurnRemainingSeconds"] =
		match.TurnStartingTime.Add(DURATION_TURN).Sub(time.Now()).Seconds()
	return result
}

// shared data and data which only specific player can read
func (match *ExMatch) ToMapForPlayer(pid int64) map[string]interface{} {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result := match.TienlenBoard.ToMapForPlayer(pid)
	result["TurnRemainingSeconds"] =
		match.TurnStartingTime.Add(DURATION_TURN).Sub(time.Now()).Seconds()
	result["TurnDuration"] = DURATION_TURN.Seconds()
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
	priorityPlayers := make([]int64, 0)
	for i := 0; i < len(match.MapSeatToPlayer); i++ {
		seats = append(seats, i)
		pObj := match.MapSeatToPlayer[i]
		if pObj != nil {
			listPids = append(listPids, pObj.Id())
			if match.MapSeatToPlayer[i].PlayerType() == "bot" {
				priorityPlayers = append(priorityPlayers, pObj.Id())
			}
		}
	}
	match.TienlenBoard = NewTienlenBoard(
		listPids, priorityPlayers, match.Lobby.Rule.BaseMoney)
	match.TienlenBoard.StartDealing()
	match.mutex.Unlock()
	match.Lobby.UpdateMatchStatus()
	time.Sleep(DURATION_CHANGE_ROUND)

	//
	for true {
		match.mutex.Lock()
		if match.TienlenBoard.CurrentTurnPlayer == 0 {
			match.mutex.Unlock()
			break
		}
		match.mutex.Unlock()
		//
		timeout := time.After(DURATION_TURN)
	ReceivingMoveLoop:
		for {
			select {
			case move := <-match.ChanMove:
				match.mutex.Lock()
				err := match.TienlenBoard.MakeMove(move)
				if err == nil {
					match.TurnStartingTime = time.Now()
				}
				match.mutex.Unlock()
				select {
				case match.ChanMoveErr <- err:
				default:
				}
				if err == nil {
					break ReceivingMoveLoop
				}
			case <-timeout:
				match.mutex.Lock()
				match.TienlenBoard.MakeNatureMove()
				match.TurnStartingTime = time.Now()
				match.mutex.Unlock()
				break ReceivingMoveLoop
			}
		}
		match.Lobby.UpdateMatchStatus()
	}
	time.Sleep(DURATION_CHANGE_ROUND)
	// TODO: update board.MapPlayerToChangedChip to lobby.Chip
	match.mutex.Lock()
	match.Lobby.MatchMaker.Mutex.Lock()
	for pid, changedMoney := range match.TienlenBoard.MapPlayerToChangedChip {
		match.Lobby.MapPidToChip[pid] += changedMoney
	}
	match.Lobby.MatchMaker.Mutex.Unlock()
	match.mutex.Unlock()
	//
	match.Lobby.UpdateMatchStatus()
	time.Sleep(DURATION_CHANGE_ROUND)
	select {
	case match.ChanAction <- gamemm2.NewAction(ACTION_FINISH_MATCH, 0, nil):
	default:
		fmt.Println("ERROR: có cái gì đó sai sai")
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
					t1, isOk := action.Data["Cards"].([]interface{})
					if !isOk {
						action.ChanResponse <- errors.New("Bài đánh sai định dạng")
					} else {
						t2 := make([]string, 0)
						for _, e := range t1 {
							es, _ := e.(string)
							t2 = append(t2, es)
						}
						cards, err := z.ToCardsFromStrings(t2)
						if err != nil {
							action.ChanResponse <- err
						} else {
							move := Move{
								PlayerId: action.PlayerId,
								Cards:    cards,
							}
							select {
							case match.ChanMove <- move:
								t := time.After(1 * time.Second)
								select {
								case me := <-match.ChanMoveErr:
									action.ChanResponse <- me
								case <-t:
									action.ChanResponse <- errors.New("<-match.ChanMoveErr timeout")
								}
							default:
								action.ChanResponse <- errors.New("Cant match.ChanMove <- move")
							}
						}
					}
				} else {
					action.ChanResponse <- errors.New("Wrong ActionName")
				}
			}(match, action)
		}
	}
}
