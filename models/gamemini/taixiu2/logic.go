package baucua

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

const (
	SELECTION_1 = "SELECTION_1"
	SELECTION_2 = "SELECTION_2"
	SELECTION_3 = "SELECTION_3"
	SELECTION_4 = "SELECTION_4"
	SELECTION_5 = "SELECTION_5"
	SELECTION_6 = "SELECTION_6"
)

var ALL_SELECTIONS map[string]bool
var MAP_DICE_TO_SELECTION map[int]string

func init() {
	ALL_SELECTIONS = map[string]bool{
		SELECTION_1: true,
		SELECTION_2: true,
		SELECTION_3: true,
		SELECTION_4: true,
		SELECTION_5: true,
		SELECTION_6: true,
	}
	MAP_DICE_TO_SELECTION = map[int]string{
		1: SELECTION_1, 2: SELECTION_2, 3: SELECTION_3,
		4: SELECTION_4, 5: SELECTION_5, 6: SELECTION_6,
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

type Bet struct {
	playerId   int64
	betTime    time.Time
	selection  string
	moneyValue int64
}

func (bet Bet) String() string {
	return fmt.Sprintf("bet|pid%v|%v|%v|%v", bet.playerId, bet.betTime.Format(time.RFC3339Nano), bet.selection, bet.moneyValue)
}

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
func CalcBalancedSumMoneyOnSelection(
	inputSelection string, mapBetInfo map[int64]map[string]int64) int64 {
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

// cloned from taixiu,
// dont need to refund, just convert from mapPlayerIdToBets to mapBetInfo
func Balance(mapPlayerIdToBets map[int64][]*Bet) map[int64]map[string]int64 {
	mapBetInfo := make(map[int64]map[string]int64)
	for pid, bets := range mapPlayerIdToBets {
		mapBetInfo[pid] = map[string]int64{
			SELECTION_1: 0,
			SELECTION_2: 0,
			SELECTION_3: 0,
			SELECTION_4: 0,
			SELECTION_5: 0,
			SELECTION_6: 0,
		}
		for _, bet := range bets {
			mapBetInfo[pid][bet.selection] += bet.moneyValue
		}
	}

	return mapBetInfo
}

// return map[pid]emwm
func CalcEndMatchWinningMoney(
	mapBetInfo map[int64]map[string]int64, shakingResult []int) map[int64]int64 {
	result := make(map[int64]int64)

	for pid, mapSelectionToMoney := range mapBetInfo {
		playerEmwm := int64(0)
		for _, diceNumber := range shakingResult {
			playerEmwm += 2 * mapSelectionToMoney[MAP_DICE_TO_SELECTION[diceNumber]]
		}
		result[pid] = playerEmwm
	}
	return result
}
