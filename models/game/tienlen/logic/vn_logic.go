package logic

import (
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/utils"
)

const (
	Single2Count                       bool    = true
	TriDoubleCardsStreakLoseMultiplier float64 = 6
	QuadrupleCardsLoseMultiplier       float64 = 7
	QuaDoubleCardsStreakLoseMultiplier float64 = 8
)

type VNLogic struct {
	c2LoseMultiplier float64
	h2LoseMultiplier float64
	d2LoseMultiplier float64
	s2LoseMultiplier float64

	cardValueOrder []string
	cardSuitOrder  []string
	typeOrder      []string
}

func NewVNLogic() *VNLogic {
	return &VNLogic{
		c2LoseMultiplier: 1,
		d2LoseMultiplier: 1,
		s2LoseMultiplier: 2,
		h2LoseMultiplier: 2,
		cardValueOrder:   []string{"3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a", "2"},
		cardSuitOrder:    []string{"s", "c", "d", "h"},
		typeOrder:        []string{},
	}
}

func (logic *VNLogic) C2LoseMultiplier() float64 {
	return logic.c2LoseMultiplier
}
func (logic *VNLogic) D2LoseMultiplier() float64 {
	return logic.d2LoseMultiplier
}
func (logic *VNLogic) S2LoseMultiplier() float64 {
	return logic.s2LoseMultiplier
}
func (logic *VNLogic) H2LoseMultiplier() float64 {
	return logic.h2LoseMultiplier
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

func (logic *VNLogic) HasInstantWin() bool {
	return true
}

func (logic *VNLogic) AllowOverrideType() bool {
	return false
}

func (logic *VNLogic) Include2InGroupCards() bool {
	return false
}
func (logic *VNLogic) LoseMultiplierByCardLeft(cards []string, willCountStuffs bool) (multiplier float64, cardTypes []string) {
	cardTypes = make([]string, 0)
	multiplier = float64(len(cards))
	if willCountStuffs == false { // thang trang
		multiplier = multiplier * 2
		return multiplier, cards
	}
	if len(cards) == 13 { // thua cong
		multiplier = multiplier * 3
		return multiplier, cards
	}
	if willCountStuffs {

		// check 2 cards
		if containCards(cards, []string{"s 2"}) {
			multiplier = multiplier + logic.S2LoseMultiplier()
			cardTypes = append(cardTypes, "s_2")
		}
		if containCards(cards, []string{"c 2"}) {
			multiplier = multiplier + logic.C2LoseMultiplier()
			cardTypes = append(cardTypes, "c_2")
		}
		if containCards(cards, []string{"d 2"}) {
			multiplier = multiplier + logic.D2LoseMultiplier()
			cardTypes = append(cardTypes, "d_2")
		}
		if containCards(cards, []string{"h 2"}) {
			multiplier = multiplier + logic.H2LoseMultiplier()
			cardTypes = append(cardTypes, "h_2")
		}

		if containXDoubleCardsInStreak(logic, cards, 4) {
			multiplier = multiplier + QuaDoubleCardsStreakLoseMultiplier
			cardTypes = append(cardTypes, MoveTypeDoubleCards4TimesOrder)
		} else if containXDoubleCardsInStreak(logic, cards, 3) {
			multiplier = multiplier + TriDoubleCardsStreakLoseMultiplier
			cardTypes = append(cardTypes, MoveTypeDoubleCards3TimesOrder)
		}
		num := numberOfQuadrupleCards(cards)
		if num > 0 {
			multiplier = multiplier + float64(num)*QuadrupleCardsLoseMultiplier
			cardTypes = append(cardTypes, MoveTypeQuadrupleCards)
		}
	}

	return multiplier, cardTypes
}

func (logic *VNLogic) GetInstantWinType(sortedCards []string) string {
	// 12 cards in order
	if is12CardsInOrderAndIncreaseBy1(logic, sortedCards) {
		return "12_cards_straight"
	}
	// 6 dup cards
	if is6DupsCards(logic, sortedCards) {
		return "6_dup"
	}

	// 5 dup cards in order
	if containXDoubleCardsInStreak(logic, sortedCards, 5) {
		return "5_pairs_straight"
	}

	// 4 triple cards
	if Is4TripleCards(logic, sortedCards) {
		return "4_trip"
	}

	// 4 12 cards
	if isFour2Cards(sortedCards) {
		return "4_2_cards"
	}

	if isSameColor(sortedCards) {
		return "same_color"
	}

	return ""
}

func (logic *VNLogic) GetCardTableMoveType(sortedCards []string) (moveType int) {
	if len(sortedCards) == 0 {
		return 0 // duoc thuong 0 lan tien ba
	}
	if len(sortedCards) == 1 { // set truong hop 2 bi phat
		suit2, value2 := components.SuitAndValueFromCard(sortedCards[0])
		if value2 != "2" {
			return 0
		}
		if suit2 == "c" || suit2 == "d" {
			return 2 // phat 2 den 1 lan
		}
		return 4 // phat 2 do 2 lan
	}

	// xet truong hop doi 2 bi phat
	if len(sortedCards) == 2 {
		suit2, value2 := components.SuitAndValueFromCard(sortedCards[0])
		suit3, value3 := components.SuitAndValueFromCard(sortedCards[1])
		if value2 == "2" && value2 == value3 {
			returnvalue := 0
			if suit2 == "c" || suit2 == "d" {
				returnvalue = returnvalue + 2
			} else {
				returnvalue = returnvalue + 4
			}
			if suit3 == "c" || suit3 == "d" {
				returnvalue = returnvalue + 2
			} else {
				returnvalue = returnvalue + 4
			}

			return returnvalue // phat doi 2
		}
		return 0
	}

	if len(sortedCards) == 4 {
		if isSameValue(sortedCards) {
			return 7 // tu quy de tu quy
		}
		if isStreak(logic, sortedCards) {
			return 0
		}
		return 0
	}

	// 6,8 cards same order or three couple increasing order
	if len(sortedCards) == 6 {
		if isStreak(logic, sortedCards) {
			return 0
		}
		if isDupCardsInOrderAndIncreaseByOne(logic, sortedCards) {
			// 3 doi thong bi phat duoc thuong 2 lan
			return 6 //phat 3 doi thong
		}
		return 0
	}
	if len(sortedCards) == 8 {
		if isStreak(logic, sortedCards) {
			return 0
		}
		if isDupCardsInOrderAndIncreaseByOne(logic, sortedCards) {
			// 4 doi thong bi phat duoc thuong 2 lan
			return 8 // phat 4 doi thong
		}
	}

	return 0
}

func (logic *VNLogic) GetMoveType(sortedCards []string) (moveType string) {
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
		if isStreak(logic, sortedCards) {
			return MoveType3OrderCards
		}
	}

	if len(sortedCards) == 4 {
		if isSameValue(sortedCards) {
			return MoveTypeQuadrupleCards
		}
		if isStreak(logic, sortedCards) {
			return MoveType4OrderCards
		}
	}

	// 5 cards same order
	if len(sortedCards) == 5 {
		if isStreak(logic, sortedCards) {
			return MoveType5OrderCards
		}
	}

	// 6,8 cards same order or three couple increasing order
	if len(sortedCards) == 6 {
		if isStreak(logic, sortedCards) {
			return MoveType6OrderCards
		}
		if isDupCardsInOrderAndIncreaseByOne(logic, sortedCards) {
			return MoveTypeDoubleCards3TimesOrder
		}
	}

	// 7 cards in order
	if len(sortedCards) == 7 {
		if isStreak(logic, sortedCards) {
			return MoveType7OrderCards
		}
	}
	if len(sortedCards) == 8 {
		if isStreak(logic, sortedCards) {
			return MoveType8OrderCards
		}
		if isDupCardsInOrderAndIncreaseByOne(logic, sortedCards) {
			return MoveTypeDoubleCards4TimesOrder
		}
	}

	//if len(sortedCards) == 9 {
	//if isStreak(logic, sortedCards) {
	//	return MjackpotCode := "all"
	//	jackpotInstance := jackpot.GetJackpot(jackpotCode, gameSession.game.currencyType)
	//	if jackpotInstance != nil {
	//		if gameSession.sessionCallback.GetNumberOfHumans() != 0 {
	//			moneyToJackpot := int64(float64(gameSession.tax) * float64(1/3.0))
	//			jackpotInstance.AddMoney(moneyToJackpot)
	//
	//			isHitJackpot := false
	//			if winnerResult.instantWinType == "12_cards_straight" {
	//				isHitJackpot = true
	//			}
	//						if isHitJackpot {
	//							if ratio, isIn := mapMoneyUnitToJackpotRatio[gameSession.betEntry.Min()]; isIn {
	//								temp := int64(float64(jackpotInstance.Value()) * ratio)
	//								jackpotInstance.AddMoney(-temp)
	//
	//								winner.ChangeMoneyAndLog(
	//									temp, gameSession.currencyType, false, "",
	//									"JACKPOT", gameSession.game.GameCode(), "")
	//
	//								jackpotInstance.NotifySomeoneHitJackpot(
	//									gameSession.game.GameCode(),
	//									temp,
	//									winner.Id(),
	//									winner.Name(),
	//								)
	//
	//								if winner.PlayerType() == "normal" {
	//									gameSession.win += temp
	//								}
	//							}
	//						}
	//		}
	//	}oveType9OrderCards
	//		}
	//	}
	if len(sortedCards) == 10 {
		if isStreak(logic, sortedCards) {
			return MoveType10OrderCards
		}
	}
	if len(sortedCards) == 11 {
		if isStreak(logic, sortedCards) {
			return MoveType11OrderCards
		}
	}

	return MoveTypeInvalid
}

func (logic *VNLogic) IsCards1BiggerThanCards2(cards1 []string, cards2 []string) bool {
	cardsType1 := logic.GetMoveType(cards1)
	cardsType2 := logic.GetMoveType(cards2)

	if cardsType1 == cardsType2 && cardsType1 == MoveTypeFullHouse {
		lastCard1 := getBiggestCardInGroupOfSameCardValueFromCards(cards1, 3)
		lastCard2 := getBiggestCardInGroupOfSameCardValueFromCards(cards2, 3)
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	} else if cardsType1 == cardsType2 && utils.ContainsByString([]string{MoveType3OrderCards, MoveType4OrderCards, MoveType5OrderCards}, cardsType1) {
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]

		suit1, value1 := components.SuitAndValueFromCard(lastCard1)
		suit2, value2 := components.SuitAndValueFromCard(lastCard2)

		value1AsInt := valueAsInt(logic, value1)
		value2AsInt := valueAsInt(logic, value2)
		if value1AsInt == value2AsInt {
			isStraightFlush1 := isStraightFlush(logic, cards1)
			isStraightFlush2 := isStraightFlush(logic, cards2)

			if (isStraightFlush1 && isStraightFlush2) ||
				(!isStraightFlush1 && !isStraightFlush2) {
				suit1AsInt := suitAsInt(logic, suit1)
				suit2AsInt := suitAsInt(logic, suit2)
				if suit1AsInt > suit2AsInt {
					return true
				} else {
					return false
				}
			} else if isStraightFlush1 {
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
	} else if cardsType1 == cardsType2 {
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	} else {
		_, value2 := components.SuitAndValueFromCard(cards2[0])
		// 2 hoac 3 doi thong bi bat boi ba doi thong, 4 doi thong, tu quy
		if ((cardsType2 == MoveTypeOneCard && value2 == "2") ||
			cardsType2 == MoveTypeDoubleCards3TimesOrder) &&
			(cardsType1 == MoveTypeQuadrupleCards ||
				cardsType1 == MoveTypeDoubleCards3TimesOrder ||
				cardsType1 == MoveTypeDoubleCards4TimesOrder) { // cay tren ban la cay 2
			return true
		}
		//  doi 2 bi bat boi 4 doi thong hoac  tu quy
		if cardsType2 == MoveTypeDoubleCards && value2 == "2" &&
			(cardsType1 == MoveTypeQuadrupleCards ||
				cardsType1 == MoveTypeDoubleCards4TimesOrder) {
			return true
		}
		//   4 doi thong bi bat boi tu quy
		if cardsType2 == MoveTypeQuadrupleCards &&
			cardsType1 == MoveTypeDoubleCards4TimesOrder {
			return true
		}
		if cardsType1 != cardsType2 {
			return false
		}
		// compare by number
		lastCard1 := cards1[len(cards1)-1]
		lastCard2 := cards2[len(cards2)-1]
		return IsCard1BiggerThanCard2(logic, lastCard1, lastCard2)
	}
}
