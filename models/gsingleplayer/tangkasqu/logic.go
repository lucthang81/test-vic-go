package tangkasqu

import (
	//	"errors"
	"fmt"
	//	"math/rand"
	//	"sort"
	//	"strings"

	z "github.com/vic/vic_go/models/cardgame"
)

const (
	T_ROYAL_FLUSH    = "TYPE_ROYAL_FLUSH"
	T_5_OF_A_KIND    = "T_5_OF_A_KIND"
	T_STRAIGHT_FLUSH = "T_STRAIGHT_FLUSH"
	T_4_OF_A_KIND    = "T_4_OF_A_KIND"
	T_FULL_HOUSE     = "T_FULL_HOUSE"
	T_FLUSH          = "T_FLUSH"
	T_STRAIGHT       = "T_STRAIGHT"
	T_3_OF_A_KIND    = "T_3_OF_A_KIND"
	// two pair, at least one pair is greater equal than 10
	T_TWO_PAIR_1GET10 = "T_TWO_PAIR_1GET10"
	T_PAIR_ACE        = "T_PAIR_ACE"
	T_NOTHING         = "T_NOTHING"
)

var MapTypeToInt = map[string]int{
	T_ROYAL_FLUSH:     10,
	T_5_OF_A_KIND:     9,
	T_STRAIGHT_FLUSH:  8,
	T_4_OF_A_KIND:     7,
	T_FULL_HOUSE:      6,
	T_FLUSH:           5,
	T_STRAIGHT:        4,
	T_3_OF_A_KIND:     3,
	T_TWO_PAIR_1GET10: 2,
	T_PAIR_ACE:        1,
	T_NOTHING:         0,
}

var MapTypeToPrize = map[string]float64{
	T_ROYAL_FLUSH:     500,
	T_5_OF_A_KIND:     200,
	T_STRAIGHT_FLUSH:  120,
	T_4_OF_A_KIND:     50,
	T_FULL_HOUSE:      7,
	T_FLUSH:           5,
	T_STRAIGHT:        3,
	T_3_OF_A_KIND:     2,
	T_TWO_PAIR_1GET10: 1,
	T_PAIR_ACE:        1,
}

// Bonus prize for top 4 types, doesnt depend on NBets
var MapNBetsToPrize2 = map[int64]float64{
	1: 100,
	2: 40,
	3: 20,
	4: 10,
}

// fullhouse predict prize, increase by NBets
var Prize3 = float64(10)

func init() {
	_ = fmt.Println
}

// 54 cards deck: Zr, Zb + 52 basic cards,
// after call this func, should call Shuffle(result)
func NewDeck(is2Jokers bool) []z.Card {
	deck := []z.Card{}
	deck = append(deck, z.Card{Rank: "Z", Suit: "r"})
	if is2Jokers {
		deck = append(deck, z.Card{Rank: "Z", Suit: "b"})
	}
	for r, _ := range z.MapRankToInt {
		for s, _ := range z.MapSuitToInt {
			deck = append(deck, z.Card{Rank: r, Suit: s})
		}
	}
	deck = z.SortedByRank(deck)
	return deck
}

// input can include wild card
func CalcRankPoker5Card(cards []z.Card) []int {
	h := z.SortedByRank(cards)
	h0, h1, h2, h3, h4 := h[0], h[1], h[2], h[3], h[4]
	if h4.Rank != "Z" { // all cards are not wild card.
		return z.CalcRankPoker5Cards(cards)
	} else { // h4 is joker
		if h3.Rank != "Z" {
			jokerSuit := "s" // spade
			if h0.Suit == h1.Suit && h1.Suit == h2.Suit && h2.Suit == h3.Suit {
				jokerSuit = h0.Suit
			}
			var bestRank []int
			for jokerRank, _ := range z.MapRankToInt {
				rank := z.CalcRankPoker5Cards([]z.Card{
					h0, h1, h2, h3, z.Card{Rank: jokerRank, Suit: jokerSuit}})
				if z.Compare2ListInt(rank, bestRank) {
					bestRank = rank
				}
			}
			return bestRank
		} else { // h3 is joker
			jokerSuit := "s"
			if h0.Suit == h1.Suit && h1.Suit == h2.Suit {
				jokerSuit = h0.Suit
			}
			var bestRank []int
			for j1Rank, _ := range z.MapRankToInt {
				for j2Rank, _ := range z.MapRankToInt {
					rank := z.CalcRankPoker5Cards([]z.Card{
						h0, h1, h2,
						z.Card{Rank: j1Rank, Suit: jokerSuit},
						z.Card{Rank: j2Rank, Suit: jokerSuit}})
					if z.Compare2ListInt(rank, bestRank) {
						bestRank = rank
					}
				}
			}
			return bestRank
		}
	}
}

func CalcRankTangkasqu5Card(cards []z.Card) []int {
	pokerRank := CalcRankPoker5Card(cards)
	// tangkasquRank
	tRank := make([]int, len(pokerRank))
	copy(tRank, pokerRank)
	if pokerRank[0] == z.PokerType[z.POKER_TYPE_ROYAL_FLUSH] {
		tRank[0] = MapTypeToInt[T_ROYAL_FLUSH]
		return tRank
	}
	if pokerRank[0] == z.PokerType[z.POKER_TYPE_QUADS] {
		h := z.SortedByRank(cards)
		h0, h1, h2, h3, h4 := h[0], h[1], h[2], h[3], h[4]
		if h0.Rank == h1.Rank && h1.Rank == h2.Rank && h3.Rank == "Z" && h4.Rank == "Z" {
			tRank[0] = MapTypeToInt[T_5_OF_A_KIND]
			return tRank
		}
		if h0.Rank == h1.Rank && h1.Rank == h2.Rank && h2.Rank == h3.Rank && h4.Rank == "Z" {
			tRank[0] = MapTypeToInt[T_5_OF_A_KIND]
			return tRank
		}
		return pokerRank
	}
	if pokerRank[0] == z.PokerType[z.POKER_TYPE_TWO_PAIR] {
		if pokerRank[1] < 10 {
			tRank[0] = MapTypeToInt[T_NOTHING]
			return tRank
		}
		return pokerRank
	}
	if pokerRank[0] == z.PokerType[z.POKER_TYPE_PAIR] {
		if pokerRank[1] != 14 {
			tRank[0] = MapTypeToInt[T_NOTHING]
			return tRank
		}
		return pokerRank
	}
	return pokerRank
}

// output: typeName, 5bestCards, rankOfFullhouse
func CalcRankTangkasqu7Card(cards []z.Card) (
	string, []z.Card, string) {
	sh := z.SortedByRank(cards)
	var bestRank []int
	var bestComb5 []z.Card
	for _, comb5 := range z.GetCombinationsForCards(sh, 5) {
		rank := CalcRankTangkasqu5Card(comb5)
		if z.Compare2ListInt(rank, bestRank) {
			bestRank = rank
			bestComb5 = comb5
		}
	}
	var bestRankName string
	for rn, v := range MapTypeToInt {
		if len(bestRank) > 0 && bestRank[0] == v {
			bestRankName = rn
			break
		}
	}
	rankOfFullhouse := ""
	if bestRankName == T_FULL_HOUSE {
		for r, i := range z.MapRankToInt {
			if len(bestRank) > 1 && bestRank[1] == i {
				rankOfFullhouse = r
			}
		}
	}
	return bestRankName, bestComb5, rankOfFullhouse
}
