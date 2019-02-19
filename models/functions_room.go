package models

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/language"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

func getCurrentRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	if player.Room() == nil {
		return map[string]interface{}{}, nil
	}
	responseData = player.Room().SerializedDataFull(player)
	return responseData, nil
}

func getRoomList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	orderBy := utils.GetStringAtPath(data, "order_by")
	limit := utils.GetIntAtPath(data, "limit")
	offset := utils.GetIntAtPath(data, "offset")

	orderBy = "requirement"

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}
	roomData, err := game.GetRoomList(gameInstance, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["game_code"] = gameCode
	responseData["currency_type"] = currencyType
	responseData["rooms"] = roomData
	return responseData, nil
}

func GetRequirementList(models *Models, data map[string]interface{}, playerId int64) (map[string]interface{}, error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")

	gameI := models.GetGame(gameCode, currencyType)
	if gameI == nil {
		return nil, errors.New("err:this_game_does_not_exist")
	}
	requirementsMap := make(map[int64]bool)
	for roomId, room := range gameI.GameData().Rooms().CoreMap() {
		_ = roomId
		requirementsMap[room.Requirement()] = true
	}
	keys := make([]int64, 0)
	for r, _ := range requirementsMap {
		keys = append(keys, r)
	}
	requirements := utils.SortedInt64(keys)
	responseData := make(map[string]interface{})
	responseData["requirements"] = requirements
	responseData["currencyType"] = currencyType
	return responseData, nil
}

// vào phòng đầu tiên sort by (number of real players, number of bots)
// nếu phòng này không có người chơi thật:
//     50% vào phòng này,
//     50% vào phòng trống,
func JoinRequirement(models *Models, data map[string]interface{}, playerId int64) (map[string]interface{}, error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	requirement := utils.GetInt64AtPath(data, "requirement")

	gameI := models.GetGame(gameCode, currencyType)
	if gameI == nil {
		return nil, errors.New("err:this_game_does_not_exist")
	}
	player, _ := models.GetPlayer(playerId)
	if player == nil {
		return nil, errors.New("err:player_does_not_exist")
	}

	rooms := make([]*game.Room, 0)
	for roomId, room := range gameI.GameData().Rooms().Copy() {
		_ = roomId
		if room.Requirement() == requirement {
			rooms = append(rooms, room)
		}
	}
	if len(rooms) == 0 {
		return nil, errors.New("err:have_0_rooms_for_this_requirement")
	}
	game.SortRoomsByNOPlayer(rooms)

	var chosenRoom *game.Room
	//	if rooms[0].GetNumberOfHumans() == 0 {
	//		if rand.Intn(2) < 1 {
	//			chosenRoom = rooms[0]
	//		} else {
	//			chosenRoom = rooms[len(rooms)-1]
	//		}
	//	} else {
	//		chosenRoom = rooms[0]
	//	}
	if len(rooms) < 4 {
		chosenRoom = rooms[0]
	} else {
		if rand.Intn(100) < 40 {
			chosenRoom = rooms[0]
		} else {
			chosenRoom = rooms[1+rand.Intn(3)]
		}
	}

	chosenRoom1, err := game.JoinRoomWithBuyInById(gameI, player, chosenRoom.Id(), 0, "")
	if err != nil {
		return nil, err
	}
	res := chosenRoom1.SerializedDataFull(player)
	return res, nil
}

func quickJoinRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameCode := utils.GetStringAtPath(data, "game_code")

	gameI := models.GetGame(gameCode, currencyType)
	if gameI == nil {
		return nil, errors.New("err:this_game_does_not_exist")
	}
	player, _ := models.GetPlayer(playerId)
	if player == nil {
		return nil, errors.New("err:player_does_not_exist")
	}

	requirement := int64(100)
	requirements := []int64{
		100, 200, 500,
		1000, 2000, 5000,
		10000, 20000, 50000,
		100000, 200000, 500000,
	}
	var goodRatio int64
	if gameCode == "tienlen" {
		goodRatio = 60
	} else if gameCode == "phom" {
		goodRatio = 80
	} else if gameCode == "bacay2" {
		goodRatio = 56
	} else if gameCode == "maubinh" {
		goodRatio = 100
	} else {
		goodRatio = 40
	}
	for i := len(requirements) - 1; i >= 0; i-- {
		r := requirements[i]
		if r*goodRatio <= player.GetMoney(currencyType) {
			requirement = r
			break
		}
	}
	data = map[string]interface{}{
		"currency_type": currencyType,
		"game_code":     gameCode,
		"requirement":   requirement,
	}
	return JoinRequirement(models, data, playerId)
}

// custom room, tax base on play duration, currency.CustomMoney
func createRoom(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	if models.ShouldStopActionsToWaitForMaintenance() {
		err = details_error.NewError(l.Get(l.M0094), map[string]interface{}{
			"duration": models.DurationUntilMaintenanceString(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		return nil, err
	}
	//
	gameCode := utils.GetStringAtPath(data, "game_code")
	requirement := utils.GetInt64AtPath(data, "requirement")
	//startingMoney := utils.GetInt64AtPath(data, "starting_money")
	startingMoney := 100 * requirement
	//password := utils.GetStringAtPath(data, "password")
	password := ""
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if !player.CheckCanCreateRoom() {
		return nil, errors.New("Bạn không có quyền tạo phòng.")
	}
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return nil, err
	}
	if player.GetMoney(currency.Money) < 8000 {
		return nil, errors.New("Bạn không đủ tiền để tạo phòng.")
	}
	gameInstance := models.GetGame(gameCode, currency.CustomMoney)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}
	//
	player.SetMoney(startingMoney, currency.CustomMoney, true)
	room, err := game.CreateRoomWithBuyIn(
		gameInstance, player, requirement, 0, 0, password)
	room.Mutex.Lock()
	room.SharedData["createrPid"] = player.Id()
	room.SharedData["startingMoney"] = startingMoney
	room.SharedData["isStopped"] = false
	room.SharedData["timeCounter"] = 0
	room.SharedData["createdTime"] = time.Now()
	// for room_join.go JoinRoomWithBuyInById:
	room.SharedData["joinedPids"] = map[int64]bool{player.Id(): true}
	// for room_join.go handleRegisterLeaveRoom:
	room.SharedData["bannedPids"] = map[int64]bool{}
	room.Mutex.Unlock()
	// periodically decrease room creater money
	// stop room when creater money = 0 or creater leave in room.handleDidEndGame
	go func() {
		timeUnit := 60 * time.Second
		nTimeUnitNotifyOutOfMoney := []int64{15, 10, 5}
		for {
			room.Mutex.Lock()
			isStopped, _ := room.SharedData["isStopped"].(bool)
			room.Mutex.Unlock()
			if isStopped == true {
				break
			}
			room.Mutex.Lock()
			timeCounter, _ := room.SharedData["timeCounter"].(int64)
			room.SharedData["timeCounter"] = timeCounter + 1
			room.Mutex.Unlock()
			var tax int64
			if timeCounter == 0 {
				tax = 134 // 8040 Kim/h
			} else if 1 <= timeCounter && timeCounter <= 10 {
				tax = 84 // 5040 Kim/h
			} else {
				tax = 17 // 1020 Kim/h
			}
			player.ChangeMoneyAndLog(
				tax, currency.Money, false, "",
				record.ACTION_CUSTOM_ROOM_PERIODIC_TAX, room.GameCode(), "")
			nTimeUnitRemaining := player.GetAvailableMoney(currency.Money) / tax
			if z.FindInt64InSlice(nTimeUnitRemaining, nTimeUnitNotifyOutOfMoney) != -1 {
				player.CreatePopUp("Bạn cần nạp thêm tiền để duy trì phòng chơi.")
			}
			//
			time.Sleep(timeUnit)
		}
	}()
	//
	player.CreateRawMessage("Tạo phòng chơi", fmt.Sprintf("Bạn đã tạo phòng chơi id %v", room.Id()))
	player.CreatePopUp(fmt.Sprintf("Bạn đã tạo phòng chơi id %v", room.Id()))
	if err != nil {
		return nil, err
	}
	responseData = room.SerializedDataFull(player)
	return responseData, nil
}

func SetMoneyCustomRoom(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := currency.CustomMoney
	roomId := utils.GetInt64AtPath(data, "room_id")
	targetPlayerId := utils.GetInt64AtPath(data, "targetPlayerId")
	moneyValue := utils.GetInt64AtPath(data, "moneyValue")

	gameObj := models.GetGame(gameCode, currencyType)
	if gameObj == nil {
		return nil, errors.New("gameObj == nil")
	}
	targetPlayer, err := models.GetPlayer(targetPlayerId)
	if err != nil {
		return nil, err
	}
	res, err := game.SetMoneyCustomRoom(
		gameObj, roomId, playerId, targetPlayer, moneyValue)
	return res, err
}

// playerId is createrId,
// targetPlayerId will be kick.
func KickPlayerCustomRoom(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := currency.CustomMoney
	roomId := utils.GetInt64AtPath(data, "room_id")
	targetPlayerId := utils.GetInt64AtPath(data, "targetPlayerId")

	gameObj := models.GetGame(gameCode, currencyType)
	if gameObj == nil {
		return nil, errors.New("gameObj == nil")
	}
	targetPlayer, err := models.GetPlayer(targetPlayerId)
	if err != nil {
		return nil, err
	}
	res, err := game.KickPlayerCustomRoom(
		gameObj, roomId, playerId, targetPlayer)
	return res, err
}

func GetMoneyHistoryInRoom(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := currency.CustomMoney
	roomId := utils.GetInt64AtPath(data, "room_id")

	gameObj := models.GetGame(gameCode, currencyType)
	if gameObj == nil {
		return nil, errors.New("gameObj == nil")
	}
	res, err := game.GetMoneyHistoryInRoom(currencyType, gameObj, roomId)
	return res, err
}

// for join custom room
func joinRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	if models.ShouldStopActionsToWaitForMaintenance() {
		err = details_error.NewError(l.Get(l.M0094), map[string]interface{}{
			"duration": models.DurationUntilMaintenanceString(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		return nil, err
	}

	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type") // usually currency.CustomMoney
	roomId := utils.GetInt64AtPath(data, "room_id")
	password := utils.GetStringAtPath(data, "password")
	buyIn := utils.GetInt64AtPath(data, "buy_in")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New("gameInstance == nil")
	}
	room, err := game.JoinRoomWithBuyInById(gameInstance, player, roomId, buyIn, password)
	if err != nil {
		return nil, err
	}
	responseData = room.SerializedDataFull(player)
	return responseData, nil
}

/*
room type quick
*/

func joinRoomByRequirement(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {

	if models.ShouldStopActionsToWaitForMaintenance() {
		err = details_error.NewError(l.Get(l.M0094), map[string]interface{}{
			"duration": models.DurationUntilMaintenanceString(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		return nil, err
	}

	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	requirement := utils.GetInt64AtPath(data, "requirement")
	buyIn := utils.GetInt64AtPath(data, "buy_in")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room, err := game.JoinRoomWithBuyInByRequirement(gameInstance, player, requirement, buyIn)
	if err != nil {
		return nil, err
	}
	responseData = room.SerializedDataFull(player)
	return responseData, nil
}

/*
common
*/

func leaveRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	/*
		remove leave room, will always register/unregister leave, and leave at start of new game
	*/

	// gameCode := utils.GetStringAtPath(data, "game_code")
	// roomId := utils.GetInt64AtPath(data, "room_id")
	// player, err := models.GetPlayer(playerId)
	// if err != nil {
	// 	return nil, err
	// }

	// gameInstance := models.GetGame(gameCode,currencyType)
	// if gameInstance == nil {
	// 	return nil, errors.New(l.Get(l.M0093))
	// }

	// err = game.LeaveRoom(gameInstance, player, roomId)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

func buyInInRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := player.Room()
	if room == nil {
		return nil, errors.New("err:player_not_in_room")
	}

	buyIn := utils.GetInt64AtPath(data, "buy_in")

	err = room.BuyIn(player, buyIn)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func registerLeaveRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	err = game.RegisterLeaveRoom(gameInstance, player)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func unregisterLeaveRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	if models.ShouldStopActionsToWaitForMaintenance() {
		err = details_error.NewError(l.Get(l.M0094), map[string]interface{}{
			"duration": models.DurationUntilMaintenanceString(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		return nil, err
	}

	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	err = game.UnregisterLeaveRoom(gameInstance, player)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func registerToBeOwner(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := player.Room()
	if room == nil {
		return nil, errors.New("err:player_not_in_room")
	}

	err = room.RegisterToBeOwner(player)
	if err != nil {
		return nil, err
	}
	responseData, err = room.GetOwnerList()
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func unregisterToBeOwner(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := player.Room()
	if room == nil {
		return nil, errors.New("err:player_not_in_room")
	}

	err = room.UnregisterToBeOwner(player)
	if err != nil {
		return nil, err
	}
	responseData, err = room.GetOwnerList()
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func getOwnerList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	currencyType := utils.GetStringAtPath(data, "currency_type")
	gameCode := utils.GetStringAtPath(data, "game_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := player.Room()
	if room == nil {
		return nil, errors.New("err:player_not_in_room")
	}

	responseData, err = room.GetOwnerList()
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func startGame(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	if models.ShouldStopActionsToWaitForMaintenance() {
		err = details_error.NewError(l.Get(l.M0094), map[string]interface{}{
			"duration": models.DurationUntilMaintenanceString(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		return nil, err
	}

	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	if !room.ContainsPlayer(player) {
		return nil, errors.New("err:player_not_in_room")
	}

	return room.StartGame(player)
}

func pingGame(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	if !room.ContainsPlayer(player) {
		return nil, errors.New("err:player_not_in_room")
	}

	return room.SerializedDataFull(player), nil
}

func kickPlayer(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	willBeKickedPlayerId := utils.GetInt64AtPath(data, "player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	willBeKickedPlayer, err := models.GetPlayer(willBeKickedPlayerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	err = room.KickPlayer(player, willBeKickedPlayer)
	return nil, err
}

func invitePlayerToRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	invitePlayerId := utils.GetInt64AtPath(data, "player_id")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	invitePlayer, err := models.GetPlayer(invitePlayerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	err = room.InvitePlayerToRoom(player, invitePlayer)
	return nil, err
}

func chatInRoom(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	roomId := utils.GetInt64AtPath(data, "room_id")
	message := utils.GetStringAtPath(data, "message")

	if len(message) == 0 {
		return nil, errors.New("err:no_message")
	}

	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		return nil, errors.New(l.Get(l.M0093))
	}

	room := gameInstance.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}

	err = room.Chat(player, message)
	return nil, err
}
