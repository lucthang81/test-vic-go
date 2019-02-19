package exgame

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
	Game           *ExGame
	Lobby          *gamemm2.Lobby
	StartedTime    time.Time
	MatchId        string
	MapPidToPlayer map[int64]*player.Player

	//	PokerBoard *PokerBoard

	// receive action from game.LoopReceiveActions
	ChanAction chan *gamemm2.Action
	// receive move info from valid action
	//	ChanMove chan Move

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
	result := map[string]interface{}{
		"": "",
	}
	return result
}

// shared data and data which only specific player can read
func (match *ExMatch) ToMapForPlayer(pid int64) map[string]interface{} {
	result := match.ToMap()
	match.mutex.Lock()
	defer match.mutex.Unlock()
	result["privateField"] = 0
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

func Start(match *ExMatch) {
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

				if action.ActionName == "EX_ACTION_0" {
					// do something
					action.ChanResponse <- nil
					match.Lobby.UpdateMatchStatus()
				} else {
					action.ChanResponse <- errors.New("hihi")
				}
			}(match, action)
		}
	}
}
