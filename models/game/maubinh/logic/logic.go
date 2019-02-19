package logic

import (
	"fmt"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/components"
)

type MBLogic interface {
	SpecialCompareMultiplier() float64
	CollapseMultiplier() float64
	WinCollapseAllMultiplier() float64

	CardValueOrder() []string
	CardSuitOrder() []string
	TypeOrder() []string

	WhiteWinMultiplier() map[string]float64
	WinMultiplierForTopPart() map[string]float64
	WinMultiplierForMiddlePart() map[string]float64
	WinMultiplierForBottomPart() map[string]float64

	HasWhiteWin() bool
	HasCountAces() bool
}

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

	SpecialCompareMultiplier float64 = 1
	CollapseMultiplier       float64 = 2
)

func ConvertOldStringsToMinahCards(ss []string) []z.Card {
	// oldRank "a", "2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k"
	// oldSuit "c", "d", "h", "s"
	// newRank "A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"
	// newSuit "s", "c", "d", "h"
	result := make([]z.Card, 0)
	for _, s := range ss {
		su, r := components.SuitAndValueFromCard(s)
		if r == "a" {
			r = "A"
		} else if r == "10" {
			r = "T"
		} else if r == "j" {
			r = "J"
		} else if r == "q" {
			r = "Q"
		} else if r == "k" {
			r = "K"
		}
		cardStr := fmt.Sprintf("%v%v", r, su)
		result = append(result, z.FNewCardFS(cardStr))
	}
	return result
}

func ConvertMinahCardsToOldStrings(cards []z.Card) []string {
	result := make([]string, 0)
	for _, card := range cards {
		r := card.Rank
		if r == "A" {
			r = "a"
		} else if r == "T" {
			r = "10"
		} else if r == "J" {
			r = "j"
		} else if r == "Q" {
			r = "q"
		} else if r == "K" {
			r = "k"
		}
		oldStr := fmt.Sprintf("%v %v", card.Suit, r)
		result = append(result, oldStr)
	}
	return result
}
