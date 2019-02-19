package logic

import ()

type VNLogic struct {
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

func NewVNLogic() *VNLogic {
	return &VNLogic{
		specialCompareMultiplier: 2,
		collapseMultiplier:       2,
		winCollapseAllMultiplier: 2,
		cardValueOrder:           []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a"},
		cardSuitOrder:            []string{"s", "c", "d", "h"},
		typeOrder:                []string{TypeHighCard, TypePair, TypeTwoPair, TypeThreeOfAKind, TypeThreeOfAces, TypeStraight, TypeFlush, TypeFullHouse, TypeFourOfAKind, TypeFourAces, TypeStraightFlush, TypeStraightFlushBottom, TypeStraightFlushTop},
		whiteWinMultiplier: map[string]float64{
			WhiteWinTypeDragonRollingStraight:   28,
			WhiteWinTypeDragonStraight:          26,
			WhiteWinTypeFivePairOneThreeOfAKind: 18,
			WhiteWinTypeThreeFlush:              18,
			WhiteWinTypeThreeStraight:           18,
			WhiteWinTypeSixPair:                 18,
		},
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

func (logic *VNLogic) SpecialCompareMultiplier() float64 {
	return logic.specialCompareMultiplier
}
func (logic *VNLogic) CollapseMultiplier() float64 {
	return logic.collapseMultiplier
}
func (logic *VNLogic) WinCollapseAllMultiplier() float64 {
	return logic.winCollapseAllMultiplier
}
func (logic *VNLogic) CardValueOrder() []string {
	return logic.cardValueOrder
}
func (logic *VNLogic) CardSuitOrder() []string {
	return logic.cardSuitOrder
}
func (logic *VNLogic) TypeOrder() []string {
	return logic.typeOrder
}
func (logic *VNLogic) WhiteWinMultiplier() map[string]float64 {
	return logic.whiteWinMultiplier
}
func (logic *VNLogic) WinMultiplierForTopPart() map[string]float64 {
	return logic.winMultiplierForTopPart
}
func (logic *VNLogic) WinMultiplierForMiddlePart() map[string]float64 {
	return logic.winMultiplierForMiddlePart
}
func (logic *VNLogic) WinMultiplierForBottomPart() map[string]float64 {
	return logic.winMultiplierForBottomPart
}
func (logic *VNLogic) HasWhiteWin() bool {
	return true
}
func (logic *VNLogic) HasCountAces() bool {
	return true
}
