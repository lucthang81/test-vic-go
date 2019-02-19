package cardgame

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

const (
	POKER_TYPE_HIGH_CARD      = "POKER_TYPE_HIGH_CARD"
	POKER_TYPE_PAIR           = "POKER_TYPE_PAIR"
	POKER_TYPE_TWO_PAIR       = "POKER_TYPE_TWO_PAIR"
	POKER_TYPE_TRIPS          = "POKER_TYPE_TRIPS"
	POKER_TYPE_STRAIGHT       = "POKER_TYPE_STRAIGHT"
	POKER_TYPE_FLUSH          = "POKER_TYPE_FLUSH"
	POKER_TYPE_FULL_HOUSE     = "POKER_TYPE_FULL_HOUSE"
	POKER_TYPE_QUADS          = "POKER_TYPE_QUADS"
	POKER_TYPE_STRAIGHT_FLUSH = "POKER_TYPE_STRAIGHT_FLUSH"
	POKER_TYPE_ROYAL_FLUSH    = "POKER_TYPE_ROYAL_FLUSH"
)

// typical rank order: A = 14 > K > .. > 3 > 2
var MapRankToInt map[string]int

// typical suit order: heart > diamond > club > spade
var MapSuitToInt map[string]int

// map poker type string to int 0 highCard -> 9 royalFlush
var PokerType map[string]int

func init() {
	fmt.Print("")
	_, _ = json.Marshal([]int{})
	_ = errors.New("")
	//
	MapRankToInt = map[string]int{
		"A": 14, "2": 2, "3": 3, "4": 4, "5": 5,
		"6": 6, "7": 7, "8": 8, "9": 9, "T": 10,
		"J": 11, "Q": 12, "K": 13}
	MapSuitToInt = map[string]int{"s": 0, "c": 1, "d": 2, "h": 3} // not importance
	PokerType = map[string]int{
		POKER_TYPE_HIGH_CARD:      0,
		POKER_TYPE_PAIR:           1,
		POKER_TYPE_TWO_PAIR:       2,
		POKER_TYPE_TRIPS:          3,
		POKER_TYPE_STRAIGHT:       4,
		POKER_TYPE_FLUSH:          5,
		POKER_TYPE_FULL_HOUSE:     6,
		POKER_TYPE_QUADS:          7,
		POKER_TYPE_STRAIGHT_FLUSH: 8,
		POKER_TYPE_ROYAL_FLUSH:    9,
	}
}

// sort by rank, suit; decrease
type ByRank []Card

func (a ByRank) Len() int      { return len(a) }
func (a ByRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRank) Less(i, j int) bool {
	if MapRankToInt[a[i].Rank] > MapRankToInt[a[j].Rank] {
		return true
	} else if MapRankToInt[a[i].Rank] == MapRankToInt[a[j].Rank] {
		if MapSuitToInt[a[i].Suit] > MapSuitToInt[a[j].Suit] {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// A is max
func SortedByRank(cards []Card) []Card {
	result := make([]Card, len(cards))
	copy(result, cards)
	sort.Sort(ByRank(result))
	return result
}

//
func CalcRankPoker5Cards(cards []Card) []int {
	h := SortedByRank(cards)
	h0 := h[0]
	h1 := h[1]
	h2 := h[2]
	h3 := h[3]
	h4 := h[4]
	h0rank := MapRankToInt[h0.Rank]
	h1rank := MapRankToInt[h1.Rank]
	h2rank := MapRankToInt[h2.Rank]
	h3rank := MapRankToInt[h3.Rank]
	h4rank := MapRankToInt[h4.Rank]
	//
	var flush = (h0.Suit == h1.Suit) && (h1.Suit == h2.Suit) && (h2.Suit == h3.Suit) && (h3.Suit == h4.Suit)
	var straightN = (h0rank == h1rank+1) && (h1rank == h2rank+1) && (h2rank == h3rank+1) && (h3rank == h4rank+1)
	var straight1 = (h0rank == 14) && (h1rank == 5) && (h2rank == 4) && (h3rank == 3) && (h4rank == 2)
	var straight = straight1 || straightN
	//
	if straight && flush {
		if straightN {
			if h0rank == 14 {
				return []int{PokerType[POKER_TYPE_ROYAL_FLUSH]}
			} else {
				return []int{PokerType[POKER_TYPE_STRAIGHT_FLUSH], h0rank}
			}
		} else {
			return []int{PokerType[POKER_TYPE_STRAIGHT_FLUSH], h1rank}
		}
	}
	//
	if (h0rank == h1rank) && (h0rank == h2rank) && (h0rank == h3rank) {
		return []int{PokerType[POKER_TYPE_QUADS], h1rank}
	}
	if (h1rank == h2rank) && (h1rank == h3rank) && (h1rank == h4rank) {
		return []int{PokerType[POKER_TYPE_QUADS], h1rank}
	}
	//
	if (h0rank == h1rank) && (h1rank == h2rank) && (h3rank == h4rank) {
		return []int{PokerType[POKER_TYPE_FULL_HOUSE], h0rank, h3rank}
	}
	if (h0rank == h1rank) && (h2rank == h3rank) && (h3rank == h4rank) {
		return []int{PokerType[POKER_TYPE_FULL_HOUSE], h2rank, h0rank}
	}
	//
	if flush {
		return []int{PokerType[POKER_TYPE_FLUSH], h0rank, h1rank, h2rank, h3rank, h4rank}
	}
	//
	if straightN {
		return []int{PokerType[POKER_TYPE_STRAIGHT], h0rank}
	}
	if straight1 {
		return []int{PokerType[POKER_TYPE_STRAIGHT], h1rank}
	}
	//
	if (h0rank == h1rank) && (h1rank == h2rank) {
		return []int{PokerType[POKER_TYPE_TRIPS], h0rank, h3rank, h4rank}
	}
	if (h1rank == h2rank) && (h2rank == h3rank) {
		return []int{PokerType[POKER_TYPE_TRIPS], h1rank, h0rank, h4rank}
	}
	if (h2rank == h3rank) && (h3rank == h4rank) {
		return []int{PokerType[POKER_TYPE_TRIPS], h2rank, h0rank, h1rank}
	}
	//
	if (h0rank == h1rank) && (h2rank == h3rank) {
		return []int{PokerType[POKER_TYPE_TWO_PAIR], h0rank, h2rank, h4rank}
	}
	if (h0rank == h1rank) && (h3rank == h4rank) {
		return []int{PokerType[POKER_TYPE_TWO_PAIR], h0rank, h3rank, h2rank}
	}
	if (h1rank == h2rank) && (h3rank == h4rank) {
		return []int{PokerType[POKER_TYPE_TWO_PAIR], h1rank, h3rank, h0rank}
	}
	//
	if h0rank == h1rank {
		return []int{PokerType[POKER_TYPE_PAIR], h0rank, h2rank, h3rank, h4rank}
	}
	if h1rank == h2rank {
		return []int{PokerType[POKER_TYPE_PAIR], h1rank, h0rank, h3rank, h4rank}
	}
	if h2rank == h3rank {
		return []int{PokerType[POKER_TYPE_PAIR], h2rank, h0rank, h1rank, h4rank}
	}
	if h3rank == h4rank {
		return []int{PokerType[POKER_TYPE_PAIR], h3rank, h0rank, h1rank, h2rank}
	}
	//
	return []int{PokerType[POKER_TYPE_HIGH_CARD], h0rank, h1rank, h2rank, h3rank, h4rank}
}

func CalcRankPoker7Cards(cards []Card) (
	[]int, []Card) {
	sh := SortedByRank(cards)
	var bestRank []int
	var bestComb5 []Card
	for _, comb5 := range GetCombinationsForCards(sh, 5) {
		rank := CalcRankPoker5Cards(comb5)
		if Compare2ListInt(rank, bestRank) {
			bestRank = rank
			bestComb5 = comb5
		}
	}
	return bestRank, bestComb5
}
