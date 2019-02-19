package tienlen3

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

	//	board.PPrint()
	// fmt.Println(board.CalcAllValidMoves())
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	//	fmt.Println("theBestMove", theBestMove)
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

	//	board.PPrint()
	// fmt.Println(board.CalcAllValidMoves())
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	//	fmt.Println("theBestMove", theBestMove)
	if len(z.Subtracted(theBestMove, []z.Card{z.FNewCardFS("Jd"), z.FNewCardFS("Jc")})) != 0 {
		t.Error()
	}

	// ________________________________________
	// case 3
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

	//	board.PPrint()
	theBestMove = board.CalcTheBestMove(time.Now().Add(5 * time.Second))
	//	fmt.Println("theBestMove", theBestMove)
	if len(z.Subtracted(theBestMove, []z.Card{z.FNewCardFS("Qd"), z.FNewCardFS("Qh")})) != 0 {
		t.Error()
	}
}

func Test2(t *testing.T) {
	var playersOrder, priorityPlayers []int64
	var baseMoney int64
	var b *TienlenBoard
	var err1, err2, err3, err4, err5, err6, err7, err8, err9 error
	var err11, err12, err13, err14, err15, err16, err17, err18, err19 error
	var err21, err22, err23, err24, err25, err26, err27, err28, err29 error

	// case1
	playersOrder = []int64{10, 11, 12}
	priorityPlayers = []int64{10, 12}
	baseMoney = 100
	b = NewTienlenBoard(playersOrder, priorityPlayers, baseMoney)
	b.StartDealing()
	b.MapPlayerToHand[10], _ = z.ToCardsFromStrings([]string{"3s", "3c", "6c", "6d", "7d", "8c", "Tc", "Th", "Qd", "Qh", "Ac", "Ah", "2s"})
	b.MapPlayerToHand[11], _ = z.ToCardsFromStrings([]string{"4c", "5d", "6h", "7s", "7h", "8s", "8h", "9s", "Td", "Js", "Jc", "Kc", "Kd"})
	b.MapPlayerToHand[12], _ = z.ToCardsFromStrings([]string{"4d", "5h", "7c", "9c", "Ts", "Jd", "Qs", "Qc", "Kh", "As", "Ad", "2d", "2h"})
	b.CurrentTurnPlayer = 10
	err1 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"3c", "3s"})})
	err2 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err3 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{"Qs", "Qc"})})
	err4 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"Qd", "Qh"})})
	err5 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{"As", "Ad"})})
	err6 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{})})
	err7 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{"4d"})})
	err8 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"6c"})})
	err9 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err11 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err12 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"6d", "8c", "7d"})})
	err13 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{"8h", "7h", "6h"})})
	err14 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err15 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{})})
	err16 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err17 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{"4c"})})
	err17 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err18 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"2s"})})
	err19 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err18 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"Ac", "Ah"})})
	err19 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err21 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err18 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"Tc", "Th"})})
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil ||
		err6 != nil || err7 != nil || err8 != nil || err9 != nil ||
		err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil ||
		err16 == nil || err17 != nil || err18 != nil || err19 != nil ||
		err21 != nil || err22 != nil || err23 != nil || err24 != nil || err25 != nil ||
		err26 != nil || err27 != nil || err28 != nil || err29 != nil {
		fmt.Println("errx", err1, err2, err3, err4, err5, err6, err7, err8, err9)
		fmt.Println("err1x", err11, err12, err13, err14, err15, err16, err17, err18, err19)
		fmt.Println("err2x", err21, err22, err23, err24, err25, err26, err27, err28, err29)
		t.Error()
	}
	//	fmt.Println(b.ToString())
	if b.CurrentTurnPlayer != 0 {
		t.Error()
	}

	// case2 punishment
	playersOrder = []int64{10, 11, 12}
	priorityPlayers = []int64{}
	baseMoney = 100
	b = NewTienlenBoard(playersOrder, priorityPlayers, baseMoney)
	b.StartDealing()
	b.MapPlayerToHand[10], _ = z.ToCardsFromStrings([]string{"3s", "3c", "4s", "4c", "5d", "5h", "6s", "6c", "7d", "7h", "2s", "2c", "2d"})
	b.MapPlayerToHand[11], _ = z.ToCardsFromStrings([]string{"6s", "6c", "6d", "6h", "7s", "7c", "7d", "7h", "8s", "8c", "8d", "8h", "2h"})
	b.MapPlayerToHand[12], _ = z.ToCardsFromStrings([]string{"Ts", "Tc", "Jd", "Jh", "Qs", "Qc", "Kd", "Kh"})
	b.CurrentTurnPlayer = 10
	err1 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"3c", "3s"})})
	err2 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err3 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err4 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"2d"})})
	err5 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{"2h"})})
	err6 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	// fmt.Println(b.ToString())
	if b.MapPlayerToChangedChip[10] != -300 {
		t.Error()
	}

	// case3 punishment
	playersOrder = []int64{10, 11, 12}
	priorityPlayers = []int64{}
	baseMoney = 100
	b = NewTienlenBoard(playersOrder, priorityPlayers, baseMoney)
	b.StartDealing()
	b.MapPlayerToHand[10], _ = z.ToCardsFromStrings([]string{"3s", "3c", "4s", "4c", "5d", "5h", "6s", "6c", "7d", "7h", "2s", "2c", "2d"})
	b.MapPlayerToHand[11], _ = z.ToCardsFromStrings([]string{"6s", "6c", "6d", "6h", "7s", "7c", "7d", "7h", "8s", "8c", "9d", "9h", "2h"})
	b.MapPlayerToHand[12], _ = z.ToCardsFromStrings([]string{"Ts", "Tc", "Jd", "Jh", "Qs", "Qc", "Kd", "Kh"})
	b.CurrentTurnPlayer = 10
	err1 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"3c", "3s"})})
	err2 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err3 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err4 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"2d"})})
	err5 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{"2h"})})
	err6 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{"Ts", "Tc", "Jd", "Jh", "Qs", "Qc", "Kd", "Kh"})})
	//	fmt.Println(b.ToString())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil ||
		err6 != nil {
		t.Error()
	}
	if b.MapPlayerToChangedChip[11] != -1600 {
		// 12 for remaining cards and 4 for TL_P_2H
		t.Error()
	}

	// case4 punishment
	playersOrder = []int64{10, 11, 12}
	priorityPlayers = []int64{}
	baseMoney = 100
	b = NewTienlenBoard(playersOrder, priorityPlayers, baseMoney)
	b.StartDealing()
	b.MapPlayerToHand[10], _ = z.ToCardsFromStrings([]string{"3s", "3c", "4s", "4c", "5d", "5h", "6s", "6c", "7d", "7h", "2s", "2c", "2d"})
	b.MapPlayerToHand[11], _ = z.ToCardsFromStrings([]string{"6s", "6c", "6d", "6h", "7s", "7c", "7d", "7h", "8s", "8c", "9d", "9h", "2h"})
	b.MapPlayerToHand[12], _ = z.ToCardsFromStrings([]string{"Ts", "Tc", "Jd", "Jh", "Qs", "Qc", "Kd", "Kh"})
	b.CurrentTurnPlayer = 10
	err1 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"3c", "3s"})})
	err2 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	err3 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{})})
	err4 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"2d"})})
	err5 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{"2h"})})
	err6 = b.MakeMove(Move{PlayerId: 12, Cards: z.ToCardsFromStrings2([]string{"Ts", "Tc", "Jd", "Jh", "Qs", "Qc"})})
	err7 = b.MakeMove(Move{PlayerId: 10, Cards: z.ToCardsFromStrings2([]string{"4s", "4c", "5d", "5h", "6s", "6c", "7d", "7h"})})
	err8 = b.MakeMove(Move{PlayerId: 11, Cards: z.ToCardsFromStrings2([]string{})})
	//	fmt.Println(b.ToString())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil ||
		err6 != nil || err7 != nil || err8 != nil {
		t.Error()
	}
	if b.MapPlayerToChangedChip[12] != -500 {
		t.Error()
	}
}
