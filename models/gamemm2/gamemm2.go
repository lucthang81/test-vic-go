// gamemm2 provides match making system
// which uses for replacing the old game.room
package gamemm2

import (
	"fmt"
)

type ServerInterface interface {
	// already run in a goroutine
	SendRequest(method string, data map[string]interface{}, toPlayerId int64)
}

var ServerObj ServerInterface

func init() {
	fmt.Print("")
}

func RegisterServer(registeredServer ServerInterface) {
	ServerObj = registeredServer
}

type GameInferface interface {
	GameCode() string
	// mimnimun number of players to start match
	MinNPlayers() int
	// maximum number of players in a lobby
	MaxNPlayers() int
	// maximum number of concurrent lobbies 1 user can join
	MaxNConcurrentLobbies() int
	// plz read Lobby.IsBuyInLobby. ex: poker IsBuyInGame == true
	IsBuyInGame() bool
	// ex: &PokerLobbyData{DealerSeat: 0}
	DefaultLobbyData() GameSpecificLobbyDataInterface

	// dont call matchMaker.Mutex in this func,
	// this lock already called in MatchMaker.StartMatch
	StartMatch(lobby *Lobby) MatchInterface
	// every action will first be send to MatchMaker.ChanAction,
	//  then if matchMaker dont handle the action:
	//  this func forward the action to game.ChanAction
	ReceiveAction(action *Action)
	//
	GetMatchMaker() *MatchMaker
}

type MatchInterface interface {
	// check whether or not player is in the match
	CheckContainingPlayer(pid int64) bool
	// included match's lock
	ToMap() map[string]interface{}
	// shared data and data which only specific player can read
	ToMapForPlayer(pid int64) map[string]interface{}
	// lobby which the match was run from
	GetLobby() *Lobby
	// every action will be send to MatchMaker.ChanAction,
	//  then can be forward to game.ChanAction,
	//  then can be forward to match.ChanAction by this func
	ReceiveAction(action *Action)
}

type GameSpecificLobbyDataInterface interface {
	// json represent
	ToString() string
}
