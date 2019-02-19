package game

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
)

func CreateRoomWithBuyIn(
	gameInstance GameInterface, player GamePlayer, requirement int64,
	buyIn int64, maxNumberOfPlayers int, password string) (
	room *Room, err error) {
	//
	// dont care below block, minah dont use GameProperties
	currencyType := gameInstance.CurrencyType()
	if gameInstance.RoomType() == RoomTypeQuick {
		return nil, errors.New("err:not_implemented")
	}
	if GameHasProperties(gameInstance, []string{GamePropertyNoCreateRoom}) {
		return nil, errors.New("err:not_implemented")
	}
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return nil, err
	}
	if !gameInstance.IsRoomRequirementValid(requirement) {
		// check in list predefineBaseMoneys
		return nil, errors.New("err:invalid_requirement")
	}
	if GameHasProperty(gameInstance, GamePropertyAlwaysHasOwner) {
		err = gameInstance.IsPlayerMoneyValidToBecomeOwner(player.GetAvailableMoney(currencyType), requirement, maxNumberOfPlayers, 1)
		if err != nil {
			return nil, err
		}
	} else {
		err = gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), requirement)
		if err != nil {
			return nil, err
		}
	}
	var validMaxNumberOfPlayers int
	if maxNumberOfPlayers == 0 {
		validMaxNumberOfPlayers = gameInstance.DefaultNumberOfPlayers()
	} else {
		if gameInstance.IsRoomMaxPlayersValid(maxNumberOfPlayers, requirement) {
			validMaxNumberOfPlayers = maxNumberOfPlayers
		} else {
			return nil, errors.New("err:invalid_max_number_of_players_in_room")
		}
	}
	//
	//
	room, err = NewRoom(
		gameInstance, player, validMaxNumberOfPlayers, requirement,
		buyIn, currencyType, password)
	if err != nil {
		return nil, err
	}
	player.SetRoom(room)
	gameInstance.GameData().rooms.Set(room.Id(), room)
	gameInstance.HandleRoomCreated(room)
	return room, nil
}

func CreateRoom(gameInstance GameInterface, player GamePlayer, requirement int64, maxNumberOfPlayers int, password string) (room *Room, err error) {
	return CreateRoomWithBuyIn(gameInstance, player, requirement, 0, maxNumberOfPlayers, password)
}

func CreateSystemRoom(gameInstance GameInterface, requirement int64, maxNumberOfPlayers int, password string) (room *Room, err error) {
	currencyType := gameInstance.CurrencyType()
	if !gameInstance.IsRoomRequirementValid(requirement) {
		return nil, errors.New("err:invalid_requirement")
	}

	var validMaxNumberOfPlayers int
	if maxNumberOfPlayers == 0 {
		validMaxNumberOfPlayers = gameInstance.DefaultNumberOfPlayers()
	} else {
		if gameInstance.IsRoomMaxPlayersValid(maxNumberOfPlayers, requirement) {
			validMaxNumberOfPlayers = maxNumberOfPlayers
		} else {
			return nil, errors.New("err:invalid_max_number_of_players_in_room")
		}
	}

	room, err = NewSystemRoom(gameInstance, nil, validMaxNumberOfPlayers, requirement, 0, currencyType, password)
	if err != nil {
		return nil, err
	}
	gameInstance.GameData().rooms.Set(room.Id(), room)
	gameInstance.HandleRoomCreated(room)
	return room, nil
}

func GetRoomList(gameInstance GameInterface, orderBy string, limit int, offset int) (roomData []map[string]interface{}, err error) {
	currencyType := gameInstance.CurrencyType()
	if gameInstance.RoomType() == RoomTypeQuick {
		return nil, errors.New("err:not_implemented")
	}

	roomData = make([]map[string]interface{}, 0, limit)
	if gameInstance.GameData().rooms.Len() > 0 {
		var count int64
		sortedRooms := make([]*Room, 0, gameInstance.GameData().rooms.Len())
		for _, room := range gameInstance.GameData().rooms.Copy() {
			if room.currencyType == currencyType {
				sortedRooms = append(sortedRooms, room)
			}
		}
		if orderBy != "" {
			if orderBy == "owner" {
				sort.Sort(ByOwnerName(sortedRooms))
			} else if orderBy == "numPlayers" {
				sort.Sort(ByNumPlayers(sortedRooms))
			} else if orderBy == "requirement" {
				sort.Sort(ByRequirement(sortedRooms))
			} else if orderBy == "-owner" {
				sort.Sort(sort.Reverse(ByOwnerName(sortedRooms)))
			} else if orderBy == "-numPlayers" {
				sort.Sort(sort.Reverse(ByNumPlayers(sortedRooms)))
			} else if orderBy == "-requirement" {
				sort.Sort(sort.Reverse(ByRequirement(sortedRooms)))
			}
		}
		for _, room := range sortedRooms {
			if room.IsRoomOnline() {
				if count >= int64(offset) {
					roomData = append(roomData, room.SerializedData())
					if len(roomData) >= limit {
						return roomData, nil
					}
				}
				count++
			}
		}
	}
	return roomData, nil
}

func JoinRoomById(gameInstance GameInterface, player GamePlayer, roomId int64, password string) (room *Room, err error) {
	return JoinRoomWithBuyInById(gameInstance, player, roomId, 0, password)
}

// main join func
// call func: handleJoinRoom
func JoinRoomWithBuyInById(
	gameInstance GameInterface, player GamePlayer,
	roomId int64, buyIn int64, password string) (
	room *Room, err error) {
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return nil, err
	}
	room = gameInstance.GameData().rooms.Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	if room.currencyType == currency.CustomMoney {
		room.Mutex.Lock()
		startingMoney, isOk1 := room.SharedData["startingMoney"].(int64)
		joinedPids, isOk := room.SharedData["joinedPids"].(map[int64]bool)
		bannedPids, isOk2 := room.SharedData["bannedPids"].(map[int64]bool)
		room.Mutex.Unlock()
		if !isOk || !isOk1 || !isOk2 {
			return nil, errors.New("createrPid joinedPids bannedPids wrong type assertion")
		}
		room.Mutex.Lock()
		isBanned := bannedPids[player.Id()]
		room.Mutex.Unlock()
		if isBanned {
			return nil, errors.New("Bạn đã bị kick, không thể vào lại phòng")
		}
		room.Mutex.Lock()
		// cộng tiền tối thiểu cho người vào bàn lần đầu
		if joinedPids[player.Id()] == false {
			joinedPids[player.Id()] = true
			player.SetMoney(startingMoney, currency.CustomMoney, true)
		}
		room.Mutex.Unlock()
	}
	return joinRoom(room, player, buyIn, password)
}

// money history for all player (even if players left)
func GetMoneyHistoryInRoom(currencyType string, gameObj GameInterface, roomId int64) (
	map[string]interface{}, error) {
	// response data to client
	hihi := map[string]interface{}{}
	room := gameObj.GameData().rooms.Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	room.Mutex.Lock()
	roomCreatedTime, isOk := room.SharedData["createdTime"].(time.Time)
	joinedPids, isOk1 := room.SharedData["joinedPids"].(map[int64]bool)
	room.Mutex.Unlock()
	if !isOk || !isOk1 {
		return nil, errors.New("createdTime joinedPids wrong type assertion")
	}
	roomCreatedTime = roomCreatedTime.UTC()
	ids := room.PlayersId()
	room.Mutex.Lock()
	for id, _ := range joinedPids {
		ids = append(ids, id)
	}
	room.Mutex.Unlock()
	for _, pid := range ids {
		moneyRows := make([]map[string]interface{}, 0)
		queryString := "SELECT id, player_id, action, game_code, change, currency_type, " +
			"value_before, value_after, additional_data, created_at " +
			"FROM currency_record " +
			"WHERE player_id = $1 AND currency_type = $2 AND created_at >= $3 ORDER BY created_at DESC"
		rows, err := dataCenter.Db().Query(queryString, pid, currencyType, roomCreatedTime)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id, player_id, change, value_before, value_after sql.NullInt64
			var action, game_code, currency_type, additional_data string
			var created_at time.Time
			err = rows.Scan(
				&id, &player_id, &action, &game_code, &change, &currency_type,
				&value_before, &value_after, &additional_data, &created_at,
			)
			if err != nil {
				rows.Close()
				return nil, err
			}
			var tempMap map[string]interface{}
			err := json.Unmarshal([]byte(additional_data), &tempMap)
			if err != nil {
				tempMap = map[string]interface{}{}
			}
			r1 := map[string]interface{}{
				"id":              id.Int64,
				"player_id":       player_id.Int64,
				"change":          change.Int64,
				"value_before":    value_before.Int64,
				"value_after":     value_after.Int64,
				"action":          action,
				"game_code":       game_code,
				"currency_type":   currency_type,
				"match_record_id": tempMap["match_record_id"],
				"created_at":      toHihiFormat(created_at.Local()),
			}
			moneyRows = append(moneyRows, r1)
		}
		rows.Close()
		hihi[fmt.Sprintf("%v", pid)] = moneyRows
	}
	return hihi, nil
}

func SetMoneyCustomRoom(
	gameObj GameInterface, roomId int64, commandSenderId int64,
	targetPlayer GamePlayer, moneyValue int64) (
	map[string]interface{}, error) {
	room := gameObj.GameData().rooms.Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	room.Mutex.Lock()
	createrPid, isOk := room.SharedData["createrPid"].(int64)
	joinedPids, isOk1 := room.SharedData["joinedPids"].(map[int64]bool)
	room.Mutex.Unlock()
	if !isOk || !isOk1 {
		return nil, errors.New("createrPid joinedPids wrong type assertion")
	}
	if commandSenderId != createrPid {
		return nil, errors.New("Chỉ có người tạo bàn được cài đặt tiền")
	}
	room.Mutex.Lock()
	isTargetInRoom := false
	for id, _ := range joinedPids {
		if targetPlayer.Id() == id {
			isTargetInRoom = true
		}
	}
	room.Mutex.Unlock()
	if !isTargetInRoom {
		return nil, errors.New("Không được cài đặt tiền cho người ngoài phòng")
	}
	targetPlayer.SetMoney(moneyValue, currency.CustomMoney, true)
	return nil, nil
}

func KickPlayerCustomRoom(
	gameObj GameInterface, roomId int64, commandSenderId int64,
	targetPlayer GamePlayer) (
	map[string]interface{}, error) {
	room := gameObj.GameData().rooms.Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	room.Mutex.Lock()
	createrPid, isOk := room.SharedData["createrPid"].(int64)
	bannedPids, isOk2 := room.SharedData["bannedPids"].(map[int64]bool)
	room.Mutex.Unlock()
	if !isOk || !isOk2 {
		return nil, errors.New("createrPid bannedPids wrong type assertion")
	}
	if commandSenderId != createrPid {
		return nil, errors.New("Chỉ có người tạo bàn được quyền đuổi")
	}
	room.Mutex.Lock()
	bannedPids[targetPlayer.Id()] = true
	room.Mutex.Unlock()
	RegisterLeaveRoom(gameObj, targetPlayer)
	return nil, nil
}

// 16:58:01 28-09-2017
func toHihiFormat(t time.Time) string {
	return fmt.Sprintf(
		"%02d:%02d:%02d    %02d-%02d-%04d",
		t.Hour(), t.Minute(), t.Second(),
		t.Day(), t.Month(), t.Year(),
	)
}

/*
room type quick
*/
func JoinRoomByRequirement(gameInstance GameInterface, player GamePlayer, requirement int64) (room *Room, err error) {
	return JoinRoomWithBuyInByRequirement(gameInstance, player, requirement, 0)
}

// bot gọi hàm này để vào room,
// 50% bot sẽ vào phòng bất kì,
// 50% bot sẽ chỉ vào phòng có 1 và chỉ 1 người chơi thật,
func JoinRoomWithBuyInByRequirement(gameInstance GameInterface, player GamePlayer, requirement int64, buyIn int64) (room *Room, err error) {
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return nil, err
	}
	currencyType := gameInstance.CurrencyType()

	var validEntry BetEntryInterface
	for _, entry := range gameInstance.BetData().Entries() {
		if entry.Min() == requirement {
			validEntry = entry
			break
		}
	}

	if validEntry == nil {
		return nil, errors.New("err:wrong room entry.Min()")
	}

	err = gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), validEntry.Min())
	if err != nil {
		return nil, err
	}

	// dead code, dont have a fucking game has GamePropertyBuyIn
	//	if GameHasProperty(gameInstance, GamePropertyBuyIn) {
	//		var isOk bool
	//		if buyIn >= validEntry.Min() &&
	//			buyIn <= validEntry.Max() &&
	//			buyIn%validEntry.Step() == 0 &&
	//			player.GetMoney(currencyType) >= buyIn {
	//			isOk = true
	//		}
	//
	//		if !isOk {
	//			return nil, errors.New("Buy In không hợp lệ")
	//		}
	//	}

	var validRoom *Room
	if player.PlayerType() == "bot" {
		if !validEntry.EnableBot() {
			return nil, errors.New(l.Get(l.M0039))
			// should be errors.New("err:Bot cant play this requirement")
		}

		possibleRooms := make([]*Room, 0)
		rooms := make([]*Room, 0)
		for roomId, room := range gameInstance.GameData().Rooms().Copy() {
			_ = roomId
			if room.Requirement() == requirement {
				rooms = append(rooms, room)
			}
		}
		SortRoomsByNOPlayer(rooms)

		for _, room := range rooms {
			//			if room.Requirement() == validEntry.Min() &&
			//				len(room.Players().coreMap) < room.MaxNumberOfPlayers() &&
			//				room.IsRoomOnline() && // always pass
			//				!room.isDelayingForNewGame { // always pass
			if true {
				numBot := 0
				numPlayer := 0
				for _, playerInRoom := range room.Players().copy() {
					if playerInRoom.PlayerType() == "bot" {
						numBot++
					} else {
						numPlayer++
					}
				}

				nBotsBound := 0   // số bot tối đa có thể trong 1 phòng
				nHumansBound := 0 // khi đã có số này người chơi thật thì không cần bot nữa
				if gameInstance.GameCode() == "tienlen" ||
					gameInstance.GameCode() == "maubinh" {
					nBotsBound = 3
					nHumansBound = 2
				} else if gameInstance.GameCode() == "bacay2" {
					nBotsBound = 7
					nHumansBound = 4
				} else if gameInstance.GameCode() == "phom" {
					nBotsBound = 3
					nHumansBound = 2
				} else if gameInstance.GameCode() == "xocdia2" {
					nBotsBound = 4
					nHumansBound = 4
				} else if gameInstance.GameCode() == "phomSolo" {
					nBotsBound = 1
					nHumansBound = 2
				} else if gameInstance.GameCode() == "tienlenSolo" {
					nBotsBound = 1
					nHumansBound = 2
				}
				if player.Id()%3 > 0 { // bot vào lung tung{
					if (numBot < nBotsBound) &&
						(numPlayer < nHumansBound) &&
						(numBot+numPlayer < room.MaxNumberOfPlayers()) {
						possibleRooms = append(possibleRooms, room)
					}
				} else { // bot chỉ vào phòng có đúng 1 người chơi
					if (numBot == 0) && (numPlayer == 1) {
						if rand.Intn(3) < 1 {
							possibleRooms = append(possibleRooms, room)
						}
					}
				}
				if len(possibleRooms) >= 4 {
					break
				}
			}
		}

		if len(possibleRooms) > 0 {
			validRoom = possibleRooms[rand.Intn(len(possibleRooms))]
		}

		if validRoom == nil {
			return nil, errors.New("err:dont have validRoom")
		} else {
			room, err = joinRoom(validRoom, player, buyIn, "")
			if err != nil {
				return nil, err
			} else {
				return room, err
			}
		}
	} else {
		return nil, errors.New("err:human_cant_call_this_func")
	}

}

/*
common
*/

func QuickJoinRoom(gameInstance GameInterface, player GamePlayer) (rom *Room, err error) {
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return nil, err
	}

	if gameInstance.RoomType() == RoomTypeList {
		return quickJoinRoomTypeList(gameInstance, player)
	} else {
		return quickJoinRoomTypeQuick(gameInstance, player)
	}
}

func RegisterLeaveRoom(gameInstance GameInterface, player GamePlayer) (err error) {
	room := player.Room()
	if room == nil {
		return errors.New("err:player_not_in_room")
	}

	action := NewRoomActionContext()
	action.actionType = "RegisterLeaveRoom"
	action.room = room
	action.player = player
	response := room.sendAction(action)
	return response.err
}

func UnregisterLeaveRoom(gameInstance GameInterface, player GamePlayer) (err error) {
	room := player.Room()
	if room == nil {
		return errors.New("err:player_not_in_room")
	}

	action := NewRoomActionContext()
	action.actionType = "UnregisterLeaveRoom"
	action.room = room
	action.player = player
	response := room.sendAction(action)
	return response.err
}

/*
detail
*/

func quickJoinRoomTypeList(gameInstance GameInterface, player GamePlayer) (room *Room, err error) {
	if GameHasProperty(gameInstance, GamePropertyBuyIn) {
		return nil, errors.New("err:not_implemented")
	}

	currencyType := gameInstance.CurrencyType()
	err = gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), gameInstance.BetData().Entries()[0].Min())
	if err != nil {
		return nil, err
	}
	// we will limit to 50 rooms randomly
	possibleRooms := make([]*Room, 0)
	lowPriorityPossibleRooms := make([]*Room, 0)
	superLowPriorityPossibleRooms := make([]*Room, 0)
	var stillHaveSlotRoom int
	for _, room := range gameInstance.GameData().rooms.Copy() {
		if gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), room.requirement) == nil &&
			len(room.players.coreMap) < room.maxNumberOfPlayers &&
			room.IsRoomOnline() &&
			!room.HasPassword() {
			stillHaveSlotRoom++
			if !room.isDelayingForNewGame {

				numPlayer := 0

				for _, playerInRoom := range room.Players().copy() {
					if playerInRoom.PlayerType() == "bot" {
					} else {
						numPlayer++
					}
				}
				if numPlayer == 0 {
					superLowPriorityPossibleRooms = append(superLowPriorityPossibleRooms, room)
				} else {
					if room.IsPlaying() {
						lowPriorityPossibleRooms = append(lowPriorityPossibleRooms, room)
					} else {
						possibleRooms = append(possibleRooms, room)
					}
				}

				if len(possibleRooms) >= 50 {
					break
				}
			}
		}
	}

	var validRoom *Room
	if len(possibleRooms) > 0 {
		validRoom = possibleRooms[rand.Intn(len(possibleRooms))]
	} else if len(lowPriorityPossibleRooms) > 0 {
		validRoom = lowPriorityPossibleRooms[rand.Intn(len(lowPriorityPossibleRooms))]
	} else if len(superLowPriorityPossibleRooms) > 0 {
		validRoom = superLowPriorityPossibleRooms[rand.Intn(len(superLowPriorityPossibleRooms))]
	}

	if validRoom == nil {
		if GameHasProperty(gameInstance, GamePropertyNoCreateRoom) {
			if stillHaveSlotRoom == 0 {
				log.LogSerious("Hết phòng %s, currency type %s, money %d", gameInstance.GameCode(), gameInstance.CurrencyType(), player.GetMoney(currencyType))
			}
			return nil, errors.New("Hiện không có phòng nào trống, xin đợi trong giây lát")
		} else {
			var validEntry BetEntryInterface
			for _, entry := range gameInstance.BetData().Entries() {
				if gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetMoney(currencyType), entry.Min()) == nil {
					validEntry = entry
					break
				}
			}

			if validEntry != nil {
				return CreateRoom(gameInstance, player, validEntry.Min(), gameInstance.MaxNumberOfPlayers(), "")
			} else {
				return nil, errors.New("Hiện không có phòng nào trống, xin đợi trong giây lát")
			}
		}

	}

	return joinRoom(validRoom, player, 0, "")
}

func quickJoinRoomTypeQuick(gameInstance GameInterface, player GamePlayer) (room *Room, err error) {
	if GameHasProperty(gameInstance, GamePropertyBuyIn) {
		return nil, errors.New("err:not_implemented")
	}

	currencyType := gameInstance.CurrencyType()
	var validRoom *Room
	var validRequirement int64
	var randomPoint int64

	err = gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), gameInstance.BetData().Entries()[0].Min())
	if err != nil {
		return nil, err
	}

	for _, entry := range gameInstance.BetData().Entries() {
		err = gameInstance.IsPlayerMoneyValidToStayInRoom(player.GetAvailableMoney(currencyType), entry.Min())
		if err != nil {
			break
		} else {
			validRequirement = entry.Min()
		}
	}
	randomPoint = int64(rand.Intn(int(validRequirement)))
	for _, entry := range gameInstance.BetData().Entries() {
		if entry.Min() > randomPoint {
			validRequirement = entry.Min()
			break
		}
	}

	possibleRooms := make([]*Room, 0)
	lowPriorityPossibleRooms := make([]*Room, 0)
	superLowPriorityPossibleRooms := make([]*Room, 0)
	for _, room := range gameInstance.GameData().Rooms().Copy() {
		if room.Requirement() == validRequirement &&
			len(room.Players().coreMap) < gameInstance.MaxNumberOfPlayers() &&
			room.IsRoomOnline() &&
			!room.isDelayingForNewGame {
			if room.Session() != nil {
				if room.Session().IsDelayingForNewGame() {
					continue
				}
			}

			numPlayer := 0

			for _, playerInRoom := range room.Players().copy() {
				if playerInRoom.PlayerType() == "bot" {
				} else {
					numPlayer++
				}
			}
			if numPlayer == 0 {
				superLowPriorityPossibleRooms = append(superLowPriorityPossibleRooms, room)
			} else {
				if room.IsPlaying() {
					lowPriorityPossibleRooms = append(lowPriorityPossibleRooms, room)
				} else {
					possibleRooms = append(possibleRooms, room)
				}
			}

			if len(possibleRooms) >= 50 {
				break
			}
		}
	}

	if len(possibleRooms) > 0 {
		validRoom = possibleRooms[rand.Intn(len(possibleRooms))]
	} else if len(lowPriorityPossibleRooms) > 0 {
		validRoom = lowPriorityPossibleRooms[rand.Intn(len(lowPriorityPossibleRooms))]
	} else if len(superLowPriorityPossibleRooms) > 0 {
		randForCreateOrJoin := rand.Intn(4)
		if randForCreateOrJoin < 3 {
			validRoom = superLowPriorityPossibleRooms[rand.Intn(len(superLowPriorityPossibleRooms))]
		}
	}

	if validRoom == nil {
		return createRoomTypeQuick(gameInstance, player, validRequirement, 0, currencyType)
	} else {
		room, err = joinRoom(validRoom, player, 0, "")
		if err != nil && err.Error() == l.Get(l.M0039) {
			return createRoomTypeQuick(gameInstance, player, validRequirement, 0, currencyType)
		} else {
			return room, err
		}

	}
}

func createRoomTypeQuick(gameInstance GameInterface, player GamePlayer, requirement int64, buyIn int64, currencyType string) (room *Room, err error) {
	room, err = NewRoom(gameInstance, player, gameInstance.DefaultNumberOfPlayers(), requirement, buyIn, currencyType, "")
	if err != nil {
		return nil, err
	}
	player.SetRoom(room)
	gameInstance.GameData().Rooms().Set(room.Id(), room)
	gameInstance.HandleRoomCreated(room)
	return room, nil
}

// call func: handleJoinRoom
func joinRoom(validRoom *Room, player GamePlayer, buyIn int64, password string) (room *Room, err error) {
	action := NewRoomActionContext()
	action.actionType = "JoinRoom"
	action.room = validRoom
	action.player = player
	action.password = password
	action.buyIn = buyIn

	response := validRoom.sendAction(action)
	return response.room, response.err
}

/*
handle method
*/

// main join func
func (room *Room) handleJoinRoom(player GamePlayer, buyIn int64, password string) (err error) {
	if room.password != password {
		return errors.New("err:invalid_password")
	}
	err = room.game.IsPlayerMoneyValidToStayInRoom(player.GetMoney(room.currencyType), room.requirement)
	if err != nil {
		return err
	}
	gameInstance := room.Game()
	validEntry := room.Game().BetData().GetEntry(room.requirement)
	if GameHasProperty(gameInstance, GamePropertyBuyIn) {
		var isOk bool
		if buyIn >= validEntry.Min() &&
			buyIn <= validEntry.Max() &&
			buyIn%validEntry.Step() == 0 &&
			player.GetMoney(room.currencyType) >= buyIn {
			isOk = true
		}

		if !isOk {
			return errors.New("Buy In không hợp lệ")
		}
	}

	err = room.addPlayer(player, buyIn, true)
	if err != nil {
		return err
	}

	return nil
}

func (room *Room) handleRegisterLeaveRoom(player GamePlayer) (err error) {
	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}
	// người đó đang trong trận thì thêm vào danh sách thoát
	// không thì cho ra luôn
	if room.IsPlaying() {
		if room.Session().GetPlayer(player.Id()) != nil {
			if !ContainPlayer(room.registerLeaveRoom, player) {
				room.registerLeaveRoom = append(room.registerLeaveRoom, player)
			}
			room.sendNotifyPlayerLeaveStatus(player)
			return nil
		}
	}
	err = room.removePlayer(player)
	return err
}

func (room *Room) handleUnregisterLeaveRoom(player GamePlayer) (err error) {
	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	room.registerLeaveRoom = RemovePlayer(room.registerLeaveRoom, player)
	room.sendNotifyPlayerLeaveStatus(player)
	return nil
}
