package tienlen

import ()

/*
all method will receive a sorted cards list
*/

func (gameInstance *TienLenGame) isInstantWin(sortedCards []string) bool {
	if len(gameInstance.logicInstance.GetInstantWinType(sortedCards)) != 0 {
		return true
	}
	return false
}

func (gameInstance *TienLenGame) playCardsOverCards(cardsOnTable []string, cards []string) (isValidMove bool) {
	// TODO: write code for vn slash
	if gameInstance.logicInstance.IsCards1BiggerThanCards2(cards, cardsOnTable) {
		return true
	} else {
		return false
	}
}

func (gameInstance *TienLenGame) GetLoseMultiplier(cardsOnTable []string) (loseMul int) {
	// TODO: write code for vn slash
	return gameInstance.logicInstance.GetCardTableMoveType(cardsOnTable)
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

func (gameInstance *TienLenGame) valueAsInt(value string) int {
	var counter int
	for _, valueOrder := range gameInstance.logicInstance.CardValueOrder() {
		if valueOrder == value {
			return counter
		}
		counter++
	}
	return counter
}

func (gameInstance *TienLenGame) suitAsInt(suit string) int {
	var counter int
	for _, suitOrder := range gameInstance.logicInstance.CardSuitOrder() {
		if suitOrder == suit {
			return counter
		}
		counter++
	}
	return counter
}
