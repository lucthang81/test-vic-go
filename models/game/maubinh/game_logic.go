package maubinh

import (
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/utils"
	"sort"
)

const (
	TypeStraightFlushTop    = "straight_flush_top"
	TypeStraightFlushBottom = "straight_flush_bottom"
	TypeStraightFlush       = "straight_flush"
	TypeFourAces            = "four_aces"
	TypeFourOfAKind         = "four_of_a_kind"
	TypeFullHouse           = "fullhouse"
	TypeFlush               = "flush"
	TypeStraight            = "straight"
	TypeThreeOfAKind        = "three_of_a_kind"
	TypeThreeOfAces         = "three_aces"
	TypeTwoPair             = "two_pair"
	TypePair                = "pair"
	TypeHighCard            = "high_card"

	WhiteWinTypeDragonRollingStraight   = "dragon_rolling_straight"
	WhiteWinTypeDragonStraight          = "dragon_straight"
	WhiteWinTypeFivePairOneThreeOfAKind = "5_pair_1_three_of_a_kind"
	WhiteWinTypeThreeFlush              = "3_flush"
	WhiteWinTypeThreeStraight           = "3_straight"
	WhiteWinTypeSixPair                 = "6_pair"

	// WhiteWinTypeSameColor        = "same_color"
	// WhiteWinType12SameColor      = "12_same_color"
	// WhiteWinTypeFourThreeOfAKind = "4_three_of_a_kind"
	// WhiteWinTypeThreeOfAKindOnly = "3_of_a_kind_only"
	// WhiteWinTypeFivePairStreak   = "5_pair_streak"

	TopPart    = "top" // 3cards
	MiddlePart = "middle"
	BottomPart = "bottom"
)

type ByCardSuitAndValue struct {
	cards        []string
	gameInstance *MauBinhGame
}

func (a *ByCardSuitAndValue) Len() int      { return len(a.cards) }
func (a *ByCardSuitAndValue) Swap(i, j int) { a.cards[i], a.cards[j] = a.cards[j], a.cards[i] }
func (a *ByCardSuitAndValue) Less(i, j int) bool {
	suitI, valueI := components.SuitAndValueFromCard(a.cards[i])
	suitJ, valueJ := components.SuitAndValueFromCard(a.cards[j])
	valueIAsInt := a.gameInstance.valueAsInt(valueI)
	valueJAsInt := a.gameInstance.valueAsInt(valueJ)
	if valueIAsInt == valueJAsInt {
		suitIAsInt := a.gameInstance.suitAsInt(suitI)
		suitJAsInt := a.gameInstance.suitAsInt(suitJ)
		return suitIAsInt < suitJAsInt
	}
	return valueIAsInt < valueJAsInt
}

func (gameInstance *MauBinhGame) sortCards(cards []string) []string {
	sortedCards := make([]string, 0)
	for _, card := range cards {
		sortedCards = append(sortedCards, card)
	}

	obj := &ByCardSuitAndValue{
		cards:        sortedCards,
		gameInstance: gameInstance,
	}

	sort.Sort(obj)
	return obj.cards
}

// min 0 max 12
func (gameInstance *MauBinhGame) valueAsInt(value string) int {
	var counter int
	for _, valueOrder := range gameInstance.logicInstance.CardValueOrder() {
		if valueOrder == value {
			return counter
		}
		counter++
	}
	return counter
}

func (gameInstance *MauBinhGame) suitAsInt(suit string) int {
	var counter int
	for _, suitOrder := range gameInstance.logicInstance.CardSuitOrder() {
		if suitOrder == suit {
			return counter
		}
		counter++
	}
	return counter
}

func (gameInstance *MauBinhGame) getWhiteWinMultiplierBetweenCards(cards1 []string, cards2 []string) float64 {
	var multiplier float64
	compare := gameInstance.getCompareBetweenWhiteWinCards(cards1, cards2)
	whiteWinType1 := gameInstance.getTypeOfWhiteWin(cards1)
	whiteWinType2 := gameInstance.getTypeOfWhiteWin(cards2)

	if compare > 0 {
		multiplier = gameInstance.getWhiteWinMultiplier(whiteWinType1)
	} else if compare < 0 {
		multiplier = -gameInstance.getWhiteWinMultiplier(whiteWinType2)
	} else {
		multiplier = 0
	}
	return multiplier
}

func (gameInstance *MauBinhGame) getMultiplierOfCardsData(cardsData map[string][]string) float64 {
	var totalMultiplier float64
	for _, positionString := range []string{TopPart, MiddlePart, BottomPart} {
		cards := cardsData[positionString]
		totalMultiplier += gameInstance.getMultiplierOfCards(cards, positionString)
	}
	return totalMultiplier
}

func (gameInstance *MauBinhGame) getMultiplierBetweenCardsData(cardsData map[string][]string, otherCardsData map[string][]string, positionString string) float64 {
	var multiplier float64
	if gameInstance.isCardsDataValid(cardsData) && !gameInstance.isCardsDataValid(otherCardsData) {
		// "player" win, "otherPlayer" lose
		multiplier = gameInstance.getMultiplierOfCards(cardsData[positionString], positionString)
	}

	if !gameInstance.isCardsDataValid(cardsData) && gameInstance.isCardsDataValid(otherCardsData) {
		// "player" lose, "otherPlayer" win
		multiplier = -gameInstance.getMultiplierOfCards(otherCardsData[positionString], positionString)
	}

	// both valid
	if gameInstance.isCardsDataValid(cardsData) && gameInstance.isCardsDataValid(otherCardsData) {
		cards1 := cardsData[positionString]
		cards2 := otherCardsData[positionString]
		multiplier = gameInstance.getMultiplierBetweenCards(cards1, cards2, positionString)

		typeOfCards1 := gameInstance.getTypeOfCards(cards1)
		typeOfCards2 := gameInstance.getTypeOfCards(cards2)
		if positionString == TopPart {
			if utils.ContainsByString([]string{TypeThreeOfAKind, TypeThreeOfAces}, typeOfCards1) &&
				utils.ContainsByString([]string{TypeThreeOfAKind, TypeThreeOfAces}, typeOfCards2) {
				multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
			}
		} else {
			if utils.ContainsByString([]string{TypeFourOfAKind, TypeFourAces}, typeOfCards1) &&
				utils.ContainsByString([]string{TypeFourOfAKind, TypeFourAces}, typeOfCards2) {
				multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
			} else if utils.ContainsByString([]string{TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop}, typeOfCards1) &&
				utils.ContainsByString([]string{TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop}, typeOfCards2) {
				multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
			}
		}
	}

	return multiplier
}

func (gameInstance *MauBinhGame) getTotalMultiplierBetweenCards(cards1 []string, cards2 []string, positionString string) float64 {
	multiplier := gameInstance.getMultiplierBetweenCards(cards1, cards2, positionString)

	typeOfCards1 := gameInstance.getTypeOfCards(cards1)
	typeOfCards2 := gameInstance.getTypeOfCards(cards2)
	if positionString == TopPart {
		if utils.ContainsByString([]string{TypeThreeOfAKind, TypeThreeOfAces}, typeOfCards1) &&
			utils.ContainsByString([]string{TypeThreeOfAKind, TypeThreeOfAces}, typeOfCards2) {
			multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
		}
	} else {
		if utils.ContainsByString([]string{TypeFourOfAKind, TypeFourAces}, typeOfCards1) &&
			utils.ContainsByString([]string{TypeFourOfAKind, TypeFourAces}, typeOfCards2) {
			multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
		} else if utils.ContainsByString([]string{TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop}, typeOfCards1) &&
			utils.ContainsByString([]string{TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop}, typeOfCards2) {
			multiplier = multiplier * gameInstance.logicInstance.SpecialCompareMultiplier()
		}
	}
	return multiplier
}

func (gameInstance *MauBinhGame) getIsCollapsingBetweenCardsData(cardsData map[string][]string, otherCardsData map[string][]string) bool {
	if !gameInstance.isCardsDataValid(cardsData) || !gameInstance.isCardsDataValid(otherCardsData) {
		// bing lủng không tính sập hầm
		return false
	}

	loseCount := 0
	winCount := 0
	for _, position := range []string{BottomPart, MiddlePart, TopPart} {
		multiplier := gameInstance.getMultiplierBetweenCardsData(cardsData, otherCardsData, position)
		if multiplier < 0 {
			loseCount++
		} else if multiplier > 0 {
			winCount++
		}
	}
	return loseCount == 3 || winCount == 3

}

func (gameInstance *MauBinhGame) getIsCollapsingBetweenCardsData2(cardsData map[string][]string, otherCardsData map[string][]string) int {
	if !gameInstance.isCardsDataValid(cardsData) && !gameInstance.isCardsDataValid(otherCardsData) {

		return 0
	}
	if !gameInstance.isCardsDataValid(cardsData) {

		return -1
	}
	if !gameInstance.isCardsDataValid(otherCardsData) {
		return 1
	}
	loseCount := 0
	winCount := 0
	for _, position := range []string{BottomPart, MiddlePart, TopPart} {
		multiplier := gameInstance.getMultiplierBetweenCardsData(cardsData, otherCardsData, position)
		if multiplier < 0 {
			loseCount++
		} else if multiplier > 0 {
			winCount++
		}
	}
	if winCount == 3 {
		return 1 // thang ca 3 chi
	} else {
		if loseCount == 3 {
			return -1
		}
	}
	return 0 // loseCount == 3 || winCount == 3

}

func (gameInstance *MauBinhGame) getMultiplierOfCards(cards []string, positionString string) float64 {
	typeOfCards := gameInstance.getTypeOfCards(cards)
	var multiplier float64
	if positionString == TopPart {
		multiplier = gameInstance.logicInstance.WinMultiplierForTopPart()[typeOfCards]
		if multiplier == 0 {
			multiplier = 1
		}
	} else if positionString == MiddlePart {
		multiplier = gameInstance.logicInstance.WinMultiplierForMiddlePart()[typeOfCards]
		if multiplier == 0 {
			multiplier = 1
		}
	} else if positionString == BottomPart {
		multiplier = gameInstance.logicInstance.WinMultiplierForBottomPart()[typeOfCards]
		if multiplier == 0 {
			multiplier = 1
		}
	}
	return multiplier
}

func (gameInstance *MauBinhGame) isCardsDataValid(cardsData map[string][]string) bool {
	if cardsData == nil {
		return false
	}
	lastScore := 0
	var lastCards []string
	for _, positionString := range []string{TopPart, MiddlePart, BottomPart} {
		cards := gameInstance.sortCards(cardsData[positionString])
		typeOfCards := gameInstance.getTypeOfCards(cards)
		typeOrder := gameInstance.getTypeOrder(typeOfCards)
		// fmt.Println("valid", positionString, typeOfCards, typeOrder)
		if lastScore < typeOrder {
			lastScore = typeOrder
		} else if lastScore > typeOrder {
			return false
		} else {
			compare := gameInstance.getCompareBetweenCards(lastCards, cards)
			if compare > 0 {
				return false
			}
		}

		if positionString == TopPart && len(cards) != 3 {
			return false
		}

		if positionString == MiddlePart && len(cards) != 5 {
			return false
		}

		if positionString == BottomPart && len(cards) != 5 {
			return false
		}

		lastCards = cards
	}
	return true
}

func (gameInstance *MauBinhGame) getCompareBetweenCards(cards1 []string, cards2 []string) int {
	typeOfCards1 := gameInstance.getTypeOfCards(cards1)
	typeOfCards2 := gameInstance.getTypeOfCards(cards2)

	orderScore1 := gameInstance.getTypeOrder(typeOfCards1)
	orderScore2 := gameInstance.getTypeOrder(typeOfCards2)

	var compare int

	if orderScore1 == orderScore2 {
		//draw order, check card value
		if typeOfCards1 == TypeStraightFlushTop {
			compare = 0
		} else {
			if typeOfCards1 == TypeFlush ||
				typeOfCards1 == TypeHighCard {
				// compare each cards
				for i := 1; i <= utils.MinInt(len(cards1), len(cards2)); i++ {
					card1 := cards1[len(cards1)-i]
					card2 := cards2[len(cards2)-i]
					value1 := gameInstance.getValueAsIntFromCard(card1)
					value2 := gameInstance.getValueAsIntFromCard(card2)

					if value1 > value2 {
						compare = 1
						break
					} else if value1 < value2 {
						compare = -1
						break
					} else {
						continue
					}
				}
			} else if typeOfCards1 == TypeStraight ||
				typeOfCards1 == TypeStraightFlush {
				var cards1IsStraightBottom bool
				var cards2IsStraightBottom bool
				if containValue(cards1, "a") &&
					containValue(cards1, "2") {
					cards1IsStraightBottom = true
				}

				if containValue(cards2, "a") &&
					containValue(cards2, "2") {
					cards2IsStraightBottom = true
				}

				if cards1IsStraightBottom && cards2IsStraightBottom {
					compare = 0
				} else if !cards1IsStraightBottom && cards2IsStraightBottom {
					compare = 1
				} else if cards1IsStraightBottom && !cards2IsStraightBottom {
					compare = -1
				} else if !cards1IsStraightBottom && !cards2IsStraightBottom {
					for i := 1; i <= utils.MinInt(len(cards1), len(cards2)); i++ {
						card1 := cards1[len(cards1)-i]
						card2 := cards2[len(cards2)-i]
						value1 := gameInstance.getValueAsIntFromCard(card1)
						value2 := gameInstance.getValueAsIntFromCard(card2)

						if value1 > value2 {
							compare = 1
							break
						} else if value1 < value2 {
							compare = -1
							break
						} else {
							continue
						}
					}
				}

			} else if typeOfCards1 == TypeTwoPair ||
				typeOfCards1 == TypePair {
				biggestCard1 := getBiggestCardInTypeForCards(cards1, typeOfCards1)
				biggestCard2 := getBiggestCardInTypeForCards(cards2, typeOfCards2)

				value1 := gameInstance.getValueAsIntFromCard(biggestCard1)
				value2 := gameInstance.getValueAsIntFromCard(biggestCard2)

				if value1 > value2 {
					compare = 1
				} else if value1 < value2 {
					compare = -1
				} else {
					_, valueString := components.SuitAndValueFromCard(biggestCard1)
					// remove the biggest cards and continue to compare
					biggestPair1 := getCardsWithValue(cards1, valueString)
					biggestPair2 := getCardsWithValue(cards2, valueString)

					temp1 := removeCardsFromCards(cards1, biggestPair1)
					temp2 := removeCardsFromCards(cards2, biggestPair2)

					// run this again, make it like comparing top part with 3 cards
					tempCompare := gameInstance.getCompareBetweenCards(temp1, temp2)
					if tempCompare == 0 {
						compare = 0
					} else if tempCompare > 0 {
						compare = 1
					} else if tempCompare < 0 {
						compare = -1
					}
				}

			} else {
				biggestCard1 := getBiggestCardInTypeForCards(cards1, typeOfCards1)
				biggestCard2 := getBiggestCardInTypeForCards(cards2, typeOfCards2)

				value1 := gameInstance.getValueAsIntFromCard(biggestCard1)
				value2 := gameInstance.getValueAsIntFromCard(biggestCard2)

				if value1 > value2 {
					compare = 1
				} else if value1 < value2 {
					compare = -1
				} else {
					compare = 0

				}
			}
		}

	} else if orderScore1 > orderScore2 {
		compare = 1
	} else {
		compare = -1
	}

	return compare
}

func (gameInstance *MauBinhGame) getCompareBetweenWhiteWinCards(cards1 []string, cards2 []string) int {
	typeOfCards1 := gameInstance.getTypeOfWhiteWin(cards1)
	typeOfCards2 := gameInstance.getTypeOfWhiteWin(cards2)

	multiplier1 := gameInstance.getWhiteWinMultiplier(typeOfCards1)
	multiplier2 := gameInstance.getWhiteWinMultiplier(typeOfCards2)

	var compare int

	if multiplier1 == multiplier2 {
		//draw order, check card value
		if typeOfCards1 == WhiteWinTypeDragonRollingStraight ||
			typeOfCards1 == WhiteWinTypeDragonStraight {
			compare = 0
		} else {
			if typeOfCards1 == WhiteWinTypeFivePairOneThreeOfAKind ||
				typeOfCards1 == WhiteWinTypeThreeFlush ||
				typeOfCards1 == WhiteWinTypeThreeStraight {
				// compare each cards
				for i := 1; i <= utils.MinInt(len(cards1), len(cards2)); i++ {
					card1 := cards1[len(cards1)-i]
					card2 := cards2[len(cards2)-i]
					value1 := gameInstance.getValueAsIntFromCard(card1)
					value2 := gameInstance.getValueAsIntFromCard(card2)

					if value1 > value2 {
						compare = 1
						break
					} else if value1 < value2 {
						compare = -1
						break
					} else {
						continue
					}
				}
			} else if typeOfCards1 == WhiteWinTypeSixPair {
				biggestCard1 := getBiggestCardInGroupOfSameCardValueFromCards(cards1, 2)
				biggestCard2 := getBiggestCardInGroupOfSameCardValueFromCards(cards2, 2)

				value1 := gameInstance.getValueAsIntFromCard(biggestCard1)
				value2 := gameInstance.getValueAsIntFromCard(biggestCard2)

				if value1 > value2 {
					compare = 1
				} else if value1 < value2 {
					compare = -1
				} else {
					compare = 0
				}

			}
		}

	} else if multiplier1 > multiplier2 {
		compare = 1
	} else {
		compare = -1
	}

	return compare
}

func (gameInstance *MauBinhGame) getTypeOrder(typeOfCards string) int {
	for index, typeInOrder := range gameInstance.logicInstance.TypeOrder() {
		if typeInOrder == typeOfCards {
			return index
		}
	}
	return -1
}

func (gameInstance *MauBinhGame) getMultiplierBetweenCards(cards1 []string, cards2 []string, positionString string) float64 {
	var multiplier float64
	compare := gameInstance.getCompareBetweenCards(cards1, cards2)

	if compare > 0 {
		multiplier = gameInstance.getMultiplierOfCards(cards1, positionString)
	} else if compare < 0 {
		multiplier = -gameInstance.getMultiplierOfCards(cards2, positionString)
	} else {
		multiplier = 0
	}

	return multiplier
}

func (gameInstance *MauBinhGame) getWhiteWinMultiplier(whiteWinType string) float64 {
	if whiteWinType == "" {
		return 0
	}

	multiplier := gameInstance.logicInstance.WhiteWinMultiplier()[whiteWinType]
	if multiplier == 0 {
		return 1
	} else {
		return multiplier
	}
}

func (gameInstance *MauBinhGame) valueOfType(typeOfCards string) int {
	for index, typeString := range gameInstance.logicInstance.TypeOrder() {
		if typeString == typeOfCards {
			return index
		}
	}
	return -1
}

func (gameInstance *MauBinhGame) getTypeOfWhiteWin(sortedCards []string) string {
	if gameInstance.isDragonRollingStraight(sortedCards) {
		return WhiteWinTypeDragonRollingStraight
	}

	if gameInstance.isDragonStraight(sortedCards) {
		return WhiteWinTypeDragonStraight
	}

	if isFivePairOneThreeOfAKind(sortedCards) {
		return WhiteWinTypeFivePairOneThreeOfAKind
	}

	if isThreeFlush(sortedCards) {
		return WhiteWinTypeThreeFlush
	}

	if gameInstance.isThreeStraight(sortedCards) {
		return WhiteWinTypeThreeStraight
	}

	if isSixPair(sortedCards) {
		return WhiteWinTypeSixPair
	}
	return ""

	// if isSameColor(sortedCards) {
	// 	return WhiteWinTypeSameColor
	// }

	// if is12SameColor(sortedCards) {
	// 	return WhiteWinType12SameColor
	// }

	// if isFourThreeOfAKind(sortedCards) {
	// 	return WhiteWinTypeFourThreeOfAKind
	// }
	// if gameInstance.isThreeOfAKindOnly(sortedCards) {
	// 	return WhiteWinTypeThreeOfAKindOnly
	// }

	// if gameInstance.isFivePairStreakOnly(sortedCards) {
	// 	return WhiteWinTypeFivePairStreak
	// }

	return ""
}
func organizeCardsInOrder(sortedCards []string, whiteWinType string) map[string][]string {
	cardData := make(map[string][]string)
	cardData[TopPart] = sortedCards[:3]
	cardData[MiddlePart] = sortedCards[3:8]
	cardData[BottomPart] = sortedCards[8:13]
	return cardData
}

func (gameInstance *MauBinhGame) organizeCardsForWhiteWin(sortedCards []string, whiteWinType string) map[string][]string {
	cardData := make(map[string][]string)
	if utils.ContainsByString([]string{WhiteWinTypeDragonRollingStraight,
		WhiteWinTypeDragonStraight,
		WhiteWinTypeSixPair}, whiteWinType) {
		cardData[TopPart] = sortedCards[:3]
		cardData[MiddlePart] = sortedCards[3:8]
		cardData[BottomPart] = sortedCards[8:13]
		return cardData
	}

	if utils.ContainsByString([]string{WhiteWinTypeFivePairOneThreeOfAKind}, whiteWinType) {
		data := make(map[string]int)
		for _, cardString := range sortedCards {
			_, value := components.SuitAndValueFromCard(cardString)
			data[value]++
		}

		var sameValue string
		for value, numCards := range data {
			if numCards == 3 {
				sameValue = value
			}
		}
		if sameValue != "" {
			cardData[TopPart] = make([]string, 0)
			cardData[MiddlePart] = make([]string, 0)
			cardData[BottomPart] = make([]string, 0)
			for _, card := range sortedCards {
				_, value := components.SuitAndValueFromCard(card)
				if value == sameValue {
					cardData[TopPart] = append(cardData[TopPart], card)
				} else {
					if len(cardData[MiddlePart]) < 5 {
						cardData[MiddlePart] = append(cardData[MiddlePart], card)
					} else {
						cardData[BottomPart] = append(cardData[BottomPart], card)
					}
				}
			}
		}
		return cardData
	}

	if whiteWinType == WhiteWinTypeThreeFlush {
		data := make(map[string]int)
		for _, cardString := range sortedCards {
			suit, _ := components.SuitAndValueFromCard(cardString)
			data[suit]++
		}

		var same3Value string
		var same5Value1 string
		var same5Value2 string
		for value, numCards := range data {
			if numCards == 3 {
				same3Value = value
			} else if numCards == 5 {
				if same5Value1 == "" {
					same5Value1 = value
				} else {
					same5Value2 = value
				}
			} else if numCards == 10 {
				same5Value1 = value
				same5Value2 = value
			} else if numCards == 8 {
				same3Value = value
				if same5Value1 == "" {
					same5Value1 = value
				} else {
					same5Value2 = value
				}
			}
		}

		cardData[TopPart] = make([]string, 0)
		cardData[MiddlePart] = make([]string, 0)
		cardData[BottomPart] = make([]string, 0)
		for _, card := range sortedCards {
			suit, _ := components.SuitAndValueFromCard(card)
			if suit == same3Value {
				cardData[TopPart] = append(cardData[TopPart], card)
			} else if suit == same5Value1 {
				cardData[MiddlePart] = append(cardData[MiddlePart], card)
			} else if suit == same5Value2 {
				cardData[BottomPart] = append(cardData[BottomPart], card)

			}
		}

		return cardData
	}

	if whiteWinType == WhiteWinTypeThreeStraight {
		straights3 := gameInstance.getStraightWithNumberOfCards(sortedCards, 3)
		for _, straight3 := range straights3 {
			temp3 := removeCardsFromCards(sortedCards, straight3)
			straights5 := gameInstance.getStraightWithNumberOfCards(temp3, 5)
			for _, straight5 := range straights5 {
				temp5 := removeCardsFromCards(temp3, straight5)
				secondStraights5 := gameInstance.getStraightWithNumberOfCards(temp5, 5)
				if len(secondStraights5) > 0 {
					cardData[TopPart] = straight3
					cardData[MiddlePart] = straight5
					cardData[BottomPart] = secondStraights5[0]
				}
			}
		}
		return cardData
	}

	return nil
}

func (gameInstance *MauBinhGame) getTypeOfCards(sortedCards []string) string {
	if gameInstance.isStraightFlushTop(sortedCards) {
		return TypeStraightFlushTop
	}

	if gameInstance.isStraightFlushBottom(sortedCards) {
		return TypeStraightFlushBottom
	}

	if gameInstance.isStraightFlush(sortedCards) {
		return TypeStraightFlush
	}

	if isFourAces(sortedCards) {
		return TypeFourAces
	}

	if isSameCards(sortedCards, 4) {
		return TypeFourOfAKind
	}

	if isFullHouse(sortedCards) {
		return TypeFullHouse
	}

	if isFlush(sortedCards) {
		return TypeFlush
	}

	if gameInstance.isStraight(sortedCards) {
		return TypeStraight
	}

	if isThreeAces(sortedCards) {
		return TypeThreeOfAces
	}

	if isSameCards(sortedCards, 3) {
		return TypeThreeOfAKind
	}

	if isTwoPair(sortedCards) {
		return TypeTwoPair
	}

	if isSameCards(sortedCards, 2) {
		return TypePair
	}

	return TypeHighCard
}

func getBiggestCardInTypeForCards(sortedCards []string, typeOfCards string) string {
	if typeOfCards == TypeStraightFlushTop ||
		typeOfCards == TypeStraightFlush ||
		typeOfCards == TypeFlush {
		return sortedCards[len(sortedCards)-1]
	}

	if typeOfCards == TypeStraightFlushBottom {
		return sortedCards[len(sortedCards)-2] // cause a is now bottom
	}

	if typeOfCards == TypeStraight {
		if containValue(sortedCards, "a") &&
			containValue(sortedCards, "2") {
			return sortedCards[len(sortedCards)-2] // cause a is now bottom
		} else {
			return sortedCards[len(sortedCards)-1]
		}
	}

	if typeOfCards == TypeFourAces ||
		typeOfCards == TypeFourOfAKind {
		return getBiggestCardInGroupOfSameCardValueFromCards(sortedCards, 4)
	}

	if typeOfCards == TypeFullHouse {
		return getBiggestCardInGroupOfSameCardValueFromCards(sortedCards, 3)
	}

	if typeOfCards == TypeThreeOfAKind {

		return getBiggestCardInGroupOfSameCardValueFromCards(sortedCards, 3)
	}

	if typeOfCards == TypeTwoPair {
		return getBiggestCardInGroupOfSameCardValueFromCards(sortedCards, 2)
	}

	if typeOfCards == TypePair {
		return getBiggestCardInGroupOfSameCardValueFromCards(sortedCards, 2)
	}

	return sortedCards[len(sortedCards)-1]
}

func (gameInstance *MauBinhGame) isDragonRollingStraight(sortedCards []string) bool {
	if gameInstance.isStraightFlush(sortedCards) {
		return true
	}
	return false
}

func (gameInstance *MauBinhGame) isDragonStraight(sortedCards []string) bool {
	if gameInstance.isStraight(sortedCards) {
		return true
	}
	return false
}

func isSameColor(sortedCards []string) bool {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if suit == "c" || suit == "s" {
			data["black"]++
		} else {
			data["red"]++
		}

	}
	for _, numCards := range data {
		if numCards == 13 {
			return true
		}
	}
	return false
}

func (gameInstance *MauBinhGame) isThreeOfAKindOnly(sortedCards []string) bool {
	if getNumberOfSameCards(sortedCards, 3) == 1 {
		if getNumberOfSameCards(sortedCards, 2) == 0 { // no pair
			if len(gameInstance.getStraightWithNumberOfCards(sortedCards, 5)) == 0 { // no streak
				if getNumberOfSameSuitCards(sortedCards, 5) == 0 { // no flush
					if getNumberOfSameCards(sortedCards, 4) == 0 { // no quadruple
						return true
					}
				}
			}

		}
	}
	return false
}

func (gameInstance *MauBinhGame) isFivePairStreakOnly(sortedCards []string) bool {
	return gameInstance.containXDoubleCardsInStreak(sortedCards, 5)
}

func is12SameColor(sortedCards []string) bool {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if suit == "c" || suit == "s" {
			data["black"]++
		} else {
			data["red"]++
		}

	}
	for _, numCards := range data {
		if numCards == 12 {
			return true
		}
	}
	return false
}

func isFivePairOneThreeOfAKind(sortedCards []string) bool {
	if getNumberOfSameCards(sortedCards, 2) == 5 &&
		getNumberOfSameCards(sortedCards, 3) == 1 {
		return true
	}
	return false
}

func isFourThreeOfAKind(sortedCards []string) bool {
	if getNumberOfSameCards(sortedCards, 3) == 4 {
		return true
	}
	return false
}

func isSixPair(sortedCards []string) bool {
	if getNumberOfSameCards(sortedCards, 2) == 6 {
		return true
	}
	return false
}

func isThreeFlush(sortedCards []string) bool {
	if (getNumberOfSameSuitCards(sortedCards, 5) == 2 &&
		getNumberOfSameSuitCards(sortedCards, 3) == 1) ||
		(getNumberOfSameSuitCards(sortedCards, 10) == 1 &&
			getNumberOfSameSuitCards(sortedCards, 3) == 1) ||
		(getNumberOfSameSuitCards(sortedCards, 5) == 1 &&
			getNumberOfSameSuitCards(sortedCards, 8) == 1) {
		return true
	}
	return false
}

func (gameInstance *MauBinhGame) isThreeStraight(sortedCards []string) bool {
	straights3 := gameInstance.getStraightWithNumberOfCards(sortedCards, 3)
	for _, straight3 := range straights3 {
		temp3 := removeCardsFromCards(sortedCards, straight3)
		straights5 := gameInstance.getStraightWithNumberOfCards(temp3, 5)
		for _, straight5 := range straights5 {
			temp5 := removeCardsFromCards(temp3, straight5)
			secondStraights5 := gameInstance.getStraightWithNumberOfCards(temp5, 5)
			if len(secondStraights5) > 0 {
				return true
			}
		}
	}
	return false
}

func isFourAces(sortedCards []string) bool {
	counter := 0
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "a" {
			counter++
		}
	}
	if counter == 4 {
		return true
	}
	return false
}

func isThreeAces(sortedCards []string) bool {
	counter := 0
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "a" {
			counter++
		}
	}
	if counter == 3 {
		return true
	}
	return false
}

func isTwoPair(sortedCards []string) bool {
	if getNumberOfSameCards(sortedCards, 2) == 2 {
		return true
	}
	return false
}

func getNumberOfSameCards(sortedCards []string, numberOfSameCards int) int {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		data[value]++
	}
	counter := 0
	for _, numCards := range data {
		if numCards == numberOfSameCards {
			counter++
		}
	}
	return counter
}

func getNumberOfSameSuitCards(sortedCards []string, numberOfSameSuitCards int) int {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		data[suit]++
	}
	counter := 0
	for _, numCards := range data {
		if numCards == numberOfSameSuitCards {
			counter++
		}
	}
	return counter
}

func getBiggestCardInGroupOfSameCardValueFromCards(sortedCards []string, groupSize int) string {
	data := make(map[string]int)
	biggestCard := ""
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		data[value]++
		if data[value] == groupSize {
			biggestCard = cardString
		}
	}
	return biggestCard
}

func (gameInstance *MauBinhGame) getStraightWithNumberOfCards(sortedCards []string, numberOfCards int) [][]string {
	results := make([][]string, 0)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)

		straight := gameInstance.getStraightStartWithCard(sortedCards, cardString, numberOfCards)
		if straight != nil {
			results = append(results, straight)
		}
		if value == "a" {
			straightBottom := make([]string, 0)
			straightBottom = append(straightBottom, cardString)
			for i := 0; i < numberOfCards-1; i++ {
				valueToFind := gameInstance.logicInstance.CardValueOrder()[i]
				cardToFind := getCardWithValue(sortedCards, valueToFind)
				if cardToFind != "" {
					straightBottom = append(straightBottom, cardToFind)
				} else {
					break
				}

			}
			if len(straightBottom) == numberOfCards {
				results = append(results, straightBottom)
			}
		}
	}

	return results
}

func (gameInstance *MauBinhGame) getStraightStartWithCard(sortedCards []string, card string, length int) (streak []string) {
	streak = []string{card}
	for i := 0; i < length-1; i++ {
		cardInStreak := gameInstance.getCardHasValueMoreThanCard(sortedCards, card, i+1)
		if cardInStreak != "" {
			streak = append(streak, cardInStreak)

			if len(streak) == length {
				return streak
			}
		} else {
			return nil
		}
	}
	return nil
}

func getNumberOfAcesInCards(cards []string) int {
	var counter int
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "a" {
			counter++
		}
	}
	return counter
}

func (gameInstance *MauBinhGame) getValueAsIntFromCard(cardString string) int {
	_, value := components.SuitAndValueFromCard(cardString)
	return gameInstance.valueAsInt(value)
}

func (gameInstance *MauBinhGame) getCardHasValueMoreThanCard(cards []string, card string, moreBy int) string {
	valueAsInt := gameInstance.getValueAsIntFromCard(card)
	for _, cardInHand := range cards {
		valueAsIntHere := gameInstance.getValueAsIntFromCard(cardInHand)
		if valueAsIntHere == valueAsInt+moreBy {
			return cardInHand
		}
	}
	return ""
}

func isFlush(sortedCards []string) bool {
	if len(sortedCards) <= 3 {
		return false
	}
	currentSuit := ""
	for _, cardString := range sortedCards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if currentSuit == "" {
			currentSuit = suit
		} else if currentSuit != suit {
			return false
		}
	}
	return true
}

func isFullHouse(sortedCards []string) bool {
	if len(sortedCards) != 5 {
		return false
	}
	containAThreeOfAKind := isSameCards(sortedCards, 3)
	containAPair := isSameCards(sortedCards, 2)
	if containAPair && containAThreeOfAKind {
		return true
	}
	return false
}

func isSameCards(sortedCards []string, numberOfSameCards int) bool {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		data[value]++
	}
	for _, numCards := range data {
		if numCards == numberOfSameCards {
			return true
		}
	}
	return false
}

func (gameInstance *MauBinhGame) isStraightFlushBottom(sortedCards []string) bool {
	if gameInstance.isStraight(sortedCards) && isFlush(sortedCards) {
		if containValue(sortedCards, "a") &&
			containValue(sortedCards, "2") {
			return true
		}
		return false
	} else {
		return false
	}
}

func (gameInstance *MauBinhGame) isStraightFlushTop(sortedCards []string) bool {
	if gameInstance.isStraight(sortedCards) && isFlush(sortedCards) {
		if containValue(sortedCards, "a") &&
			containValue(sortedCards, "k") {
			return true
		}
		return false
	} else {
		return false
	}
}

func (gameInstance *MauBinhGame) isStraightFlush(sortedCards []string) bool {
	if gameInstance.isStraight(sortedCards) && isFlush(sortedCards) {
		return true
	} else {
		return false
	}

}

func (gameInstance *MauBinhGame) isStraight(sortedCards []string) bool {
	if len(sortedCards) <= 3 {
		return false
	}

	if getNumberOfSameCards(sortedCards, 2) > 0 {
		return false
	}

	// check for a
	lastCardString := sortedCards[len(sortedCards)-1]
	_, lastValue := components.SuitAndValueFromCard(lastCardString)
	if lastValue == "a" {
		// need check reverse straight too
		firstCardString := sortedCards[0]
		_, firstValue := components.SuitAndValueFromCard(firstCardString)
		if firstValue == "2" {
			if len(sortedCards) == 2 {
				return true
			}
			temp := make([]string, 0)
			for index, cardString := range sortedCards {
				if index != len(sortedCards)-1 {
					temp = append(temp, cardString)
				}
			}

			if gameInstance.isStraight(temp) {
				return true
			} else {
				return false
			}
		}
	}

	currentValueInt := -1
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		valueInt := gameInstance.valueAsInt(value)
		if currentValueInt == -1 {
			currentValueInt = valueInt
		} else if currentValueInt == valueInt-1 {
			currentValueInt = valueInt
		} else {
			return false
		}
	}
	return true
}

func (gameInstance *MauBinhGame) containXDoubleCardsInStreak(cards []string, times int) bool {
	data := make(map[int]int)
	pairValue := make([]int, 0)
	for _, cardString := range cards {
		valueAsInt := gameInstance.getValueAsIntFromCard(cardString)
		data[valueAsInt]++
		if data[valueAsInt] > 1 && valueAsInt != 12 { // ignore 2
			if !utils.Contains(pairValue, valueAsInt) {
				pairValue = append(pairValue, valueAsInt)
			}
		}
	}
	currentIntValue := -1
	count := 0
	for _, valueAsInt := range pairValue {
		if currentIntValue == -1 {
			currentIntValue = valueAsInt
		} else {
			if valueAsInt-currentIntValue == 1 {
				// in order
				count++
				if count == times-1 {
					return true
				}
			} else {
				// not in order anymore
				count = 0
			}
		}
		currentIntValue = valueAsInt
	}
	return false
}

func containValue(cards []string, valueToFind string) bool {
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == valueToFind {
			return true
		}
	}
	return false
}

func getCardWithValue(cards []string, valueToFind string) string {
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == valueToFind {
			return cardString
		}
	}
	return ""
}

func getCardsWithValue(cards []string, valueToFind string) []string {
	results := make([]string, 0)
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == valueToFind {
			results = append(results, cardString)
		}
	}
	return results
}

func removeCardsFromCards(totalCards []string, removedCards []string) (cards []string) {
	cards = make([]string, 0)
	for _, card := range totalCards {
		notFound := true
		for _, removedCard := range removedCards {
			if card == removedCard {
				notFound = false
				break
			}
		}
		if notFound {
			cards = append(cards, card)
		}
	}
	return cards
}
