package logic

type TLLogic interface {
	C2LoseMultiplier() float64
	D2LoseMultiplier() float64
	S2LoseMultiplier() float64
	H2LoseMultiplier() float64

	CardValueOrder() []string
	CardSuitOrder() []string
	TypeOrder() []string

	HasInstantWin() bool

	AllowOverrideType() bool
	Include2InGroupCards() bool

	GetMoveType(cards []string) string
	GetInstantWinType(cards []string) string
	LoseMultiplierByCardLeft(cards []string, willCountStuffs bool) (multiplier float64, cardTypes []string)
	IsCards1BiggerThanCards2(cards1 []string, cards2 []string) bool
	GetCardTableMoveType(cards1 []string) int
}

const (
	MoveTypeInvalid = "invalid"
	MoveTypeOneCard = "one_card"

	MoveTypeDoubleCards    = "dup_cards"
	MoveTypeTripleCards    = "tri_cards"
	MoveTypeQuadrupleCards = "qua_cards"

	MoveType3OrderCards  = "3_order"
	MoveType4OrderCards  = "4_order"
	MoveType5OrderCards  = "5_order"
	MoveType6OrderCards  = "6_order"
	MoveType7OrderCards  = "7_order"
	MoveType8OrderCards  = "8_order"
	MoveType9OrderCards  = "9_order"
	MoveType10OrderCards = "10_order"
	MoveType11OrderCards = "11_order"
	MoveType12OrderCards = "12_order"

	MoveType5FlushCards = "5_flush"

	MoveTypeStraightFlush = "straight_flush"

	MoveTypeFullHouse              = "fullhouse"
	MoveTypeDoubleCards3TimesOrder = "double_cards_3_order"
	MoveTypeDoubleCards4TimesOrder = "double_cards_4_order"
)
