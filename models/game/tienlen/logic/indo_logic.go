package logic

import (
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/utils"
)

type IndoLogic struct {
	c2LoseMultiplier float64
	h2LoseMultiplier float64
	d2LoseMultiplier float64
	s2LoseMultiplier float64

	cardValueOrder []string
	cardSuitOrder  []string
	typeOrder      []string
}

func NewIndoLogic() *IndoLogic {
	return &IndoLogic{
		c2LoseMultiplier: 2,
		d2LoseMultiplier: 2,
		s2LoseMultiplier: 4,
		h2LoseMultiplier: 4,
		cardValueOrder:   []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a"},
		cardSuitOrder:    []string{"d", "c", "h", "s"},
		typeOrder: []string{MoveTypeInvalid, MoveTypeDoubleCards, MoveTypeTripleCards, MoveTypeQuadrupleCards,
			MoveType5OrderCards, MoveType5FlushCards, MoveTypeFullHouse, MoveTypeStraightFlush},
	}
}

func (logic *IndoLogic) C2LoseMultiplier() float64 {
	return logic.c2LoseMultiplier
}
func (logic *IndoLogic) D2LoseMultiplier() float64 {
	return logic.d2LoseMultiplier
}
func (logic *IndoLogic) S2LoseMultiplier() float64 {
	return logic.s2LoseMultiplier
}
func (logic *IndoLogic) H2LoseMultiplier() float64 {
	return logic.h2LoseMultiplier
}

func (logic *IndoLogic) CardValueOrder() []string {
	return logic.cardValueOrder
}
func (logic *IndoLogic) CardSuitOrder() []string {
	return logic.cardSuitOrder
}
func (logic *IndoLogic) TypeOrder() []string {
	return logic.typeOrder
}

func (logic *IndoLogic) HasInstantWin() bool {
	return false
}
func (logic *IndoLogic) AllowOverrideType() bool {
	return true
}
func (logic *IndoLogic) Include2InGroupCards() bool {
	return true
}

func (logic *IndoLogic) LoseMultiplierByCardLeft(cards []string, willCountStuffs bool) (multiplier float64, cardTypes []string) {
	cardTypes = make([]string, 0)

	var twoMultiplier float64
	if containCards(cards, []string{"s 2"}) {
		twoMultiplier += logic.S2LoseMultiplier()
		cardTypes = append(cardTypes, "s_2")
	}
	if containCards(cards, []string{"c 2"}) {
		twoMultiplier += logic.C2LoseMultiplier()
		cardTypes = append(cardTypes, "c_2")
	}
	if containCards(cards, []string{"d 2"}) {
		twoMultiplier += logic.D2LoseMultiplier()
		cardTypes = append(cardTypes, "d_2")
	}
	if containCards(cards, []string{"h 2"}) {
		twoMultiplier += logic.H2LoseMultiplier()
		cardTypes = append(cardTypes, "h_2")
	}

	var cardsNumMultiplier float64
	if len(cards) <= 9 {
		cardsNumMultiplier = 1
	} else if len(cards) <= 12 {
		cardsNumMultiplier = 2
	} else {
		cardsNumMultiplier = 3
	}

	multiplier = (float64(len(cards)) + twoMultiplier) * cardsNumMultiplier
	return multiplier, cardTypes
}

func (logic *IndoLogic) GetInstantWinType(sortedCards []string) string {
	return ""
}
func (logic *IndoLogic) GetCardTableMoveType(sortedCards []string) (moveType int) {
	return 0 // duoc thuong 0 lan tien ba

}
func (logic *IndoLogic) GetMoveType(sortedCards []string) (moveType string) {
	if len(sortedCards) == 0 {
		return MoveTypeInvalid
	}
	if len(sortedCards) == 1 {
		return MoveTypeOneCard
	}

	// 2 cards same value
	if len(sortedCards) == 2 {
		if components.IsCardValueEqual(sortedCards[0], sortedCards[1]) {
			return MoveTypeDoubleCards
		}
		return MoveTypeInvalid
	}

	// 3,4 cards in order, or same value
	if len(sortedCards) == 3 {
		if isSameValue(sortedCards) {
			return MoveTypeTripleCards
		}
	}

	if len(sortedCards) == 4 {
		if isSameValue(sortedCards) {
			return MoveTypeQuadrupleCards
		}
	}

	// 5 cards same order
	if len(sortedCards) == 5 {
		if isStraightFlush(logic, sortedCards) {
			return MoveTypeStraightFlush
		}

		if isFullHouse(sortedCards) {
			return MoveTypeFullHouse
		}

		if isStreak(logic, sortedCards) {
			return MoveType5OrderCards
		}

		if isFlush(sortedCards) {
			return MoveType5FlushCards
		}
	}

	return MoveTypeInvalid
}

func (logic *IndoLogic) IsCards1BiggerThanCards2(cards1 []string, cards2 []string) bool {
	if len(cards1) != len(cards2) {
		return false
	}

	cardsType1 := logic.GetMoveType(cards1)
	cardsType2 := logic.GetMoveType(cards2)

	if cardsType1 == MoveTypeInvalid || cardsType2 == MoveTypeInvalid {
		return false
	}

	if cardsType1 != cardsType2 {
		if len(cards1) == 5 {
			typeIndex1 := getTypeIndex(logic, cardsType1)
			typeIndex2 := getTypeIndex(logic, cardsType2)
			return typeIndex1 > typeIndex2
		} else {
			return false
		}
	}
	if utils.ContainsByString([]string{MoveTypeOneCard,
		MoveTypeDoubleCards, MoveTypeTripleCards, MoveTypeQuadrupleCards}, cardsType1) {
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]
		return logic.isCard1BiggerThanCard2SingleValueGroup(lastCard1, lastCard2)
	} else if cardsType1 == MoveTypeFullHouse {
		lastCard1 := getBiggestCardInGroupOfSameCardValueFromCards(cards1, 3)
		lastCard2 := getBiggestCardInGroupOfSameCardValueFromCards(cards2, 3)
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	} else if cardsType1 == MoveType5OrderCards {
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]
		firstCard1 := cards1[0]
		firstCard2 := cards2[0]
		if getValueAsIntFromCard(logic, firstCard1) == 0 &&
			getValueAsIntFromCard(logic, lastCard1) == 12 {
			lastCard1 = cards1[len(cards1)-2]
		}
		if getValueAsIntFromCard(logic, firstCard2) == 0 &&
			getValueAsIntFromCard(logic, lastCard2) == 12 {
			lastCard2 = cards2[len(cards2)-2]
		}
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	} else {
		// compare by number
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	}
}

func (logic *IndoLogic) isCard1BiggerThanCard2SingleValueGroup(card1 string, card2 string) bool {
	suit1, value1 := components.SuitAndValueFromCard(card1)
	suit2, value2 := components.SuitAndValueFromCard(card2)

	value1AsInt := valueAsInt(logic, value1)
	value2AsInt := valueAsInt(logic, value2)

	if value1AsInt == 0 {
		value1AsInt = 13 // hard code biggest value
	}

	if value2AsInt == 0 {
		value2AsInt = 13 // hard code biggest value
	}

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
