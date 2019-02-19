package maubinh

// session
import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/utils"
)

func init() {

}

type FinishOrganizeCardsAction struct {
	playerId  int64
	cardsData map[string]interface{}
}

type UploadCardsAction struct {
	playerId  int64
	cardsData map[string]interface{}
}

type MauBinhResult struct {
	playerId     int64
	username     string
	resultType   string
	whiteWinType string
	change       int64
	money        int64
	moneyOnTable int64
	result       string

	moneyGain int64
	moneyLost int64

	multiplier    float64
	aceMultiplier float64
	aceChange     int64

	organizedCardsData map[string][]string
	valid              bool
	cards              []string
	compareData        map[string]map[string]map[string]interface{}
	isCollapsing       bool
	winCollapseAll     bool
	compareData2       map[int64]map[string]int64
}

func (result *MauBinhResult) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = result.playerId
	data["username"] = result.username
	data["organized_cards_data"] = result.organizedCardsData
	data["is_valid"] = result.valid
	data["compare_data"] = result.compareData
	data["compare_data_2"] = result.compareData2
	data["result_type"] = result.resultType
	data["change"] = result.moneyGain - result.moneyLost
	data["money"] = result.money
	data["money_on_table"] = result.moneyOnTable
	data["white_win_type"] = result.whiteWinType
	data["ace_multiplier"] = result.aceMultiplier
	data["ace_change"] = result.aceChange
	data["is_collapsing"] = result.isCollapsing

	data["win_collapse_all"] = result.winCollapseAll
	if result.moneyGain-result.moneyLost > 0 {
		data["result"] = "win"
	} else if result.moneyGain-result.moneyLost < 0 {
		data["result"] = "lose"
	} else {
		data["result"] = "draw"
	}
	data["multiplier"] = result.multiplier + result.aceMultiplier
	data["cards"] = result.cards
	return data
}

type BetweenPlayersResult struct {
	player1Id int64
	player2Id int64
}

func hasCalculatedResultBetweenPlayers(results []*BetweenPlayersResult, player1Id int64, player2Id int64) bool {
	for _, result := range results {
		if (result.player1Id == player1Id && result.player2Id == player2Id) ||
			(result.player1Id == player2Id && result.player2Id == player1Id) {
			return true
		}
	}
	return false
}

func recordAlreadyCalculatedResultBetweenPlayers(results []*BetweenPlayersResult, player1Id int64, player2Id int64) []*BetweenPlayersResult {
	result := &BetweenPlayersResult{
		player1Id: player1Id,
		player2Id: player2Id,
	}
	return append(results, result)
}

type MauBinhSession struct {
	game                       *MauBinhGame
	currencyType               string
	owner                      game.GamePlayer
	playersData                []*PlayerData
	results                    []map[string]interface{}
	additionalResultsForRecord []map[string]interface{}
	resultObjects              []*MauBinhResult

	players []game.GamePlayer
	deck    *components.CardGameDeck

	cards                 map[int64][]string
	organizedCardsData    map[int64]map[string][]string
	finishOrganizingCards map[int64]bool
	whiteWin              map[int64]string
	matchId               string

	finished      bool
	goingToFinish bool // to send data back to client to prepare for display result

	playersGain map[int64]int64

	betEntry game.BetEntryInterface

	sessionStartDate time.Time
	startTurnDate    time.Time
	turnTime         time.Duration

	sessionCallback game.ActivityGameSessionCallback

	timeOutForTurn *utils.TimeOut

	// event chan
	finishOrganizeCardsChan     chan *FinishOrganizeCardsAction
	uploadCardsChan             chan *UploadCardsAction
	startOrganizeCardsAgainChan chan int64
	endGameChan                 chan bool
	startGameChan               chan bool
	forceEndChan                chan bool // make sure the main loop will be close

	// response from event chan
	finishOrganizeCardsResponseChan     chan error
	startOrganizeCardsAgainResponseChan chan error
	uploadCardsResponseChan             chan error

	// log to file
	logFile *log.LogObject

	// record
	playersIdWhenStart    []int64
	playersMoneyWhenStart map[int64]int64
	win                   int64
	lose                  int64
	botWin                int64
	botLose               int64
	tax                   int64
	totalBet              int64

	resultInCollapse        map[string]int64
	winCollapseAllPlayerId  int64
	loseCollapseAllPhayerId int64
	forceEndGame            bool

	mapPlayerIdToIsNotAfk map[int64]bool

	mutex sync.RWMutex
}

func NewMauBinhSession(game *MauBinhGame, currencyType string, owner game.GamePlayer, players []game.GamePlayer) *MauBinhSession {
	maubinhSession := &MauBinhSession{
		game:                       game,
		currencyType:               currencyType,
		owner:                      owner,
		players:                    players,
		playersIdWhenStart:         getIdFromPlayersMap(players),
		playersGain:                make(map[int64]int64),
		results:                    make([]map[string]interface{}, 0),
		additionalResultsForRecord: make([]map[string]interface{}, 0),
		organizedCardsData:         make(map[int64]map[string][]string),
		finishOrganizingCards:      make(map[int64]bool),
		whiteWin:                   make(map[int64]string),
		finished:                   false,
		goingToFinish:              false,
		sessionStartDate:           time.Now(),
		// chan
		finishOrganizeCardsChan:             make(chan *FinishOrganizeCardsAction),
		uploadCardsChan:                     make(chan *UploadCardsAction),
		startOrganizeCardsAgainChan:         make(chan int64),
		startGameChan:                       make(chan bool),
		endGameChan:                         make(chan bool),
		finishOrganizeCardsResponseChan:     make(chan error),
		startOrganizeCardsAgainResponseChan: make(chan error),
		uploadCardsResponseChan:             make(chan error),
		forceEndChan:                        make(chan bool),

		mapPlayerIdToIsNotAfk: map[int64]bool{},
	}
	for _, player := range players {
		maubinhSession.mutex.Lock()
		maubinhSession.finishOrganizingCards[player.Id()] = false
		maubinhSession.mutex.Unlock()
	}

	maubinhSession.logFile = log.NewLogObject(maubinhSession.getFileName())
	go maubinhSession.startEventLoop()
	return maubinhSession
}

func NewMauBinhSessionFromData(models game.ModelsInterface, game *MauBinhGame, sessionCallback game.ActivityGameSessionCallback, data map[string]interface{}) (session *MauBinhSession, err error) {
	session, err = loadSession(models, sessionCallback, game, data)
	session.logToFile("Load session from data %	", data)
	if err != nil {
		return nil, err
	}
	session.sessionCallback = sessionCallback
	// check if already start or already end
	if session.isEveryoneFinishOrganizedCards() {
		go session.sendEndGameEvent()
	} else {
		session.start()
	}

	return session, nil
}

func (gameSession *MauBinhSession) CleanUp() {
	gameSession.resultObjects = nil
	gameSession.results = nil
	gameSession.finishOrganizeCardsChan = nil
	gameSession.startOrganizeCardsAgainChan = nil
	gameSession.startGameChan = nil
	gameSession.endGameChan = nil
	gameSession.finishOrganizeCardsResponseChan = nil
	gameSession.uploadCardsResponseChan = nil
	gameSession.startOrganizeCardsAgainResponseChan = nil
	gameSession.forceEndChan = nil
}

func (gameSession *MauBinhSession) startEventLoop() {
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
		case <-gameSession.endGameChan:
			gameSession.handleEndGameEvent()
			return // end this
		case uploadCardsAction := <-gameSession.uploadCardsChan:
			err := gameSession.handleUploadCards(uploadCardsAction)
			gameSession.uploadCardsResponseChan <- err
		case finishOrganizeCardsAction := <-gameSession.finishOrganizeCardsChan:
			err := gameSession.handleFinishOrganizeCards(finishOrganizeCardsAction)
			gameSession.finishOrganizeCardsResponseChan <- err
		case playerId := <-gameSession.startOrganizeCardsAgainChan:
			err := gameSession.handleStartOrganizeCardsAgain(playerId)
			gameSession.startOrganizeCardsAgainResponseChan <- err
		case <-gameSession.forceEndChan:
			return // end this
		}
	}
}

func (gameSession *MauBinhSession) start() {
	gameSession.notifyStartGameSession()
	go gameSession.sendStartGameEvent()
}

// play logic

func (gameSession *MauBinhSession) HandlePlayerOffline(player game.GamePlayer) {

}
func (gameSession *MauBinhSession) HandlePlayerOnline(player game.GamePlayer) {

}
func (gameSession *MauBinhSession) HandlePlayerAddedToGame(player game.GamePlayer) {

}

func (gameSession *MauBinhSession) HandlePlayerRemovedFromGame(player game.GamePlayer) {

}
func (gameSession *MauBinhSession) IsDelayingForNewGame() bool {
	return false
}

func (gameSession *MauBinhSession) IsPlaying() bool {
	return !gameSession.finished
}

/*

game method

*/

func (gameSession *MauBinhSession) uploadCards(player game.GamePlayer, cardsData map[string]interface{}) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	gameSession.mutex.Lock()
	gameSession.mapPlayerIdToIsNotAfk[player.Id()] = true
	gameSession.mutex.Unlock()

	action := &UploadCardsAction{
		playerId:  player.Id(),
		cardsData: cardsData,
	}
	go gameSession.sendUploadCardsAction(action)
	err = <-gameSession.uploadCardsResponseChan
	return err
}

func (gameSession *MauBinhSession) finishOrganizedCards(player game.GamePlayer, cardsData map[string]interface{}) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	gameSession.mutex.Lock()
	gameSession.mapPlayerIdToIsNotAfk[player.Id()] = true
	gameSession.mutex.Unlock()

	action := &FinishOrganizeCardsAction{
		playerId:  player.Id(),
		cardsData: cardsData,
	}
	go gameSession.sendFinishOrganizeCardsAction(action)
	err = <-gameSession.finishOrganizeCardsResponseChan
	return err
}

func (gameSession *MauBinhSession) startOrganizeCardsAgain(player game.GamePlayer) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	gameSession.mutex.Lock()
	gameSession.mapPlayerIdToIsNotAfk[player.Id()] = true
	gameSession.mutex.Unlock()

	go gameSession.sendStartOrganizeCardsAgainAction(player.Id())
	err = <-gameSession.startOrganizeCardsAgainResponseChan
	return err
}

func (gameSession *MauBinhSession) notifyFinishGameSession(results []map[string]interface{}, delayUntilsNewGameSeconds int) {
	gameSession.sessionCallback.DidEndGame(gameSession.ResultSerializedData(), delayUntilsNewGameSeconds)
	gameSession.logToFile("finish game")
}

func (gameSession *MauBinhSession) notifyGameStateChange() {
	gameSession.sessionCallback.DidChangeGameState(gameSession)

}

func (gameSession *MauBinhSession) notifyStartGameSession() {
	gameSession.sessionCallback.DidStartGame(gameSession)
}

func (session *MauBinhSession) containPlayer(playerId int64) bool {
	for _, player := range session.players {
		if player.Id() == playerId {
			return true
		}
	}
	return false
}

func (session *MauBinhSession) getPlayerDataForPlayer(playerId int64) *PlayerData {
	for _, playerData := range session.playersData {
		if playerData.id == playerId {
			return playerData
		}
	}
	return nil
}

func (session *MauBinhSession) isEveryoneFinishOrganizedCards() bool {
	for _, finish := range session.finishOrganizingCards {
		if !finish {
			return false
		}
	}
	return true
}

func (session *MauBinhSession) isPlayerFinishOrganizedCards(playerId int64) bool {
	return session.finishOrganizingCards[playerId]
}

func (session *MauBinhSession) getCardsForPlayer(playerId int64) []string {
	return session.cards[playerId]
}

func (session *MauBinhSession) updateGameRecordForPlayer(player game.GamePlayer, result string, totalMoneyChange int64) {
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

// func (session *MauBinhSession) update

func (gameSession *MauBinhSession) updateDataForPlayerWithResult(player game.GamePlayer, result string, shouldBlock bool) (money int64, change int64) {
	bet := gameSession.betEntry.Min()
	// increase money
	if result == "draw" {
		change = 0
	} else {
		if bet != 0 {
			if result == "win" {
				change = game.MoneyAfterTax(bet, gameSession.betEntry)
				gameSession.sessionCallback.IncreaseMoney(player, change, shouldBlock)
			} else if result == "lose" {
				change = -bet
				gameSession.sessionCallback.DecreaseMoney(player, bet, shouldBlock)
			}
		} else {
			money = player.GetMoney(gameSession.currencyType)
			change = 0
		}
	}

	return money, change
}

// change in here should already include tax
func (gameSession *MauBinhSession) updateDataForPlayerWithChange(player game.GamePlayer, change int64, shouldBlock bool) (money int64) {
	if change != 0 {
		if change > 0 {
			gameSession.sessionCallback.IncreaseMoney(player, change, shouldBlock)
		} else {
			gameSession.sessionCallback.DecreaseMoney(player, -change, shouldBlock)
		}
	}

	money = player.GetMoney(gameSession.currencyType)
	return money
}

func (gameSession *MauBinhSession) calculateAces() {
	acesMap := make(map[int64]int)
	var maxAces int
	for _, player := range gameSession.players {
		acesNum := getNumberOfAcesInCards(gameSession.cards[player.Id()])
		if acesNum > maxAces {
			maxAces = acesNum
		}
		acesMap[player.Id()] = acesNum
	}

	if maxAces == 4 {
		return
	}

	if len(gameSession.players) == 4 {
		if maxAces == 3 {
			for playerId, acesNum := range acesMap {
				result := gameSession.resultObjectForPlayer(playerId)
				if acesNum == 3 {
					result.aceMultiplier = 8
				} else if acesNum == 1 {
					result.aceMultiplier = 0
				} else {
					result.aceMultiplier = -4
				}
			}
		} else if maxAces == 2 {
			var has1AcesCase bool
			for _, acesNum := range acesMap {
				if acesNum == 1 {
					has1AcesCase = true
					break
				}
			}

			if has1AcesCase {
				for playerId, acesNum := range acesMap {
					result := gameSession.resultObjectForPlayer(playerId)
					if acesNum == 2 {
						result.aceMultiplier = 4
					} else if acesNum == 1 {
						result.aceMultiplier = 0
					} else {
						result.aceMultiplier = -4
					}
				}
			} else {
				for playerId, acesNum := range acesMap {
					result := gameSession.resultObjectForPlayer(playerId)
					if acesNum == 2 {
						result.aceMultiplier = 4
					} else {
						result.aceMultiplier = -4
					}
				}
			}
		}
	} else {
		alreadyCalculatedResults := make([]*BetweenPlayersResult, 0)
		for playerId, acesNum := range acesMap {
			result := gameSession.resultObjectForPlayer(playerId)
			for otherPlayerId, otherAcesNum := range acesMap {
				if otherPlayerId != playerId {
					if hasCalculatedResultBetweenPlayers(alreadyCalculatedResults, playerId, otherPlayerId) {
						continue
					}
					alreadyCalculatedResults = recordAlreadyCalculatedResultBetweenPlayers(alreadyCalculatedResults, playerId, otherPlayerId)
					otherResult := gameSession.resultObjectForPlayer(otherPlayerId)

					offset := acesNum - otherAcesNum
					result.aceMultiplier += float64(offset) * 1
					otherResult.aceMultiplier += -float64(offset) * 1
				}
			}
		}
	}

	// calculate money
	// lost money
	var totalLost int64
	for _, player := range gameSession.players {
		result := gameSession.resultObjectForPlayer(player.Id())
		if result.aceMultiplier < 0 {
			absLost := -moneyAfterApplyMultiplier(gameSession.betEntry.Min(), result.aceMultiplier)
			player.LockMoney(gameSession.currencyType)
			if player.GetMoney(gameSession.currencyType) < absLost {
				absLost = player.GetMoney(gameSession.currencyType)
			}
			totalLost += absLost
			result.moneyLost += absLost
			result.aceChange = -absLost
			gameSession.updateDataForPlayerWithChange(player, -absLost, false)
			player.UnlockMoney(gameSession.currencyType)
			gameSession.recordMoneyForOnePlayer(-absLost, player, false)
		}
	}

	// gain
	var totalAceMultiplierWin float64
	for _, player := range gameSession.players {
		result := gameSession.resultObjectForPlayer(player.Id())
		if result.aceMultiplier > 0 {
			totalAceMultiplierWin += result.aceMultiplier
		}
	}

	for _, player := range gameSession.players {
		result := gameSession.resultObjectForPlayer(player.Id())
		if result.aceMultiplier > 0 {
			gain := int64(float64(totalLost) / totalAceMultiplierWin * result.aceMultiplier)
			moneyAfterTax := game.MoneyAfterTax(gain, gameSession.betEntry)
			result.moneyGain += moneyAfterTax
			result.aceChange = moneyAfterTax
			gameSession.updateDataForPlayerWithChange(player, moneyAfterTax, true)
			gameSession.recordMoneyForOnePlayer(gain, player, true)
		}
	}
}

func (gameSession *MauBinhSession) GetPlayer(playerId int64) (player game.GamePlayer) {
	for _, player := range gameSession.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (gameSession *MauBinhSession) GetPlayerIndex(playerId int64) int {
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
func (session *MauBinhSession) SerializedData() (data map[string]interface{}) {
	data = session.serializedDataForAll()
	// all the different data
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	playersCardsData := make(map[string][]string)
	for playerId, cards := range session.cards {
		playersCardsData[fmt.Sprintf("%d", playerId)] = cards
	}
	data["matchId"] = session.matchId
	data["players_id_when_start"] = session.playersIdWhenStart

	data["players_cards"] = playersCardsData
	data["session_start_date"] = session.sessionStartDate.Format(time.RFC3339Nano)
	temp := session.sessionStartDate.Add(
		session.game.turnTimeInSeconds).Sub(
		time.Now())
	data["remainingDuration"] = temp.Seconds()

	organizedCardsDataJsonData := make(map[string]interface{})
	for playerId, organizedCardsData := range session.organizedCardsData {
		organizedCardsDataJsonData[fmt.Sprintf("%d", playerId)] = organizedCardsData
	}
	data["organized_cards_data"] = organizedCardsDataJsonData

	finishOrganizedCardsJsonData := make(map[string]interface{})
	for playerId, finish := range session.finishOrganizingCards {
		finishOrganizedCardsJsonData[fmt.Sprintf("%d", playerId)] = finish
	}
	data["finish_organizing_cards"] = finishOrganizedCardsJsonData

	playersMoneyWhenStartData := make(map[string]int64)
	for playerId, money := range session.playersMoneyWhenStart {
		playersMoneyWhenStartData[fmt.Sprintf("%d", playerId)] = money
	}
	data["players_money_when_start"] = playersMoneyWhenStartData
	data["bet"] = session.betEntry.Min()
	data["additional_results"] = session.additionalResultsForRecord
	return data
}

func (session *MauBinhSession) serializedDataForRecord() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["players_id_when_start"] = session.playersIdWhenStart
	data["bet"] = session.betEntry.Min()
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
			resultsForRecord = append(resultsForRecord, validResult)
		}
	}
	data["results"] = resultsForRecord

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

func (session *MauBinhSession) ResultSerializedData() (data map[string]interface{}) {
	data = session.serializedDataForAll()
	data["results"] = session.results
	return data
}

func (session *MauBinhSession) SerializedDataForPlayer(player game.GamePlayer) (data map[string]interface{}) {
	data = session.serializedDataForAll()
	session.mutex.Lock()
	defer session.mutex.Unlock()
	data["matchId"] = session.matchId
	// different data for different player
	// since this is maubinh, only give the cards of the current player back
	if player != nil && session.cards[player.Id()] != nil {
		data["cards"] = session.cards[player.Id()]
		data["white_win_type"] = session.whiteWin[player.Id()]
	}

	if player != nil && session.organizedCardsData[player.Id()] != nil {
		temp := make(map[string][]string)
		for k, v := range session.organizedCardsData[player.Id()] {
			tempL1 := []string{}
			for _, e := range v {
				tempL1 = append(tempL1, e)
			}
			temp[k] = tempL1
		}
		data["cards_data"] = temp
	}

	if player != nil {
		data["finish_organizing_cards"] = session.finishOrganizingCards[player.Id()]
	}

	// if player != nil && player.PlayerType() == "bot" {
	// 	newPlayersData := make([]map[string]interface{}, 0)
	// 	for _, playerData := range session.playersData {
	// 		playerId := playerData.id
	// 		newPlayerData := make(map[string]interface{})
	// 		for key, value := range playerData {
	// 			newPlayerData[key] = value
	// 		}
	// 		newPlayerData["cards"] = session.cards[playerId]
	// 		newPlayerData["cards_data"] = session.organizedCardsData[player.Id()]
	// 		newPlayerData["finish_organizing_cards"] = session.finishOrganizingCards[player.Id()]
	// 		newPlayersData = append(newPlayersData, newPlayerData)
	// 	}
	// 	data["players_data"] = newPlayersData
	// }

	return data
}

func (session *MauBinhSession) serializedDataForAll() (data map[string]interface{}) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data = make(map[string]interface{})
	data["matchId"] = session.matchId
	data["game_code"] = session.game.gameCode
	data["owner_id"] = session.owner.Id()
	data["players_id"] = getIdFromPlayersMap(session.players)
	data["finished"] = session.finished
	data["gtf"] = session.goingToFinish
	//data["result_in_collapse"] = session.resultInCollapse
	data["win_collapse_all_player"] = session.winCollapseAllPlayerId
	data["lose_collaps_all_phayer"] = session.loseCollapseAllPhayerId
	data["force_end_game"] = session.forceEndGame

	playersDataRaw := make([]map[string]interface{}, 0)
	for _, playerData := range session.playersData {
		playerId := playerData.id
		playerDataRaw := playerData.SerializedData()

		playerDataRaw["money"] = session.GetPlayer(playerId).GetMoney(session.currencyType)

		playerDataRaw["finish_organizing_cards"] = session.isPlayerFinishOrganizedCards(playerId)

		if !session.startTurnDate.IsZero() && len(session.results) == 0 {
			turnTime := session.game.turnTimeInSeconds.Seconds() - time.Now().Sub(session.startTurnDate).Seconds()
			playerDataRaw["turn_time"], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", turnTime), 10)
		} else {
			playerDataRaw["turn_time"] = 0
		}
		playersDataRaw = append(playersDataRaw, playerDataRaw)
	}
	data["players_data"] = playersDataRaw

	if len(session.results) != 0 {
		data["results"] = session.results
	}

	playersGainData := make(map[string]int64)
	for playerId, playerGain := range session.playersGain {
		playersGainData[fmt.Sprintf("%d", playerId)] = playerGain
	}
	data["players_gain"] = playersGainData

	return data
}

func loadSession(models game.ModelsInterface, sessionCallback game.ActivityGameSessionCallback, gameInstance *MauBinhGame, data map[string]interface{}) (session *MauBinhSession, err error) {

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

func containPlayer(players []game.GamePlayer, player game.GamePlayer) bool {
	for _, playerInList := range players {
		if playerInList.Id() == player.Id() {
			return true
		}
	}
	return false
}

func (gameSession *MauBinhSession) isPlayerAlreadyHasResult(playerId int64) bool {
	for _, resultData := range gameSession.results {
		resultPlayerId := utils.GetInt64AtPath(resultData, "id")
		if resultPlayerId == playerId {
			return true
		}
	}
	return false
}

func (gameSession *MauBinhSession) resultOfPlayer(playerId int64) map[string]interface{} {
	for _, resultData := range gameSession.results {
		resultPlayerId := utils.GetInt64AtPath(resultData, "id")
		if resultPlayerId == playerId {
			return resultData
		}
	}
	return nil
}

func (gameSession *MauBinhSession) resultObjectForPlayer(playerId int64) *MauBinhResult {
	for _, result := range gameSession.resultObjects {
		if result.playerId == playerId {
			return result
		}
	}
	player := gameSession.GetPlayer(playerId)
	if player == nil {
		return nil
	}

	result := &MauBinhResult{
		playerId:           playerId,
		username:           player.Name(),
		cards:              gameSession.cards[playerId],
		organizedCardsData: gameSession.organizedCardsData[playerId],
		compareData:        make(map[string]map[string]map[string]interface{}),
	}
	// fmt.Println("result for player", playerId, result.organizedCardsData)
	defer func() {
		result = nil
	}()
	result.compareData[TopPart] = make(map[string]map[string]interface{})
	result.compareData[MiddlePart] = make(map[string]map[string]interface{})
	result.compareData[BottomPart] = make(map[string]map[string]interface{})

	gameSession.resultObjects = append(gameSession.resultObjects, result)
	return result
}

func moneyAfterApplyMultiplier(money int64, multiplier float64) int64 {
	return int64(utils.Round(float64(money) * multiplier))
}

func (gameSession *MauBinhSession) getFileName() string {
	if gameSession.sessionStartDate.IsZero() {
		gameSession.sessionStartDate = time.Now()
	}
	return fmt.Sprintf("maubinh_%s", gameSession.sessionStartDate.Format(time.RFC3339Nano))
}

func (gameSession *MauBinhSession) logToFile(data string, a ...interface{}) {
	gameSession.logFile.Log(data, a...)
}

func (gameSession *MauBinhSession) recordMoney(money int64, winPlayer game.GamePlayer, losePlayer game.GamePlayer) {
	// record
	gain := game.MoneyAfterTax(money, gameSession.betEntry)
	gameSession.tax += game.TaxFromMoney(money, gameSession.betEntry)
	if winPlayer.PlayerType() == "bot" {
		gameSession.botWin += gain
	} else {
		gameSession.win += gain
	}

	if losePlayer.PlayerType() == "bot" {
		gameSession.botLose += money
	} else {
		gameSession.lose += money
	}
}

func (gameSession *MauBinhSession) recordMoneyForOnePlayer(money int64, player game.GamePlayer, needAddTax bool) {
	// record
	var gain int64
	if needAddTax {
		gain = game.MoneyAfterTax(money, gameSession.betEntry)
		gameSession.tax += game.TaxFromMoney(money, gameSession.betEntry)
	} else {
		gain = money
	}

	if money > 0 {
		if player.PlayerType() == "bot" {
			gameSession.botWin += gain
		} else {
			gameSession.win += gain
		}
	} else if money < 0 {
		if player.PlayerType() == "bot" {
			gameSession.botLose += -money
		} else {
			gameSession.lose += -money
		}
	}
}
