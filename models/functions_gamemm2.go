package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/models/gamemm2"
	"github.com/vic/vic_go/models/gamemm2/lieng"
	"github.com/vic/vic_go/models/gamemm2/poker"
	"github.com/vic/vic_go/models/gamemm2/tienlen3"
	//	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
	_ = time.Now()
}

func GetJoinedLobbies(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	result := map[string]interface{}{}
	for _, game := range models.gamesmm2 {
		gameCode := game.GameCode()
		lids := []int64{}
		matchMaker := game.GetMatchMaker()
		matchMaker.Mutex.RLock()
		if matchMaker.MapPidToLobbies[playerId] != nil {
			for lid, _ := range matchMaker.MapPidToLobbies[playerId] {
				lids = append(lids, lid)
			}
		}
		matchMaker.Mutex.RUnlock()
		result[gameCode] = lids
	}
	return result, nil
}

func GetLobbyStatus(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	gameCode := utils.GetStringAtPath(data, "GameCode")
	lobbyId := utils.GetInt64AtPath(data, "LobbyId")
	game := models.GetGameMM2(gameCode)
	if game == nil {
		return nil, errors.New("Invalid GameCode")
	}
	matchMaker := game.GetMatchMaker()
	matchMaker.Mutex.RLock()
	lobby := matchMaker.MapLidToLobby[lobbyId]
	matchMaker.Mutex.RUnlock()
	if lobby != nil {
		sTime := time.Now()
		_ = sTime
		//		fmt.Println("cp1", sTime)
		lobby.UpdateLobbyStatus()
		time.Sleep(200 * time.Millisecond)
		lobby.UpdateMatchStatus()
		return nil, nil
	} else {
		return nil, errors.New("lobby == nil")
	}

}

// poker
func PokerGeneralF(
	models *Models, data map[string]interface{}, playerId int64, actionName string) (
	map[string]interface{}, error) {
	game := models.GetGameMM2(poker.EXAMPLE_GAME_CODE)
	pokerG, isOk := game.(*poker.ExGame)
	if !isOk {
		return nil, errors.New("ERROR PokerGeneralF type assertion")
	}
	action := gamemm2.NewAction(actionName, playerId, data)
	err := poker.DoPlayerAction(pokerG, action)
	return nil, err
}

func PokerChooseRule(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return PokerGeneralF(models, data, playerId, poker.ACTION_G_CHOOSE_RULE)
}

func PokerBuyIn(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return PokerGeneralF(models, data, playerId, gamemm2.ACTION_MM_BUY_IN)
}

func PokerFindLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return PokerGeneralF(models, data, playerId, gamemm2.ACTION_MM_FIND_LOBBY)
}

func PokerLeaveLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return PokerGeneralF(models, data, playerId, gamemm2.ACTION_MM_LEAVE_LOBBY)
}

func PokerMakeMove(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return PokerGeneralF(models, data, playerId, poker.ACTION_M_MAKE_MOVE)
}

// tienlen3
func TL3GeneralF(
	models *Models, data map[string]interface{}, playerId int64, actionName string) (
	map[string]interface{}, error) {
	game := models.GetGameMM2(tienlen3.EXAMPLE_GAME_CODE)
	tienlen3G, isOk := game.(*tienlen3.ExGame)
	if !isOk {
		return nil, errors.New("ERROR TL3GeneralF type assertion")
	}
	action := gamemm2.NewAction(actionName, playerId, data)
	err := tienlen3.DoPlayerAction(tienlen3G, action)
	return nil, err
}

func TL3ChooseRule(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return TL3GeneralF(models, data, playerId, tienlen3.ACTION_G_CHOOSE_RULE)
}

func TL3FindLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return TL3GeneralF(models, data, playerId, gamemm2.ACTION_MM_FIND_LOBBY)
}

func TL3ChooseRuleAndFindLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	_, e := TL3GeneralF(models, data, playerId, tienlen3.ACTION_G_CHOOSE_RULE)
	if e != nil {
		return nil, e
	}
	return TL3GeneralF(models, data, playerId, gamemm2.ACTION_MM_FIND_LOBBY)
}

func TL3LeaveLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return TL3GeneralF(models, data, playerId, gamemm2.ACTION_MM_LEAVE_LOBBY)
}

func TL3MakeMove(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return TL3GeneralF(models, data, playerId, tienlen3.ACTION_M_MAKE_MOVE)
	return nil, nil
}

// poker
func LiengGeneralF(
	models *Models, data map[string]interface{}, playerId int64, actionName string) (
	map[string]interface{}, error) {
	game := models.GetGameMM2(lieng.EXAMPLE_GAME_CODE)
	liengG, isOk := game.(*lieng.ExGame)
	if !isOk {
		return nil, errors.New("ERROR liengGeneralF type assertion")
	}
	action := gamemm2.NewAction(actionName, playerId, data)
	err := lieng.DoPlayerAction(liengG, action)
	return nil, err
}

func LiengChooseRule(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return LiengGeneralF(models, data, playerId, lieng.ACTION_G_CHOOSE_RULE)
}

func LiengBuyIn(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return LiengGeneralF(models, data, playerId, gamemm2.ACTION_MM_BUY_IN)
}

func LiengFindLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return LiengGeneralF(models, data, playerId, gamemm2.ACTION_MM_FIND_LOBBY)
}

func LiengLeaveLobby(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return LiengGeneralF(models, data, playerId, gamemm2.ACTION_MM_LEAVE_LOBBY)
}

func LiengMakeMove(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	return LiengGeneralF(models, data, playerId, lieng.ACTION_M_MAKE_MOVE)
}
