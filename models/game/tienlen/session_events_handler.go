package tienlen

import (
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	//"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/game/tienlen/logic"
	//	"github.com/vic/vic_go/models/player"
	//"github.com/vic/vic_go/models/try"
	"github.com/vic/vic_go/language"
	sc "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

var mapMoneyUnitToJackpotRatio map[int64]float64

func init() {
	mapMoneyUnitToJackpotRatio = map[int64]float64{
		100: 0.005, 200: 0.005, 500: 0.005,
		1000: 0.01, 2000: 0.01, 5000: 0.01,
		10000: 0.025, 20000: 0.025, 50000: 0.025,
		100000: 0.05, 200000: 0.05, 500000: 0.05,
	}
}

func (gameSession *TienLenSession) sendStartGameEvent() {
	gameSession.startGameChan <- true
}

func (gameSession *TienLenSession) sendNextTurnEvent() {
	gameSession.nextTurnChan <- true
}

func (gameSession *TienLenSession) sendEndTurnNaturallyEvent(turnData *TurnData) {
	gameSession.endTurnNaturallyChan <- turnData
}

func (gameSession *TienLenSession) sendPlayCardsEvent(playCardsAction *PlayCardsAction) {
	gameSession.playCardsChan <- playCardsAction
}

func (gameSession *TienLenSession) sendSkipTurnEvent(playerId int64) {
	gameSession.skipTurnChan <- playerId
}

func (gameSession *TienLenSession) sendForceEndGameEvent() {
	gameSession.forceEndChan <- true
}

func (gameSession *TienLenSession) handleStartGameEvent() {
	if gameSession.finished {
		return
	}
	gameSession.tax = 0
	gameSession.logToFile("start game with players %v cards %v", getIdFromPlayersMap(gameSession.players), gameSession.cards)
	// calculate who go first
	gameSession.matchId = fmt.Sprintf("#%v", int32(time.Now().Unix()))
	var firstTurnPlayerId int64
	gameSession.playerLoseWinInTurn = map[int64]int64{}

	minCard := ""
	var currentPlayerIdTurn int64
	for playerId, cards := range gameSession.cards {
		// first card is smallest
		if minCard == "" {
			minCard = cards[0]
			currentPlayerIdTurn = playerId
		} else {
			if logic.IsCard1BiggerThanCard2(gameSession.game.logicInstance, minCard, cards[0]) {
				minCard = cards[0]
				currentPlayerIdTurn = playerId
			}
		}
	}
	firstTurnPlayerId = currentPlayerIdTurn

	gameSession.createNewRoundData(firstTurnPlayerId)
	gameSession.turnCounter = 0
	gameSession.logToFile("first turn player id %d", firstTurnPlayerId)

	if gameSession.game.logicInstance.HasInstantWin() {
		// calculate instant win
		for _, player := range gameSession.playersInCurrentRound {
			instantType := gameSession.game.logicInstance.GetInstantWinType(gameSession.cards[player.Id()])
			if len(instantType) != 0 {
				gameSession.handleInstantWin(player, instantType)
				return
			}
		}
	}

	// start first turn
	go gameSession.sendNextTurnEvent()
}

func (gameSession *TienLenSession) handleInstantWin(winner game.GamePlayer, instantWinType string) {
	if gameSession.finished {
		return
	}
	if !gameSession.containPlayer(winner.Id()) {
		return
	}

	gameSession.logToFile("instant win: %d", winner.Id())
	gameSession.logToFile("cards %v", gameSession.cards)

	result := &TienLenResult{}
	defer func() {
		result = nil
	}()
	result.playerId = winner.Id()
	result.username = winner.Name()
	result.result = "win"
	result.cards = gameSession.cards[winner.Id()]
	result.winType = "instant"
	result.instantWinType = instantWinType
	gameSession.handleReceiveResultEvent(result, true)
}

func (gameSession *TienLenSession) handleReceiveResultEvent(winnerResult *TienLenResult, isInstantWin bool) {

	var totalGain int64
	winner := gameSession.GetPlayer(winnerResult.playerId)
	for _, player := range gameSession.players {
		if player.Id() != winner.Id() {
			cards := gameSession.cards[player.Id()]

			// instant win will not count stuffs
			var shouldCountStuffs bool
			if isInstantWin {
				shouldCountStuffs = false
			} else {
				shouldCountStuffs = true
			}

			loseMultiplier, cardTypes := gameSession.game.logicInstance.LoseMultiplierByCardLeft(cards, shouldCountStuffs)
			moneyLost := moneyAfterApplyMultiplier(gameSession.betEntry.Min(), loseMultiplier)

			player.LockMoney(gameSession.currencyType)
			winner.LockMoney(gameSession.currencyType)

			moneyActuallyLost := utils.MinInt64(moneyLost, gameSession.sessionCallback.GetMoneyOnTable(player.Id()))
			moneyGain := game.MoneyAfterTax(moneyActuallyLost, gameSession.betEntry)
			//var  float64
			var playerLoseInturn, playerWinInturn, winnerLoseInturn, winnerWinInturn, p1 int64
			if _, ok := gameSession.playerLoseWinInTurn[player.Id()]; ok {
				playerLoseInturn = 0
				playerWinInturn = 0
				if gameSession.playerLoseWinInTurn[player.Id()] < 0 {
					p1 = moneyAfterApplyMultiplier(gameSession.betEntry.Min(), float64(-1.0*gameSession.playerLoseWinInTurn[player.Id()]))
					playerLoseInturn = utils.MinInt64(p1, gameSession.sessionCallback.GetMoneyOnTable(player.Id()))
					gameSession.playerLoseWinInTurn[player.Id()] = 0
				} else if gameSession.playerLoseWinInTurn[player.Id()] > 0 {
					p1 = moneyAfterApplyMultiplier(gameSession.betEntry.Min(), float64(gameSession.playerLoseWinInTurn[player.Id()]))
					m1 := utils.MinInt64(int64(p1), gameSession.sessionCallback.GetMoneyOnTable(player.Id()))
					playerWinInturn = game.MoneyAfterTax(m1, gameSession.betEntry)
					gameSession.playerLoseWinInTurn[player.Id()] = 0
					gameSession.tax += m1 - playerWinInturn
				}
				delete(gameSession.playerLoseWinInTurn, player.Id())
				//do something here
			}

			totalPlayerChange := playerWinInturn - moneyActuallyLost - playerLoseInturn

			if totalPlayerChange > 0 {
				gameSession.sessionCallback.IncreaseMoney(player,
					totalPlayerChange, false)
			}
			if totalPlayerChange < 0 {

				gameSession.sessionCallback.DecreaseMoney(player,
					int64(-1)*totalPlayerChange, false)

			}

			if _, ok := gameSession.playerLoseWinInTurn[winner.Id()]; ok {
				winnerLoseInturn = 0
				winnerWinInturn = 0
				if gameSession.playerLoseWinInTurn[winner.Id()] < 0 {
					p1 = moneyAfterApplyMultiplier(gameSession.betEntry.Min(), float64(-1.0*gameSession.playerLoseWinInTurn[winner.Id()]))
					winnerLoseInturn = utils.MinInt64(p1, gameSession.sessionCallback.GetMoneyOnTable(winner.Id()))
					gameSession.playerLoseWinInTurn[winner.Id()] = 0
				} else if gameSession.playerLoseWinInTurn[winner.Id()] > 0 {
					p1 = moneyAfterApplyMultiplier(gameSession.betEntry.Min(), float64(gameSession.playerLoseWinInTurn[winner.Id()]))
					m := utils.MinInt64(p1, gameSession.sessionCallback.GetMoneyOnTable(winner.Id()))
					winnerWinInturn = game.MoneyAfterTax(m, gameSession.betEntry)
					gameSession.playerLoseWinInTurn[winner.Id()] = 0
					gameSession.tax += m - winnerWinInturn
				}
				delete(gameSession.playerLoseWinInTurn, winner.Id())
				//do something here
			}
			totalWinnerChange := moneyGain + winnerWinInturn - winnerLoseInturn

			if totalWinnerChange > 0 {
				gameSession.sessionCallback.IncreaseMoney(winner,
					totalWinnerChange, false)
			}
			if totalWinnerChange < 0 {

				gameSession.sessionCallback.DecreaseMoney(winner,
					int64(-1)*totalWinnerChange, false)

			}
			money := player.GetMoney(gameSession.currencyType)
			player.UnlockMoney(gameSession.currencyType)
			winner.UnlockMoney(gameSession.currencyType)

			gameSession.sessionCallback.SendNotifyMoneyChange(player.Id(), -moneyActuallyLost, "tienlen_freeze_lose_lose", map[string]interface{}{
				"card_types_left": cardTypes,
			})

			gameSession.playersGain[player.Id()] = gameSession.playersGain[player.Id()] + totalPlayerChange
			gameSession.playersGain[winner.Id()] = gameSession.playersGain[winner.Id()] + totalWinnerChange

			totalGain += moneyGain
			gameSession.tax += game.TaxFromMoney(moneyActuallyLost, gameSession.betEntry)
			if player.PlayerType() == "bot" {
				if totalPlayerChange > 0 {
					gameSession.botWin += totalPlayerChange
				} else if totalPlayerChange < 0 {
					gameSession.botLose -= totalPlayerChange
				}
			} else {
				if totalPlayerChange > 0 {
					gameSession.win += totalPlayerChange
				} else if totalPlayerChange < 0 {
					gameSession.lose -= totalPlayerChange
				}
			}
			if winner.PlayerType() == "bot" {
				if totalWinnerChange > 0 {
					gameSession.botWin += totalWinnerChange
				} else if totalWinnerChange < 0 {
					gameSession.botLose -= totalWinnerChange
				}
			} else {
				if totalWinnerChange > 0 {
					gameSession.win += totalWinnerChange
				} else if totalWinnerChange < 0 {
					gameSession.lose -= totalWinnerChange
				}
			}

			gameSession.removePlayerFromRound(player)
			result := &TienLenResult{}
			result.result = "lose"
			result.rank = -1 // last place
			result.change = -moneyActuallyLost
			result.cards = cards
			result.cardTypesLeft = cardTypes
			result.money = money
			result.username = player.Name()

			result.displayName = player.DisplayName()
			result.playerId = player.Id()

			resultData := result.SerializedData()
			gameSession.results = append(gameSession.results, resultData)
			gameSession.removePlayerFromRound(gameSession.GetPlayer(player.Id()))
			gameSession.updateGameRecordForPlayer(player, result.result, result.change)
		}
	}

	winnerResult.money = winner.GetMoney(gameSession.currencyType)
	winnerResult.change = totalGain
	winnerResult.result = "win"
	resultData := winnerResult.SerializedData()
	gameSession.results = append(gameSession.results, resultData)
	gameSession.updateGameRecordForPlayer(winner, winnerResult.result, winnerResult.change)
	gameSession.removePlayerFromRound(gameSession.GetPlayer(winnerResult.playerId))

	gameSession.logToFile("End session with results %v", gameSession.results)
	gameSession.finished = true

	//jackpot
	isTesting := false
	if isTesting {
		fmt.Println("DEBUG checkpoint 1")
	}
	//	jackpotCode := "all"
	//	jackpotInstance := jackpot.GetJackpot(jackpotCode, gameSession.game.currencyType)
	//	if jackpotInstance != nil {
	//		if gameSession.sessionCallback.GetNumberOfHumans() != 0 {
	//			moneyToJackpot := int64(float64(gameSession.tax) * float64(1/3.0))
	//			jackpotInstance.AddMoney(moneyToJackpot)
	//
	//			isHitJackpot := false
	//			if winnerResult.instantWinType == "12_cards_straight" {
	//				isHitJackpot = true
	//			}
	//						if isHitJackpot {
	//							if ratio, isIn := mapMoneyUnitToJackpotRatio[gameSession.betEntry.Min()]; isIn {
	//								temp := int64(float64(jackpotInstance.Value()) * ratio)
	//								jackpotInstance.AddMoney(-temp)
	//
	//								winner.ChangeMoneyAndLog(
	//									temp, gameSession.currencyType, false, "",
	//									"JACKPOT", gameSession.game.GameCode(), "")
	//
	//								jackpotInstance.NotifySomeoneHitJackpot(
	//									gameSession.game.GameCode(),
	//									temp,
	//									winner.Id(),
	//									winner.Name(),
	//								)
	//
	//								if winner.PlayerType() == "normal" {
	//									gameSession.win += temp
	//								}
	//							}
	//						}
	//		}
	//	}

	// event EVENTSC_TIENLEN_BAIDEP
	if isTesting {
		fmt.Println("DEBUG checkpoint 2")
	}
	sc.GlobalMutex.Lock()
	event := sc.MapEventSCs[sc.EVENTSC_TIENLEN_BAIDEP]
	sc.GlobalMutex.Unlock()
	if isTesting {
		fmt.Println("DEBUG checkpoint 3")
	}
	if event != nil {
		event.Mutex.Lock()
		isLimited := false
		if event.MapPlayerIdToValue[winner.Id()] >= event.LimitNOBonus {
			isLimited = true
		}
		event.Mutex.Unlock()
		if isTesting {
			fmt.Println("DEBUG checkpoint 4")
		}
		if winnerResult.instantWinType != "" && !isLimited &&
			gameSession.currencyType != currency.CustomMoney {
			event.ChangeValue(winner.Id(), 1)
			winner.ChangeMoneyAndLog(
				2*gameSession.betEntry.Min(), gameSession.currencyType, false, "",
				sc.ACTION_BONUS_EVENT, gameSession.game.GameCode(), "")
			if winner.PlayerType() == "normal" {
				gameSession.win += 2 * gameSession.betEntry.Min()
			}
		}
	}
	if isTesting {
		fmt.Println("DEBUG checkpoint 5")
	}

	// log to match_record
	var matchId int64
	if gameSession.sessionCallback.GetNumberOfHumans() != 0 {
		matchId = record.LogMatchRecord(gameSession.game.gameCode,
			gameSession.currencyType,
			gameSession.betEntry.Min(),
			gameSession.totalBet,
			gameSession.tax,
			gameSession.win,
			gameSession.lose,
			gameSession.botWin,
			gameSession.botLose,
			gameSession.serializedDataForRecord(),
		)
	}

	if gameSession.game.currencyType == currency.Money {
		for _, player := range gameSession.players {
			bet := gameSession.betEntry.Min()
			player.IncreaseBet(bet)
			player.IncreaseVipPointForMatch(bet, matchId, gameSession.game.gameCode)
		}
	}
	//
	event_player.GlobalMutex.Lock()
	e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
	event_player.GlobalMutex.Unlock()
	if e != nil {
		for _, r1p := range gameSession.results {
			playerId, _ := r1p["id"].(int64)
			change, _ := r1p["change"].(int64)
			e.GiveAPiece(playerId, false,
				gameSession.currencyType == currency.TestMoney, change)
		}
	}
	// bot budget
	gameSession.game.botBudget += (gameSession.botWin - gameSession.botLose)

	if utils.AbsInt64((gameSession.win+gameSession.botWin+gameSession.tax)-
		(gameSession.lose+gameSession.botLose)) > 4 {
		log.LogSerious("ERROR did not balance result %s, match id %d, win %d, lose %d, %v", gameSession.game.gameCode, matchId,
			gameSession.win+gameSession.botWin+gameSession.tax,
			gameSession.lose+gameSession.botLose, gameSession.serializedDataForRecord())
	}

	// đuổi người afk
	//	session := gameSession
	//	afkPids := make([]int64, 0)
	//	session.mutex.Lock()
	//	for pid, isAfk := range session.mapPlayerIdToIsAfkLastTurn {
	//		if isAfk {
	//			afkPids = append(afkPids, pid)
	//		}
	//	}
	//	session.mutex.Unlock()
	//	for _, pid := range afkPids {
	//		pObj := session.GetPlayer(pid)
	//		if pObj != nil {
	//			game.RegisterLeaveRoom(session.game, pObj)
	//			pObj2, _ := player.GetPlayer(pid)
	//			if pObj2 != nil {
	//				pObj2.CreatePopUp("Bạn đã không thao tác quá lâu")
	//			}
	//		}
	//	}

	gameSession.notifyFinishGameSession(gameSession.results)
	//	time.Sleep(7 * time.Second)
	if isTesting {
		fmt.Println("DEBUG checkpoint 6")
	}

	go gameSession.sendForceEndGameEvent()
}

func (gameSession *TienLenSession) handleNextTurnEvent() {
	gameSession.logToFile("start turn for player %d cards %v", gameSession.currentPlayerTurn.Id(), gameSession.cards[gameSession.currentPlayerTurn.Id()])
	gameSession.logToFile("card on table %v", gameSession.cardsOnTable)
	gameSession.startTurnDate = time.Now()
	gameSession.notifyGameStateChange()
	go gameSession.runTimeOutForEndTurnNaturally()

}

func (gameSession *TienLenSession) handleEndTurnNaturallyEvent(turnData *TurnData) {
	if gameSession.finished {
		return
	}
	if gameSession.turnCounter == turnData.turnCounter &&
		gameSession.roundCounter == turnData.roundCounter {
		if gameSession.currentPlayerTurn != nil &&
			gameSession.currentPlayerTurn.Id() == turnData.playerId {
			// play smallest card and then end turn
			if gameSession.turnCounter == 0 {
				// need to play a card since this is first turn
				firstCard := gameSession.cards[gameSession.currentPlayerTurn.Id()][0]
				playCardsAction := &PlayCardsAction{}
				playCardsAction.playerId = gameSession.currentPlayerTurn.Id()
				playCardsAction.cards = []string{firstCard}
				gameSession.handlePlayCardsEvent(playCardsAction)
			} else {
				// this is a skip turn
				gameSession.handleSkipTurnEvent(gameSession.currentPlayerTurn.Id())
			}

			//			gameSession.mutex.Lock()
			//			gameSession.mapPlayerIdToIsAfkLastTurn[turnData.playerId] = true
			//			gameSession.mutex.Unlock()
		}
	}
}

func (gameSession *TienLenSession) handlePlayCardsEvent(playCardsAction *PlayCardsAction) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	player := gameSession.GetPlayer(playCardsAction.playerId)
	if player == nil {
		return errors.New("err:player_not_in_game")
	}
	cards := logic.SortCards(gameSession.game.logicInstance, playCardsAction.cards)

	playerCards := gameSession.cards[player.Id()]
	if !logic.ContainCards(playerCards, cards) {
		return errors.New(l.Get(l.M0011))
	}

	if gameSession.timeOutForTurn == nil || gameSession.currentPlayerTurn.Id() != player.Id() {
		return errors.New(l.Get(l.M0012))
	}
	gameSession.resultInTurn = []string{}
	//resultInTurn :=[]
	if len(gameSession.cardsOnTable) == 0 {
		// no cards on table, this is the first move of the turn
		if gameSession.lastMatchResult == nil &&
			gameSession.turnCounter == 0 &&
			len(playerCards) == 13 {
			// need to play smallest on first turn
			if !logic.ContainCards(cards, []string{playerCards[0]}) {
				return errors.New(l.Get(l.M0013))
			}
		}

		moveType := gameSession.game.logicInstance.GetMoveType(cards)
		if moveType == logic.MoveTypeInvalid {
			return errors.New(l.Get(l.M0014))
		}
		gameSession.allMovesOfCurrentTurn = make([][]string, 0)
	} else {
		isValid := gameSession.game.playCardsOverCards(gameSession.cardsOnTable, cards)
		if !isValid {
			return errors.New(l.Get(l.M0014))
		}
		loseMultiplier := int64(gameSession.game.GetLoseMultiplier(gameSession.cardsOnTable))
		loser := gameSession.getPreviousPlayer()
		//fmt.Print("dang danh")
		if loseMultiplier > 0 && len(cards) > 3 {

			var moneyActuallyLost, tableMoneyUnit int64

			// fmt.Print("dang danh 2")
			tableMoneyUnit = gameSession.betEntry.Min()
			if _, ok := gameSession.playerLoseWinInTurn[loser.Id()]; ok {
				gameSession.playerLoseWinInTurn[loser.Id()] -= loseMultiplier
				//do something here
			} else {
				gameSession.playerLoseWinInTurn[loser.Id()] = -1 * loseMultiplier
			}
			if _, ok := gameSession.playerLoseWinInTurn[player.Id()]; ok {
				gameSession.playerLoseWinInTurn[player.Id()] += loseMultiplier
				//do something here
			} else {
				gameSession.playerLoseWinInTurn[player.Id()] = loseMultiplier
			}
			reason := gameSession.game.logicInstance.GetMoveType(cards)
			moneyActuallyLost = loseMultiplier * tableMoneyUnit
			//moneyGain = game.MoneyAfterTax(moneyActuallyLost, gameSession.betEntry)
			// tra ve nguyen nhan bi tru tien cho tat cac cac user
			// mang gom nguyen nhan tru,  id nguoi bi tru, tien bi tru
			gameSession.resultInTurn = []string{reason, fmt.Sprintf("%v", loser.Id()),
				fmt.Sprintf("%v", moneyActuallyLost), fmt.Sprintf("%v", player.Id())}
			fmt.Printf(" locw %v \r\n", loseMultiplier)
			/* var  moneyGain int64
			loser.LockMoney(gameSession.currencyType)
			player.LockMoney(gameSession.currencyType)


				gameSession.sessionCallback.DecreaseMoney(loser, moneyActuallyLost, false)
				gameSession.sessionCallback.IncreaseMoney(player, moneyGain, false)



			loser.UnlockMoney(gameSession.currencyType)
			player.UnlockMoney(gameSession.currencyType)

			gameSession.playersGain[loser.Id()] = gameSession.playersGain[loser.Id()] - moneyActuallyLost
			gameSession.playersGain[player.Id()] = gameSession.playersGain[player.Id()] + moneyGain

			if loser.PlayerType() == "bot" {
				gameSession.botLose += moneyActuallyLost
			} else {
				gameSession.lose += moneyActuallyLost
			}
			if player.PlayerType() == "bot" {
				gameSession.botWin += moneyGain
			} else {
				gameSession.win += moneyGain
			}
			gameSession.tax += moneyActuallyLost - moneyGain*/
			// fmt.Print("dang danh 2 - thoat")

		}
	}

	playerCards = removeCardsFromCards(playerCards, cards)
	gameSession.cards[player.Id()] = playerCards
	//gameSession.totalCardLeft = len(playerCards)
	gameSession.cardsOnTable = cards
	gameSession.ownerOfCardsOnTable = player
	gameSession.mutex.Lock()
	gameSession.currentPlayerTurn = player
	gameSession.mutex.Unlock()
	gameSession.allMovesOfCurrentTurn = append(gameSession.allMovesOfCurrentTurn, cards)

	gameSession.logToFile("player %d play cards %v", gameSession.currentPlayerTurn.Id(), cards)
	// check the current players card

	var forcedNextPlayer game.GamePlayer
	if len(gameSession.cards[player.Id()]) == 0 {
		// win already

		winner := player
		result := &TienLenResult{}
		winRank := 0
		result.rank = winRank
		result.money = winner.GetMoney(gameSession.currencyType)
		result.playerId = winner.Id()
		result.username = winner.Name()

		forcedNextPlayer = gameSession.getNextPlayerInRound()
		gameSession.handleReceiveResultEvent(result, false)
	} else {
		gameSession.calculateNextTurnMove(forcedNextPlayer)
		go gameSession.sendNextTurnEvent()
	}

	return nil
}

func (gameSession *TienLenSession) handleSkipTurnEvent(playerId int64) (err error) {
	// need to handle check if first move, if yes then cannot skip
	if gameSession.turnCounter == 0 {
		return errors.New(l.Get(l.M0015))
	}

	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}
	if gameSession.currentPlayerTurn.Id() != playerId {
		return errors.New(l.Get(l.M0012))
	}
	gameSession.resultInTurn = []string{}
	gameSession.logToFile("player %d skip turn", gameSession.currentPlayerTurn.Id())
	// remove player from current round
	skipPlayer := gameSession.currentPlayerTurn
	theoricallyNextPlayer := gameSession.getNextPlayerInRound()
	gameSession.removePlayerFromRound(skipPlayer)

	gameSession.calculateNextTurnMove(theoricallyNextPlayer)
	go gameSession.sendNextTurnEvent()
	return nil

}

func (gameSession *TienLenSession) runTimeOutForEndTurnNaturally() {
	if gameSession.timeOutForTurn != nil {
		gameSession.timeOutForTurn.SetShouldHandle(false)
	}

	gameSession.timeOutForTurn = utils.NewTimeOut(gameSession.game.turnTimeInSeconds)
	turnData := &TurnData{
		turnCounter:  gameSession.turnCounter,
		roundCounter: gameSession.roundCounter,
		playerId:     gameSession.currentPlayerTurn.Id(),
	}
	if gameSession.timeOutForTurn.Start() && !gameSession.finished {
		// log.Log("endnatural")
		gameSession.sendEndTurnNaturallyEvent(turnData)
	}

}
