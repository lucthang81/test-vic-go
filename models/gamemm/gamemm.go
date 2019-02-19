package gamemm

import (
	"fmt"
)

type ServerInterface interface {
	// already run in a goroutine
	SendRequest(requestType string, data map[string]interface{}, toPlayerId int64)
	SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64)
	SendRequestsToAll(requestType string, data map[string]interface{})
}

var ServerObj ServerInterface

func init() {
	fmt.Print("")
}

func RegisterServer(registeredServer ServerInterface) {
	ServerObj = registeredServer
}

type GameInferface interface {
	GetGameCode() string
	SerializeData() map[string]interface{}

	StartMatch(lobby *Lobby) MatchInterface
	GetMatchMaker() *MatchMaker
	ReceiveAction(action *Action) *ActionResponse
	GetRuleForPlayer(pid int64) Rule
}

type MatchInterface interface {
	GetLobby() *Lobby
	ReceiveAction(action *Action) *ActionResponse
}
