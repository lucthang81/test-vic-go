package maubinh

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/game"
	"math/rand"
	"sort"
)

type ByRandom []string

func (a ByRandom) Len() int      { return len(a) }
func (a ByRandom) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRandom) Less(i, j int) bool {
	return rand.Intn(2) == 0
}

// chia bài xịn hơn cho bot @@ ??
func (gameSession *MauBinhSession) generateBetterCardsData(playerId int64, percent int) {
	defer func() {
		if r := recover(); r != nil {
			log.SendMailWithCurrentStack(fmt.Sprintf("cheat mb error %v", r))
		}
	}()

	gameSession.mutex.Lock()
	defer gameSession.mutex.Unlock()

	gameInstance := gameSession.game
	validPlayers := make([]game.GamePlayer, 0)
	for _, player := range gameSession.players {
		whiteWin := gameSession.whiteWin[player.Id()]
		var isValid bool
		isValid = true
		if whiteWin != "" {
			isValid = false
		}

		cardsData := gameSession.organizedCardsData[player.Id()]
		if !gameSession.game.isCardsDataValid(cardsData) {
			isValid = false
		}
		if isValid == false {
			if playerId == player.Id() {
				return
			}
			continue
		} else {
			validPlayers = append(validPlayers, player)
		}
	}

	if len(validPlayers) <= 1 {
		return
	}

	deck := gameSession.deck
	deck.PutCardBack(gameSession.cards[playerId])
	cardsData := make(map[string][]string)
	var willAbort bool
	for _, positionString := range []string{BottomPart, MiddlePart, TopPart} {
		chance := rand.Intn(100) + 1
		if chance > percent {
			cardsData[positionString] = gameSession.organizedCardsData[playerId][positionString]
			deck.DrawSpecificCards(cardsData[positionString])
			continue
		}

		var bestCards []string
		for _, player := range validPlayers {
			if player.Id() == playerId {
				continue
			}
			cards := gameSession.organizedCardsData[player.Id()][positionString]
			if len(bestCards) == 0 {
				bestCards = cards
			} else {
				if gameSession.game.getTotalMultiplierBetweenCards(bestCards, cards, positionString) < 0 {
					bestCards = cards
				}
			}
		}

		bestCardsType := gameInstance.getTypeOfCards(bestCards)
		bestTypeOrder := gameSession.game.getTypeOrder(bestCardsType)
		betterTypes := make([]string, 0)
		for index, cardsType := range gameInstance.logicInstance.TypeOrder() {
			if bestTypeOrder <= index {
				betterTypes = append(betterTypes, cardsType)
			}
		}
		if len(betterTypes) == 0 {
			continue
		}

		// find better cards
		cards := gameSession.getBetterCards(bestCards, betterTypes, positionString, cardsData)
		if positionString == BottomPart || positionString == MiddlePart {
			if len(cards) == 0 {
				willAbort = true
				break
			}

			if len(cards) != 5 {
				missingCount := 5 - len(cards)
				if positionString == BottomPart {
					for i := 0; i < 250; i++ {
						randomCards := deck.DrawRandomCards(missingCount)
						group := append(cards, randomCards...)
						if gameInstance.getCompareBetweenCards(group, bestCards) > 0 {
							cards = group
							break
						}
					}
				} else if positionString == MiddlePart {
					bottomCards := cardsData[BottomPart]
					for i := 0; i < 250; i++ {
						randomCards := deck.DrawRandomCards(missingCount)
						group := append(cards, randomCards...)
						if gameInstance.getCompareBetweenCards(group, bestCards) > 0 &&
							gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
							cards = group
							break
						}
					}
				}
			}

			if len(cards) != 5 {
				willAbort = true
				break
			}
		} else {
			if len(cards) != 3 {
				missingCount := 3 - len(cards)
				middleCards := cardsData[MiddlePart]
				for i := 0; i < 250; i++ {
					randomCards := deck.DrawRandomCards(missingCount)
					group := append(cards, randomCards...)
					if gameInstance.getCompareBetweenCards(group, bestCards) > 0 &&
						gameInstance.getCompareBetweenCards(middleCards, group) > 0 {
						cards = group
						break
					}
				}
			}
		}

		cardsData[positionString] = cards
	}

	if willAbort {
		cardsDeck := make([]string, 0)
		for cardString, available := range deck.Cards() {
			if available {
				cardsDeck = append(cardsDeck, cardString)
			}
		}
		// log.LogSerious("abort cheat mb, deck %v", cardsDeck)
		deck.DrawSpecificCards(gameSession.cards[playerId])
		return
	}
	cards := make([]string, 0)
	cards = append(cards, cardsData[TopPart]...)
	cards = append(cards, cardsData[MiddlePart]...)
	cards = append(cards, cardsData[BottomPart]...)
	cards = gameSession.game.sortCards(cards)
	gameSession.organizedCardsData[playerId] = cardsData
	gameSession.cards[playerId] = cards
}

func (gameSession *MauBinhSession) getBetterCards(betterCards []string, betterTypes []string, position string, cardsData map[string][]string) []string {
	gameInstance := gameSession.game
	deck := gameSession.deck
	cardsDeck := make([]string, 0)
	for cardString, available := range deck.Cards() {
		if available {
			cardsDeck = append(cardsDeck, cardString)
		}
	}
	cardsGroup := gameInstance.NewCardGroups(cardsDeck)

	if position == BottomPart {
		sort.Sort(ByRandom(betterTypes))
		for _, cardType := range betterTypes {
			if cardType == TypeStraightFlushTop ||
				cardType == TypeStraightFlushBottom ||
				cardType == TypeStraightFlush {
				for _, group := range cardsGroup.straightFlushGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeFourAces ||
				cardType == TypeFourOfAKind {
				for _, group := range cardsGroup.fourOfAKindGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						deck.DrawSpecificCards(group)
						return group
					}
				}
			} else if cardType == TypeFullHouse {
				for _, group := range cardsGroup.fullHouseGroups {
					if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
						deck.DrawSpecificCards(group)
						return group
					}
				}
			} else if cardType == TypeFlush {
				for _, group := range cardsGroup.flushGroups {
					if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
						deck.DrawSpecificCards(group)
						return group
					}
				}
			} else if cardType == TypeStraight {
				for _, group := range cardsGroup.straightGroups {
					if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
						deck.DrawSpecificCards(group)
						return group
					}
				}
			} else if cardType == TypeThreeOfAces ||
				cardType == TypeThreeOfAKind {
				for _, group := range cardsGroup.threeOfAKindGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						deck.DrawSpecificCards(group)
						return group
					}
				}
			} else {
				// ignore too small type
			}
		}
	} else if position == MiddlePart {
		bottomCards := cardsData[BottomPart]
		bottomType := gameInstance.getTypeOfCards(bottomCards)
		bottomTypeOrder := gameInstance.getTypeOrder(bottomType)
		alteredBetterTypes := make([]string, 0)
		for _, cardsType := range betterTypes {
			typeOrder := gameInstance.getTypeOrder(cardsType)
			if bottomTypeOrder >= typeOrder {
				alteredBetterTypes = append(alteredBetterTypes, cardsType)
			}
		}

		sort.Sort(ByRandom(alteredBetterTypes))
		for _, cardType := range alteredBetterTypes {
			if cardType == TypeStraightFlushTop ||
				cardType == TypeStraightFlushBottom ||
				cardType == TypeStraightFlush {
				for _, group := range cardsGroup.straightFlushGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
							if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
								deck.DrawSpecificCards(group)
								return group
							}
						}
					}
				}
			} else if cardType == TypeFourAces ||
				cardType == TypeFourOfAKind {
				for _, group := range cardsGroup.fourOfAKindGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeFullHouse {
				for _, group := range cardsGroup.fullHouseGroups {
					if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
						if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeFlush {
				for _, group := range cardsGroup.flushGroups {
					if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
						if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeStraight {
				for _, group := range cardsGroup.straightGroups {
					if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
						if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeThreeOfAces ||
				cardType == TypeThreeOfAKind {
				for _, group := range cardsGroup.threeOfAKindGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(bottomCards, group) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypeTwoPair {
				for _, group := range cardsGroup.twoPairGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(bottomCards, group) >= 0 { // two pair can be the same in both group
							if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
								deck.DrawSpecificCards(group)
								return group
							}
						}
					}
				}
			} else {
				// ignore too small group
			}
		}
	} else if position == TopPart {
		middleCards := cardsData[MiddlePart]
		middleType := gameInstance.getTypeOfCards(middleCards)
		middleTypeOrder := gameInstance.getTypeOrder(middleType)
		alteredBetterTypes := make([]string, 0)
		for _, cardsType := range betterTypes {
			typeOrder := gameInstance.getTypeOrder(cardsType)
			if middleTypeOrder >= typeOrder {
				alteredBetterTypes = append(alteredBetterTypes, cardsType)
			}
		}

		sort.Sort(ByRandom(alteredBetterTypes))
		for _, cardType := range alteredBetterTypes {
			if cardType == TypeThreeOfAces ||
				cardType == TypeThreeOfAKind {
				for _, group := range cardsGroup.threeOfAKindGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(middleCards, group) > 0 {
							deck.DrawSpecificCards(group)
							return group
						}
					}
				}
			} else if cardType == TypePair {
				for _, group := range cardsGroup.pairGroups {
					if gameInstance.getTypeOfCards(group) == cardType {
						if gameInstance.getCompareBetweenCards(middleCards, group) > 0 {
							if gameInstance.getCompareBetweenCards(group, betterCards) > 0 {
								deck.DrawSpecificCards(group)
								return group
							}
						}
					}
				}
			} else if cardType == TypeHighCard {
				// ignore high cards, will auto put stuff in later
			}
		}
	}
	return []string{}
}
