package taixiu

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

const (
	SELECTION_XIU = "SELECTION_XIU" // nhỏ 3-10
	SELECTION_TAI = "SELECTION_TAI" // lớn 11-18
)

// map {SELECTION_TAI: true, SELECTION_XIU: true}
var ALL_SELECTIONS map[string]bool

func init() {
	ALL_SELECTIONS = map[string]bool{
		SELECTION_TAI: true,
		SELECTION_XIU: true,
	}

	fmt.Print("")
	rand.Seed(time.Now().Unix())
}

func RandomShake() []int {
	result := []int{1, 1, 1}
	result[0] = 1 + rand.Intn(6)
	result[1] = 1 + rand.Intn(6)
	result[2] = 1 + rand.Intn(6)
	return result
}

// return n: n in [a, b) && n >= 1
func RandRange16(a int, b int) int {
	var r int
	if b-a <= 0 {
		r = a
	} else {
		r = a + rand.Intn(b-a)
	}
	if r <= 0 {
		return 1
	} else if r >= 7 {
		return 6
	} else {
		return r
	}
}

//
func CheatShake(selection string) []int {
	result := []int{1, 1, 1}
	if selection == SELECTION_TAI {
		result[0] = RandRange16(1, 7)
		result[1] = RandRange16(11-6-result[0], 7)
		result[2] = RandRange16(11-result[1]-result[0], 7)
	} else {
		result[0] = RandRange16(1, 7)
		result[1] = RandRange16(1, 10-1-result[0]+1)
		result[2] = RandRange16(1, 10-result[1]-result[0]+1)
	}
	return result
}

// return SELECTION_XIU or SELECTION_TAI
func GetTypeOutcome(outcome []int) string {
	if outcome[0]+outcome[1]+outcome[2] <= 10 {
		return SELECTION_XIU
	} else {
		return SELECTION_TAI
	}
}

type Bet struct {
	playerId   int64
	betTime    time.Time
	selection  string
	moneyValue int64
}

func (bet Bet) String() string {
	return fmt.Sprintf("bet|pid%v|%v|%v|%v", bet.playerId, bet.betTime.Format(time.RFC3339Nano), bet.selection, bet.moneyValue)
}

// use before balance
func CalcSumMoneyOnSelection(selection string, mapPlayerIdToBets map[int64][]*Bet) int64 {
	result := int64(0)
	for _, bets := range mapPlayerIdToBets {
		for _, bet := range bets {
			if bet.selection == selection {
				result += bet.moneyValue
			}
		}
	}
	return result
}

// use before balance
func CalcSumMoneyOnSelectionForPlayer(selection string, mapPlayerIdToBets map[int64][]*Bet, playerId int64) int64 {
	result := int64(0)
	bets := mapPlayerIdToBets[playerId]
	for _, bet := range bets {
		if bet.selection == selection {
			result += bet.moneyValue
		}
	}
	return result
}

// use before balance
func CalcNOPOnSelection(selection string, mapPlayerIdToBets map[int64][]*Bet) int {
	result := 0
	for _, bets := range mapPlayerIdToBets {
		for _, bet := range bets {
			if bet.selection == selection {
				result += 1
				break
			}
		}
	}
	return result
}

// use after balance
func CalcSumMoneyForPlayer(pid int64, mapBetInfo map[int64]map[string]int64) int64 {
	result := int64(0)
	for _, amount := range mapBetInfo[pid] {
		result += amount
	}
	return result
}

// use after balance
func CalcBalancedSumMoneyOnSelection(inputSelection string, mapBetInfo map[int64]map[string]int64) int64 {
	result := int64(0)
	for _, mapSelectionToAmount := range mapBetInfo {
		for selection, amount := range mapSelectionToAmount {
			if selection == inputSelection {
				result += amount
			}
		}
	}
	return result
}

// sort bets by betTime, late bet is first element
type BetsSortByTime []*Bet

func (a BetsSortByTime) Len() int { return len(a) }
func (a BetsSortByTime) Swap(i int, j int) {
	temp := a[i]
	a[i] = a[j]
	a[j] = temp
}
func (a BetsSortByTime) Less(i int, j int) bool {
	if a[i].betTime.Sub(a[j].betTime) > 0 {
		return true
	} else {
		return false
	}
}

func SortedByTime(bets []*Bet) []*Bet {
	result := make([]*Bet, len(bets))
	copy(result, bets)
	sort.Sort(BetsSortByTime(result))
	return result
}

// cân đối và trả tiền cược không được cân
// trả về mapBetInfo, map[pId]tiềnThừa, refundSelection
func Balance(mapPlayerIdToBets map[int64][]*Bet) (map[int64]map[string]int64, map[int64]int64, string) {
	mapBetInfo := make(map[int64]map[string]int64)
	mapPlayerIdToRefund := make(map[int64]int64)

	for pid, bets := range mapPlayerIdToBets {
		mapBetInfo[pid] = map[string]int64{
			SELECTION_TAI: 0,
			SELECTION_XIU: 0,
		}
		for _, bet := range bets {
			mapBetInfo[pid][bet.selection] += bet.moneyValue
		}
	}

	moneyOnTai := CalcSumMoneyOnSelection(SELECTION_TAI, mapPlayerIdToBets)
	moneyOnXiu := CalcSumMoneyOnSelection(SELECTION_XIU, mapPlayerIdToBets)
	var refundSelection string
	var refundAmount int64
	if moneyOnTai >= moneyOnXiu {
		refundSelection = SELECTION_TAI
		refundAmount = moneyOnTai - moneyOnXiu
	} else {
		refundSelection = SELECTION_XIU
		refundAmount = moneyOnXiu - moneyOnTai
	}
	betsOnRefundSelection := []*Bet{}
	for _, bets := range mapPlayerIdToBets {
		for _, bet := range bets {
			if bet.selection == refundSelection {
				betsOnRefundSelection = append(betsOnRefundSelection, bet)
			}
		}
	}
	sortedBetsOnRS := SortedByTime(betsOnRefundSelection)
	for _, bet := range sortedBetsOnRS {
		if refundAmount == 0 {
			break
		} else {
			if _, isIn := mapPlayerIdToRefund[bet.playerId]; !isIn {
				mapPlayerIdToRefund[bet.playerId] = 0
			}
			if refundAmount > bet.moneyValue {
				mapPlayerIdToRefund[bet.playerId] += bet.moneyValue
				mapBetInfo[bet.playerId][refundSelection] -= bet.moneyValue
				refundAmount -= bet.moneyValue
			} else {
				mapPlayerIdToRefund[bet.playerId] += refundAmount
				mapBetInfo[bet.playerId][refundSelection] -= refundAmount
				refundAmount = 0
			}
		}
	}

	return mapBetInfo, mapPlayerIdToRefund, refundSelection
}
