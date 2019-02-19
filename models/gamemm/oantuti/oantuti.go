package oantuti

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemm"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

const (
	GAME_CODE_OANTUTI = "oantuti"

	ACTION_CHOOSE_RULE   = "ACTION_CHOOSE_RULE"
	ACTION_GET_USER_INFO = "ACTION_GET_USER_INFO"
	ACTION_GET_TOP       = "ACTION_GET_TOP"

	OANTUTI_JACKPOT_CODE = "OANTUTI_JACKPOT_CODE"
)

type OantutiGame struct {
	nPlayersToStartPlaying int
	gameCode               string
	tax                    float64

	matchMaker *gamemm.MatchMaker

	mapPidToRule map[int64]gamemm.Rule

	topStreak       string
	topLosingStreak string

	jackpot *jackpot.Jackpot

	ChanAction chan *gamemm.Action

	mutex sync.Mutex
}

func NewOantutiGame() *OantutiGame {
	game := &OantutiGame{
		gameCode:               GAME_CODE_OANTUTI,
		nPlayersToStartPlaying: 2,
		tax: 0.03,

		mapPidToRule: make(map[int64]gamemm.Rule),

		jackpot: jackpot.GetJackpot(OANTUTI_JACKPOT_CODE, currency.Money),

		ChanAction: make(chan *gamemm.Action),
	}
	game.matchMaker = gamemm.NewMatchMaker(game, game.nPlayersToStartPlaying)

	go LoopUpdateTop(game)
	go LoopReceiveActions(game)

	return game
}

func (game *OantutiGame) GetGameCode() string {
	return GAME_CODE_OANTUTI
}

func (game *OantutiGame) StartMatch(lobby *gamemm.Lobby) gamemm.MatchInterface {
	match := NewOantutiMatch(game, lobby)
	return match
}

func (game *OantutiGame) GetMatchMaker() *gamemm.MatchMaker {
	return game.matchMaker
}

func (game *OantutiGame) SerializeData() map[string]interface{} {
	return map[string]interface{}{
		"gameCode":     game.gameCode,
		"currencyType": currency.Money,
		"tax":          game.tax,
	}
}

func (game *OantutiGame) GetRuleForPlayer(pid int64) gamemm.Rule {
	pRule := gamemm.Rule{MoneyType: currency.Money, RequirementMoney: 1000}
	game.mutex.Lock()
	if rule, isIn := game.mapPidToRule[pid]; isIn == true {
		pRule = rule
	}
	game.mutex.Unlock()
	return pRule
}

type TopRow struct {
	PlayerId          int64
	PlayerName        string
	PlayerDisplayName string
	Streak            int64
}

type OrderByStreakDesc []TopRow

func (a OrderByStreakDesc) Len() int {
	return len(a)
}
func (a OrderByStreakDesc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a OrderByStreakDesc) Less(i, j int) bool {
	return a[i].Streak > a[j].Streak
}

func LoopUpdateTop(game *OantutiGame) {
	for {
		top.GlobalMutex.Lock()
		trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
		trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
		top.GlobalMutex.Unlock()
		for k, topMap := range []map[int64]int64{
			trackerW.MapPlayerIdToCurrentValue,
			trackerL.MapPlayerIdToCurrentValue,
		} {
			a := make([]TopRow, 0)
			if k == 0 {
				trackerW.Mutex.Lock()
			} else {
				trackerL.Mutex.Lock()
			}
			for pid, streak := range topMap {
				a = append(a, TopRow{PlayerId: pid, Streak: streak})
			}
			if k == 0 {
				trackerW.Mutex.Unlock()
			} else {
				trackerL.Mutex.Unlock()
			}
			sort.Sort(OrderByStreakDesc(a))
			if len(a) <= 10 {
				a = a[:len(a)]
			} else {
				a = a[:10]
			}
			for i, _ := range a {
				pObj, _ := player.GetPlayer(a[i].PlayerId)
				if pObj != nil {
					a[i].PlayerName = pObj.Name()
					a[i].PlayerDisplayName = pObj.DisplayName()
				}
			}
			bytes, _ := json.Marshal(a)
			if k == 0 {
				game.topStreak = string(bytes)
			} else {
				game.topLosingStreak = string(bytes)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (game *OantutiGame) ReceiveAction(action *gamemm.Action) *gamemm.ActionResponse {
	game.ChanAction <- action
	timeout := time.After(4 * time.Second)
	select {
	case res := <-action.ChanResponse:
		return res
	case <-timeout:
		res := &gamemm.ActionResponse{Err: errors.New("err:gameOnetuti_time_out")}
		return res
	}

	//	t1 := time.After(3 * time.Second)
	//	select {
	//	case game.ChanAction <- action:
	//		t2 := time.After(3 * time.Second)
	//		select {
	//		case res := <-action.ChanResponse:
	//			return res.Err
	//		case <-t2:
	//			res := &gamemm.ActionResponse{Err: errors.New("err:gameOnetuti_time_out")}
	//			return res
	//		}
	//	case <-t1:
	//		return errors.New("err:sending_time_out")
	//	}
}

func LoopReceiveActions(game *OantutiGame) {
	for {
		action := <-game.ChanAction
		go func() {
			defer func() {
				if r := recover(); r != nil {
					bytes := debug.Stack()
					fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
				}
			}()

			playerObj, err := player.GetPlayer(action.PlayerId)
			if err != nil {
				action.ChanResponse <- &gamemm.ActionResponse{
					Err: errors.New("Cant find player for this id")}
			} else {
				if action.ActionName == ACTION_CHOOSE_RULE {
					rule := gamemm.Rule{
						MoneyType:        currency.Money,
						RequirementMoney: 1000,
					}
					rm := utils.GetInt64AtPath(action.Data, "RequirementMoney")
					if rm == 1000 || rm == 2000 || rm == 5000 {
						// rm == 10000 || rm == 20000 || rm == 50000 {
						rule.RequirementMoney = rm
						game.mutex.Lock()
						game.mapPidToRule[action.PlayerId] = rule
						game.mutex.Unlock()
						action.ChanResponse <- &gamemm.ActionResponse{Err: nil}
					} else {
						action.ChanResponse <- &gamemm.ActionResponse{
							Err: errors.New("Bạn chọn sai mức tiền cược")}
					}
				} else if action.ActionName == ACTION_GET_USER_INFO {
					game.matchMaker.Mutex.Lock()
					isFindingMatch := game.matchMaker.MapPidToIsQueuing[action.PlayerId]
					var nPlayer int
					for _, v := range game.matchMaker.MapPidToIsQueuing {
						if v == true {
							nPlayer += 1
						}
					}
					game.matchMaker.Mutex.Unlock()
					game.mutex.Lock()
					RequirementMoney := game.mapPidToRule[action.PlayerId].RequirementMoney
					game.mutex.Unlock()

					top.GlobalMutex.Lock()
					trackerW := top.MapEvents[top.NORMAL_TRACK_OANTUTI_WINNING_STREAK]
					trackerL := top.MapEvents[top.NORMAL_TRACK_OANTUTI_LOSING_STREAK]
					top.GlobalMutex.Unlock()
					trackerW.Mutex.Lock()
					streak := trackerW.MapPlayerIdToCurrentValue[action.PlayerId]
					trackerW.Mutex.Unlock()
					trackerL.Mutex.Lock()
					losingStreak := trackerL.MapPlayerIdToCurrentValue[action.PlayerId]
					trackerL.Mutex.Unlock()

					action.ChanResponse <- &gamemm.ActionResponse{Err: nil}
					gamemm.ServerObj.SendRequest(
						"OantutiUserInfo",
						map[string]interface{}{
							"nPlayer":          100 + nPlayer,
							"isFindingMatch":   isFindingMatch,
							"streak":           streak,
							"losingStreak":     losingStreak,
							"RequirementMoney": RequirementMoney,
						},
						action.PlayerId)
				} else if action.ActionName == ACTION_GET_TOP {
					action.ChanResponse <- &gamemm.ActionResponse{Err: nil}
					gamemm.ServerObj.SendRequest(
						"OantutiTop",
						map[string]interface{}{
							"topStreak":       game.topStreak,
							"topLosingStreak": game.topLosingStreak,
						},
						action.PlayerId)
				} else { // các hành động trong trận đấu
					game.matchMaker.Mutex.Lock()
					match := game.matchMaker.MapPidToMatch[playerObj.Id()]
					game.matchMaker.Mutex.Unlock()
					if match == nil {
						action.ChanResponse <- &gamemm.ActionResponse{
							Err: errors.New("Bạn không ở trong trận đấu nào")}
					} else {
						match.ReceiveAction(action)
					}
				}
			}
		}()
	}
}

////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// send command to game.matchMaker.ChanAction,
// in case command is gameplay, forward to the game.ChanAction
func DoPlayerAction(game *OantutiGame, action *gamemm.Action) error {
	t1 := time.After(3 * time.Second)
	select {
	case game.matchMaker.ChanAction <- action:
		t2 := time.After(3 * time.Second)
		select {
		case res := <-action.ChanResponse:
			return res.Err
		case <-t2:
			return errors.New("err:receiving_time_out")
		}
	case <-t1:
		return errors.New("err:sending_time_out")
	}
}

// aaa,
func (game *OantutiGame) ChooseRule(player *player.Player, RequirementMoney int64) error {
	action := &gamemm.Action{
		ActionName: ACTION_CHOOSE_RULE,
		PlayerId:   player.Id(),
		Data: map[string]interface{}{
			"RequirementMoney": RequirementMoney,
		},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) FindMatch(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   gamemm.ACTION_FIND_MATCH,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) StopFindingMatch(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   gamemm.ACTION_STOP_FINDING_MATCH,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) ChooseHandPaper(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   ACTION_HAND_PAPER,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) ChooseHandRock(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   ACTION_HAND_ROCK,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) ChooseHandScissors(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   ACTION_HAND_SCISSORS,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) GetUserInfo(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   ACTION_GET_USER_INFO,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}

// aaa,
func (game *OantutiGame) GetTop(player *player.Player) error {
	action := &gamemm.Action{
		ActionName:   ACTION_GET_TOP,
		PlayerId:     player.Id(),
		Data:         map[string]interface{}{},
		ChanResponse: make(chan *gamemm.ActionResponse),
	}
	return DoPlayerAction(game, action)
}
