package lieng

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
	var r1, r2, r3, r4, r5, r6, r7, r8 []int
	r1 = CalcLiengRank(z.ToCardsFromStrings2([]string{"As", "Ad", "Ac"}))
	r2 = CalcLiengRank(z.ToCardsFromStrings2([]string{"As", "2d", "3c"}))
	r3 = CalcLiengRank(z.ToCardsFromStrings2([]string{"As", "Qd", "Kc"}))
	r4 = CalcLiengRank(z.ToCardsFromStrings2([]string{"9s", "8d", "7c"}))
	r5 = CalcLiengRank(z.ToCardsFromStrings2([]string{"As", "6d", "2c"}))
	r6 = CalcLiengRank(z.ToCardsFromStrings2([]string{"2s", "5d", "2c"}))
	r7 = CalcLiengRank(z.ToCardsFromStrings2([]string{"As", "6d", "3c"}))
	r8 = CalcLiengRank(z.ToCardsFromStrings2([]string{"9s", "Jd", "Kc"}))
	//	fmt.Println(r1, r2, r3, r4, r5, r6, r7, r8)
	if r1[0] != PokerType[LIENG_TYPE_TRIPS] {
		t.Error()
	}
	if r2[0] != PokerType[LIENG_TYPE_STRAIGHT] {
		t.Error()
	}
	if r3[0] != PokerType[LIENG_TYPE_STRAIGHT] {
		t.Error()
	}
	if r4[0] != PokerType[LIENG_TYPE_STRAIGHT] {
		t.Error()
	}
	if z.Compare2ListInt(r3, r4) != true {
		t.Error()
	}
	if z.Compare2ListInt(r5, r6) != true {
		t.Error()
	}
	if z.Compare2ListInt(r6, r7) != true {
		t.Error()
	}
	if z.Compare2ListInt(r5, r8) != true || z.Compare2ListInt(r8, r6) != true {
		t.Error()
	}
}

func Test2(t *testing.T) {
	var playersOrder []int64
	var mapPidToChip map[int64]int64
	var dealerButtonPid, ante int64
	var b *PokerBoard

	// case 1
	playersOrder = []int64{10, 11}
	mapPidToChip = map[int64]int64{10: 500, 11: 2000}
	dealerButtonPid = 11
	ante = 100
	b = NewLiengBoard(
		playersOrder, mapPidToChip, dealerButtonPid, ante)
	b.StartDealingHoleCards()
	b.StartPreFlop()
	if b.InRoundCurrentTurnPlayer != 10 {
		t.Error()
	}
	if b.Pots[0].Value != 200 {
		t.Error()
	}
	//	fmt.Println(b.ToString())
}
