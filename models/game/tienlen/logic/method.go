package logic

import (
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/utils"
	"reflect"
	"sort"
)

type ByCardSuitAndValue struct {
	cards []string
	logic TLLogic
}

func (a *ByCardSuitAndValue) Len() int      { return len(a.cards) }
func (a *ByCardSuitAndValue) Swap(i, j int) { a.cards[i], a.cards[j] = a.cards[j], a.cards[i] }
func (a *ByCardSuitAndValue) Less(i, j int) bool {
	suitI, valueI := components.SuitAndValueFromCard(a.cards[i])
	suitJ, valueJ := components.SuitAndValueFromCard(a.cards[j])
	valueIAsInt := valueAsInt(a.logic, valueI)
	valueJAsInt := valueAsInt(a.logic, valueJ)
	if valueIAsInt == valueJAsInt {
		suitIAsInt := suitAsInt(a.logic, suitI)
		suitJAsInt := suitAsInt(a.logic, suitJ)
		return suitIAsInt < suitJAsInt
	}
	return valueIAsInt < valueJAsInt
}

func sortCards(logic TLLogic, cards []string) []string {
	sortedCards := make([]string, 0)
	for _, card := range cards {
		sortedCards = append(sortedCards, card)
	}

	obj := &ByCardSuitAndValue{
		cards: sortedCards,
		logic: logic,
	}

	sort.Sort(obj)
	return obj.cards
}

func containCards(totalCards []string, checkCards []string) bool {
	for _, cardString := range checkCards {
		var found bool
		for _, cardStringInTotal := range totalCards {
			if cardString == cardStringInTotal {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func getValueFromCard(card string) string {
	_, value := components.SuitAndValueFromCard(card)
	return value
}

func getValueAsIntFromCard(logicInstance TLLogic, card string) int {
	_, value := components.SuitAndValueFromCard(card)
	return valueAsInt(logicInstance, value)
}

func valueAsInt(logic TLLogic, value string) int {
	var counter int
	for _, valueOrder := range logic.CardValueOrder() {
		if valueOrder == value {
			return counter
		}
		counter++
	}
	return counter
}

func suitAsInt(logic TLLogic, suit string) int {
	var counter int
	for _, suitOrder := range logic.CardSuitOrder() {
		if suitOrder == suit {
			return counter
		}
		counter++
	}
	return counter
}

func getTypeIndex(logic TLLogic, typeString string) int {
	var counter int
	for _, typeInList := range logic.TypeOrder() {
		if typeInList == typeString {
			return counter
		}
		counter++
	}
	return -1
}

func numberOf2Cards(cards []string) int {
	var counter int
	for _, card := range cards {

		if getValueFromCard(card) == "2" {
			counter++
		}
	}
	return counter
}

func numberOfQuadrupleCards(cards []string) (counter int) {
	var sameCardCount int
	lastCardValue := "-1"

	for _, card := range cards {
		value := getValueFromCard(card)
		if value != lastCardValue {
			lastCardValue = value
			sameCardCount = 1
		} else {
			sameCardCount++
			if sameCardCount == 4 {
				counter++
			}
		}
	}
	return counter
}

func containXDoubleCardsInStreak(logicInstance TLLogic, cards []string, times int) bool {
	data := make(map[int]int)
	pairValue := make([]int, 0)
	for _, cardString := range cards {
		valueAsInt := getValueAsIntFromCard(logicInstance, cardString)
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

func isStreak(logic TLLogic, sortedCards []string) bool {
	if isContainCardsWithSameValue(sortedCards) {
		return false
	}

	if !logic.Include2InGroupCards() {
		if numberOf2Cards(sortedCards) > 0 {
			return false
		}
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

			if isStreak(logic, temp) {
				return true
			} else {
				return false
			}
		}
	}

	currentValueInt := -1
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		valueInt := valueAsInt(logic, value)
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

func isStraightFlush(logic TLLogic, sortedCards []string) bool {
	if isStreak(logic, sortedCards) && isFlush(sortedCards) {
		return true
	}
	return false

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

func isDupCardsInOrderAndIncreaseByOne(logic TLLogic, sortedCards []string) bool {
	if len(sortedCards)%2 == 1 {
		return false
	}

	// cannot have 2 at the last position
	lastCardString := sortedCards[len(sortedCards)-1]
	_, lastCardValue := components.SuitAndValueFromCard(lastCardString)
	if lastCardValue == "2" {
		return false
	}

	currentIntValue := -1
	counter := 0
	firstPositionIntValue := -1
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if counter%2 == 0 {
			firstPositionIntValue = valueAsInt
		} else {
			if firstPositionIntValue == valueAsInt {
				if currentIntValue == -1 {
					currentIntValue = valueAsInt
				} else {
					if valueAsInt-currentIntValue == 1 {
						currentIntValue = valueAsInt
					} else {
						// not in order anymore
						return false
					}
				}
			} else {
				// not couple anymore
				return false
			}
		}

		counter++
	}
	return true
}

func is5DupsCardsInOrder(logic TLLogic, sortedCards []string) bool {
	return containXDoubleCardsInStreak(logic, sortedCards, 5)
}

func is12CardsInOrderAndIncreaseBy1(logic TLLogic, sortedCards []string) bool {
	if len(sortedCards) != 13 {
		return false
	}
	minDupCardsTime := 1
	if !logic.Include2InGroupCards() {
		if numberOf2Cards(sortedCards) > 0 {
			minDupCardsTime = 0
		}
	}

	currentIntValue := -1
	duplicateCardsTime := 0
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if currentIntValue == -1 {
			currentIntValue = valueAsInt
		} else {
			if valueAsInt-currentIntValue == 1 {
				currentIntValue = valueAsInt
			} else if valueAsInt-currentIntValue == 0 {
				duplicateCardsTime++
				if duplicateCardsTime > minDupCardsTime {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

func is6DupsCards(logic TLLogic, sortedCards []string) bool {
	currentIntValue := -1
	duplicateCardsTime := 0
	lastDupValue := -1
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if currentIntValue == -1 {
			currentIntValue = valueAsInt
		} else {
			if valueAsInt-currentIntValue == 0 {
				if lastDupValue == valueAsInt {
					// 3 cards same as each other, will not count
				} else {
					duplicateCardsTime++
					if duplicateCardsTime == 6 {
						return true
					}
					lastDupValue = valueAsInt
				}
			}
			currentIntValue = valueAsInt
		}
	}
	return false
}

func isSameColor(sortedCards []string) bool {
	if len(sortedCards) != 13 {
		return false
	}
	isCurrentColorBlack := true
	isFirstCard := true
	for _, cardString := range sortedCards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if suit == "s" || suit == "c" {
			if isFirstCard {
				isFirstCard = false
				isCurrentColorBlack = true
			} else {
				if !isCurrentColorBlack {
					return false
				}
			}
		} else {
			if isFirstCard {
				isFirstCard = false
				isCurrentColorBlack = false
			} else {
				if isCurrentColorBlack {
					return false
				}
			}
		}
	}
	return true
}

func Is4TripleCards(logic TLLogic, sortedCards []string) bool {
	currentIntValue := -1
	tripleCardsTime := 0
	sameCardsCounter := 0
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if currentIntValue == -1 {
			currentIntValue = valueAsInt
		} else {
			if valueAsInt == currentIntValue {
				sameCardsCounter++
				if sameCardsCounter == 2 { // same 2 times already, include the first card we will have 3 cards
					tripleCardsTime++
					if tripleCardsTime == 4 {
						return true
					}
				}
			} else {
				sameCardsCounter = 0
			}
			currentIntValue = valueAsInt
		}
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

func is3Or4DupsCardsInOrderAndHas3s(logic TLLogic, sortedCards []string) bool {
	if sortedCards[0] != "d 3" {
		return false
	}
	currentIntValue := -1
	duplicateCardsTime := 0
	lastDupValue := -1
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if currentIntValue == -1 {
			currentIntValue = valueAsInt
		} else {
			if valueAsInt-currentIntValue == 0 {
				if lastDupValue == -1 {
					duplicateCardsTime++
					lastDupValue = valueAsInt
					if lastDupValue != 0 { // first dup cards is not 3, for sure it will be false since we need 3s
						return false
					}

				} else {
					if valueAsInt == lastDupValue {
						// 3 cards same as each other, will not count, will not care
					} else if valueAsInt-lastDupValue == 1 {
						// increase by 1, so ok
						duplicateCardsTime++
						if duplicateCardsTime == 3 { // got 3 is enough
							return true
						}
						lastDupValue = valueAsInt
					}
				}
			}
			currentIntValue = valueAsInt
		}
	}
	return false
}

func isFour2Cards(cards []string) bool {
	twoCardsCounter := 0
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "2" {
			twoCardsCounter++
			if twoCardsCounter == 4 {
				return true
			}
		}
	}
	return false
}

func isFour3Cards(cards []string) bool {
	threeCardsCounter := 0
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "3" {
			threeCardsCounter++
			if threeCardsCounter == 4 {
				return true
			}
		}
	}
	return false
}

func isContainCardsWithSameValue(sortedCards []string) bool {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		data[value]++
	}
	for _, numCards := range data {
		if numCards > 1 {
			return true
		}
	}
	return false
}

func getSameCardsMoves(sortedCards []string) (moves [][]string) {
	data := make(map[string][]string)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		if data[value] == nil {
			data[value] = make([]string, 0)
		}
		data[value] = append(data[value], cardString)
	}
	moves = make([][]string, 0)
	for _, move := range data {
		if len(move) > 1 {
			moves = append(moves, move)
			if len(move) > 2 {
				for i := 2; i < len(move); i++ {
					moves = append(moves, getAllSubsetWithLength(move, i)...)
				}
			}
		}
	}
	return
}

func getNumberOf2Cards(logic TLLogic, sortedCards []string) int {
	counter := 0
	for _, cardString := range sortedCards {
		valueAsInt := getValueAsIntFromCard(logic, cardString)
		if valueAsInt == 12 {
			counter++
		}
	}
	return counter
}

func getPairStreakMoves(logic TLLogic, sameCardsMoves [][]string) (streaks [][]string) {
	streaks = make([][]string, 0)
	for _, sameCardsMove := range sameCardsMoves {
		if len(sameCardsMove) == 2 {
			singleCard := sameCardsMove[0]
			valueAsInt := getValueAsIntFromCard(logic, singleCard)
			if valueAsInt == 12 {
				continue
			}
			subStreaks := make([][]string, 0)
			subStreaks = append(subStreaks, sameCardsMove)
			for i := 1; i < 12; i++ {
				nextSameCardsMoves := getCardsForPairStreakHasValueMoreThanCard(logic, sameCardsMoves, singleCard, i)
				if len(nextSameCardsMoves) == 0 {
					// do nothing
					break
				} else {
					newSubStreaks := make([][]string, 0)
					for _, nextSameCardsMove := range nextSameCardsMoves {
						for _, subStreak := range subStreaks {

							streak := make([]string, 0)
							streak = append(streak, subStreak...)
							streak = append(streak, nextSameCardsMove...)
							newSubStreaks = append(newSubStreaks, streak)
						}
					}
					subStreaks = newSubStreaks
				}
			}
			streaks = append(streaks, subStreaks...)
		}

	}

	validStreaks := make([][]string, 0)
	for _, streak := range streaks {
		if len(streak) >= 6 {
			validStreaks = append(validStreaks, streak)
		}
	}
	return validStreaks

}

func getAllSubsetWithLength(originSet []string, subsetLength int) [][]string {
	if len(originSet) < subsetLength {
		return nil
	}
	results := make([][]string, 0)
	if subsetLength == 1 {
		for _, element := range originSet {
			subset := []string{element}
			results = append(results, subset)
		}
		return results
	} else {
		shorterSubsetResults := getAllSubsetWithLength(originSet, subsetLength-1)
		for _, shorterSubset := range shorterSubsetResults {
			for _, element := range originSet {
				if !utils.ContainsByString(shorterSubset, element) {
					temp := make([]string, len(shorterSubset))
					copy(temp, shorterSubset)
					temp = append(temp, element)
					if !containSlice(results, temp) {
						results = append(results, temp)
					}
				}
			}
		}
		return results
	}
	return nil

}

func getCardsForStreakHasValueMoreThanCard(logic TLLogic, sortedCards []string, card string, moreBy int) (cards []string) {
	cards = make([]string, 0)
	valueAsInt := getValueAsIntFromCard(logic, card)
	for _, cardInHand := range sortedCards {
		valueAsIntHere := getValueAsIntFromCard(logic, cardInHand)
		if cardInHand != card && valueAsIntHere == valueAsInt+moreBy && valueAsIntHere != 12 {
			cards = append(cards, cardInHand)
		}
	}
	return cards
}

func getCardsForPairStreakHasValueMoreThanCard(logic TLLogic, sameCardsMoves [][]string, card string, moreBy int) (cards [][]string) {
	cards = make([][]string, 0)
	valueAsInt := getValueAsIntFromCard(logic, card)
	for _, sameCardsMove := range sameCardsMoves {
		if len(sameCardsMove) == 2 {
			singleCard := sameCardsMove[0]
			valueAsIntHere := getValueAsIntFromCard(logic, singleCard)
			if valueAsIntHere == valueAsInt+moreBy && valueAsInt != 12 {
				cards = append(cards, sameCardsMove)
			}
		}
	}
	return cards
}

func isSameValue(cards []string) bool {
	currentValue := ""
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if currentValue == "" {
			currentValue = value
		} else {
			if currentValue != value {
				return false
			}
		}
	}
	return true
}

func isCard1BiggerThanCard2(logic TLLogic, card1 string, card2 string) bool {
	suit1, value1 := components.SuitAndValueFromCard(card1)
	suit2, value2 := components.SuitAndValueFromCard(card2)

	value1AsInt := valueAsInt(logic, value1)
	value2AsInt := valueAsInt(logic, value2)
	if value1AsInt == value2AsInt {
		suit1AsInt := suitAsInt(logic, suit1)
		suit2AsInt := suitAsInt(logic, suit2)
		if suit1AsInt > suit2AsInt {
			return true
		} else {
			return false
		}
	} else if value1AsInt > value2AsInt {
		return true
	} else {
		return false
	}
	return false
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

func containSlice(arrayOfSlices [][]string, slice []string) bool {
	for _, sliceElement := range arrayOfSlices {
		if reflect.DeepEqual(sliceElement, slice) {
			return true
		}
	}
	return false
}

func cloneSlice(slice []string) []string {
	cloned := make([]string, len(slice))
	copy(cloned, slice)
	return cloned
}
