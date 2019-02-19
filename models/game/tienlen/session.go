package tienlen

// session
import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/utils"
)

func init() {
	_ = z.Card{}
	_ = sync.Mutex{}
}

const (
	kStartPhase  = "start_phase"
	kChangePhase = "change_phase"
	kFinishPhase = "finish_phase"
)

type PlayCardsAction struct {
	playerId int64
	cards    []string
}
type TurnData struct {
	turnCounter  int
	roundCounter int
	playerId     int64
}

type TienLenResult struct {
	playerId       int64
	username       string
	displayName    string
	rank           int
	cards          []string
	loseType       string
	winType        string
	change         int64
	money          int64
	moneyOnTable   int64
	result         string
	cardTypesLeft  []string
	instantWinType string
}

func (result *TienLenResult) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = result.playerId
	data["username"] = result.username
	data["display_name"] = result.displayName
	data["rank"] = result.rank
	data["cards"] = result.cards
	data["lose_type"] = result.loseType
	data["win_type"] = result.winType
	data["change"] = result.change
	data["money"] = result.money
	data["money_on_table"] = result.moneyOnTable
	data["instant_win_type"] = result.instantWinType
	data["result"] = result.result
	data["card_types_left"] = result.cardTypesLeft
	return data
}

type TienLenSession struct {
	game                       *TienLenGame
	currencyType               string
	owner                      game.GamePlayer
	playersData                []*PlayerData
	results                    []map[string]interface{}
	additionalResultsForRecord []map[string]interface{}

	players []game.GamePlayer

	cards                 map[int64][]string
	cardsOnTable          []string
	allMovesOfCurrentTurn [][]string
	matchId               string

	finished bool

	playersGain map[int64]int64
	betEntry    game.BetEntryInterface

	sessionStartDate      time.Time
	startTurnDate         time.Time
	turnTime              time.Duration
	currentPlayerTurn     game.GamePlayer
	ownerOfCardsOnTable   game.GamePlayer
	turnCounter           int
	roundCounter          int
	playersInCurrentRound []game.GamePlayer

	lastMatchResult map[int64]*game.GameResult

	// mutex          *sync.Mutex
	// timeOutForTurn *utils.TimeOut

	sessionCallback game.ActivityGameSessionCallback

	timeOutForTurn *utils.TimeOut

	// event chan
	playCardsChan        chan *PlayCardsAction
	skipTurnChan         chan int64
	nextTurnChan         chan bool
	endTurnNaturallyChan chan *TurnData
	startGameChan        chan bool
	forceEndChan         chan bool // make sure the main loop will be close

	// response from event chan
	playCardsResponseChan chan error
	skipTurnResponseChan  chan error

	// log to file
	logFile *log.LogObject

	// record
	playersIdWhenStart []int64
	cardsWhenStartGame map[int64][]string

	playerLoseWinInTurn   map[int64]int64
	playersMoneyWhenStart map[int64]int64
	win                   int64
	lose                  int64
	botWin                int64
	botLose               int64
	tax                   int64
	totalBet              int64
	resultInTurn          []string
	//player_id_recent_win  int64
	//totalCardLeft         int

	//	mapPlayerIdToIsAfkLastTurn map[int64]bool

	mutex sync.RWMutex
}

func NewTienLenSession(game *TienLenGame, currencyType string, owner game.GamePlayer, players []game.GamePlayer) *TienLenSession {
	tienLenSession := &TienLenSession{
		game:                       game,
		currencyType:               currencyType,
		owner:                      owner,
		players:                    players,
		playersIdWhenStart:         getIdFromPlayersMap(players),
		playersGain:                make(map[int64]int64),
		results:                    make([]map[string]interface{}, 0),
		additionalResultsForRecord: make([]map[string]interface{}, 0),
		finished:                   false,
		sessionStartDate:           time.Now(),
		// chan
		playCardsChan:         make(chan *PlayCardsAction),
		skipTurnChan:          make(chan int64),
		nextTurnChan:          make(chan bool),
		endTurnNaturallyChan:  make(chan *TurnData),
		startGameChan:         make(chan bool),
		playCardsResponseChan: make(chan error),
		skipTurnResponseChan:  make(chan error),
		forceEndChan:          make(chan bool),

		//		mapPlayerIdToIsAfkLastTurn: map[int64]bool{},
	}
	tienLenSession.logFile = log.NewLogObject(tienLenSession.getFileName())
	go tienLenSession.startEventLoop()
	return tienLenSession
}

func (gameSession *TienLenSession) CleanUp() {
	gameSession.playersGain = nil
	gameSession.playCardsChan = nil
	gameSession.skipTurnChan = nil
	gameSession.nextTurnChan = nil
	gameSession.endTurnNaturallyChan = nil
	gameSession.startGameChan = nil
	gameSession.playCardsResponseChan = nil
	gameSession.skipTurnChan = nil
	gameSession.skipTurnResponseChan = nil
	gameSession.forceEndChan = nil
}

func (gameSession *TienLenSession) startEventLoop() {
	defer func() {
		gameSession.CleanUp()
		if r := recover(); r != nil {
			gameSession.logToFile("recovered in f %v", r)
			gameSession.logFile.Save()
			log.SendMailWithCurrentStack(gameSession.logFile.Content())
		}
	}()
	for {
		select {
		case <-gameSession.startGameChan:
			gameSession.handleStartGameEvent()
		case <-gameSession.nextTurnChan:
			gameSession.handleNextTurnEvent()
		case turnData := <-gameSession.endTurnNaturallyChan:
			gameSession.handleEndTurnNaturallyEvent(turnData)
		case playCardsAction := <-gameSession.playCardsChan:
			err := gameSession.handlePlayCardsEvent(playCardsAction)
			gameSession.playCardsResponseChan <- err
		case playerId := <-gameSession.skipTurnChan:
			err := gameSession.handleSkipTurnEvent(playerId)
			gameSession.skipTurnResponseChan <- err
		case <-gameSession.forceEndChan:
			return // force end, did nothing
		}
	}
}

func (gameSession *TienLenSession) start() {
	gameSession.notifyStartGameSession()
	go gameSession.sendStartGameEvent()
}

/*
play logic
*/

func (gameSession *TienLenSession) calculateNextTurnMove(forcedPlayerToGoNext game.GamePlayer) {
	// if the round only have 1 player left
	if len(gameSession.playersInCurrentRound) == 1 {
		lastPlayer := gameSession.playersInCurrentRound[0]
		ownerOfCardsOntable := gameSession.ownerOfCardsOnTable
		if ownerOfCardsOntable != nil && lastPlayer.Id() == ownerOfCardsOntable.Id() {
			// the lastPlayer win this round
			gameSession.logToFile("player %d win this round", gameSession.ownerOfCardsOnTable.Id())
			gameSession.createNewRoundData(lastPlayer.Id())
		} else {
			// the owner of cards on table already win with result in this round
			// the last player need to play to gain first turn next round or can skip
			gameSession.advanceTurnCounter(forcedPlayerToGoNext)
		}
	} else if len(gameSession.playersInCurrentRound) == 0 {
		// 1 of the player has win, all the other skip,
		winnerOfRound := gameSession.ownerOfCardsOnTable
		startingPlayer := gameSession.getNextPlayerInLine(winnerOfRound)
		gameSession.createNewRoundData(startingPlayer.Id())
	} else {
		// next round normally
		gameSession.advanceTurnCounter(forcedPlayerToGoNext)
	}
}

func (gameSession *TienLenSession) HandlePlayerOffline(player game.GamePlayer) {

}
func (gameSession *TienLenSession) HandlePlayerOnline(player game.GamePlayer) {

}
func (gameSession *TienLenSession) HandlePlayerAddedToGame(player game.GamePlayer) {

}
func (gameSession *TienLenSession) HandlePlayerRemovedFromGame(player game.GamePlayer) {

}
func (gameSession *TienLenSession) IsPlaying() bool {
	return !gameSession.finished
}

func (gameSession *TienLenSession) IsDelayingForNewGame() bool {
	return false
}

/*

game method

*/

func (gameSession *TienLenSession) skipTurn(player game.GamePlayer) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	//	gameSession.mutex.Lock()
	//	gameSession.mapPlayerIdToIsAfkLastTurn[player.Id()] = false
	//	gameSession.mutex.Unlock()

	go gameSession.sendSkipTurnEvent(player.Id())
	err = <-gameSession.skipTurnResponseChan
	return err
}

func (gameSession *TienLenSession) playCards(player game.GamePlayer, cards []string) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	//	gameSession.mutex.Lock()
	//	gameSession.mapPlayerIdToIsAfkLastTurn[player.Id()] = false
	//	gameSession.mutex.Unlock()

	playCardsAction := &PlayCardsAction{
		playerId: player.Id(),
		cards:    cards,
	}
	go gameSession.sendPlayCardsEvent(playCardsAction)
	err = <-gameSession.playCardsResponseChan
	return err
}

func (gameSession *TienLenSession) notifyFinishGameSession(results []map[string]interface{}) {
	gameSession.sessionCallback.DidEndGame(gameSession.ResultSerializedData(), gameSession.game.delayAfterEachGameInSeconds)
	gameSession.logToFile("finish game")
}

func (gameSession *TienLenSession) notifyGameStateChange() {
	gameSession.sessionCallback.DidChangeGameState(gameSession)

}

func (gameSession *TienLenSession) notifyStartGameSession() {
	gameSession.sessionCallback.DidStartGame(gameSession)
}

func (session *TienLenSession) containPlayer(playerId int64) bool {
	for _, player := range session.players {
		if player.Id() == playerId {
			return true
		}
	}
	return false
}

func (session *TienLenSession) getPlayerDataForPlayer(playerId int64) *PlayerData {
	for _, playerData := range session.playersData {
		if playerData.id == playerId {
			return playerData
		}
	}
	return nil
}

func (session *TienLenSession) updateGameRecordForPlayer(player game.GamePlayer, result string, totalMoneyChange int64) {
	// increase exp
	if totalMoneyChange > 0 {
		player.IncreaseExp(totalMoneyChange)
	} else {
		if result != "quit" {
			player.IncreaseExp(10)
		}
	}
	// record result
	player.RecordGameResult(session.game.gameCode, result, totalMoneyChange, session.currencyType)
}

func (gameSession *TienLenSession) GetPlayer(playerId int64) (player game.GamePlayer) {
	for _, player := range gameSession.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (gameSession *TienLenSession) GetPlayerIndex(playerId int64) int {
	counter := 0
	for _, player := range gameSession.players {
		if player.Id() == playerId {
			return counter
		}
		counter++
	}
	return -1
}

// save load
func (session *TienLenSession) SerializedData() (data map[string]interface{}) {
	data = session.serializedDataForAll()
	// all the different data
	playersCardsleft := make(map[string]int)
	playersCardsData := make(map[string][]string)

	for playerId, cards := range session.cards {
		splayerId := fmt.Sprintf("%d", playerId)
		playersCardsData[splayerId] = cards
		playersCardsleft[splayerId] = len(cards)
	}
	data["matchId"] = session.matchId
	// data["players_cards"] = playersCardsData

	playersCardsData = make(map[string][]string)
	for playerId, cards := range session.cardsWhenStartGame {
		playersCardsData[fmt.Sprintf("%d", playerId)] = cards
	}
	data["players_cards_when_start"] = playersCardsData
	data["players_id_when_start"] = session.playersIdWhenStart

	playersMoneyWhenStartData := make(map[string]int64)
	for playerId, money := range session.playersMoneyWhenStart {
		playersMoneyWhenStartData[fmt.Sprintf("%d", playerId)] = money
	}
	data["players_money_when_start"] = playersMoneyWhenStartData
	data["additional_results"] = session.additionalResultsForRecord
	data["result_in_turn"] = session.resultInTurn
	data["total_card_left"] = playersCardsleft
	data["session_start_date"] = session.sessionStartDate.Format(time.RFC3339Nano)
	return data
}

func (session *TienLenSession) serializedDataForRecord() (data map[string]interface{}) {
	data = make(map[string]interface{})
	playersCardsData := make(map[string][]string)
	for playerId, cards := range session.cardsWhenStartGame {
		playersCardsData[fmt.Sprintf("%d", playerId)] = cards
	}
	data["players_cards_when_start"] = playersCardsData
	data["matchId"] = session.matchId
	resultsForRecord := make([]map[string]interface{}, 0)
	for _, playerId := range session.playersIdWhenStart {
		var validResult map[string]interface{}
		for _, result := range session.results {
			if utils.GetInt64AtPath(result, "id") == playerId {
				validResult = result
				break
			}
		}

		if len(validResult) == 0 {
			for _, result := range session.additionalResultsForRecord {
				if utils.GetInt64AtPath(result, "id") == playerId {
					validResult = result
					break
				}
			}
		}
		if len(validResult) != 0 {
			validResult["money_on_table"] = session.sessionCallback.GetMoneyOnTable(playerId)
			resultsForRecord = append(resultsForRecord, validResult)
		}
	}
	data["results"] = resultsForRecord

	data["players_id_when_start"] = session.playersIdWhenStart

	playersMoneyWhenStartData := make(map[string]int64)
	for playerId, money := range session.playersMoneyWhenStart {
		playersMoneyWhenStartData[fmt.Sprintf("%d", playerId)] = money
	}
	data["players_money_when_start"] = playersMoneyWhenStartData

	var normalCount, botCount int
	playersIpWhenStart := make(map[string]string)
	for _, player := range session.players {
		if player.PlayerType() != "bot" {
			normalCount++
		} else {
			botCount++
		}
		playersIpWhenStart[fmt.Sprintf("%d", player.Id())] = player.IpAddress()
	}
	data["players_ip_when_start"] = playersIpWhenStart
	data["normal_count"] = normalCount
	data["bot_count"] = botCount
	return data
}

func (session *TienLenSession) ResultSerializedData() (data map[string]interface{}) {
	data = session.SerializedData()

	playersCardsData := make(map[string][]string)
	for playerId, cards := range session.cards {
		splayerId := fmt.Sprintf("%d", playerId)
		playersCardsData[splayerId] = cards
	}
	data["players_cards"] = playersCardsData

	data["results"] = session.results
	return data
}

func (session *TienLenSession) SerializedDataForPlayer(player game.GamePlayer) (data map[string]interface{}) {
	data = session.serializedDataForAll()
	data["matchId"] = session.matchId
	// different data for different player
	// since this is tienlen, only give the cards of the current player back
	if player != nil && session.cards[player.Id()] != nil {
		data["cards"] = session.cards[player.Id()]
	}

	if player != nil && player.PlayerType() == "bot" {
		newPlayersData := make([]map[string]interface{}, 0)
		for _, playerData := range session.playersData {
			playerId := playerData.id
			newPlayerData := playerData.SerializedData()
			newPlayerData["money"] = session.GetPlayer(playerId).GetMoney(session.currencyType)
			newPlayerData["money_on_table"] = session.sessionCallback.GetMoneyOnTable(playerId)
			newPlayerData["cards"] = session.cards[playerId]
			newPlayersData = append(newPlayersData, newPlayerData)
		}
		data["players_data"] = newPlayersData
	}

	if player != nil && player.PlayerType() == "bot" {
		session.mutex.Lock()
		//
		IsFirstTurnInMatch := true
		MapPlayerToHand := make(map[int64][]string)
		for pid, cards := range session.cards {
			temp := make([]string, len(cards))
			copy(temp, cards)
			MapPlayerToHand[pid] = temp
			IsFirstTurnInMatch = IsFirstTurnInMatch && (len(temp) == 13)
		}
		data["MapPlayerToHand"] = MapPlayerToHand
		//
		Order := make([]int64, 0)
		for _, e := range session.playersData {
			Order = append(Order, e.id)
		}
		data["Order"] = Order
		//
		if session.currentPlayerTurn != nil {
			data["CurrentTurnPlayer"] = session.currentPlayerTurn.Id()
			PlayersInRound := make([]int64, 0)
			temp := getIdFromPlayersMap(session.playersInCurrentRound)
			for _, pid := range Order {
				if z.FindInt64InSlice(pid, temp) != -1 {
					PlayersInRound = append(PlayersInRound, pid)
				}
			}
			data["PlayersInRound"] = PlayersInRound
		}
		//
		CurrentComboOnBoard := make([]string, len(session.cardsOnTable))
		copy(CurrentComboOnBoard, session.cardsOnTable)
		data["CurrentComboOnBoard"] = CurrentComboOnBoard
		//
		data["IsFirstTurnInMatch"] = IsFirstTurnInMatch
		//
		data["IsFirstTurnInRound"] = len(CurrentComboOnBoard) == 0
		//
		session.mutex.Unlock()
	}
	return data
}

func (session *TienLenSession) serializedDataForAll() (data map[string]interface{}) {
	session.mutex.RLock()
	data = make(map[string]interface{})
	data["matchId"] = session.matchId
	data["game_code"] = session.game.gameCode
	data["owner_id"] = session.owner.Id()
	data["players_id"] = getIdFromPlayersMap(session.players)

	playersCardsleft := make(map[int64]int)
	playersDataRaw := make([]map[string]interface{}, 0)
	for _, playerData := range session.playersData {
		playerId := playerData.id

		playersCardsleft[playerId] = len(session.cards[playerId])

		playerDataRaw := playerData.SerializedData()
		playerDataRaw["money"] = session.GetPlayer(playerId).GetMoney(session.currencyType)
		playerDataRaw["money_on_table"] = session.sessionCallback.GetMoneyOnTable(playerId)

		playersDataRaw = append(playersDataRaw, playerDataRaw)
	}
	data["total_card_left"] = playersCardsleft
	data["players_data"] = playersDataRaw

	if session.currentPlayerTurn != nil {
		data["player_ids_in_round"] = getIdFromPlayersMap(session.playersInCurrentRound)
		data["current_player_id_turn"] = session.currentPlayerTurn.Id()
		data["turn_time"], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", session.game.turnTimeInSeconds.Seconds()-time.Now().Sub(session.startTurnDate).Seconds()), 10)
	}
	data["turn_counter"] = session.turnCounter
	data["round_counter"] = session.roundCounter

	if len(session.results) != 0 {
		data["results"] = session.results
	}

	if len(session.cardsOnTable) != 0 {
		data["cards_on_table"] = session.cardsOnTable

		if session.ownerOfCardsOnTable != nil {
			data["owner_id_of_cards_on_table"] = session.ownerOfCardsOnTable.Id()
		}
	}

	if len(session.allMovesOfCurrentTurn) != 0 {
		data["all_moves_of_current_turn"] = session.allMovesOfCurrentTurn
	}

	playersGainData := make(map[string]int64)
	for playerId, playerGain := range session.playersGain {
		playersGainData[fmt.Sprintf("%d", playerId)] = playerGain
	}
	data["players_gain"] = playersGainData

	data["bet"] = session.betEntry.Min()

	// anh 3 them o cho nay
	data["result_in_turn"] = session.resultInTurn
	session.mutex.RUnlock()

	return data
}

func loadSession(models game.ModelsInterface, sessionCallback game.ActivityGameSessionCallback, gameInstance *TienLenGame, data map[string]interface{}) (session *TienLenSession, err error) {

	return nil, nil
}

// helper

func getIdFromPlayersMap(players []game.GamePlayer) []int64 {
	keys := make([]int64, 0, len(players))
	for _, player := range players {
		keys = append(keys, player.Id())
	}
	return keys
}

func (gameSession *TienLenSession) isPlayerInCurrentRound(playerId int64) bool {
	for _, player := range gameSession.playersInCurrentRound {
		if player.Id() == playerId {
			return true
		}
	}
	return false
}

// func (session *TienLenSession) getPlayerFromTurnCounter(turnCounter int) game.GamePlayer {
// 	playerIndexInSlice := (session.initialTurnStartAtIndex + turnCounter) % len(session.players)
// 	return session.players[playerIndexInSlice]
// }
func (gameSession *TienLenSession) advanceTurnCounter(forcedPlayerToGoNext game.GamePlayer) {
	gameSession.turnCounter++
	gameSession.mutex.Lock()
	if forcedPlayerToGoNext != nil {
		// check if the forcedPlayerToGoNext is still valid (still in the current round)
		// (since he can "freeze lose already")

		if gameSession.isPlayerInCurrentRound(forcedPlayerToGoNext.Id()) {
			gameSession.currentPlayerTurn = forcedPlayerToGoNext
		} else {
			nextPlayer := gameSession.getNextPlayerInRoundFromThisPlayer(forcedPlayerToGoNext)
			gameSession.currentPlayerTurn = nextPlayer
		}
	} else {
		gameSession.currentPlayerTurn = gameSession.getNextPlayerInRound()
	}
	gameSession.mutex.Unlock()
}

func (gameSession *TienLenSession) getPreviousPlayer() game.GamePlayer {
	for index, player := range gameSession.playersInCurrentRound {
		if player.Id() == gameSession.currentPlayerTurn.Id() {
			return gameSession.playersInCurrentRound[utils.MaxInt(0, index-1)]
		}
	}
	return nil
}

func (gameSession *TienLenSession) getNextPlayerInRound() game.GamePlayer {
	for index, player := range gameSession.playersInCurrentRound {
		if player.Id() == gameSession.currentPlayerTurn.Id() {
			return gameSession.playersInCurrentRound[(index+1)%len(gameSession.playersInCurrentRound)]
		}
	}
	return nil
}

func (gameSession *TienLenSession) getNextPlayerInRoundFromThisPlayer(startingPlayer game.GamePlayer) game.GamePlayer {
	index := gameSession.GetPlayerIndex(startingPlayer.Id())
	for true {
		nextPlayer := gameSession.players[(index+1)%len(gameSession.players)]
		if nextPlayer.Id() == startingPlayer.Id() {
			return nil
		}
		if len(gameSession.cards[nextPlayer.Id()]) != 0 && gameSession.isPlayerInCurrentRound(nextPlayer.Id()) { // not win yet and still in round
			return nextPlayer
		}
		index++
	}
	return nil
}

func (gameSession *TienLenSession) getNextPlayerInLine(startingPlayer game.GamePlayer) game.GamePlayer {
	index := gameSession.GetPlayerIndex(startingPlayer.Id())
	for true {
		nextPlayer := gameSession.players[(index+1)%len(gameSession.players)]
		if nextPlayer.Id() == startingPlayer.Id() {
			return nil
		}
		if len(gameSession.cards[nextPlayer.Id()]) != 0 { // not win yet
			return nextPlayer
		}
		index++
	}
	return nil
}

func (gameSession *TienLenSession) getNextRank() int {
	maxRank := -1
	for _, resultData := range gameSession.results {
		rank := utils.GetIntAtPath(resultData, "rank")
		if rank > maxRank {
			maxRank = rank
		}
	}
	return maxRank + 1
}

func (gameSession *TienLenSession) createNewRoundData(startingPlayerId int64) {
	gameSession.playersInCurrentRound = make([]game.GamePlayer, 0)
	startingIndex := gameSession.GetPlayerIndex(startingPlayerId)
	for i := 0; i < len(gameSession.players); i++ {
		playerIndex := (startingIndex + i) % len(gameSession.players)
		player := gameSession.players[playerIndex]
		if len(gameSession.cards[player.Id()]) > 0 && !gameSession.isPlayerAlreadyHasResult(player.Id()) {
			gameSession.playersInCurrentRound = append(gameSession.playersInCurrentRound, player)
		}
	}
	gameSession.mutex.Lock()
	gameSession.currentPlayerTurn = gameSession.playersInCurrentRound[0]
	gameSession.mutex.Unlock()
	gameSession.turnCounter = 0
	gameSession.roundCounter++
	gameSession.allMovesOfCurrentTurn = make([][]string, 0)
	gameSession.cardsOnTable = make([]string, 0)
	gameSession.ownerOfCardsOnTable = nil

}

func (gameSession *TienLenSession) removePlayerFromRound(playerToRemove game.GamePlayer) {
	playersInCurrentRound := make([]game.GamePlayer, 0)
	for _, player := range gameSession.playersInCurrentRound {
		if player.Id() != playerToRemove.Id() {
			playersInCurrentRound = append(playersInCurrentRound, player)
		}
	}
	gameSession.playersInCurrentRound = playersInCurrentRound
}

func (gameSession *TienLenSession) getLastPlayerWhoDidNotHaveResult() (lastPlayer game.GamePlayer) {
	for _, player := range gameSession.players {
		var found bool
		for _, resultData := range gameSession.results {
			resultPlayerId := utils.GetInt64AtPath(resultData, "id")
			if resultPlayerId == player.Id() {
				found = true
				break
			}
		}
		if !found {
			lastPlayer = player
		}
	}
	return lastPlayer
}

func (gameSession *TienLenSession) isPlayerAlreadyHasResult(playerId int64) bool {
	for _, resultData := range gameSession.results {
		resultPlayerId := utils.GetInt64AtPath(resultData, "id")
		if resultPlayerId == playerId {
			return true
		}
	}
	return false
}

func (gameSession *TienLenSession) resultOfPlayer(playerId int64) map[string]interface{} {
	for _, resultData := range gameSession.results {
		resultPlayerId := utils.GetInt64AtPath(resultData, "id")
		if resultPlayerId == playerId {
			return resultData
		}
	}
	return nil
}

func moneyAfterApplyMultiplier(money int64, multiplier float64) int64 {
	return int64(utils.Round(float64(money) * multiplier))
}

func (gameSession *TienLenSession) getFileName() string {
	if gameSession.sessionStartDate.IsZero() {
		gameSession.sessionStartDate = time.Now()
	}
	return fmt.Sprintf("tienlen_%s", gameSession.sessionStartDate.Format(time.RFC3339Nano))
}

func (gameSession *TienLenSession) logToFile(data string, a ...interface{}) {
	// logString := gameSession.logFile.Log(data, a...)
	// fmt.Println(logString)
}
