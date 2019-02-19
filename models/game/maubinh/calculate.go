package maubinh

import (
	"fmt"
	"github.com/vic/vic_go/models/components"
	"sort"
)

type ByCardsGroup struct {
	groups       [][]string
	gameInstance *MauBinhGame
}

func (a *ByCardsGroup) Len() int      { return len(a.groups) }
func (a *ByCardsGroup) Swap(i, j int) { a.groups[i], a.groups[j] = a.groups[j], a.groups[i] }
func (a *ByCardsGroup) Less(i, j int) bool {
	group1 := a.groups[i]
	group2 := a.groups[j]
	return a.gameInstance.getCompareBetweenCards(group1, group2) < 0
}

func (gameInstance *MauBinhGame) SortGroups(groups [][]string) [][]string {
	obj := &ByCardsGroup{
		groups:       make([][]string, len(groups)),
		gameInstance: gameInstance,
	}
	copy(obj.groups, groups)
	sort.Sort(obj)
	return obj.groups
}

func (gameInstance *MauBinhGame) ReverseSortGroups(groups [][]string) [][]string {
	obj := &ByCardsGroup{
		groups:       make([][]string, len(groups)),
		gameInstance: gameInstance,
	}
	copy(obj.groups, groups)
	sort.Sort(sort.Reverse(obj))
	return obj.groups
}

type CardGroups struct {
	sameCardsGroups     map[string][]string
	fourOfAKindGroups   [][]string
	fullHouseGroups     [][]string
	threeOfAKindGroups  [][]string
	pairGroups          [][]string
	twoPairGroups       [][]string
	flushGroups         [][]string
	straightGroups      [][]string
	straightFlushGroups [][]string

	groupsInOrder [][][]string
}

func (gameInstance *MauBinhGame) NewCardGroups(cards []string) *CardGroups {
	groups := &CardGroups{}

	groups.sameCardsGroups = getGroupsOfSameValueCards(cards)
	groups.fourOfAKindGroups = getGroupsOfSameCardsWithSpecificLength(groups.sameCardsGroups, 4)
	groups.threeOfAKindGroups = getGroupsOfSameCardsWithSpecificLength(groups.sameCardsGroups, 3)
	groups.pairGroups = getGroupsOfSameCardsWithSpecificLength(groups.sameCardsGroups, 2)
	groups.flushGroups = getBiggestFlushGroups(cards)
	groups.straightGroups = gameInstance.getStraightGroups(cards)
	groups.fullHouseGroups = getFullHouseGroups(cards, groups.threeOfAKindGroups)
	groups.twoPairGroups = getTwoPairGroups(cards, groups.pairGroups)
	groups.straightFlushGroups = getStraightFlushGroups(groups.straightGroups)

	groups.groupsInOrder = make([][][]string, 0)
	if len(groups.straightFlushGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.straightFlushGroups)
	}

	if len(groups.fourOfAKindGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.fourOfAKindGroups)
	}

	if len(groups.fullHouseGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.fullHouseGroups)
	}

	if len(groups.flushGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.flushGroups)
	}

	if len(groups.straightGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.straightGroups)
	}

	if len(groups.threeOfAKindGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.threeOfAKindGroups)
	}

	if len(groups.twoPairGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.twoPairGroups)
	}

	if len(groups.pairGroups) > 0 {
		groups.groupsInOrder = append(groups.groupsInOrder, groups.pairGroups)
	}

	// //fmt.Println("----data---")
	// //fmt.Println("cards", cards)
	// //fmt.Println("straight flush", groups.straightFlushGroups)
	// //fmt.Println("four of a kind", groups.fourOfAKindGroups)
	// //fmt.Println("full house", groups.fullHouseGroups)
	// //fmt.Println("three of a kind", groups.threeOfAKindGroups)
	// //fmt.Println("pair", groups.pairGroups)
	// //fmt.Println("flush", groups.flushGroups)
	// //fmt.Println("straight", groups.straightGroups)
	// //fmt.Println("-------")
	return groups
}

func (gameInstance *MauBinhGame) CalculateCardsDataBigToSmall(cards []string) map[string][]string {
	// find group of straight flush

	cardsData := make(map[string][]string)

	for len(cards) > 0 {
		cards, cardsData = gameInstance.putBiggestGroupToCorrectPartBigToSmall(cards, cardsData)
	}
	return cardsData
}

func (gameInstance *MauBinhGame) putBiggestGroupToCorrectPartBigToSmall(cards []string, cardsData map[string][]string) (cardsLeft []string, cardsDataOutput map[string][]string) {
	groups := gameInstance.NewCardGroups(cards)
	if len(cardsData[BottomPart]) == 0 {
		// work on bottom part
		if len(groups.straightFlushGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.straightFlushGroups, BottomPart)
			cardsData[BottomPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}

		if len(groups.fourOfAKindGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.fourOfAKindGroups, BottomPart)
			cardsData[BottomPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}
	}

	if len(cardsData[MiddlePart]) == 0 {
		if len(groups.straightFlushGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.straightFlushGroups, MiddlePart)
			cardsData[MiddlePart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}

		if len(groups.fourOfAKindGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.fourOfAKindGroups, MiddlePart)
			cardsData[MiddlePart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}

	}

	if len(cardsData[BottomPart]) == 0 {
		for _, possibleGroups := range groups.groupsInOrder {
			groupToUse := gameInstance.getBiggestGroupInGroup(possibleGroups, BottomPart)
			cardsData[BottomPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}
	}

	if len(cardsData[MiddlePart]) == 0 {
		for _, possibleGroups := range groups.groupsInOrder {
			groupToUse := gameInstance.getBiggestGroupInGroup(possibleGroups, BottomPart)
			cardsData[MiddlePart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}
	}

	if len(cardsData[TopPart]) == 0 {
		if len(groups.threeOfAKindGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.threeOfAKindGroups, TopPart)
			cardsData[TopPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}

		if len(groups.pairGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.pairGroups, TopPart)
			cardsData[TopPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData
		}
	}

	// fill in what is missing
	if len(cardsData[BottomPart]) != 5 {
		lastCard := cards[len(cards)-1]
		cardsData[BottomPart] = append(cardsData[BottomPart], lastCard)
		cards = removeCardsFromCards(cards, []string{lastCard})
		return cards, cardsData
	}

	if len(cardsData[MiddlePart]) != 5 {
		lastCard := cards[len(cards)-1]
		cardsData[MiddlePart] = append(cardsData[MiddlePart], lastCard)
		cards = removeCardsFromCards(cards, []string{lastCard})
		return cards, cardsData
	}
	if len(cardsData[TopPart]) != 3 {
		lastCard := cards[len(cards)-1]
		cardsData[TopPart] = append(cardsData[TopPart], lastCard)
		cards = removeCardsFromCards(cards, []string{lastCard})
		return cards, cardsData
	}

	return nil, nil
}

func (gameInstance *MauBinhGame) CalculateCardsData(cards []string) map[string][]string {
	// find group of straight flush
	//fmt.Println("start with cards", cards)
	cardsData := make(map[string][]string)
	var shouldEnd bool
	for !shouldEnd {
		cards, cardsData, shouldEnd = gameInstance.putBiggestGroupToCorrectPart(cards, cardsData)
	}
	//fmt.Println("out first stage", cardsData, cards)
	// check if still not enough
	if len(cardsData[BottomPart]) == 5 &&
		len(cardsData[MiddlePart]) == 5 &&
		len(cardsData[TopPart]) == 3 {
		return cardsData
	}

	// will now check and fill in
	// the point will be to make balance parts, try not to make bottom too big and then the other 2 very small etc
	// so go from top first, and check each times to see if we have enough stuffs to fill other part

	// if bottom or middle fill straight flush then it is 5
	// if fill in four aces, then we only need to put any cards in
	// so only check if these part did not have anythings

	if len(cardsData[TopPart]) == 0 { // no three groups that can fit
		var numPartMissing int
		if len(cardsData[BottomPart]) == 0 {
			numPartMissing++
		}

		if len(cardsData[MiddlePart]) == 0 {
			numPartMissing++
		}
		if numPartMissing == 0 {
			// fill anything that can be fill in that top part
			// if high cards then put all the biggest cards in
			groups := gameInstance.NewCardGroups(cards)
			var groupForTopPart []string
			if len(groups.groupsInOrder) > 0 {
				for _, posibleGroups := range groups.groupsInOrder {
					groupToUse := gameInstance.getBiggestGroupInGroup(posibleGroups, TopPart)
					if len(groupToUse) < 3 {
						groupForTopPart = groupToUse
						break
					}

				}

			}

			cards = removeCardsFromCards(cards, groupForTopPart)
			if len(groupForTopPart) != 3 {
				// fill in biggest high cards
				missingCount := 3 - len(groupForTopPart)
				for i := 0; i < missingCount; i++ {
					lengthCards := len(cards)
					card := cards[lengthCards-1]
					groupForTopPart = append(groupForTopPart, card)
					cards = removeCardsFromCards(cards, []string{card})
				}
			}

			cardsData[TopPart] = groupForTopPart
			// any order to fill for middle and bot is fine

			for _, positionString := range []string{MiddlePart, BottomPart} {
				if len(cardsData[positionString]) != 5 {
					missingCount := 5 - len(cardsData[positionString])
					for i := 0; i < missingCount; i++ {
						lengthCards := len(cards)
						card := cards[lengthCards-1]
						cardsData[positionString] = append(cardsData[positionString], card)
						cards = removeCardsFromCards(cards, []string{card})
					}
				}
			}
			return cardsData

		} else {
			groups := gameInstance.NewCardGroups(cards)
			for _, posibleGroups := range groups.groupsInOrder {
				sortedGroups := gameInstance.ReverseSortGroups(posibleGroups)
				var shouldBreak bool
				for _, group := range sortedGroups {
					if len(group) < 3 {
						temp := removeCardsFromCards(cards, group)
						if gameInstance.stillHaveGroupsBiggerThanGroup(group, temp, numPartMissing) {
							cardsData[TopPart] = group
							cards = temp
							shouldBreak = true
							break
						}
					} else {
						break
					}
				}
				if shouldBreak {
					break
				}
			}
			//fmt.Println("finish for top", cardsData, cards)

			// done what we can for top part already
			// now if top part is still empty, it mean high cards
			// we will now work on middle and then bottom

			if len(cardsData[MiddlePart]) == 0 {
				groups = gameInstance.NewCardGroups(cards)
				numPartMissing = 0
				if len(cardsData[BottomPart]) == 0 {
					numPartMissing++
				}
				for _, posibleGroups := range groups.groupsInOrder {
					var shouldBreak bool
					sortedGroups := gameInstance.ReverseSortGroups(posibleGroups)
					for _, group := range sortedGroups {
						temp := removeCardsFromCards(cards, group)
						if gameInstance.stillHaveGroupsBiggerThanGroup(group, temp, numPartMissing) {
							cardsData[MiddlePart] = group
							cards = temp
							shouldBreak = true
							break
						}
					}
					if shouldBreak {
						break
					}
				}
			}
			//fmt.Println("finish for middle", cardsData, cards)

			// "IF" middle part is empty, it will be high cards now...
			// go on to bottom
			if len(cardsData[BottomPart]) == 0 {
				groups = gameInstance.NewCardGroups(cards)
				//fmt.Println(groups.groupsInOrder)
				if len(groups.groupsInOrder) > 0 {
					biggestGroup := gameInstance.getBiggestGroupInGroup(groups.groupsInOrder[0], BottomPart)
					cardsData[BottomPart] = biggestGroup
					cards = removeCardsFromCards(cards, biggestGroup)
				}
			}
			//fmt.Println("finish for bottom", cardsData, cards)

			// now we will fill in missing part
			// start from top -> bottom, also need to make sure that we don't make later part bigger than earlier part

			orderToFill := make([]string, 0)
			if len(cardsData[TopPart]) != 3 {
				// check if there is a posibility that middle and bottom will be bigger
				if len(cardsData[TopPart]) == 0 {
					// top part is high card -> check if middle or bototm is high card too
					if len(cardsData[MiddlePart]) == 0 || len(cardsData[BottomPart]) == 0 {
						// will not fill top part first
						if len(cardsData[BottomPart]) == 0 {
							orderToFill = append(orderToFill, BottomPart)
							orderToFill = append(orderToFill, MiddlePart)
							orderToFill = append(orderToFill, TopPart)
						} else {
							orderToFill = append(orderToFill, MiddlePart)
							orderToFill = append(orderToFill, TopPart)
							orderToFill = append(orderToFill, BottomPart)
						}
					} else {
						orderToFill = append(orderToFill, TopPart)
						compare := gameInstance.getCompareBetweenCards(cardsData[BottomPart], cardsData[MiddlePart])
						if compare == 0 {
							orderToFill = append(orderToFill, BottomPart)
							orderToFill = append(orderToFill, MiddlePart)
						} else {

							orderToFill = append(orderToFill, MiddlePart)
							orderToFill = append(orderToFill, BottomPart)
						}
					}
				} else {
					// top part is a pair -> check if middle or bototm is pair too
					compare := gameInstance.getCompareBetweenCards(cardsData[TopPart], cardsData[MiddlePart])
					if compare == 0 {
						orderToFill = append(orderToFill, MiddlePart)
						orderToFill = append(orderToFill, TopPart)
						orderToFill = append(orderToFill, BottomPart)
					} else {
						// top part is smaller than middle part
						orderToFill = append(orderToFill, TopPart)
						compare = gameInstance.getCompareBetweenCards(cardsData[MiddlePart], cardsData[BottomPart])
						if compare == 0 {
							orderToFill = append(orderToFill, BottomPart)
							orderToFill = append(orderToFill, MiddlePart)
						} else {

							orderToFill = append(orderToFill, MiddlePart)
							orderToFill = append(orderToFill, BottomPart)
						}
					}
				}
			} else {
				compare := gameInstance.getCompareBetweenCards(cardsData[BottomPart], cardsData[MiddlePart])
				if compare == 0 {
					orderToFill = append(orderToFill, BottomPart)
					orderToFill = append(orderToFill, MiddlePart)
				} else {
					orderToFill = append(orderToFill, MiddlePart)
					orderToFill = append(orderToFill, BottomPart)
				}
			}

			//fmt.Println("fill order", cardsData, cards, orderToFill)
			for len(cards) != 0 {
				for _, positionString := range orderToFill {
					if len(cards) == 0 {
						break
					}
					cardsInPart := cardsData[positionString]
					maxLength := 5
					if positionString == TopPart {
						maxLength = 3
					}

					if len(cardsInPart) < maxLength {
						card := cards[len(cards)-1]
						cardsData[positionString] = append(cardsData[positionString], card)
						cards = removeCardsFromCards(cards, []string{card})
					}

				}
			}

		}

	} else {
		// top part has three group, and we should have two other group to fill in for bottom and and middle
		if len(cardsData[MiddlePart]) == 0 {
			groups := gameInstance.NewCardGroups(cards)
			numPartMissing := 0
			if len(cardsData[BottomPart]) == 0 {
				numPartMissing++
			}
			for _, posibleGroups := range groups.groupsInOrder {
				var shouldBreak bool
				sortedGroups := gameInstance.ReverseSortGroups(posibleGroups)
				for _, group := range sortedGroups {
					temp := removeCardsFromCards(cards, group)
					if gameInstance.stillHaveGroupsBiggerThanGroup(group, temp, numPartMissing) {
						cardsData[MiddlePart] = group
						cards = temp
						shouldBreak = true
						break
					}
				}
				if shouldBreak {
					break
				}
			}
		}

		// go on to bottom
		if len(cardsData[BottomPart]) == 0 {
			groups := gameInstance.NewCardGroups(cards)
			for _, posibleGroups := range groups.groupsInOrder {
				biggestGroup := gameInstance.getBiggestGroupInGroup(posibleGroups, BottomPart)
				cardsData[BottomPart] = biggestGroup
				cards = removeCardsFromCards(cards, biggestGroup)
			}
		}

		// now we will fill in missing part
		// start from top -> bottom, also need to make sure that we don't make later part bigger than earlier part

		orderToFill := make([]string, 0)

		compare := gameInstance.getCompareBetweenCards(cardsData[BottomPart], cardsData[MiddlePart])
		if compare == 0 {
			orderToFill = append(orderToFill, BottomPart)
			orderToFill = append(orderToFill, MiddlePart)
		} else {
			orderToFill = append(orderToFill, MiddlePart)
			orderToFill = append(orderToFill, BottomPart)
		}

		for len(cards) != 0 {
			for _, positionString := range orderToFill {
				if len(cards) == 0 {
					break
				}
				cardsInPart := cardsData[positionString]
				maxLength := 5
				if positionString == TopPart {
					maxLength = 3
				}

				if len(cardsInPart) < maxLength {
					card := cards[len(cards)-1]
					cardsData[positionString] = append(cardsData[positionString], card)
					cards = removeCardsFromCards(cards, []string{card})
				}

			}
		}
	}

	return cardsData
}

func (gameInstance *MauBinhGame) putBiggestGroupToCorrectPart(cards []string, cardsData map[string][]string) (cardsLeft []string, cardsDataOutput map[string][]string, shouldEnd bool) {
	groups := gameInstance.NewCardGroups(cards)
	if len(cardsData[BottomPart]) == 0 {
		// work on bottom part
		if len(groups.straightFlushGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.straightFlushGroups, BottomPart)
			cardsData[BottomPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData, false
		}

		if len(groups.fourOfAKindGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.fourOfAKindGroups, BottomPart)
			cardsData[BottomPart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData, false
		}
	}

	if len(cardsData[MiddlePart]) == 0 {
		if len(groups.straightFlushGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.straightFlushGroups, MiddlePart)
			cardsData[MiddlePart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData, false
		}

		if len(groups.fourOfAKindGroups) > 0 {
			groupToUse := gameInstance.getBiggestGroupInGroup(groups.fourOfAKindGroups, MiddlePart)
			cardsData[MiddlePart] = groupToUse
			cards = removeCardsFromCards(cards, groupToUse)
			return cards, cardsData, false
		}

	}

	// check three of a kind for top
	// only put in if we have somethings to put in bottom and middle part
	if len(groups.threeOfAKindGroups) > 0 {
		var numPartMissing int
		if len(cardsData[BottomPart]) == 0 {
			numPartMissing++
		}

		if len(cardsData[MiddlePart]) == 0 {
			numPartMissing++
		}

		sortedGroups := gameInstance.ReverseSortGroups(groups.threeOfAKindGroups)
		for _, group := range sortedGroups {
			temp := removeCardsFromCards(cards, group)
			if gameInstance.stillHaveGroupsBiggerThanGroup(group, temp, numPartMissing) {
				cardsData[TopPart] = group
				cards = temp
				return cards, cardsData, true
			}
		}
	}

	return cards, cardsData, true
}

func (gameInstance *MauBinhGame) stillHaveGroupsBiggerThanGroup(groupToCompare []string, cards []string, numberToCheck int) bool {
	if numberToCheck == 0 {
		return true
	}
	tempGroups := gameInstance.NewCardGroups(cards)
	for _, possibleGroups := range tempGroups.groupsInOrder {
		for _, tempGroup := range possibleGroups {
			if gameInstance.getCompareBetweenCards(groupToCompare, tempGroup) < 0 { // middle or bottom is same
				if numberToCheck == 1 {
					//fmt.Println("got group bigger than group", tempGroup, groupToCompare)
					return true
				} else {
					temp2 := removeCardsFromCards(cards, tempGroup)
					result := gameInstance.stillHaveGroupsBiggerThanGroup(groupToCompare, temp2, numberToCheck-1)
					if result {
						return true
					} else {
						continue
					}
				}
			} else {
				return false
			}
		}
	}
	return false
}

func (gameInstance *MauBinhGame) getBiggestGroupInGroup(groups [][]string, positionString string) []string {
	var groupToUse []string
	for _, group := range groups {
		if len(groupToUse) == 0 {
			groupToUse = group
		} else {
			if gameInstance.getMultiplierBetweenCards(group, groupToUse, positionString) > 0 {
				groupToUse = group
			}
		}
	}
	return groupToUse
}

func (gameInstance *MauBinhGame) getStraightFlushStartWithCard(cards []string, card string, length int) (streak []string) {
	if gameInstance.getValueAsIntFromCard(card) == 12 {
		suit, _ := components.SuitAndValueFromCard(card)
		secondCard := fmt.Sprintf("%s 2", suit)
		if containCard(cards, secondCard) {
			streak = []string{card, secondCard}
			for i := 0; i < length-2; i++ {
				cardInStreak := gameInstance.getSameSuitCardHasValueMoreThanCard(cards, secondCard, i+1)
				if cardInStreak != "" {
					streak = append(streak, cardInStreak)

					if len(streak) == length-1 {
						return streak
					}
				} else {
					return nil
				}
			}
		}
		return nil
	} else {
		streak = []string{card}
		for i := 0; i < length-1; i++ {
			cardInStreak := gameInstance.getSameSuitCardHasValueMoreThanCard(cards, card, i+1)
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
}

func getFullHouseGroups(cards []string, threeOfAKindGroups [][]string) [][]string {
	if len(threeOfAKindGroups) == 0 {
		return nil
	}
	results := make([][]string, 0)
	for _, threeOfAKindGroup := range threeOfAKindGroups {
		_, threeOfAKindValue := components.SuitAndValueFromCard(threeOfAKindGroup[0])
		temp := removeCardsFromCards(cards, threeOfAKindGroup)
		sameCardsGroups := getGroupsOfSameValueCards(temp)
		pairGroups := getGroupsOfSameCardsWithSpecificLength(sameCardsGroups, 2)
		if len(pairGroups) > 0 {
			for _, pair := range pairGroups {
				card := pair[0]
				_, value := components.SuitAndValueFromCard(card)
				if value != threeOfAKindValue {
					group := make([]string, 0)
					group = append(group, threeOfAKindGroup...)
					group = append(group, pair...)
					results = append(results, group)
				}
			}
		}
	}
	return results
}

func getTwoPairGroups(cards []string, pairGroups [][]string) [][]string {
	if len(pairGroups) == 0 {
		return nil
	}
	results := make([][]string, 0)
	for _, pairCore := range pairGroups {
		_, pairValue := components.SuitAndValueFromCard(pairCore[0])
		temp := removeCardsFromCards(cards, pairCore)
		sameCardsGroups := getGroupsOfSameValueCards(temp)
		subPairGroups := getGroupsOfSameCardsWithSpecificLength(sameCardsGroups, 2)
		if len(subPairGroups) > 0 {
			for _, pair := range subPairGroups {
				card := pair[0]
				_, value := components.SuitAndValueFromCard(card)
				if value != pairValue {
					group := make([]string, 0)
					group = append(group, pairCore...)
					group = append(group, pair...)
					results = append(results, group)
				}
			}
		}
	}
	return results
}

func getStraightFlushGroups(straightsGroup [][]string) [][]string {
	results := make([][]string, 0)
	for _, straight := range straightsGroup {
		currentSuit := ""
		isOk := true
		for _, cardString := range straight {
			suit, _ := components.SuitAndValueFromCard(cardString)
			if currentSuit == "" {
				currentSuit = suit
			} else if currentSuit != suit {
				isOk = false
				break
			}
		}
		if isOk {
			results = append(results, straight)
		}
	}
	return results
}

func (gameInstance *MauBinhGame) getStraightGroups(cards []string) [][]string {
	groups := make([][]string, 0)
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == "a" { // ace
			straight := []string{cardString}
			secondCards := getAllCardsWithValue(cards, "2")
			if len(secondCards) > 0 {
				straight = append(straight, secondCards[0])
			}

			currentValue := "2"
			for i := 0; i < 3; i++ { // need 3 more
				nextValue := gameInstance.nextValue(currentValue)
				currentValue = nextValue
				nextCards := getAllCardsWithValue(cards, nextValue)
				if len(nextCards) > 0 {
					straight = append(straight, nextCards[0])
				} else {
					break
				}
			}

			if len(straight) == 5 {
				groups = append(groups, straight)
			}

		} else {
			straight := []string{cardString}
			currentValue := value
			for i := 0; i < 4; i++ { // need 4 more
				nextValue := gameInstance.nextValue(currentValue)
				currentValue = nextValue
				nextCards := getAllCardsWithValue(cards, nextValue)
				if len(nextCards) > 0 {
					straight = append(straight, nextCards[0])
				} else {
					break
				}
			}

			if len(straight) == 5 {
				groups = append(groups, straight)
			}
		}
	}

	// see if straight can be dup
	results := make([][]string, 0)
	for _, coreStraight := range groups {
		moreStraights := make([][]string, 0)
		moreStraights = append(moreStraights, coreStraight)
		for index, cardString := range coreStraight {
			if index == 0 {
				continue // ignore index 0 since that is the start, we branch off using the start above already
			}
			_, value := components.SuitAndValueFromCard(cardString)
			moreCards := getAllCardsWithValue(cards, value)
			if len(moreCards) == 1 {
				// intent blank
			} else {
				for _, straight := range moreStraights {
					for _, newCard := range moreCards {
						if newCard != straight[index] {
							newStraight := make([]string, len(straight))
							copy(newStraight, straight)
							newStraight[index] = newCard
							moreStraights = append(moreStraights, newStraight)
						}
					}
				}
			}
		}

		results = append(results, moreStraights...)
	}
	return results
}

func getFlushGroups(cards []string) [][]string {
	data := make(map[string][]string)
	for _, cardString := range cards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if data[suit] == nil {
			data[suit] = make([]string, 0)
		}
		data[suit] = append(data[suit], cardString)
	}

	results := make([][]string, 0)
	for _, cards := range data {
		subsets := getAllSubsetWithLength(cards, 5)
		if len(subsets) > 0 {
			results = append(results, subsets...)
		}
	}
	return results
}

func getBiggestFlushGroups(cards []string) [][]string {
	data := make(map[string][]string)
	for _, cardString := range cards {
		suit, _ := components.SuitAndValueFromCard(cardString)
		if data[suit] == nil {
			data[suit] = make([]string, 0)
		}
		data[suit] = append(data[suit], cardString)
	}

	results := make([][]string, 0)
	for _, cards := range data {
		if len(cards) < 5 {
			continue
		}
		startIndex := len(cards) - 5
		group := make([]string, 0)
		for index, card := range cards {
			if index >= startIndex {
				group = append(group, card)
			}
		}
		results = append(results, group)
	}
	return results
}

func getGroupsOfSameValueCards(cards []string) map[string][]string {
	data := make(map[string][]string)
	for _, cardString := range cards {
		_, value := components.SuitAndValueFromCard(cardString)
		if data[value] == nil {
			data[value] = make([]string, 0)
		}
		data[value] = append(data[value], cardString)
	}
	return data
}

func getGroupsOfSameCardsWithSpecificLength(sameCardsGroups map[string][]string, length int) [][]string {
	results := make([][]string, 0)
	for _, group := range sameCardsGroups {
		if len(group) >= length {
			results = append(results, getAllSubsetWithLength(group, length)...)
		}
	}
	return results
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
				if !containString(shorterSubset, element) {
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

func containString(array []string, element string) bool {
	for _, arrayElement := range array {
		if arrayElement == element {
			return true
		}
	}
	return false
}

func containSlice(arrayOfSlices [][]string, slice []string) bool {
	for _, sliceElement := range arrayOfSlices {
		if isEqual(sliceElement, slice) {
			return true
		}
	}
	return false
}

func isEqual(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for _, element1 := range slice1 {
		var found bool
		for _, element2 := range slice2 {
			if element1 == element2 {
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

func (gameInstance *MauBinhGame) getGroupsOfStraightFlush(cards []string) [][]string {
	results := make([][]string, 0)
	for _, cardString := range cards {
		straightFlushGroup := gameInstance.getStraightFlushStartWithCard(cards, cardString, 5)
		if straightFlushGroup != nil {
			results = append(results, straightFlushGroup)
		}
	}
	return results
}

func (gameInstance *MauBinhGame) getSameSuitCardHasValueMoreThanCard(cards []string, card string, moreBy int) string {
	suit, value := components.SuitAndValueFromCard(card)
	valueInt := gameInstance.valueAsInt(value)
	for _, cardInHand := range cards {
		suitInHand, valueInHand := components.SuitAndValueFromCard(cardInHand)
		if suit == suitInHand {
			valueAsIntInHand := gameInstance.valueAsInt(valueInHand)
			if valueAsIntInHand == valueInt+moreBy {
				return cardInHand
			}
		}
	}
	return ""
}

func containCard(cards []string, card string) bool {
	for _, cardString := range cards {
		if cardString == card {
			return true
		}
	}
	return false
}

func getGroupsOfSameCards(sortedCards []string, groupSize int) [][]string {
	data := make(map[string]int)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		data[value]++
	}
	results := make([][]string, 0)
	for cardValue, numCards := range data {
		if numCards == groupSize {
			results = append(results, getAllCardsWithValue(sortedCards, cardValue))
		}
	}
	return results
}

func getAllCardsWithValue(sortedCards []string, valueToGet string) []string {
	results := make([]string, 0)
	for _, cardString := range sortedCards {
		_, value := components.SuitAndValueFromCard(cardString)
		if value == valueToGet {
			results = append(results, cardString)
		}
	}
	return results
}

func (gameInstance *MauBinhGame) nextValue(value string) string {
	if value == "a" {
		return ""
	}
	for index, valueInOrder := range gameInstance.logicInstance.CardValueOrder() {
		if valueInOrder == value {
			return gameInstance.logicInstance.CardValueOrder()[index+1]
		}
	}
	return ""
}
