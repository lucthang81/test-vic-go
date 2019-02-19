package tienlen3

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/vic/vic_go/models/gamemm2"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

const (
	EXAMPLE_GAME_CODE = "tienlen3"

	ACTION_G_CHOOSE_RULE = "ACTION_G_CHOOSE_RULE"
	ACTION_GAME0         = "ACTION_GAME0"
	ACTION_GAME1         = "ACTION_GAME1"
)

type ExLobbyData struct {
}

func (d *ExLobbyData) ToString() string {
	bs, err := json.Marshal(d)
	if err != nil {
		temp := map[string]error{"errorJson": err}
		bs2, _ := json.Marshal(temp)
		return string(bs2)
	} else {
		return string(bs)
	}
}

func NewExgame() *ExGame {
	gameObj := &ExGame{
		gameCode:              EXAMPLE_GAME_CODE,
		minNPlayers:           2,
		maxNPlayers:           4,
		maxNConcurrentLobbies: 1,
		isBuyInGame:           false,

		ChanAction: make(chan *gamemm2.Action),
	}
	gameObj.MatchMaker = gamemm2.NewMatchMaker(gameObj)
	go LoopReceiveActions(gameObj)
	return gameObj
}

type ExGame struct {
	gameCode              string
	currencyType          string
	minNPlayers           int
	maxNPlayers           int
	maxNConcurrentLobbies int
	isBuyInGame           bool

	MatchMaker *gamemm2.MatchMaker

	// receive action from MatchMaker.LoopReceiveActions
	ChanAction chan *gamemm2.Action
}

// will be called in MatchMaker.StartMatch,
// already embraced in MatchMaker.Mutex.Lock
func (game *ExGame) StartMatch(lobby *gamemm2.Lobby) gamemm2.MatchInterface {
	match := &ExMatch{
		Game:             game,
		Lobby:            lobby,
		StartedTime:      time.Now(),
		MatchId:          fmt.Sprintf("#%v", time.Now().UnixNano()),
		MapPidToPlayer:   make(map[int64]*player.Player),
		MapSeatToPlayer:  make(map[int]*player.Player),
		TurnStartingTime: time.Now(),
		ChanAction:       make(chan *gamemm2.Action),
		ChanMove:         make(chan Move),
		ChanMoveErr:      make(chan error),
	}
	for k, v := range lobby.MapPidToPlayer {
		match.MapPidToPlayer[k] = v
	}
	for k, v := range lobby.MapSeatToPlayer {
		match.MapSeatToPlayer[k] = v
	}
	go Start(match)
	go InMatchLoopReceiveActions(match)

	return match
}

// forward action to game.LoopReceiveActions
func (game *ExGame) ReceiveAction(action *gamemm2.Action) {
	t1 := time.After(3 * time.Second)
	select {
	case game.ChanAction <- action:
	case <-t1:
		fmt.Println("ERROR: poker game.ReceiveAction TimeOut")
		_ = errors.New("")
	}
}

// every actions will be pass to MatchMaker.ChanAction,
//    w8 responseError on action.ChanResponse,
// action can be forward to game.ChanAction,
//    and can be forward from game.ChanAction to match.ChanAction
func DoPlayerAction(game *ExGame, action *gamemm2.Action) error {
	t1 := time.After(3 * time.Second)
	select {
	case game.MatchMaker.ChanAction <- action:
		t2 := time.After(3 * time.Second)
		select {
		case res := <-action.ChanResponse:
			return res
		case <-t2:
			return errors.New("err:DoPlayerAction_Receiving_TimeOut")
		}
	case <-t1:
		return errors.New("err:DoPlayerAction_Sending_TimeOut")
	}
}

func (game *ExGame) GameCode() string {
	return game.gameCode
}
func (game *ExGame) CurrencyType() string {
	return game.currencyType
}

func (game *ExGame) MinNPlayers() int {
	return game.minNPlayers
}

func (game *ExGame) MaxNPlayers() int {
	return game.maxNPlayers
}

func (game *ExGame) MaxNConcurrentLobbies() int {
	return game.maxNConcurrentLobbies
}

func (game *ExGame) IsBuyInGame() bool {
	return game.isBuyInGame
}

func (game *ExGame) DefaultLobbyData() gamemm2.GameSpecificLobbyDataInterface {
	d := &ExLobbyData{}
	return d
}

func (game *ExGame) GetMatchMaker() *gamemm2.MatchMaker {
	return game.MatchMaker
}

//
func LoopReceiveActions(game *ExGame) {
	for {
		action := <-game.ChanAction
		go func() {
			defer func() {
				if r := recover(); r != nil {
					bytes := debug.Stack()
					fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
				}
			}()

			playerObj, _ := player.GetPlayer(action.PlayerId)
			_ = playerObj

			if action.ActionName == ACTION_G_CHOOSE_RULE {
				BaseMoney := utils.GetInt64AtPath(action.Data, "BaseMoney")
				MoneyType := utils.GetStringAtPath(action.Data, "MoneyType")
				rule := gamemm2.Rule{
					BaseMoney:    BaseMoney,
					MoneyType:    MoneyType,
					MinimumMoney: 40 * BaseMoney,
					MaximumBuyIn: BaseMoney * 1000,
				}
				game.MatchMaker.Mutex.Lock()
				game.MatchMaker.MapPidToRule[action.PlayerId] = rule
				game.MatchMaker.Mutex.Unlock()
				action.ChanResponse <- nil
			} else {
				lobbyId := utils.GetInt64AtPath(action.Data, "LobbyId")
				game.MatchMaker.Mutex.Lock()
				lobby := game.MatchMaker.MapLidToLobby[lobbyId]
				game.MatchMaker.Mutex.Unlock()
				if lobby != nil && lobby.Match != nil {
					lobby.Match.ReceiveAction(action)
				} else {
					action.ChanResponse <- errors.New("lobby == nil || lobby.Match == nil")
				}
			}
		}()
	}
}
