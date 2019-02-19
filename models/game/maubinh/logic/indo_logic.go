package logic

import ()

type IndoLogic struct {
	specialCompareMultiplier float64
	collapseMultiplier       float64
	winCollapseAllMultiplier float64

	cardValueOrder []string
	cardSuitOrder  []string
	typeOrder      []string

	whiteWinMultiplier         map[string]float64
	winMultiplierForTopPart    map[string]float64
	winMultiplierForMiddlePart map[string]float64
	winMultiplierForBottomPart map[string]float64
}

func NewIndoLogic() *IndoLogic {
	return &IndoLogic{
		specialCompareMultiplier: 2,
		collapseMultiplier:       2,
		winCollapseAllMultiplier: 2,
		cardValueOrder:           []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a"},
		cardSuitOrder:            []string{"d", "c", "h", "s"},
		typeOrder:                []string{TypeHighCard, TypePair, TypeTwoPair, TypeThreeOfAKind, TypeThreeOfAces, TypeStraight, TypeFlush, TypeFullHouse, TypeFourOfAKind, TypeFourAces, TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop},
		whiteWinMultiplier:       map[string]float64{},
		winMultiplierForTopPart: map[string]float64{ // 3 cards
			TypeThreeOfAKind: 6,
			TypeThreeOfAces:  6,
		},
		winMultiplierForMiddlePart: map[string]float64{ // 5 cards
			TypeFullHouse:           4,
			TypeFourOfAKind:         8,
			TypeFourAces:            8,
			TypeStraightFlushTop:    20,
			TypeStraightFlushBottom: 20,
			TypeStraightFlush:       20,
		},
		winMultiplierForBottomPart: map[string]float64{ // 5 cards
			TypeFourOfAKind:         4,
			TypeFourAces:            4,
			TypeStraightFlushTop:    10,
			TypeStraightFlushBottom: 10,
			TypeStraightFlush:       10,
		},
	}
}

func (logic *IndoLogic) SpecialCompareMultiplier() float64 {
	return logic.specialCompareMultiplier
}
func (logic *IndoLogic) CollapseMultiplier() float64 {
	return logic.collapseMultiplier
}
func (logic *IndoLogic) WinCollapseAllMultiplier() float64 {
	return logic.winCollapseAllMultiplier
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
func (logic *IndoLogic) WhiteWinMultiplier() map[string]float64 {
	return logic.whiteWinMultiplier
}
func (logic *IndoLogic) WinMultiplierForTopPart() map[string]float64 {
	return logic.winMultiplierForTopPart
}
func (logic *IndoLogic) WinMultiplierForMiddlePart() map[string]float64 {
	return logic.winMultiplierForMiddlePart
}
func (logic *IndoLogic) WinMultiplierForBottomPart() map[string]float64 {
	return logic.winMultiplierForBottomPart
}
func (logic *IndoLogic) HasWhiteWin() bool {
	return false
}
func (logic *IndoLogic) HasCountAces() bool {
	return false
}
