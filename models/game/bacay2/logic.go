package bacay2

import (
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/models/components"
)

const (
	COMPARE_RESULT_GREATER = "greater"
	COMPARE_RESULT_EQUAL   = "equal"
	COMPARE_RESULT_LESS    = "less"
)

// với bacay chỉ chia bài từ A đến 9
var valueStrInt64Map = map[string]int{
	"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9,
	"10": 10, "j": 10, "q": 10, "k": 10, "a": 1}
var rankOrder = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a"}
var suitsOrder = []string{"s", "c", "h", "d"}

func ConvertValueToInt(value string) int {
	v, isExisted := valueStrInt64Map[value]
	if isExisted {
		return v
	} else {
		fmt.Printf("%v WARNING: %v is not a card value.\n", time.Now().Format(time.RFC3339), value)
		return 0
	}
}

// Return the lowest index of arg0 in where arg1 is found,
// return -1 on failure
func Find(es []string, ipE string) int {
	for i, e := range es {
		if e == ipE {
			return i
		}
	}
	return -1
}

// Return the lowest index of arg0 in where arg1 is found,
// return -1 on failure
func Find2(es []int, ipE int) int {
	for i, e := range es {
		if e == ipE {
			return i
		}
	}
	return -1
}

// Return the lowest index of arg0 in where arg1 is found,
// return -1 on failure
func Find3(es []int64, ipE int64) int {
	for i, e := range es {
		if e == ipE {
			return i
		}
	}
	return -1
}

// return COMPARE_RESULT_GREATER / COMPARE_RESULT_EQUAL / COMPARE_RESULT_LESS
func CompareTwoCard(card1 string, card2 string) string {
	s1, r1 := components.SuitAndValueFromCard(card1)
	s2, r2 := components.SuitAndValueFromCard(card2)
	if Find(suitsOrder, s1) > Find(suitsOrder, s2) {
		return COMPARE_RESULT_GREATER
	} else if Find(suitsOrder, s1) < Find(suitsOrder, s2) {
		return COMPARE_RESULT_LESS
	} else {
		if Find(rankOrder, r1) > Find(rankOrder, r2) {
			return COMPARE_RESULT_GREATER
		} else if Find(rankOrder, r1) < Find(rankOrder, r2) {
			return COMPARE_RESULT_LESS
		} else {
			return COMPARE_RESULT_EQUAL
		}
	}
}

func GetMaxCard(cards []string) (result string) {
	result = cards[0]
	for _, card := range cards {
		if CompareTwoCard(card, result) == COMPARE_RESULT_GREATER {
			result = card
		}
	}
	return result
}

func CalcScore(ThreeCards []string) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = errors.New("Wrong input")
		}
	}()
	result = 0
	for _, card := range ThreeCards {
		_, val := components.SuitAndValueFromCard(card)
		result += ConvertValueToInt(val)
	}
	result = result % 10
	if result == 0 {
		result = 10
	}
	return result, nil
}

func CompareTwoBacayHand(hand1 []string, hand2 []string) (result string, err error) {
	score1, err1 := CalcScore(hand1)
	score2, err2 := CalcScore(hand2)
	if (err1 != nil) || (err2 != nil) {
		return "", errors.New("Wrong input")
	}
	if score1 > score2 {
		return COMPARE_RESULT_GREATER, nil
	} else if score1 < score2 {
		return COMPARE_RESULT_LESS, nil
	} else {
		maxCard1 := GetMaxCard(hand1)
		maxCard2 := GetMaxCard(hand2)
		if CompareTwoCard(maxCard1, maxCard2) == COMPARE_RESULT_GREATER {
			return COMPARE_RESULT_GREATER, nil
		} else if CompareTwoCard(maxCard1, maxCard2) == COMPARE_RESULT_LESS {
			return COMPARE_RESULT_LESS, nil
		} else {
			return COMPARE_RESULT_EQUAL, nil
		}

	}
}

func GetMaxHandIndex(hands [][]string) (int, error) {
	maxIndex := 0
	maxHand := hands[0]

	for index, hand := range hands {
		compareResult, err := CompareTwoBacayHand(hand, maxHand)
		if err != nil {
			return -1, err
		}
		if compareResult == COMPARE_RESULT_GREATER {
			maxIndex = index
			maxHand = hand
		}
	}

	return maxIndex, nil
}
