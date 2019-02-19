package maubinh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	sc "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/player"
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

func (gameSession *MauBinhSession) sendStartGameEvent() {
	gameSession.startGameChan <- true
}

func (gameSession *MauBinhSession) sendEndGameEvent() {
	gameSession.endGameChan <- true
}

func (gameSession *MauBinhSession) sendFinishOrganizeCardsAction(action *FinishOrganizeCardsAction) {
	gameSession.finishOrganizeCardsChan <- action
}

func (gameSession *MauBinhSession) sendUploadCardsAction(action *UploadCardsAction) {
	gameSession.uploadCardsChan <- action
}

func (gameSession *MauBinhSession) sendStartOrganizeCardsAgainAction(playerId int64) {
	gameSession.startOrganizeCardsAgainChan <- playerId
}

func (gameSession *MauBinhSession) sendForceEndGameEvent() {
	gameSession.forceEndChan <- true
}

func (gameSession *MauBinhSession) handleStartGameEvent() {
	gameSession.logToFile("start maubinh with players %v cards %v", getIdFromPlayersMap(gameSession.players), gameSession.cards)
	gameSession.tax = 0
	gameSession.matchId = fmt.Sprintf("#%v", int32(time.Now().Unix()))

	if gameSession.game.logicInstance.HasWhiteWin() {
		var whiteWinCount int
		for _, player := range gameSession.players {
			cards := gameSession.cards[player.Id()]
			whiteWinType := gameSession.game.getTypeOfWhiteWin(cards)
			if whiteWinType != "" {
				gameSession.whiteWin[player.Id()] = whiteWinType
				gameSession.mutex.Lock()
				gameSession.finishOrganizingCards[player.Id()] = true
				gameSession.organizedCardsData[player.Id()] = gameSession.game.organizeCardsForWhiteWin(cards, whiteWinType)
				gameSession.mutex.Unlock()
				whiteWinCount++
			} else {
				gameSession.mutex.Lock()
				gameSession.organizedCardsData[player.Id()] = organizeCardsInOrder(cards, whiteWinType)
				gameSession.mutex.Unlock()
			}
		}
		gameSession.forceEndGame = false
		if len(gameSession.players)-whiteWinCount <= 1 {
			// end game right away
			gameSession.forceEndGame = true
			gameSession.handleEndGameEvent()
			go gameSession.sendForceEndGameEvent()
		} else {
			gameSession.startTurnDate = time.Now()
			gameSession.notifyGameStateChange()
			go gameSession.runTimeOutForEndTurnNaturally()
		}
	} else {
		for _, player := range gameSession.players {
			cards := gameSession.cards[player.Id()]
			gameSession.mutex.Lock()
			gameSession.organizedCardsData[player.Id()] = organizeCardsInOrder(cards, "")
			gameSession.mutex.Unlock()
		}
		gameSession.startTurnDate = time.Now()
		gameSession.notifyGameStateChange()
		go gameSession.runTimeOutForEndTurnNaturally()
	}

}

func (gameSession *MauBinhSession) calculateResultsBetweenPlayers() (endByWhiteWin bool, needToCountAces bool) {
	gameSession.mutex.Lock()
	defer gameSession.mutex.Unlock()

	betEntry := gameSession.betEntry
	gameSession.resultObjects = make([]*MauBinhResult, 0)
	numberOfWhiteWin := 0
	var isWhiteWin bool

	if gameSession.game.logicInstance.HasWhiteWin() {
		for _, player := range gameSession.players {
			cards := gameSession.cards[player.Id()]
			result := gameSession.resultObjectForPlayer(player.Id())
			result.compareData2 = map[int64]map[string]int64{}
			whiteWinType := gameSession.game.getTypeOfWhiteWin(cards)
			if whiteWinType != "" {
				isWhiteWin = true
				numberOfWhiteWin++
			}
		}

		if numberOfWhiteWin >= len(gameSession.players)-1 {
			endByWhiteWin = true
		} else {
			endByWhiteWin = false
		}
	} else {
		isWhiteWin = false
	}

	comparePlayers := make([]game.GamePlayer, 0)

	if isWhiteWin {
		alreadyCalculatedResults := make([]*BetweenPlayersResult, 0)
		for _, player := range gameSession.players {
			cards := gameSession.cards[player.Id()]
			whiteWinType := gameSession.game.getTypeOfWhiteWin(cards)

			result := gameSession.resultObjectForPlayer(player.Id())
			result.resultType = "white_win"
			result.whiteWinType = whiteWinType
			if whiteWinType != "" {
				for _, otherPlayer := range gameSession.players {
					if otherPlayer.Id() != player.Id() {
						if hasCalculatedResultBetweenPlayers(alreadyCalculatedResults, player.Id(), otherPlayer.Id()) {
							continue
						} else {
							alreadyCalculatedResults = recordAlreadyCalculatedResultBetweenPlayers(alreadyCalculatedResults, player.Id(), otherPlayer.Id())

							otherCards := gameSession.cards[otherPlayer.Id()]
							otherResult := gameSession.resultObjectForPlayer(otherPlayer.Id())
							otherResult.resultType = "white_win"
							multiplier := gameSession.game.getWhiteWinMultiplierBetweenCards(cards, otherCards)

							if multiplier == 0 {
								// draw
								// do nothing
							} else {
								result.multiplier += multiplier
								otherResult.multiplier -= multiplier

								if multiplier > 0 {
									money := moneyAfterApplyMultiplier(gameSession.betEntry.Min(), multiplier)
									otherPlayer.LockMoney(gameSession.currencyType)
									player.LockMoney(gameSession.currencyType)
									otherMoney := gameSession.sessionCallback.GetMoneyOnTable(otherPlayer.Id())
									if otherMoney < money {
										money = otherMoney
									}
									gain := game.MoneyAfterTax(money, gameSession.betEntry)
									gameSession.updateDataForPlayerWithChange(player, gain, false)
									gameSession.updateDataForPlayerWithChange(otherPlayer, -money, false)
									otherPlayer.UnlockMoney(gameSession.currencyType)
									player.UnlockMoney(gameSession.currencyType)

									result.moneyGain += gain
									otherResult.moneyLost += money

									// record
									gameSession.tax += game.TaxFromMoney(money, gameSession.betEntry)
									if player.PlayerType() == "bot" {
										gameSession.botWin += gain
									} else {
										gameSession.win += gain
									}

									if otherPlayer.PlayerType() == "bot" {
										gameSession.botLose += money
									} else {
										gameSession.lose += money
									}
								} else if multiplier < 0 {
									money := moneyAfterApplyMultiplier(gameSession.betEntry.Min(), -multiplier)
									player.LockMoney(gameSession.currencyType)
									otherPlayer.LockMoney(gameSession.currencyType)
									playerMoney := gameSession.sessionCallback.GetMoneyOnTable(player.Id())
									if playerMoney < money {
										money = playerMoney
									}
									gain := game.MoneyAfterTax(money, gameSession.betEntry)
									gameSession.updateDataForPlayerWithChange(player, -money, false)
									gameSession.updateDataForPlayerWithChange(otherPlayer, gain, false)
									otherPlayer.UnlockMoney(gameSession.currencyType)
									player.UnlockMoney(gameSession.currencyType)

									otherResult.moneyGain += gain
									result.moneyLost += money

									// record
									gameSession.tax += game.TaxFromMoney(money, gameSession.betEntry)
									if player.PlayerType() == "bot" {
										gameSession.botWin += gain
									} else {
										gameSession.win += gain
									}

									if otherPlayer.PlayerType() == "bot" {
										gameSession.botLose += money
									} else {
										gameSession.lose += money
									}
								}
							}

							result.money = player.GetMoney(gameSession.currencyType)
							otherResult.money = otherPlayer.GetMoney(gameSession.currencyType)
						}

					}
				}
			} else {
				if !containPlayer(comparePlayers, player) {
					comparePlayers = append(comparePlayers, player)
				}
			}
		}
	} else {
		for _, player := range gameSession.players {
			comparePlayers = append(comparePlayers, player)
		}
	}

	if betEntry.CheatCode() != "" {
		tokens := strings.Split(betEntry.CheatCode(), ";")
		if len(tokens) > 0 {
			cheatType := tokens[0]
			if cheatType == "botstrong" {
				if len(tokens) == 2 {
					percent, _ := strconv.ParseInt(tokens[1], 10, 64)
					if percent > 0 {
						for _, player := range comparePlayers {
							if player.PlayerType() == "bot" {
								gameSession.generateBetterCardsData(player.Id(), int(percent))
								break
							}
						}
					}
				}
			} else if cheatType == "balance" {
				if len(tokens) == 4 {
					budgetOffset, _ := strconv.ParseInt(tokens[1], 10, 64)
					minPercent, _ := strconv.ParseInt(tokens[2], 10, 64)
					maxPercent, _ := strconv.ParseInt(tokens[3], 10, 64)
					var percent int64
					if gameSession.game.botBudget > budgetOffset {
						percent = minPercent
					} else if gameSession.game.botBudget < -budgetOffset {
						percent = maxPercent
					} else {
						percent = (minPercent + maxPercent) / 2
					}

					if percent > 0 {
						for _, player := range comparePlayers {
							if player.PlayerType() == "bot" {
								gameSession.generateBetterCardsData(player.Id(), int(percent))
								break
							}
						}
					}
				}
			}
		}
	}
	//gameSession.resultInCollapse = map[string]int64{}

	// check win collapse all
	var winCollapseAllPlayerId int64
	if !isWhiteWin && len(comparePlayers) > 2 {
		for _, player := range comparePlayers {
			cardsData := gameSession.organizedCardsData[player.Id()]

			collapseCounter := 0
			loseCollapseCounter := 0
			for _, otherPlayer := range comparePlayers {
				if otherPlayer.Id() != player.Id() {
					otherCardsData := gameSession.organizedCardsData[otherPlayer.Id()]
					iret := gameSession.game.getIsCollapsingBetweenCardsData2(cardsData, otherCardsData)
					if iret == 1 {
						collapseCounter++
					} else if iret == -1 {
						loseCollapseCounter++
					}
				}
			}
			if collapseCounter == len(comparePlayers)-1 {
				winCollapseAllPlayerId = player.Id()

			} else if loseCollapseCounter == len(comparePlayers)-1 {
				gameSession.loseCollapseAllPhayerId = player.Id()
			}

		}
	}
	//keywl := ""

	gameSession.winCollapseAllPlayerId = winCollapseAllPlayerId

	// calculate result normally against each player, for each part
	for _, positionString := range []string{BottomPart, MiddlePart, TopPart} {
		alreadyCalculatedResults := make([]*BetweenPlayersResult, 0)
		for _, player := range comparePlayers {
			cardsData := gameSession.organizedCardsData[player.Id()]
			result := gameSession.resultObjectForPlayer(player.Id())
			//result.compareData2 := map[int64]map[string]int64{}
			for _, otherPlayer := range comparePlayers {
				if otherPlayer.Id() != player.Id() {
					if hasCalculatedResultBetweenPlayers(alreadyCalculatedResults, player.Id(), otherPlayer.Id()) {
						continue
					}
					collapseMult := 0.
					allCollapseMult := 0.
					loseAllCollapseMult := 0.
					otherCardsData := gameSession.organizedCardsData[otherPlayer.Id()]
					otherResult := gameSession.resultObjectForPlayer(otherPlayer.Id())
					data := make(map[string]interface{})
					otherData := make(map[string]interface{})
					// calculate for each position to see if collapse

					isCollapsing := gameSession.game.getIsCollapsingBetweenCardsData(cardsData, otherCardsData)
					multiplier1 := gameSession.game.getMultiplierBetweenCardsData(cardsData, otherCardsData, positionString)
					multiplier := multiplier1

					if isCollapsing { //detect sap

						//gameSession.resultInCollapse[keywl] = true // otherPlayer.Id()
						multiplier = multiplier * gameSession.game.logicInstance.CollapseMultiplier()
						collapseMult = multiplier - multiplier1

					}
					if player.Id() == winCollapseAllPlayerId ||
						otherPlayer.Id() == winCollapseAllPlayerId {
						multiplier = multiplier * gameSession.game.logicInstance.WinCollapseAllMultiplier()
						allCollapseMult = multiplier - multiplier1 - collapseMult
					}
					if player.Id() == gameSession.loseCollapseAllPhayerId ||
						otherPlayer.Id() == gameSession.loseCollapseAllPhayerId {
						multiplier = multiplier * gameSession.game.logicInstance.WinCollapseAllMultiplier()
						loseAllCollapseMult = multiplier - multiplier1 - collapseMult
					}

					if multiplier < 0 { //playerid thua
						result.isCollapsing = true
						multiplier1 *= -1.
						collapseMult *= -1.
						allCollapseMult *= -1.
						loseAllCollapseMult *= -1.
						if _, isin := result.compareData2[otherPlayer.Id()]; isin == false {
							result.compareData2[otherPlayer.Id()] = map[string]int64{}
						}
						result.compareData2[otherPlayer.Id()][positionString] = -1 * int64(multiplier1)

						if _, isin := otherResult.compareData2[player.Id()]; isin == false {
							otherResult.compareData2[player.Id()] = map[string]int64{}
						}
						otherResult.compareData2[player.Id()][positionString] = int64(multiplier1)

						if _, isin := result.compareData2[otherPlayer.Id()]["c"]; isin == false {
							result.compareData2[otherPlayer.Id()]["c"] = -1 * int64(collapseMult)
							result.compareData2[otherPlayer.Id()]["a"] = -1 * int64(allCollapseMult)
							result.compareData2[otherPlayer.Id()]["la"] = -1 * int64(loseAllCollapseMult)
							otherResult.compareData2[player.Id()]["c"] = int64(collapseMult)
							otherResult.compareData2[player.Id()]["a"] = int64(allCollapseMult)
							otherResult.compareData2[player.Id()]["la"] = int64(loseAllCollapseMult)
						} else {
							result.compareData2[otherPlayer.Id()]["c"] += -1 * int64(collapseMult)
							result.compareData2[otherPlayer.Id()]["a"] += -1 * int64(allCollapseMult)
							result.compareData2[otherPlayer.Id()]["la"] += -1 * int64(loseAllCollapseMult)
							otherResult.compareData2[player.Id()]["c"] += int64(collapseMult)
							otherResult.compareData2[player.Id()]["a"] += int64(allCollapseMult)
							otherResult.compareData2[player.Id()]["la"] += int64(loseAllCollapseMult)
						}
						//  compareData2      map[int64][]int64
					} else {
						otherResult.isCollapsing = true
						if _, isin := result.compareData2[otherPlayer.Id()]; isin == false {
							result.compareData2[otherPlayer.Id()] = map[string]int64{}
						}
						result.compareData2[otherPlayer.Id()][positionString] = int64(multiplier1)

						if _, isin := otherResult.compareData2[player.Id()]; isin == false {
							otherResult.compareData2[player.Id()] = map[string]int64{}
						}
						otherResult.compareData2[player.Id()][positionString] = -1 * int64(multiplier1)
						if _, isin := result.compareData2[otherPlayer.Id()]["c"]; isin == false {
							result.compareData2[otherPlayer.Id()]["c"] = int64(collapseMult)
							result.compareData2[otherPlayer.Id()]["a"] = int64(allCollapseMult)
							result.compareData2[otherPlayer.Id()]["la"] = int64(loseAllCollapseMult)
							otherResult.compareData2[player.Id()]["c"] = -1 * int64(collapseMult)
							otherResult.compareData2[player.Id()]["a"] = -1 * int64(allCollapseMult)
							otherResult.compareData2[player.Id()]["la"] = -1 * int64(loseAllCollapseMult)
						} else {
							result.compareData2[otherPlayer.Id()]["c"] += int64(collapseMult)
							result.compareData2[otherPlayer.Id()]["a"] += int64(allCollapseMult)
							result.compareData2[otherPlayer.Id()]["la"] += int64(loseAllCollapseMult)
							otherResult.compareData2[player.Id()]["c"] += -1 * int64(collapseMult)
							otherResult.compareData2[player.Id()]["a"] += -1 * int64(allCollapseMult)
							otherResult.compareData2[player.Id()]["la"] += -1 * int64(loseAllCollapseMult)
						}
					}
					if multiplier == 0 {
						// draw
						// do nothing
						data["multiplier"] = 0
						data["change"] = 0
						otherData["multiplier"] = 0
						otherData["change"] = 0
					} else {
						var change int64
						var otherChange int64
						if multiplier > 0 {
							money := moneyAfterApplyMultiplier(gameSession.betEntry.Min(), multiplier)
							otherPlayer.LockMoney(gameSession.currencyType)
							player.LockMoney(gameSession.currencyType)
							otherMoney := gameSession.sessionCallback.GetMoneyOnTable(otherPlayer.Id())
							if otherMoney < money {
								money = otherMoney
							}

							gain := game.MoneyAfterTax(money, gameSession.betEntry)
							gameSession.updateDataForPlayerWithChange(player, gain, false)
							gameSession.updateDataForPlayerWithChange(otherPlayer, -money, false)
							otherPlayer.UnlockMoney(gameSession.currencyType)
							player.UnlockMoney(gameSession.currencyType)

							result.moneyGain += gain
							otherResult.moneyLost += money
							change = gain
							otherChange = -money

							// record
							gameSession.recordMoney(money, player, otherPlayer)
						} else if multiplier < 0 {
							money := moneyAfterApplyMultiplier(gameSession.betEntry.Min(), -multiplier)

							player.LockMoney(gameSession.currencyType)
							otherPlayer.LockMoney(gameSession.currencyType)
							playerMoney := gameSession.sessionCallback.GetMoneyOnTable(player.Id())
							if playerMoney < money {
								money = playerMoney
							}

							gain := game.MoneyAfterTax(money, gameSession.betEntry)
							gameSession.updateDataForPlayerWithChange(player, -money, false)
							gameSession.updateDataForPlayerWithChange(otherPlayer, gain, false)
							player.UnlockMoney(gameSession.currencyType)
							otherPlayer.UnlockMoney(gameSession.currencyType)

							otherResult.moneyGain += gain
							result.moneyLost += money
							change = -money
							otherChange = gain

							// record
							gameSession.recordMoney(money, otherPlayer, player)
						}
						result.multiplier += multiplier
						otherResult.multiplier -= multiplier
						data["multiplier"] = multiplier
						data["change"] = change
						otherData["multiplier"] = -multiplier
						otherData["change"] = otherChange
					}

					result.compareData[positionString][fmt.Sprintf("%d", otherPlayer.Id())] = data
					otherResult.compareData[positionString][fmt.Sprintf("%d", player.Id())] = otherData

					alreadyCalculatedResults = recordAlreadyCalculatedResultBetweenPlayers(alreadyCalculatedResults, player.Id(), otherPlayer.Id())

				}

			}

		}

	}
	// work on ace
	if gameSession.game.logicInstance.HasCountAces() {
		gameSession.calculateAces()
	}

	// serialize result
	gameSession.results = make([]map[string]interface{}, 0)
	collapseCount := 0

	for _, result := range gameSession.resultObjects {
		if result.isCollapsing {
			collapseCount++
		}
	}
	if len(gameSession.resultObjects) > 2 && collapseCount == len(gameSession.resultObjects)-1 {
		// has collapse all win
		for _, result := range gameSession.resultObjects {
			if !result.isCollapsing {
				result.winCollapseAll = true
			}
		}
	}

	for _, result := range gameSession.resultObjects {
		if result.aceMultiplier != 0 {
			needToCountAces = true
		}
		if result.isCollapsing {
			collapseCount++
		}
		result.valid = gameSession.game.isCardsDataValid(gameSession.organizedCardsData[result.playerId])
		player := gameSession.GetPlayer(result.playerId)
		result.money = gameSession.GetPlayer(result.playerId).GetMoney(gameSession.currencyType)
		result.moneyOnTable = gameSession.sessionCallback.GetMoneyOnTable(result.playerId)
		data := result.SerializedData()
		gameSession.results = append(gameSession.results, result.SerializedData())
		if player != nil {
			gameSession.updateGameRecordForPlayer(player, utils.GetStringAtPath(data, "result"), utils.GetInt64AtPath(data, "change"))
		}
	}

	//jackpot
	jackpotCode := "all"
	jackpotInstance := jackpot.GetJackpot(jackpotCode, gameSession.game.currencyType)
	if jackpotInstance != nil {
		var winner game.GamePlayer
		maxWinningMoney := int64(-999999999)
		for _, result := range gameSession.resultObjects {
			if result.moneyGain-result.moneyLost > maxWinningMoney {
				maxWinningMoney = result.moneyGain - result.moneyLost
				winner = gameSession.GetPlayer(result.playerId)
			}
		}

		if gameSession.sessionCallback.GetNumberOfHumans() != 0 {
			moneyToJackpot := int64(float64(gameSession.tax) * float64(1/3.0))
			jackpotInstance.AddMoney(moneyToJackpot)
		}

		isHitJackpot := false
		if endByWhiteWin {
			cards := gameSession.cards[winner.Id()]
			whiteWinType := gameSession.game.getTypeOfWhiteWin(cards)
			if whiteWinType == WhiteWinTypeDragonRollingStraight {
				isHitJackpot = true
			}

			// event EVENTSC_MAUBINH_BAIDEP
			sc.GlobalMutex.Lock()
			event := sc.MapEventSCs[sc.EVENTSC_MAUBINH_BAIDEP]
			sc.GlobalMutex.Unlock()
			if event != nil {
				event.Mutex.Lock()
				isLimited := false
				if event.MapPlayerIdToValue[winner.Id()] >= event.LimitNOBonus {
					isLimited = true
				}
				event.Mutex.Unlock()
				if whiteWinType != "" && !isLimited &&
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
		}
		if isHitJackpot {
			if ratio, isIn := mapMoneyUnitToJackpotRatio[gameSession.betEntry.Min()]; isIn {
				temp := int64(float64(jackpotInstance.Value()) * ratio)
				jackpotInstance.AddMoney(-temp)
				winner.ChangeMoneyAndLog(
					temp, gameSession.currencyType, false, "",
					"JACKPOT", gameSession.game.GameCode(), "")

				jackpotInstance.NotifySomeoneHitJackpot(
					gameSession.game.GameCode(),
					temp,
					winner.Id(),
					winner.Name(),
				)

				if winner.PlayerType() == "normal" {
					gameSession.win += temp
				}
			}
		}

	}

	return endByWhiteWin, needToCountAces
}

func (gameSession *MauBinhSession) handleEndGameEvent() {
	gameSession.logToFile("handle end session")
	gameSession.goingToFinish = true
	gameSession.notifyGameStateChange()
	gameSession.goingToFinish = false

	gameSession.finished = true
	endByWhiteWin, needToCountAces := gameSession.calculateResultsBetweenPlayers()

	var delayUntilsNewGame int
	if endByWhiteWin {
		delayUntilsNewGame = 7
	} else {
		delayUntilsNewGame = 15
	}

	if needToCountAces {
		delayUntilsNewGame += 3
	}

	gameSession.logToFile("End session with results %v", gameSession.results)
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
	//
	event_player.GlobalMutex.Lock()
	e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
	event_player.GlobalMutex.Unlock()
	if e != nil {
		for _, r1p := range gameSession.resultObjects {
			e.GiveAPiece(r1p.playerId, false,
				gameSession.currencyType == currency.TestMoney, r1p.moneyGain)
		}
	}
	// bot budget
	gameSession.game.botBudget += (gameSession.botWin - gameSession.botLose)

	// đuổi người afk
	session := gameSession
	afkPids := make([]int64, 0)
	session.mutex.Lock()
	for _, pid := range session.playersIdWhenStart {
		if session.mapPlayerIdToIsNotAfk[pid] == false {
			afkPids = append(afkPids, pid)
		}
	}
	session.mutex.Unlock()
	for _, pid := range afkPids {
		pObj := session.GetPlayer(pid)
		if pObj != nil {
			game.RegisterLeaveRoom(session.game, pObj)
			pObj2, _ := player.GetPlayer(pid)
			if pObj2 != nil {
				pObj2.CreatePopUp("Bạn đã không thao tác quá lâu")
			}
		}
	}

	gameSession.notifyFinishGameSession(gameSession.results, delayUntilsNewGame)
	//	time.Sleep(time.Duration(delayUntilsNewGame) * time.Second)

	// gameSession.recordTypeAfterGame(matchId)

	if gameSession.game.currencyType == currency.Money {
		for _, player := range gameSession.players {
			bet := gameSession.betEntry.Min()
			player.IncreaseBet(bet)
			player.IncreaseVipPointForMatch(bet, matchId, gameSession.game.gameCode)
		}
	}
	if utils.AbsInt64((gameSession.win+gameSession.botWin+gameSession.tax)-
		(gameSession.lose+gameSession.botLose)) > 100 {
		log.LogSerious("ERROR did not balance result %s, match id %d, win %d, lose %d, %v",
			gameSession.game.gameCode,
			matchId,
			gameSession.win+gameSession.botWin+gameSession.tax,
			gameSession.lose+gameSession.botLose, gameSession.serializedDataForRecord())
	}

}

func (gameSession *MauBinhSession) handleUploadCards(action *UploadCardsAction) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	player := gameSession.GetPlayer(action.playerId)
	if player == nil {
		return errors.New("err:player_not_in_game")
	}

	if gameSession.whiteWin[player.Id()] != "" {
		return errors.New("Bạn đã đạt mậu binh, không cần làm gì nữa cả")
	}

	cardsData := action.cardsData
	parsedCardsData := make(map[string][]string)
	defer func() {
		parsedCardsData = nil
	}()
	for positionString, cardsDataInterface := range cardsData {
		parsedCardsData[positionString] = gameSession.game.sortCards(utils.GetStringSliceFromScanResult(cardsDataInterface))
	}

	// valid check
	if len(parsedCardsData[TopPart]) != 3 ||
		len(parsedCardsData[MiddlePart]) != 5 ||
		len(parsedCardsData[BottomPart]) != 5 {
		return errors.New(l.Get(l.M0022))
	}

	playerCards := gameSession.cards[player.Id()]
	for _, positionString := range []string{TopPart, MiddlePart, BottomPart} {
		for _, cardString := range parsedCardsData[positionString] {
			if !utils.ContainsByString(playerCards, cardString) {
				return errors.New(l.Get(l.M0022))
			}
			playerCards = removeCardsFromCards(playerCards, []string{cardString})
		}
	}

	gameSession.mutex.Lock()
	gameSession.organizedCardsData[player.Id()] = parsedCardsData
	gameSession.mutex.Unlock()
	return nil
}

func (gameSession *MauBinhSession) handleFinishOrganizeCards(action *FinishOrganizeCardsAction) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	if gameSession.whiteWin[action.playerId] != "" {
		return errors.New("Bạn đã đạt mậu binh, không cần làm gì nữa cả")
	}

	player := gameSession.GetPlayer(action.playerId)
	if player == nil {
		return errors.New("err:player_not_in_game")
	}

	cardsData := action.cardsData
	parsedCardsData := make(map[string][]string)
	defer func() {
		parsedCardsData = nil
	}()
	for positionString, cardsDataInterface := range cardsData {
		parsedCardsData[positionString] = gameSession.game.sortCards(utils.GetStringSliceFromScanResult(cardsDataInterface))
	}

	// valid check
	if len(parsedCardsData[TopPart]) != 3 ||
		len(parsedCardsData[MiddlePart]) != 5 ||
		len(parsedCardsData[BottomPart]) != 5 {
		return errors.New(l.Get(l.M0022))
	}

	playerCards := gameSession.cards[player.Id()]
	for _, positionString := range []string{TopPart, MiddlePart, BottomPart} {
		for _, cardString := range parsedCardsData[positionString] {
			if !utils.ContainsByString(playerCards, cardString) {
				// fmt.Println("notfound", playerCards, cardString)
				return errors.New(l.Get(l.M0022))
			}
			playerCards = removeCardsFromCards(playerCards, []string{cardString})
		}
	}

	if !gameSession.game.isCardsDataValid(parsedCardsData) {
		return errors.New(l.Get(l.M0022))
	}

	gameSession.mutex.Lock()
	gameSession.organizedCardsData[player.Id()] = parsedCardsData
	gameSession.finishOrganizingCards[player.Id()] = true
	gameSession.mutex.Unlock()
	if gameSession.isEveryoneFinishOrganizedCards() {
		go gameSession.sendEndGameEvent()
	}
	gameSession.notifyGameStateChange()
	return nil
}

func (gameSession *MauBinhSession) handleStartOrganizeCardsAgain(playerId int64) (err error) {
	if gameSession.finished {
		return errors.New(l.Get(l.M0010))
	}

	if gameSession.whiteWin[playerId] != "" {
		return errors.New("Bạn đã đạt mậu binh, không cần làm gì nữa cả")
	}

	player := gameSession.GetPlayer(playerId)
	if player == nil {
		return errors.New("err:player_not_in_game")
	}
	gameSession.mutex.Lock()
	gameSession.finishOrganizingCards[player.Id()] = false
	gameSession.mutex.Unlock()
	gameSession.notifyGameStateChange()
	return nil
}

func (gameSession *MauBinhSession) runTimeOutForEndTurnNaturally() {
	if gameSession.timeOutForTurn != nil {
		gameSession.timeOutForTurn.SetShouldHandle(false)
	}
	gameSession.timeOutForTurn = utils.NewTimeOut(gameSession.game.turnTimeInSeconds)
	if gameSession.timeOutForTurn.Start() && !gameSession.finished {
		gameSession.sendEndGameEvent()
	}

}
