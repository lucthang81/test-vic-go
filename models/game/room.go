package game

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/congrat_queue"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
)

// store game result

const (
	RoomTypeSystem = "system"
	RoomTypeNormal = "normal"
)

func init() {
	_ = rand.Intn(5)
	_ = currency.Money
	_ = congrat_queue.GetQueue
}

type GameResult struct {
	result string // win / lost / draw
	change int64  // money
	rank   int    // first prize is 0, seconds will be 1, etc
	bet    int64
}

func NewGameResult(result string, change int64, rank int, bet int64) *GameResult {
	return &GameResult{
		result: result,
		change: change,
		rank:   rank,
		bet:    bet,
	}
}

func (result *GameResult) Rank() int {
	return result.rank
}

func (result *GameResult) Bet() int64 {
	return result.bet
}

func (result *GameResult) SerializedData() map[string]interface{} {
	return map[string]interface{}{
		"result": result.result,
		"change": result.change,
		"rank":   result.rank,
	}
}

type RoomActionContext struct {
	actionType                  string
	room                        *Room
	newRoomId                   int64
	player                      GamePlayer
	password                    string
	bet                         int64
	buyIn                       int64
	requirement                 int64
	willBeKickPlayer            GamePlayer
	session                     GameSessionInterface
	result                      map[string]interface{}
	delayUntilsNewActionSeconds int

	responseChan chan *RoomResponseContext
}

type RoomResponseContext struct {
	room *Room
	data map[string]interface{}
	err  error
}

func NewRoomActionContext() *RoomActionContext {
	return &RoomActionContext{
		responseChan: make(chan *RoomResponseContext),
	}
}

type Room struct {
	id                 int64
	stateString        string
	requirement        int64
	currencyType       string
	maxNumberOfPlayers int
	gameCode           string

	password string

	roomType string

	models  ModelsInterface
	game    GameInterface
	owner   GamePlayer
	players *IntGamePlayerMap // int is the order, not playerId

	bets *utils.Int64Int64Map

	lastMatchResults map[int64]*GameResult

	session GameSessionInterface

	onlinePlayers         *Int64GamePlayerMap
	timeoutOfflinePlayers *utils.Int64TimeOutMap

	autoStartTimer        *utils.TimeOut
	startAutoStartTimerAt time.Time

	registerLeaveRoom          []GamePlayer
	registerToBeOwner          []GamePlayer
	registerJoinOwnerOtherRoom *utils.Int64Int64Map // key is playerId, value is roomId
	willNotBeOwnerNextRound    bool

	startLoopTime    time.Time
	alreadyCloseRoom bool
	actionChan       chan *RoomActionContext
	counter          int
	logFile          *log.LogObject
	initActionStack  string

	isDelayingForNewGame bool

	// log
	subLog string

	//
	SharedData map[string]interface{}
	Mutex      sync.RWMutex
}

func NewSystemRoom(game GameInterface, firstPlayer GamePlayer, maxNumberOfPlayers int, requirement int64, buyIn int64, currencyType string, password string) (room *Room, err error) {
	room = &Room{
		id:                 getNewRoomId(),
		gameCode:           game.GameCode(),
		currencyType:       currencyType,
		game:               game,
		requirement:        requirement,
		maxNumberOfPlayers: maxNumberOfPlayers,

		players:                    NewIntGamePlayerMap(),
		bets:                       utils.NewInt64Int64Map(),
		registerLeaveRoom:          make([]GamePlayer, 0),
		registerToBeOwner:          make([]GamePlayer, 0),
		registerJoinOwnerOtherRoom: utils.NewInt64Int64Map(),

		onlinePlayers:         NewInt64GamePlayerMap(),
		timeoutOfflinePlayers: utils.NewInt64TimeOutMap(),

		password:   password,
		roomType:   RoomTypeSystem,
		actionChan: make(chan *RoomActionContext),

		SharedData: make(map[string]interface{}),
	}

	if firstPlayer != nil {
		room.addPlayer(firstPlayer, buyIn, false)
	} else {
		room.players = NewIntGamePlayerMap()
		room.bets = utils.NewInt64Int64Map()
	}

	if room.gameCode == "xocdia2" {
		room.Mutex.Lock()
		room.SharedData["lastBetInfo"] = make(map[int64]map[string]int64)
		temp := z.NewSizedList(32)
		room.SharedData["outcomeHistory"] = &temp
		room.SharedData["hostId"] = int64(0)
		room.SharedData["hostMinMoney"] = int64(1000) * room.requirement
		room.Mutex.Unlock()
	} else if room.gameCode == "phom" ||
		room.gameCode == "phomSolo" {
		room.Mutex.Lock()
		room.SharedData["testField"] = "talaTung208"
		room.SharedData["lastWinnerId"] = int64(0)
		room.Mutex.Unlock()
	}

	room.logFile = log.NewLogObject(utils.RandSeq(20))
	go room.StartEventLoop()
	return room, nil
}

func NewRoom(game GameInterface, player GamePlayer, maxNumberOfPlayers int, requirement int64, buyIn int64, currencyType string, password string) (room *Room, err error) {
	room, err = NewSystemRoom(game, player, maxNumberOfPlayers, requirement, buyIn, currencyType, password)

	if err != nil {
		return nil, err
	}
	room.roomType = RoomTypeNormal

	if GameHasProperties(game, []string{GamePropertyAlwaysHasOwner}) {
		room.makeOwner(player)
	}
	return room, nil
}

func (room *Room) StartEventLoop() {
	defer func() {
		room.actionChan = nil
		if r := recover(); r != nil {
			log.SendMailWithCurrentStack(fmt.Sprintf("room event error %v, init stack %v \n", r, room.initActionStack))
		}
	}()

	for {
		action, ok := <-room.actionChan
		if !ok {
			room.logToFile("process action close, counter %d", room.counter)
			break
		}
		room.logToFile("process action %s, counter %d", action.actionType, room.counter)
		responseContext := &RoomResponseContext{}
		if action.actionType == "JoinRoom" {
			err := room.handleJoinRoom(action.player, action.buyIn, action.password)
			responseContext.room = room
			responseContext.err = err
		} else if action.actionType == "RegisterLeaveRoom" {
			err := room.handleRegisterLeaveRoom(action.player)
			responseContext.err = err
		} else if action.actionType == "UnregisterLeaveRoom" {
			err := room.handleUnregisterLeaveRoom(action.player)
			responseContext.err = err
		} else if action.actionType == "RegisterToBeOwner" {
			err := room.handleRegisterToBeOwner(action.player)
			responseContext.err = err
		} else if action.actionType == "UnregisterToBeOwner" {
			err := room.handleUnregisterToBeOwner(action.player)
			responseContext.err = err
		} else if action.actionType == "AssignOwner" {
			err := room.handleAssignOwner(action.player)
			responseContext.err = err
		} else if action.actionType == "RemoveOwner" {
			err := room.handleRemoveOwner()
			responseContext.err = err
		} else if action.actionType == "StartGame" {
			data, err := room.startGame(action.player)
			responseContext.err = err
			responseContext.data = data
		} else if action.actionType == "BuyIn" {
			err := room.handleBuyIn(action.player, action.buyIn)
			responseContext.err = err
		} else if action.actionType == "KickPlayer" {
			err := room.kickPlayer(action.player, action.willBeKickPlayer)
			responseContext.err = err
		} else if action.actionType == "DidEndGame" {
			room.handleDidEndGame(action.result, action.delayUntilsNewActionSeconds)
		} else if action.actionType == "AcceptToBeOwnerOtherRoom" {
			err := room.handleAcceptToBeOwnerOtherRoom(action.player, action.newRoomId)
			responseContext.err = err
		} else {
			err := errors.New("wrong_room_action")
			responseContext.err = err
		}
		action.responseChan <- responseContext
		room.counter++
	}
}

func (room *Room) sendAction(action *RoomActionContext) *RoomResponseContext {
	if room.alreadyCloseRoom {
		responseContext := &RoomResponseContext{}
		responseContext.err = errors.New("err:room_already_close")
		return responseContext
	}

	// if action.actionType == "StartGame" {
	// 	fmt.Println("startGame???", log.GetStack())
	// }

	room.logToFile("receive action %s", action.actionType)
	initStack := log.GetStack()
	room.initActionStack = fmt.Sprintf("actionType %s %v", action.actionType, initStack)

	timeout := utils.NewTimeOut(60 * time.Second)
	defer timeout.SetShouldHandle(false)
	go func(timeout *utils.TimeOut, action string, counter int, initStack string, room *Room) {
		if timeout.Start() {
			message := fmt.Sprintf("%s room stuck %s id:%d, action %s, counter %d \n %s \n init action stack %s",
				time.Now().Format(time.ANSIC), room.gameCode, room.Id(), action, counter, room.logFile.Content(), initStack)
			log.SendMailWithCurrentStack(message)
		}
	}(timeout, action.actionType, room.counter, initStack, room)

	go room.actualSendRequest(action)
	room.logToFile("done actual send action %s", action.actionType)
	response := <-action.responseChan
	room.logToFile("receive result %s", action.actionType)
	return response
}

func (room *Room) actualSendRequest(action *RoomActionContext) {
	if room.actionChan == nil {
		room.logToFile("nil actionChan %s", action.actionType)
		responseContext := &RoomResponseContext{}
		responseContext.err = errors.New("err:room_already_close")
		go func(actionInBlock *RoomActionContext) {
			actionInBlock.responseChan <- responseContext
			room.logToFile("done give response nil actionChan %s", actionInBlock.actionType)
		}(action)
		return
	}
	defer func(actionInDefer *RoomActionContext) {
		if r := recover(); r != nil {
			room.logToFile("recover already close actionChan %s", actionInDefer.actionType)
			// only reason will be room already close
			responseContext := &RoomResponseContext{}
			responseContext.err = errors.New("err:room_already_close")
			actionInDefer.responseChan <- responseContext
			room.logToFile("done give response close actionChan %s", actionInDefer.actionType)
		}
	}(action)
	room.actionChan <- action
	room.logToFile("done send action %s", action.actionType)
}

func (room *Room) StateString() string {
	return room.stateString
}

func (room *Room) Requirement() int64 {
	return room.requirement
}

func (room *Room) SetRequirement(requirement int64) {
	room.requirement = requirement
}

func (room *Room) Id() int64 {
	return room.id
}

func (room *Room) MaxNumberOfPlayers() int {
	return room.maxNumberOfPlayers
}

func (room *Room) SetMaxNumberOfPlayers(maxNumberOfPlayers int) {
	room.maxNumberOfPlayers = maxNumberOfPlayers
}

func (room *Room) GameCode() string {
	return room.gameCode
}

func (room *Room) Game() GameInterface {
	return room.game
}

func (room *Room) CurrencyType() string {
	return room.currencyType
}

func (room *Room) Owner() GamePlayer {
	return room.owner
}

func (room *Room) Players() *IntGamePlayerMap {
	return room.players
}

func (room *Room) Session() GameSessionInterface {
	return room.session
}

func (room *Room) SetSession(session GameSessionInterface) {
	room.session = session
}

func (room *Room) HasPassword() bool {
	return len(room.password) > 0
}

func (room *Room) SetRoomType(roomType string) {
	room.roomType = roomType
}

func (room *Room) RoomType() string {
	return room.roomType
}

func (room *Room) PlayersId() []int64 {
	playersId := make([]int64, 0, len(room.players.coreMap))
	for _, player := range room.players.copy() {
		playersId = append(playersId, player.Id())
	}
	return playersId
}

func (room *Room) ContainsPlayer(player GamePlayer) bool {
	for _, currentPlayer := range room.players.copy() {
		if currentPlayer.Id() == player.Id() {
			return true
		}
	}
	return false
}

func (room *Room) ContainsPlayerId(playerId int64) bool {
	for _, currentPlayer := range room.players.copy() {
		if currentPlayer.Id() == playerId {
			return true
		}
	}
	return false
}

func (room *Room) addPlayer(player GamePlayer, buyIn int64, shouldNotify bool) (err error) {
	if player.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return err
	}
	if room.ContainsPlayer(player) {
		err = details_error.NewError("err:already_in_room", room.SerializedDataFull(player))
		return err
	}

	if room.MaxNumberOfPlayers() == len(room.players.coreMap) {
		return errors.New(l.Get(l.M0039))
	}

	if GameHasProperties(room.game, []string{GamePropertyAlwaysHasOwner}) {
		if len(room.players.coreMap) == 0 {
			err = room.game.IsPlayerMoneyValidToBecomeOwner(player.GetAvailableMoney(room.currencyType),
				room.requirement, room.maxNumberOfPlayers, 1)
			if err != nil {
				return err
			} else {
				if !room.isPlayerOwner(player) {
					player.SetRoom(room)
					room.makeOwner(player)
				}
			}
		}
	}

	player.SetRoom(room)

	var index int
	for i := 0; i < room.game.MaxNumberOfPlayers(); i++ {
		if room.players.get(i) == nil {
			index = i
			room.players.set(i, player)
			break
		}
	}

	if player.IsOnline() {
		room.onlinePlayers.Set(player.Id(), player)
	}

	// put money on table
	if GameHasProperties(room.Game(), []string{GamePropertyBuyIn}) {
		err = player.FreezeMoney(buyIn, room.currencyType, room.GetRoomIdentifierString(), true)
		if err != nil {
			return err
		}
	} else {
		if room.isPlayerOwner(player) {
			amount := room.game.MoneyOnTableForOwner(room.Requirement(), room.MaxNumberOfPlayers(), room.players.Len())
			err = player.FreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), true)
			if err != nil {
				return err
			}
		} else {
			amount := room.game.MoneyOnTable(room.Requirement(), room.MaxNumberOfPlayers(), len(room.players.coreMap))
			err = player.FreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), true)
			if err != nil {
				return err
			}
		}
	}

	if room.session != nil {
		room.session.HandlePlayerAddedToGame(player)
	}

	// stop auto start timer if has
	room.updateAutoStartTimer() // in addPlayer func
	if shouldNotify {
		room.sendNotifyNewPlayerJoinRoom(player, index)
	}

	return nil
}
func (room *Room) removePlayer(player GamePlayer) (err error) {
	if player.Name() == game_config.LogUsername() {
		log.LogSeriousWithStack("remove player %s", player.Name())
	}
	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	if len(room.players.coreMap) == 1 {
		player.SetRoom(nil)
		room.players = NewIntGamePlayerMap()

		player.FreezeMoney(0, room.currencyType, room.GetRoomIdentifierString(), true)
		room.removePlayerDataWhenRemovePlayer(player)
		room.updateAutoStartTimer() // in func removePlayer

		room.sendNotifyPlayerLeaveRoom(player)
		room.cleanupAndRemoveRoom()

		if room.Session() != nil {
			room.Session().HandlePlayerRemovedFromGame(player)
		}
		return nil
	} else {
		needNewOwner := false
		ownerChange := false

		if GameHasProperties(room.game, []string{GamePropertyAlwaysHasOwner}) {
			if room.isPlayerOwner(player) {
				// assign another player as owner
				needNewOwner = true
			}
		}

		var orderToRemove int
		orderToRemove = room.getOrderOfPlayer(player)

		if needNewOwner {
			for i := 1; i < room.MaxNumberOfPlayers(); i++ {
				nextOrder := orderToRemove + i
				if nextOrder >= room.MaxNumberOfPlayers() {
					nextOrder = nextOrder - room.MaxNumberOfPlayers()
				}
				playerAtOrder := room.players.Get(nextOrder)

				if playerAtOrder != nil {
					if room.game.IsPlayerMoneyValidToBecomeOwner(room.GetTotalPlayerMoney(playerAtOrder.Id()),
						room.requirement,
						room.maxNumberOfPlayers,
						len(room.players.coreMap)) != nil {
						continue
					}
					// turn to owner, so get money on table back, cause owner do not need to bet if everyone is playing against him
					// get everyone money back too
					// fmt.Println("make new owner", playerAtOrder.Id())
					room.makeOwner(playerAtOrder)
					// log.Log("change owner from %d to %d", player.Id(), oldPlayer.Id())
					ownerChange = true
					break
				}
			}

			if GameHasProperties(room.game, []string{GamePropertyAlwaysHasOwner}) {
				if !ownerChange {
					// cannot select any owner, we now just kick all player
					for _, playerToKick := range room.players.copy() {
						if player.Id() != playerToKick.Id() {
							go RegisterLeaveRoom(room.Game(), playerToKick)
						}
					}
				}
			}
		}

		player.SetRoom(nil)
		room.players.delete(orderToRemove)

		player.FreezeMoney(0, room.currencyType, room.GetRoomIdentifierString(), true)

		room.removePlayerDataWhenRemovePlayer(player)

		room.removePromoteRegisterOwner()

		room.updateAutoStartTimer() // in func removePlayer

		if ownerChange {
			room.sendNotifyNewOwnerInRoom(room.owner)
		}
		room.sendNotifyPlayerLeaveRoom(player)

		if room.Session() != nil {
			room.Session().HandlePlayerRemovedFromGame(player)
		}
		return nil
	}

}

func (room *Room) removePlayerDataWhenRemovePlayer(player GamePlayer) {
	room.bets.Delete(player.Id())
	room.onlinePlayers.Delete(player.Id())
	room.timeoutOfflinePlayers.Delete(player.Id())
	room.registerLeaveRoom = RemovePlayer(room.registerLeaveRoom, player)
	room.registerToBeOwner = RemovePlayer(room.registerToBeOwner, player)
	room.registerJoinOwnerOtherRoom.Delete(player.Id())

	if room.isPlayerOwner(player) {
		room.removeOwner()
	}
	room.save()
}

func (room *Room) IncreaseAndFreezeThoseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}
	defer room.save()
	playerInstance.LockMoney(room.currencyType)
	playerInstance.IncreaseMoney(amount, room.currencyType, false)
	playerInstance.IncreaseFreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), false)
	playerInstance.UnlockMoney(room.currencyType)
	return err
}

func (room *Room) IncreaseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}
	defer room.save()
	_, err = playerInstance.IncreaseMoney(amount, room.currencyType, shouldLock)
	return err
}

func (room *Room) DecreaseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error) {
	if amount < 0 {
		return errors.New("err:decrease_negative")
	}
	defer room.save()

	// decrease, will only decrease freeze money
	_, err = playerInstance.DecreaseFromFreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), shouldLock)
	return err
}

func (room *Room) kickPlayer(player GamePlayer, willBeKickedPlayer GamePlayer) (err error) {
	if room.IsPlaying() {
		return errors.New("err:already_start_playing")
	}

	if !GameHasProperty(room.game, GamePropertyCanKick) {
		return errors.New("err:cannot_kick_in_this_game")
	}
	if !room.isPlayerOwner(player) {
		return errors.New("err:no_permission")
	}

	if room.isPlayerOwner(willBeKickedPlayer) {
		return errors.New("err:cannot_kick_owner")
	}
	if !room.ContainsPlayer(willBeKickedPlayer) {
		return errors.New("err:player_not_in_room")
	}
	return room.removePlayer(willBeKickedPlayer)

}

func (room *Room) invitePlayerToRoom(player GamePlayer, invitePlayer GamePlayer) (err error) {
	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	if room.ContainsPlayer(invitePlayer) {
		return errors.New("err:player_already_in_room")
	}

	if invitePlayer.Room() != nil {
		err = details_error.NewError(l.Get(l.M0038), player.Room().SerializedDataFull(player))
		return err
	}

	err = room.game.IsPlayerMoneyValidToStayInRoom(invitePlayer.GetMoney(room.currencyType), room.requirement)

	if err != nil {
		return err
	}

	if room.MaxNumberOfPlayers() == len(room.players.coreMap) {
		return errors.New(l.Get(l.M0039))
	}

	room.sendNotifyPlayerReceiveInvitationToJoinRoom(invitePlayer.Id(), player)
	return nil
}

func (room *Room) chat(player GamePlayer, message string) (err error) {
	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	room.sendNotifyMessageInRoom(player, message)
	return nil
}

func (room *Room) HandlePlayerOffline(player GamePlayer) {
	room.onlinePlayers.Delete(player.Id())
	room.sendNotifyPlayerStatus(player, false)
	if room.session != nil {
		room.session.HandlePlayerOffline(player)
	}

	go room.waitForOfflinePlayer(player)
}

func (room *Room) HandlePlayerOnline(player GamePlayer) {
	room.onlinePlayers.Set(player.Id(), player)
	if room.timeoutOfflinePlayers.Get(player.Id()) != nil {
		room.timeoutOfflinePlayers.Get(player.Id()).SetShouldHandle(false)
		room.timeoutOfflinePlayers.Delete(player.Id())
	}
	room.sendNotifyPlayerStatus(player, true)
}

func (room *Room) IsRoomOnline() bool {
	if room.actionChan == nil {
		return false
	}

	if room.roomType == RoomTypeSystem {
		return true
	}

	if GameHasProperty(room.game, GamePropertyCasino) {
		if room.Owner() != nil {
			return true
		}
	}

	return room.onlinePlayers.Len() >= 1
}

// điều kiện bắt đầu ván chơi
func (room *Room) IsRoomReadyForAutoStart() bool {
	// numPlayer = nHuman + nBot
	numPlayer := len(room.players.Copy())
	if numPlayer >= 2 {
		if room.GetNumberOfHumans() == 0 {
			return true
		} else {
			return true
		}
	} else { // numPlayer <= 1
		return false
	}
}

func (room *Room) startGame(player GamePlayer) (data map[string]interface{}, err error) {
	if GameHasProperties(room.game, []string{GamePropertyNoStart}) {
		return nil, errors.New("err:cannot_start_manually_in_this_game")
	}
	// fmt.Println("============abc=======", room.GameCode(), room.Game().Properties())
	/*
		if !GameHasProperties(room.game, []string{GamePropertyOwnerAssignByGame}) {
			if !room.isPlayerOwner(player) {
				return nil, errors.New("err:no_permission")
			}
		}
	*/
	if len(room.players.coreMap) <= 1 {
		return nil, errors.New("err:not_enough_player")
	}

	if room.IsPlaying() {
		return nil, errors.New("err:playing")
	}

	// remove start game auto timer
	room.stopAutoStartTimer()

	room.NotifyAboutToStartGame()
	playersInGame := room.getPlayersSliceInOrder()
	session, err := room.game.StartGame(room, room.owner,
		playersInGame, room.requirement, room.bets.Copy(), room.lastMatchResults)
	if err != nil {
		return nil, err
	}
	room.session = session

	room.sendNotifyIsPlayingRoom()

	if player != nil {
		return session.SerializedDataForPlayer(player), nil
	} else {
		return session.SerializedData(), nil
	}
}

func (room *Room) getOrderOfPlayer(playerInstance GamePlayer) int {
	for order, player := range room.players.copy() {
		if player.Id() == playerInstance.Id() {
			return order
		}
	}
	return -1
}

func (room *Room) GetPlayerAtIndex(index int) (playerInstance GamePlayer) {
	playerInstance = room.players.Get(index)
	return playerInstance
}

func (room *Room) getListOfPlayerOrders() []int {
	orders := make([]int, 0)
	for order, _ := range room.players.copy() {
		orders = append(orders, order)
	}

	sort.Sort(utils.ByInt(orders))
	return orders
}

func (room *Room) getPlayersSliceInOrder() []GamePlayer {
	playersInGame := make([]GamePlayer, 0)
	// do this to keep the order
	temp := make(map[int]GamePlayer)
	for order, player := range room.players.copy() {
		temp[order] = player
	}

	for i := 0; i < len(room.players.coreMap); i++ {
		minOrder := 100000
		for order, _ := range temp {
			if order < minOrder {
				minOrder = order
			}
		}
		playersInGame = append(playersInGame, temp[minOrder])
		delete(temp, minOrder)
	}
	return playersInGame
}

func (room *Room) cleanupAndRemoveRoom() (err error) {
	if room.roomType == RoomTypeSystem {
		return nil
	}

	if room.game != nil {
		if room.game.GameData() != nil {
			if room.game.GameData().rooms.Len() > 0 {
				room.game.GameData().rooms.Delete(room.Id())
				if room.Session() != nil {
					room.Session().CleanUp()
					room.session = nil
				}
			}
		}
	}

	room.removeFromCache()
	room.alreadyCloseRoom = true
	close(room.actionChan)
	return nil
}

func (room *Room) getPlayer(playerId int64) (player GamePlayer) {
	for _, player := range room.players.copy() {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (room *Room) sendNotifyNewPlayerJoinRoom(newPlayer GamePlayer, index int) {
	roomData := room.SerializedData()
	playerData := newPlayer.SerializedDataMinimal()
	data := make(map[string]interface{})
	data["player"] = playerData
	data["room"] = roomData
	data["index"] = index
	data["bet"] = room.bets.Get(newPlayer.Id())
	data["money_on_table"] = room.GetMoneyOnTable(newPlayer.Id())
	data["nHuman"] = room.GetNumberOfHumans()
	server.SendRequests("new_player_join_room", data, room.PlayersId())
}

func (room *Room) sendNotifyPlayerLeaveRoom(player GamePlayer) {
	roomData := room.SerializedData()
	playerData := player.SerializedDataMinimal()
	data := make(map[string]interface{})
	data["player"] = playerData
	data["room"] = roomData
	server.SendRequests("player_leave_room", data, room.PlayersId())
	server.SendRequest("player_leave_room", data, player.Id())
}

func (room *Room) sendNotifyNewOwnerInRoom(player GamePlayer) {
	var playerData map[string]interface{}
	if player == nil {
		playerData = nil
	} else {
		playerData = player.SerializedDataMinimal()
	}

	for _, destinationPlayer := range room.players.copy() {
		roomData := room.SerializedDataWithFields(destinationPlayer, []string{})

		data := make(map[string]interface{})
		data["player"] = playerData
		data["room"] = roomData
		server.SendRequest("new_owner_in_room", data, destinationPlayer.Id())
	}

}

func (room *Room) sendNotifyOwnerStatusChange() {
	for _, player := range room.players.copy() {
		roomData := room.SerializedDataWithFields(player, []string{"session"})
		data := make(map[string]interface{})
		data["room"] = roomData
		server.SendRequest("player_owner_status_change", data, player.Id())
	}
}

func (room *Room) sendNotifyPlayerUpdateBet(player GamePlayer) {
	roomData := room.SerializedData()
	playerData := player.SerializedDataMinimal()
	data := make(map[string]interface{})
	data["player"] = playerData
	data["room"] = roomData
	data["bet"] = room.bets.Get(player.Id())
	data["money_on_table"] = room.GetMoneyOnTable(player.Id())
	server.SendRequests("player_update_bet", data, room.PlayersId())
}

func (room *Room) sendNotifyPlayerStatus(player GamePlayer, isOnline bool) {
	roomData := room.SerializedData()
	playerData := player.SerializedDataMinimal()
	data := make(map[string]interface{})
	data["player"] = playerData
	data["room"] = roomData
	if isOnline {
		data["status"] = "online"
	} else {
		data["status"] = "offline"
	}
	server.SendRequests("player_status_change", data, room.PlayersId())
}

func (room *Room) sendNotifyPlayerLeaveStatus(player GamePlayer) {
	roomData := room.SerializedDataWithFields(player, []string{"session"})
	data := make(map[string]interface{})
	data["room"] = roomData
	server.SendRequest("player_leave_status_change", data, player.Id())
}

func (room *Room) sendNotifyRequirementChange() {
	roomData := room.SerializedData()
	data := make(map[string]interface{})
	data["room"] = roomData
	server.SendRequests("room_requirement_change", data, room.PlayersId())
}

func (room *Room) sendNotifyOwnerPingRoom() {
	roomData := room.SerializedData()
	playersData := make([]map[string]interface{}, 0)
	for _, player := range room.players.copy() {
		playerData := make(map[string]interface{})
		playerData["id"] = player.Id()
		playerData["bet"] = room.bets.Get(player.Id())
		playerData["money_on_table"] = room.GetMoneyOnTable(player.Id())
		playersData = append(playersData, playerData)
	}
	data := make(map[string]interface{})
	data["players_data"] = playersData
	data["room"] = roomData
	server.SendRequests("owner_ping_room", data, room.PlayersId())

}

func (room *Room) sendNotifyIsPlayingRoom() {
	roomData := room.SerializedData()
	data := make(map[string]interface{})
	data["room"] = roomData
	server.SendRequests("room_is_playing", data, room.PlayersId())
}

func (room *Room) sendNotifyPlayerReceiveInvitationToJoinRoom(playerId int64, fromPlayer GamePlayer) {
	roomData := room.SerializedDataWithFields(fromPlayer, []string{"password"})
	data := make(map[string]interface{})
	data["room"] = roomData
	data["from_player"] = fromPlayer.SerializedDataMinimal()
	server.SendRequest("player_receive_invitation_to_room", data, playerId)
}

func (room *Room) SendNotifyMoneyChange(playerId int64, change int64, reason string, additionalData map[string]interface{}) {
	room.sendNotifyMoneyChange(playerId, change, reason, additionalData)
}

func (room *Room) sendNotifyMoneyChange(playerId int64, change int64, reason string, additionalData map[string]interface{}) {
	data := make(map[string]interface{})
	data["player_id"] = playerId
	data["room_id"] = room.Id()
	data["money_on_table"] = room.GetMoneyOnTable(playerId)
	data["change"] = change
	data["reason"] = reason
	data["data"] = additionalData
	server.SendRequests("player_money_change_in_room", data, room.PlayersId())
}

func (room *Room) sendNotifyMessageInRoom(fromPlayer GamePlayer, message string) {
	roomData := room.SerializedDataWithFields(fromPlayer, []string{"password"})
	data := make(map[string]interface{})
	data["room"] = roomData
	data["from_player"] = fromPlayer.SerializedDataMinimal()
	data["message"] = message
	server.SendRequests("room_receive_message", data, room.PlayersId())
}

func (room *Room) isVip() bool {
	return room.requirement >= room.game.VipThreshold()
}

func (room *Room) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = room.Id()
	data["game_code"] = room.gameCode
	data["currency_type"] = room.currencyType

	if room.owner != nil {
		data["owner_id"] = room.owner.Id()
		data["owner"] = room.owner.SerializedDataMinimal()
		data["name"] = room.owner.Name()
	} else {
		data["name"] = "Không có nhà cái"
	}

	data["register_leave_ids"] = GetIdFromPlayers(room.registerLeaveRoom)
	data["register_owner_ids"] = GetIdFromPlayers(room.registerToBeOwner)
	data["will_not_be_owner_next_round"] = room.willNotBeOwnerNextRound
	data["requirement"] = room.requirement
	data["max_number_of_players"] = room.maxNumberOfPlayers
	data["number_of_players"] = len(room.players.coreMap)
	data["is_vip"] = room.isVip()
	data["time_until_start"] = room.getTimeUntilsAutoStartGame()
	data["is_playing"] = room.IsPlaying()
	if len(room.password) > 0 {
		data["has_password"] = true
	} else {
		data["has_password"] = false
	}

	room.Mutex.Lock()
	if room.SharedData != nil {
		data["createrPid"] = room.SharedData["createrPid"]
		data["startingMoney"] = room.SharedData["startingMoney"]
	}
	room.Mutex.Unlock()

	return data
}

func (room *Room) SerializedDataMinimal() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = room.Id()
	data["game_code"] = room.gameCode
	data["currency_type"] = room.currencyType
	data["requirement"] = room.requirement
	data["max_number_of_players"] = room.maxNumberOfPlayers
	data["number_of_players"] = len(room.players.coreMap)
	if room.owner != nil {
		data["owner_id"] = room.owner.Id()
		data["name"] = room.owner.Name()
	} else {
		data["name"] = "Không có nhà cái"
	}
	data["is_vip"] = room.isVip()
	data["is_playing"] = room.IsPlaying()
	if len(room.password) > 0 {
		data["has_password"] = true
	} else {
		data["has_password"] = false
	}
	return data
}

func (room *Room) SerializedDataWithFields(currentPlayer GamePlayer, fields []string) (data map[string]interface{}) {
	data = room.SerializedData()

	if utils.ContainsByString(fields, "players_id") {
		playersIdData := make(map[string]int64)
		for order, player := range room.players.copy() {
			playersIdData[fmt.Sprintf("%d", order)] = player.Id()
		}
		data["players_id"] = playersIdData
	}
	if utils.ContainsByString(fields, "bets") {
		betsData := make(map[string]interface{})
		for playerId, bet := range room.bets.Copy() {
			betsData[fmt.Sprintf("%d", playerId)] = bet
		}
		data["bets"] = betsData
	}

	if utils.ContainsByString(fields, "players") {
		playersData := make(map[string]map[string]interface{})
		for order, player := range room.players.copy() {
			playerData := player.SerializedDataMinimal()
			playersData[fmt.Sprintf("%d", order)] = playerData

		}
		data["players"] = playersData
	}

	if utils.ContainsByString(fields, "players_short") {
		// for room.SerializedDataFull
		data["players"] = room.GetPlayersDataForDisplay(currentPlayer)
	}

	if utils.ContainsByString(fields, "session") && room.session != nil {
		if currentPlayer != nil {
			pIMObj := room.session.GetPlayer(currentPlayer.Id())
			if pIMObj == nil {
				data["session"] = room.session.SerializedData()
			} else {
				data["session"] = room.session.SerializedDataForPlayer(pIMObj)
			}
		}
	}

	if utils.ContainsByString(fields, "last_match_results") && room.lastMatchResults != nil {
		lastMatchResultsData := make(map[string]interface{})
		for playerId, result := range room.lastMatchResults {
			lastMatchResultsData[fmt.Sprintf("%d", playerId)] = result.SerializedData()
		}
		data["last_match_results"] = lastMatchResultsData
	}

	if utils.ContainsByString(fields, "moneys_on_table") {
		moneysOnTableData := make(map[string]int64)
		for _, player := range room.players.copy() {
			moneysOnTableData[fmt.Sprintf("%d", player.Id())] = room.GetMoneyOnTable(player.Id())
		}
		data["moneys_on_table"] = moneysOnTableData
	}

	if utils.ContainsByString(fields, "password") {
		data["password"] = room.password
	}

	return data
}

func (room *Room) SerializedDataFull(player GamePlayer) (data map[string]interface{}) {
	return room.SerializedDataWithFields(player, []string{"players_short", "bets", "moneys_on_table", "ready_players_id", "session"})
}

func (room *Room) GetPlayersDataForDisplay(currentPlayer GamePlayer) (
	playersData map[string]map[string]interface{}) {
	playersData = make(map[string]map[string]interface{})
	var counter int

	for i := 0; i < room.maxNumberOfPlayers; i++ {
		player := room.players.get(i)
		if player != nil {
			if currentPlayer.Id() == player.Id() || room.isPlayerOwner(player) {
				playerData := player.SerializedDataMinimal()
				playersData[fmt.Sprintf("%d", i)] = playerData

				counter++
			} else {
				if counter >= 10 {
					continue
				}
				playerData := player.SerializedDataMinimal()
				playersData[fmt.Sprintf("%d", i)] = playerData
				counter++
			}
		}
	}
	return playersData
}

func (room *Room) MoneyDidChange(session GameSessionInterface, playerId int64, change int64, reason string, additionalData map[string]interface{}) {
	room.sendNotifyMoneyChange(playerId, change, reason, additionalData)
}

func (room *Room) removePlayersDidNotMeetRequirement(requirement int64) {
	willBeKick := make([]GamePlayer, 0)
	for _, player := range room.players.copy() {
		if room.isPlayerOwner(player) {
			// skip this
		} else {
			if room.game.IsPlayerMoneyValidToStayInRoom(room.GetTotalPlayerMoney(player.Id()), room.requirement) != nil {
				// kick the player
				willBeKick = append(willBeKick, player)
			}
		}
	}

	for _, player := range willBeKick {
		room.removePlayer(player)
	}

	if room.Owner() != nil {
		if room.game.IsPlayerMoneyValidToBecomeOwner(room.GetTotalPlayerMoney(room.owner.Id()),
			room.requirement, room.MaxNumberOfPlayers(), len(room.players.coreMap)) != nil {
			room.removePlayer(room.Owner())
		}
	}
}

func (room *Room) removeRegisterLeavePlayers() {
	for _, player := range room.registerLeaveRoom {
		newRoomId := room.registerJoinOwnerOtherRoom.Get(player.Id())
		if newRoomId == 0 {
			room.removePlayer(player)
		} else {
			room.removePlayer(player)
			newRoom, err := JoinRoomById(room.game, player, newRoomId, "")
			if err == nil {
				server.SendRequest("join_room_as_owner", newRoom.SerializedDataFull(player), player.Id())
				newRoom.RegisterToBeOwner(player)
			}
			room.registerJoinOwnerOtherRoom.Delete(player.Id())
		}

	}
}

// now, no game have GamePropertyRegisterOwner
func (room *Room) removePromoteRegisterOwner() {
	if GameHasProperties(room.game, []string{GamePropertyRegisterOwner}) {
		var shouldNotify bool
		if room.Owner() != nil {
			if room.willNotBeOwnerNextRound {
				room.removeOwner()
				room.willNotBeOwnerNextRound = false
				shouldNotify = true
			} else {
				if room.game.IsPlayerMoneyValidToBecomeOwner(room.GetTotalPlayerMoney(room.owner.Id()),
					room.requirement, room.MaxNumberOfPlayers(), len(room.players.coreMap)) != nil {
					room.removeOwner()
					room.willNotBeOwnerNextRound = false
					shouldNotify = true
				}
			}
		}
		room.filterOwnerList()
		if room.Owner() == nil {
			if len(room.registerToBeOwner) > 0 {
				room.makeOwner(room.registerToBeOwner[0])
				room.registerToBeOwner = RemovePlayer(room.registerToBeOwner, room.owner)
				shouldNotify = true
			}
		}

		if shouldNotify {
			room.sendNotifyNewOwnerInRoom(room.owner)
		}
	}
}

// now, all games dont have property
func (room *Room) rotateOwner() {
	if GameHasProperties(room.game, []string{GamePropertyRotateOwner}) {
		var shouldNotify bool
		var ownerChange bool
		currencyType := room.currencyType
		ownerOrder := room.getOrderOfPlayer(room.Owner())
		for i := 1; i < room.MaxNumberOfPlayers(); i++ {
			nextOrder := ownerOrder + i
			if nextOrder >= room.MaxNumberOfPlayers() {
				nextOrder = nextOrder - room.MaxNumberOfPlayers()
			}
			playerAtOrder := room.players.Get(nextOrder)

			if playerAtOrder != nil {
				playerInGameMoney := room.GetMoneyOnTable(playerAtOrder.Id()) + playerAtOrder.GetAvailableMoney(currencyType)
				if room.game.IsPlayerMoneyValidToBecomeOwner(playerInGameMoney,
					room.requirement,
					room.maxNumberOfPlayers,
					len(room.players.coreMap)) != nil {
					continue
				}
				// turn to owner, so get money on table back, cause owner do not need to bet if everyone is playing against him
				// get everyone money back too
				room.makeOwner(playerAtOrder)
				// log.Log("change owner from %d to %d", player.Id(), oldPlayer.Id())
				shouldNotify = true
				ownerChange = true
				break
			}
		}

		if !ownerChange {
			playerInGameMoney := room.GetMoneyOnTable(room.Owner().Id()) + room.Owner().GetAvailableMoney(currencyType)
			if room.game.IsPlayerMoneyValidToBecomeOwner(playerInGameMoney,
				room.requirement,
				room.maxNumberOfPlayers,
				len(room.players.coreMap)) != nil {
				// no one can be owner, will kick all

				for _, player := range room.Players().copy() {
					go RegisterLeaveRoom(room.Game(), player)
				}
			} else {
				// do nothing, keep the current owner
			}
		}

		if shouldNotify {
			room.sendNotifyNewOwnerInRoom(room.owner)
		}
	}
}

// save load room

// empty func, do nothing
func (room *Room) save() (err error) {
	// data := room.SerializedDataWithFields(nil, []string{"moneys_on_table"})
	// payload, err := json.Marshal(data)
	// if err != nil {
	// 	log.LogSerious("error marshal room data %s", err)
	// 	return err
	// }
	// stringValue := string(payload)
	// err = dataCenter.SaveKeyValueToCache(GetRoomGroupString(room.game), room.GetRoomIdString(), stringValue)
	// if err != nil {
	// 	log.LogSerious("err when save room %s", err.Error())
	// 	return err
	// }
	// return nil
	return nil
}

func (room *Room) removeFromCache() (err error) {
	return dataCenter.RemoveKeyValueFromCache(GetRoomGroupString(room.game), room.GetRoomIdString())
}

func GetRoomGroupString(game GameInterface) string {
	return fmt.Sprintf("room_%s", game.GameCode())
}

func (room *Room) GetRoomIdString() string {
	return fmt.Sprintf("%d", room.Id())
}

// is reason string in decrease money from freeze
func (room *Room) GetRoomIdentifierString() string {
	return fmt.Sprintf("room_%s_%d", room.game.GameCode(), room.Id())
}

func getIdFromPlayersMap(players []GamePlayer) []int64 {
	keys := make([]int64, 0, len(players))
	for _, player := range players {
		keys = append(keys, player.Id())
	}
	return keys
}

func (room *Room) isPlayerOwner(player GamePlayer) bool {
	if room.owner != nil {
		if room.owner.Id() == player.Id() {
			return true
		}
	}
	return false
}

func (room *Room) isBetValid(bet int64) bool {
	for _, entry := range room.game.BetData().Entries() {
		minBet := entry.Min()
		maxBet := entry.Max()
		step := entry.Step()

		if bet >= minBet && bet <= maxBet && (bet-minBet)%step == 0 {
			return true
		}
	}
	return false
}

func (room *Room) makeOwner(owner GamePlayer) {
	if owner == nil {
		return
	}

	if room.owner != nil {
		room.removeOwner()
	}

	room.owner = owner
	room.bets.Set(owner.Id(), 0)
	if GameHasProperty(room.game, GamePropertyBuyIn) {
		// ignore
	} else {
		amount := room.game.MoneyOnTableForOwner(room.requirement, room.MaxNumberOfPlayers(), len(room.players.coreMap))
		owner.FreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), true)
	}
}

func (room *Room) removeOwner() {
	if room.owner != nil {
		if room.ContainsPlayer(room.owner) {
			amount := room.game.MoneyOnTable(room.requirement, room.MaxNumberOfPlayers(), room.players.Len())
			room.owner.FreezeMoney(amount, room.currencyType, room.GetRoomIdentifierString(), true)
		}
		room.owner = nil
	}
}

type BetEntryPlayer struct {
	playerId int64
	bet      int64
}

type ByPlayerBet []*BetEntryPlayer

func (p ByPlayerBet) Len() int           { return len(p) }
func (p ByPlayerBet) Less(i, j int) bool { return p[i].bet < p[j].bet }
func (p ByPlayerBet) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (room *Room) filterOwnerList() {
	willBeRemove := make([]GamePlayer, 0)
	for _, player := range room.registerToBeOwner {
		err := room.game.IsPlayerMoneyValidToBecomeOwner(player.GetMoney(room.currencyType),
			room.Requirement(), room.MaxNumberOfPlayers(), len(room.players.coreMap))
		if err != nil {
			willBeRemove = append(willBeRemove, player)
		}
	}

	for _, player := range willBeRemove {
		room.registerToBeOwner = RemovePlayer(room.registerToBeOwner, player)
	}
}

func (room *Room) IsPlaying() bool {
	session := room.Session()
	if session == nil {
		return false
	}
	return session.IsPlaying()
}

func (room *Room) waitForOfflinePlayer(player GamePlayer) {
	timeout := utils.NewTimeOut(2 * time.Minute)
	if room.timeoutOfflinePlayers.Get(player.Id()) != nil {
		room.timeoutOfflinePlayers.Get(player.Id()).SetShouldHandle(false)
	}
	room.timeoutOfflinePlayers.Set(player.Id(), timeout)
	if timeout.Start() {
		if !player.IsOnline() {
			if GameHasProperties(room.game, []string{GamePropertyCasino}) &&
				room.isPlayerOwner(player) {
				// do nothing
			} else {
				room.timeoutOfflinePlayers.Delete(player.Id())
				RegisterLeaveRoom(room.game, player)
			}
		}
	}
}

func (room *Room) updateAutoStartTimer() {
	if room.IsRoomReadyForAutoStart() {
		if room.autoStartTimer == nil {
			room.waitToStartGameAutomatically()
		}
	} else {
		room.stopAutoStartTimer()
	}
}

func (room *Room) waitToStartGameAutomatically() {
	if room.autoStartTimer != nil {
		room.autoStartTimer.SetShouldHandle(false)
	}
	room.startAutoStartTimerAt = time.Now()

	go func(roomInBlock *Room) {
		room.autoStartTimer = utils.NewTimeOut(game_config.AutoStartAfter())
		if roomInBlock.autoStartTimer != nil && roomInBlock.autoStartTimer.Start() {
			roomInBlock.autoStartTimer = nil
			roomInBlock.startAutoStartTimerAt = time.Time{}
			roomInBlock.StartGame(roomInBlock.owner)
		}
	}(room)
}

func (room *Room) stopAutoStartTimer() {
	room.startAutoStartTimerAt = time.Time{}
	if room.autoStartTimer != nil {
		room.autoStartTimer.SetShouldHandle(false)
		room.autoStartTimer = nil
	}
}

func (room *Room) getTimeUntilsAutoStartGame() float64 {
	if !room.startAutoStartTimerAt.IsZero() {
		timeDuration := game_config.AutoStartAfter() - time.Now().Sub(room.startAutoStartTimerAt)
		return timeDuration.Seconds()
	} else {
		return 0
	}
	return 0
}

func (room *Room) DebugLog() string {
	// return room.logFile.Content()
	return ""
}

func (room *Room) UnlockMutex() {

}

func (room *Room) logToFile(format string, a ...interface{}) string {
	// line := fmt.Sprintf(format, a...)
	// fmt.Println(line)
	// if room.logFile != nil {
	// 	room.logFile.Log(format, a...)
	// }
	return ""
}
func (room *Room) StartGame(player GamePlayer) (data map[string]interface{}, err error) {
	action := NewRoomActionContext()
	action.actionType = "StartGame"
	action.player = player
	response := room.sendAction(action)
	return response.data, response.err
}

func (room *Room) KickPlayer(player GamePlayer, willBeKickedPlayer GamePlayer) (err error) {
	action := NewRoomActionContext()
	action.actionType = "KickPlayer"
	action.player = player
	action.willBeKickPlayer = willBeKickedPlayer
	response := room.sendAction(action)
	return response.err
}

func (room *Room) BuyIn(player GamePlayer, buyIn int64) (err error) {
	action := NewRoomActionContext()
	action.actionType = "BuyIn"
	action.player = player
	action.buyIn = buyIn
	response := room.sendAction(action)
	return response.err
}

func (room *Room) AssignOwner(player GamePlayer) (err error) {
	action := NewRoomActionContext()
	action.actionType = "AssignOwner"
	action.player = player
	response := room.sendAction(action)
	return response.err
}

func (room *Room) RemoveOwner() (err error) {
	action := NewRoomActionContext()
	action.actionType = "RemoveOwner"
	response := room.sendAction(action)
	return response.err
}

func (room *Room) RegisterToBeOwner(player GamePlayer) (err error) {
	action := NewRoomActionContext()
	action.actionType = "RegisterToBeOwner"
	action.player = player
	response := room.sendAction(action)
	return response.err
}

func (room *Room) Xocdia2BecomeHost(player GamePlayer) (err error) {
	action := NewRoomActionContext()
	action.actionType = "Xocdia2BecomeHost"
	action.player = player
	response := room.sendAction(action)
	return response.err
}

func (room *Room) UnregisterToBeOwner(player GamePlayer) (err error) {
	action := NewRoomActionContext()
	action.actionType = "UnregisterToBeOwner"
	action.player = player
	response := room.sendAction(action)
	return response.err
}

func (room *Room) InvitePlayerToRoom(player GamePlayer, invitePlayer GamePlayer) (err error) {
	return room.invitePlayerToRoom(player, invitePlayer)
}

func (room *Room) Chat(player GamePlayer, message string) (err error) {
	return room.chat(player, message)
}

func (room *Room) GetOwnerList() (data map[string]interface{}, err error) {
	results := make([]map[string]interface{}, 0)
	for _, player := range room.registerToBeOwner {
		results = append(results, player.SerializedDataMinimal())
	}
	data = make(map[string]interface{})
	data["will_not_be_owner_next_round"] = room.willNotBeOwnerNextRound
	data["results"] = results
	return data, nil
}

/*

Game Session callback

*/
func (room *Room) NotifyAboutToStartGame() {
	method := fmt.Sprintf("%s_%s_about_to_start_game", room.gameCode, room.currencyType)
	for _, playerId := range room.PlayersId() {
		server.SendRequest(method, map[string]interface{}{"seconds": game_config.AutoStartAfter().Seconds()}, playerId)
	}
}

func (room *Room) SendAMap(method string, anyMap map[string]interface{}) {
	for _, playerId := range room.PlayersId() {
		server.SendRequest(method, anyMap, playerId)
	}
}

func (room *Room) DidStartGame(session GameSessionInterface) {
	method := fmt.Sprintf("%s_%s_start_game_session", room.gameCode, room.currencyType)
	for _, playerId := range room.PlayersId() {
		server.SendRequest(method, session.SerializedDataForPlayer(room.getPlayer(playerId)), playerId)
	}
}

func (room *Room) DidChangeGameState(session GameSessionInterface) {
	method := fmt.Sprintf("%s_%s_change_game_session", room.gameCode, room.currencyType)
	for _, playerId := range room.PlayersId() {
		server.SendRequest(method, session.SerializedDataForPlayer(room.getPlayer(playerId)), playerId)
	}
}

func (room *Room) DidEndGame(result map[string]interface{}, delayUntilsNewActionSeconds int) {
	if len(room.players.coreMap) != 0 {
		action := NewRoomActionContext()
		action.actionType = "DidEndGame"
		action.result = result
		action.delayUntilsNewActionSeconds = delayUntilsNewActionSeconds
		room.sendAction(action)
	}
}

func (room *Room) SendMessageToPlayer(session GameSessionInterface, playerId int64, method string, data map[string]interface{}) {
	server.SendRequest(method, data, playerId)
}

/*
handle method
*/
func (room *Room) handleDidEndGame(result map[string]interface{}, delayUntilsNewActionSeconds int) {
	isDebuging := false
	if isDebuging {
		fmt.Println("handleDidEndGame cp0", room.gameCode)
	}

	method := fmt.Sprintf("%s_%s_finish_game_session", room.gameCode, room.currencyType)
	server.SendRequests(method, result, room.PlayersId())
	// delay (for client to display result etc) before room can start any action
	if delayUntilsNewActionSeconds != 0 {
		room.isDelayingForNewGame = true
		utils.Delay(delayUntilsNewActionSeconds)
		room.isDelayingForNewGame = false
	}

	// record last result
	//	if len(result) > 0 {
	//		resultsData := utils.GetMapSliceAtPath(result, "results")
	//		room.lastMatchResults = make(map[int64]*GameResult)
	//		for _, resultData := range resultsData {
	//			playerId := utils.GetInt64AtPath(resultData, "id")
	//			username := utils.GetStringAtPath(resultData, "username")
	//			result := &GameResult{
	//				result: utils.GetStringAtPath(resultData, "result"),
	//				change: utils.GetInt64AtPath(resultData, "change"),
	//				rank:   utils.GetIntAtPath(resultData, "rank"),
	//				bet:    utils.GetInt64AtPath(resultData, "bet"),
	//			}
	//			room.lastMatchResults[utils.GetInt64AtPath(resultData, "id")] = result
	//
	//			if playerId != 0 && len(username) > 0 {
	//				if result.change >= room.game.VipThreshold() && room.Game().CurrencyType() == currency.Money {
	//					congrat_queue.AddWinCongrat(username, room.gameCode, result.change)
	//				}
	//			}
	//		}
	//	}

	// remove player  if do not meet requirement
	room.removePlayersDidNotMeetRequirement(room.requirement)

	// below func do nothing
	//	room.rotateOwner()

	// remove register leave players
	room.removeRegisterLeavePlayers()

	// below func do nothing
	//	room.removePromoteRegisterOwner()

	if isDebuging {
		fmt.Println("handleDidEndGame cp2", room.gameCode)
	}

	// always true, all games dont have properties
	//if !GameHasProperties(room.game, []string{GamePropertyPersistentSession}) {
	room.session = nil
	//}

	// put money on table
	if GameHasProperty(room.game, GamePropertyBuyIn) {
		// keep the same money on table
	} else {
		for _, player := range room.players.copy() {
			if room.isPlayerOwner(player) {
				moneyToMove := room.game.MoneyOnTableForOwner(room.requirement, room.MaxNumberOfPlayers(), len(room.players.coreMap))
				if isDebuging {
					fmt.Println("handleDidEndGame cp31", room.gameCode)
				}
				player.FreezeMoney(moneyToMove, room.currencyType, room.GetRoomIdentifierString(), true)
				if isDebuging {
					fmt.Println("handleDidEndGame cp41", room.gameCode)
				}
			} else {
				moneyToMove := room.game.MoneyOnTable(room.requirement, room.MaxNumberOfPlayers(), len(room.players.coreMap))
				if isDebuging {
					fmt.Println("handleDidEndGame cp32", room.gameCode)
				}
				player.FreezeMoney(moneyToMove, room.currencyType, room.GetRoomIdentifierString(), true)
				if isDebuging {
					fmt.Println("handleDidEndGame cp42", room.gameCode)
				}
			}
		}
	}
	if isDebuging {
		fmt.Println("handleDidEndGame cp5", room.gameCode)
	}

	//
	if room.currencyType != currency.CustomMoney {
		room.updateAutoStartTimer()
	} else {
		room.Mutex.Lock()
		createrId, _ := room.SharedData["createrPid"].(int64)
		room.Mutex.Unlock()
		creater := room.getPlayer(createrId)
		if z.FindInt64InSlice(createrId, room.PlayersId()) == -1 {
			// dont start new match
			room.SendAMap("NotifyRoomStoped", map[string]interface{}{
				"msg": "Bàn chơi dừng hoạt động do người tạo phòng đã thoát.",
			})
			room.Mutex.Lock()
			room.SharedData["isStopped"] = true
			room.Mutex.Unlock()
		} else if creater != nil && creater.GetMoney(currency.Money) == 0 {
			// dont start new match
			room.SendAMap("NotifyRoomStoped", map[string]interface{}{
				"msg": "Bàn chơi dừng hoạt động do người tạo phòng hết tiền.",
			})
			room.Mutex.Lock()
			room.SharedData["isStopped"] = true
			room.Mutex.Unlock()
		} else {
			room.updateAutoStartTimer()
		}
	}

	//
	room.sendNotifyOwnerPingRoom()
	if isDebuging {
		fmt.Println("handleDidEndGame cp6", room.gameCode)
	}
}

func (room *Room) handleBuyIn(player GamePlayer, buyIn int64) (err error) {
	if !GameHasProperties(room.game, []string{GamePropertyBuyIn}) {
		return errors.New("err:not_implemented")
	}

	if room.IsPlaying() {
		if room.Session().GetPlayer(player.Id()) != nil {
			return errors.New("err:already_start_playing")
		}
		// player is not currently playing in the session can just buy in
	}

	betEntry := room.game.BetData().GetEntry(room.requirement)

	currentBuyIn := player.GetFreezeValueForReason(room.currencyType, room.GetRoomIdentifierString())
	totalBuyIn := currentBuyIn + buyIn
	if totalBuyIn < betEntry.Min() || totalBuyIn > betEntry.Max() || buyIn > player.GetAvailableMoney(room.currencyType) {
		return errors.New("err:invalid_buy_in")
	}

	player.IncreaseFreezeMoney(buyIn, room.currencyType, room.GetRoomIdentifierString(), true)
	return nil
}

func (room *Room) handleRegisterToBeOwner(player GamePlayer) (err error) {
	if !GameHasProperties(room.game, []string{GamePropertyRegisterOwner}) {
		return errors.New("err:not_implemented")
	}

	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	playerMoney := player.GetAvailableMoney(room.currencyType) + player.GetFreezeValueForReason(room.currencyType, room.GetRoomIdentifierString())
	err = room.game.IsPlayerMoneyValidToBecomeOwner(playerMoney,
		room.requirement,
		room.MaxNumberOfPlayers(),
		len(room.players.coreMap))
	if err != nil {
		return err
	}

	if room.IsPlaying() {
		if !room.isPlayerOwner(player) {
			if !ContainPlayer(room.registerToBeOwner, player) {
				room.registerToBeOwner = append(room.registerToBeOwner, player)
				room.sendNotifyOwnerStatusChange()
			}
		} else {
			room.willNotBeOwnerNextRound = false
			room.sendNotifyOwnerStatusChange()
		}
	} else {
		if room.Owner() != nil {
			if !room.isPlayerOwner(player) {
				if !ContainPlayer(room.registerToBeOwner, player) {
					room.registerToBeOwner = append(room.registerToBeOwner, player)
					room.sendNotifyOwnerStatusChange()
				}
			}
		} else {
			room.makeOwner(player)
			room.sendNotifyNewOwnerInRoom(player)
		}
	}

	return err
}

func (room *Room) handleUnregisterToBeOwner(player GamePlayer) (err error) {
	if !GameHasProperties(room.game, []string{GamePropertyRegisterOwner}) {
		return errors.New("err:not_implemented")
	}

	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	if room.isPlayerOwner(player) {
		if room.IsPlaying() {
			room.willNotBeOwnerNextRound = true
			room.sendNotifyOwnerStatusChange()
		} else {
			room.removeOwner()
			room.willNotBeOwnerNextRound = false
			room.filterOwnerList()

			if len(room.registerToBeOwner) > 0 {
				room.owner = room.registerToBeOwner[0]
				room.registerToBeOwner = RemovePlayer(room.registerToBeOwner, room.owner)
			}
			room.sendNotifyNewOwnerInRoom(room.owner)
			room.sendNotifyOwnerStatusChange()
		}
	} else {
		room.registerToBeOwner = RemovePlayer(room.registerToBeOwner, player)
		room.sendNotifyOwnerStatusChange()
	}
	return nil
}

func (room *Room) handleAssignOwner(player GamePlayer) error {
	if player == nil {
		return errors.New("err:input_player_is_nil")
	}
	if !GameHasProperties(room.game, []string{GamePropertyOwnerAssignByGame}) {
		return errors.New("err:not_implemented")
	}

	if !room.ContainsPlayer(player) {
		return errors.New("err:player_not_in_room")
	}

	room.makeOwner(player)
	room.sendNotifyNewOwnerInRoom(player)
	return nil
}

func (room *Room) handleRemoveOwner() (err error) {
	if !GameHasProperties(room.game, []string{GamePropertyOwnerAssignByGame}) {
		return errors.New("err:not_implemented")
	}

	room.removeOwner()
	room.sendNotifyNewOwnerInRoom(nil)
	return err
}

func (room *Room) handleAcceptToBeOwnerOtherRoom(player GamePlayer, newRoomId int64) (err error) {
	err = room.handleRegisterLeaveRoom(player)
	if err != nil {
		return err
	}

	if room.ContainsPlayer(player) {
		// still in room, just register leave
		room.registerJoinOwnerOtherRoom.Set(player.Id(), newRoomId)
		return nil
	} else {
		// already leave
		newRoom, err := JoinRoomById(room.game, player, newRoomId, "")
		if err != nil {
			return err
		}
		server.SendRequest("join_room_as_owner", newRoom.SerializedDataFull(player), player.Id())
		err = newRoom.RegisterToBeOwner(player)
		if err != nil {
			return err
		}
	}
	return nil
}

func (room *Room) Server() ServerInterface {
	return server
}

func (room *Room) GetListRegisterLeaveRoom() []GamePlayer {
	return room.registerLeaveRoom
}

// number of human players
func (room *Room) GetNumberOfHumans() int {
	numBot := 0
	numPlayer := 0
	for _, player := range room.Players().copy() {
		if player.PlayerType() == "bot" {
			numBot++
		} else {
			numPlayer++
		}
	}
	return numPlayer
}

// number of players, include bot
func (room *Room) GetNumberOfPlayers() int {
	return len(room.Players().copy())
}

//
type ByNOP []*Room

func (a ByNOP) Len() int      { return len(a) }
func (a ByNOP) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNOP) Less(i, j int) bool {
	// return true trong hàm nhỏ này ưu tiên đứng lên đầu
	if len(a[i].players.coreMap) == a[i].MaxNumberOfPlayers() {
		return false
	}
	if a[i].GetNumberOfHumans() > a[j].GetNumberOfHumans() {
		return true
	} else if a[i].GetNumberOfHumans() < a[j].GetNumberOfHumans() {
		return false
	} else {
		return a[i].GetNumberOfPlayers() > a[j].GetNumberOfPlayers()
	}
}

// sort rooms by (number of real players, number of bots)
func SortRoomsByNOPlayer(rooms []*Room) {
	sort.Sort(ByNOP(rooms))
}
