package poker

import (
	"encoding/json"
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
	_, _ = json.Marshal([]int{})
	_ = z.NewDeck()
}

func TestHihi(t *testing.T) {
	var r int64
	r = 1 + 1
	if r != 2 {
		t.Error()
	}

}

func Test1(t *testing.T) {
	var playersOrder []int64
	var mapPidToChip map[int64]int64
	var dealerButtonPid, bigBlind, smallBlind, ante int64
	var b *PokerBoard

	// case 1
	playersOrder = []int64{10, 11}
	mapPidToChip = map[int64]int64{10: 3000, 11: 300}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	if b.SmallBlindPid != 11 {
		t.Error()
	}
	if b.BigBlindPid != 10 {
		t.Error()
	}
	if b.InRoundCurrentTurnPlayer != 11 {
		t.Error()
	}
	if b.Pots[0].Value != 0 {
		t.Error()
	}
	//	fmt.Println(b.ToString())

	// case 2
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 3000, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 10
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	//	fmt.Println(b.ToString())
	if b.SmallBlindPid != 12 {
		t.Error()
	}
	if b.BigBlindPid != 13 {
		t.Error()
	}
	if b.InRoundCurrentTurnPlayer != 10 {
		t.Error()
	}
	if b.Pots[0].Value != 40 {
		t.Error()
	}
	var err1, err2, err3, err4 error
	err1 = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	err2 = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_FOLD})
	err3 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	err4 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_CHECK})
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		t.Error()
	}
	if b.InRoundCurrentTurnPlayer != 0 {
		t.Error()
	}
	if b.Pots[0].Value != 640 {
		t.Error()
	}
	b.StartShowdown()
	fmt.Println(b.ToString())

	// case 3
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 700, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	var err5, err6, err7 error
	err1 = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	err2 = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	err3 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_RAISE, Value: 600})
	err4 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_RAISE, Value: 1000})
	err5 = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_CALL, IsAllIn: true, Value: b.MapPlayerToChip[10]})
	err6 = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_FOLD})
	err7 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_CALL,
		Value: b.InRoundMoveLimit.AmountToCall})
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil ||
		err5 != nil || err6 != nil || err7 != nil {
		t.Error(err1, err2, err3, err4, err5, err6, err7)
	}
	if b.InRoundCurrentTurnPlayer != 0 {
		t.Error()
	}
	if b.Pots[0].Value != 2300 || b.Pots[1].Value != 600 {
		t.Error()
	}
	//	fmt.Println(b.ToString())
	b.StartFlopOrTurnOrRiver(ROUND_FLOP)
	var err8, err9, err10 error
	//	err8 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_FOLD})
	//	err8 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_FOLD})
	err8 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_BET, Value: 500})
	err9 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_RAISE, Value: 1500})
	err10 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_FOLD})
	if err8 != nil || err9 != nil || err10 != nil {
		t.Error(err8, err9, err10)
	}
	//	fmt.Println(b.ToString())
	b.StartFlopOrTurnOrRiver(ROUND_TURN)
	b.StartFlopOrTurnOrRiver(ROUND_RIVER)
	b.StartShowdown()
	sumLost := int64(0)
	sumWon := int64(0)
	for pid, _ := range b.MapPlayerToLostChip {
		sumLost += b.MapPlayerToLostChip[pid]
		sumWon += b.MapPlayerToWonChip[pid]
	}
	if !(-sumLost == sumWon) {
		t.Error(sumLost, sumWon)
	}
	//	fmt.Println(b.ToString())

	// case 4
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 700, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	_ = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_FOLD})
	_ = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_FOLD})
	_ = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_FOLD})
	if b.InRoundCurrentTurnPlayer != 0 {
		t.Error()
	}
	//	fmt.Println(b.ToString())

	// case 5
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 700, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	_ = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_RAISE, IsAllIn: true, Value: b.MapPlayerToChip[10]})
	_ = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	_ = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_FOLD})
	_ = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	b.StartFlopOrTurnOrRiver(ROUND_FLOP)
	_ = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_FOLD})
	if b.InRoundCurrentTurnPlayer != 0 {
		t.Error()
	}
	//	fmt.Println(b.ToString())

	// case 6
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 700, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	_ = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_RAISE, IsAllIn: true, Value: b.MapPlayerToChip[10]})
	_ = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	_ = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_FOLD})
	_ = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_CALL, Value: b.InRoundMoveLimit.AmountToCall})
	b.StartFlopOrTurnOrRiver(ROUND_FLOP)
	_ = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_BET, IsAllIn: true, Value: b.MapPlayerToChip[13]})
	err1 = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_FOLD})
	if err1 != nil {
		t.Error()
	}
	b.StartShowdown()
	//	fmt.Println(b.ToString())

	// case 7
	playersOrder = []int64{10, 11, 12, 13}
	mapPidToChip = map[int64]int64{10: 700, 11: 5000, 12: 9000, 13: 7000}
	dealerButtonPid = 11
	bigBlind, smallBlind, ante = 200, 100, 0
	b = NewPokerBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante, smallBlind, bigBlind)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	err1 = b.MakeMove(Move{PlayerId: 10, MoveType: MOVE_RAISE, IsAllIn: true, Value: b.MapPlayerToChip[10]})
	err2 = b.MakeMove(Move{PlayerId: 11, MoveType: MOVE_FOLD})
	err3 = b.MakeMove(Move{PlayerId: 12, MoveType: MOVE_RAISE, IsAllIn: true, Value: b.MapPlayerToChip[12] + b.InRoundMapPlayerToStatus[12].ChipOnRound})
	//	err4 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_CALL, IsAllIn: true, Value: b.MapPlayerToChip[13]})
	err4 = b.MakeMove(Move{PlayerId: 13, MoveType: MOVE_FOLD})
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		t.Error(err1, err2, err3, err4)
	}
	//	b.StartShowdown()
	//	fmt.Println(b.ToString())
}
