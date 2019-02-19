package tienlen2

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
}

func TestCheckIsStraight(t *testing.T) {
	var cards []z.Card
	cards, _ = z.ToCardsFromStrings([]string{"5c", "4d", "3h"})
	if checkIsStraight(cards) != true {
		t.Error()
	}
	cards, _ = z.ToCardsFromStrings([]string{"Ac", "Qd", "Kh"})
	cards = SortedByRank(cards)
	if checkIsStraight(cards) != true {
		t.Error()
	}
	cards, _ = z.ToCardsFromStrings([]string{"2c", "Ad", "Kh"})
	if checkIsStraight(cards) != false {
		t.Error()
	}
	cards, _ = z.ToCardsFromStrings([]string{"Kc", "Qd", "Qh"})
	if checkIsStraight(cards) != false {
		t.Error()
	}
}

func TestGetNext(t *testing.T) {
	array := []int64{9, 8, 7, 6, 5, 4}
	if SliceGetNext(9, array) != 8 {
		t.Error()
	}
	if SliceGetNext(7, array) != 6 {
		t.Error()
	}
	if SliceGetNext(5, array) != 4 {
		t.Error()
	}
	if SliceGetNext(4, array) != 9 {
		t.Error()
	}
}

func TestCalcSplitWays(t *testing.T) {
	//	var ways [][][]z.Card
	//	var hand []z.Card
	//
	//	//	hand, _ = z.ToCardsFromStrings([]string{"7c", "8d", "6h", "8h", "9c", "9s"})
	//	hand, _ = z.ToCardsFromStrings([]string{"7c", "6h", "8h", "9c", "9s"})
	//
	//	ways = CalcSplitWays(hand)
	//	for _, way := range ways {
	//		_ = way
	//		//		fmt.Println("way", way)
	//	}
	//
	//	minWay := CalcWayToMinNSingleCards(hand)
	//	fmt.Println("minWay", minWay)
}

func TestTienlenBoard(t *testing.T) {
	var hand1, hand2, hand3 []z.Card
	var board *TienlenBoard
	var isMoveSuccessfully bool
	var theBestMove []z.Card
	_ = []interface{}{hand1, hand2, hand3, board, isMoveSuccessfully, theBestMove}

	// ________________________________________
	// case 1
	hand1, _ = z.ToCardsFromStrings([]string{"7c", "Jc", "6c", "6d"})
	hand2, _ = z.ToCardsFromStrings([]string{"8d", "8h", "9c", "9s"})
	hand3, _ = z.ToCardsFromStrings([]string{"Td", "Th"})
	board = &TienlenBoard{
		Order:               []int64{1, 2, 3},
		IsFirstTurnInMatch:  false,
		IsFirstTurnInRound:  true,
		PlayersInRound:      []int64{1, 2},
		MapPlayerToHand:     map[int64][]z.Card{1: hand1, 2: hand2, 3: hand3},
		CurrentTurnPlayer:   1,
		CurrentComboOnBoard: []z.Card{},
	}

	board.PPrint()
	// fmt.Println(board.CalcAllValidMoves())
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	fmt.Println("theBestMove", theBestMove)
	if len(z.Subtracted(theBestMove, []z.Card{z.FNewCardFS("7c")})) != 0 {
		t.Error()
	}

	// ________________________________________
	// case 2
	hand1, _ = z.ToCardsFromStrings([]string{"Jc", "Jd", "6c", "6d"})
	hand2, _ = z.ToCardsFromStrings([]string{"9d", "9c"})
	board = &TienlenBoard{
		Order:               []int64{1, 2},
		IsFirstTurnInMatch:  false,
		IsFirstTurnInRound:  false,
		PlayersInRound:      []int64{1, 2},
		MapPlayerToHand:     map[int64][]z.Card{1: hand1, 2: hand2},
		CurrentTurnPlayer:   1,
		CurrentComboOnBoard: []z.Card{}, //z.FNewCardFS("3d")
	}

	board.PPrint()
	// fmt.Println(board.CalcAllValidMoves())
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	fmt.Println("theBestMove", theBestMove)
	if len(z.Subtracted(theBestMove, []z.Card{z.FNewCardFS("7c")})) != 0 {
		t.Error()
	}

	// ________________________________________
	// case 3
	st := time.Now()
	hand1, _ = z.ToCardsFromStrings([]string{"4h", "4d", "Qh", "Qd"})
	hand2, _ = z.ToCardsFromStrings([]string{"Tc", "Th"}) //
	board = &TienlenBoard{
		Order:               []int64{1, 2},
		IsFirstTurnInMatch:  false,
		IsFirstTurnInRound:  false,
		PlayersInRound:      []int64{1, 2},
		MapPlayerToHand:     map[int64][]z.Card{1: hand1, 2: hand2},
		CurrentTurnPlayer:   1,
		CurrentComboOnBoard: []z.Card{},
	}

	board.PPrint()
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	fmt.Println("theBestMove", theBestMove)
	if len(z.Subtracted(theBestMove, []z.Card{z.FNewCardFS("4d")})) != 0 {
		t.Error()
	}
	ft := time.Now()
	fmt.Println("dur", ft.Sub(st))
}
