package xocdia2

import (
	//"errors"
	"fmt"
	"math/rand"
)

const (
	SELECTION_EVEN  = "SELECTION_EVEN"
	SELECTION_ODD   = "SELECTION_ODD"
	SELECTION_0_RED = "SELECTION_0_RED"
	SELECTION_1_RED = "SELECTION_1_RED"
	SELECTION_3_RED = "SELECTION_3_RED"
	SELECTION_4_RED = "SELECTION_4_RED"

	OUTCOME_0_RED = "OUTCOME_0_RED"
	OUTCOME_1_RED = "OUTCOME_1_RED"
	OUTCOME_2_RED = "OUTCOME_2_RED"
	OUTCOME_3_RED = "OUTCOME_3_RED"
	OUTCOME_4_RED = "OUTCOME_4_RED"
)

var mapSelectionToRate map[string]float64      // map cửa đến số tiền nhận nếu thắng
var mapOutcomeToSelections map[string][]string // map outcome đến các cửa thắng
var mapSelectionToOutcomes map[string][]string // map cửa đến các outcome tốt
var allSelections []string

func init() {
	fmt.Print("")
	mapOutcomeToSelections = map[string][]string{
		OUTCOME_0_RED: []string{SELECTION_EVEN, SELECTION_0_RED},
		OUTCOME_1_RED: []string{SELECTION_ODD, SELECTION_1_RED},
		OUTCOME_2_RED: []string{SELECTION_EVEN},
		OUTCOME_3_RED: []string{SELECTION_ODD, SELECTION_3_RED},
		OUTCOME_4_RED: []string{SELECTION_EVEN, SELECTION_4_RED},
	}
	mapSelectionToOutcomes = map[string][]string{
		SELECTION_EVEN:  []string{OUTCOME_0_RED, OUTCOME_2_RED, OUTCOME_4_RED},
		SELECTION_ODD:   []string{OUTCOME_1_RED, OUTCOME_3_RED},
		SELECTION_0_RED: []string{OUTCOME_0_RED},
		SELECTION_1_RED: []string{OUTCOME_1_RED},
		SELECTION_3_RED: []string{OUTCOME_3_RED},
		SELECTION_4_RED: []string{OUTCOME_4_RED},
	}
	allSelections = []string{SELECTION_EVEN, SELECTION_ODD, SELECTION_0_RED,
		SELECTION_1_RED, SELECTION_3_RED, SELECTION_4_RED,
	}
}

// sum money 1 player bet on all selections
func GetSumBet(playerBetInfo map[string]int64) int64 {
	result := int64(0)
	for selection, _ := range mapSelectionToOutcomes {
		result += playerBetInfo[selection]
	}
	return result
}

// copy for 1 player
func CopyBet(playerId int64, lastBetInfo map[int64]map[string]int64) map[string]int64 {
	result := make(map[string]int64)
	for selection, _ := range mapSelectionToOutcomes {
		result[selection] = lastBetInfo[playerId][selection]
	}
	return result
}

func Copy2xBet(playerId int64, lastBetInfo map[int64]map[string]int64) map[string]int64 {
	result := make(map[string]int64)
	for selection, _ := range mapSelectionToOutcomes {
		result[selection] = int64(2) * lastBetInfo[playerId][selection]
	}
	return result
}

// copy all map bet info
func CopyAllBet(betInfo map[int64]map[string]int64) map[int64]map[string]int64 {
	result := make(map[int64]map[string]int64)
	for playerId, _ := range betInfo {
		result[playerId] = CopyBet(playerId, betInfo)
	}
	return result
}

func RandomShake() string {
	r := rand.Float64()
	if (0 <= r) && (r < 1/16.0) {
		return OUTCOME_0_RED
	} else if (1/16.0 <= r) && (r < 5/16.0) {
		return OUTCOME_1_RED
	} else if (5/16.0 <= r) && (r < 11/16.0) {
		return OUTCOME_2_RED
	} else if (11/16.0) <= r && (r < 15/16.0) {
		return OUTCOME_3_RED
	} else {
		return OUTCOME_4_RED
	}
}

// in OUTCOME_0_RED, out SELECTION_EVEN
func GetTypeOutcome(shakingResult string) string {
	if shakingResult == OUTCOME_0_RED ||
		shakingResult == OUTCOME_2_RED ||
		shakingResult == OUTCOME_4_RED {
		return SELECTION_EVEN
	} else {
		return SELECTION_ODD
	}
}

func CheatShake(typeOutcome string) string {
	if typeOutcome == SELECTION_EVEN {
		r := rand.Intn(8)
		if r < 6 {
			return OUTCOME_2_RED
		} else if r < 7 {
			return OUTCOME_0_RED
		} else {
			return OUTCOME_4_RED
		}
	} else {
		r := rand.Intn(2)
		if r < 1 {
			return OUTCOME_1_RED
		} else {
			return OUTCOME_3_RED
		}
	}
}

func GetAllBetOn1Selection(betInfo map[int64]map[string]int64, selection string) int64 {
	_, isIn := mapSelectionToOutcomes[selection]
	if !isIn {
		return int64(0)
	} else {
		result := int64(0)
		for playerId, _ := range betInfo {
			result += betInfo[playerId][selection]
		}
		return result
	}
}

func CheckIsIn(sub string, list []string) bool {
	for _, e := range list {
		if sub == e {
			return true
		}
	}
	return false
}

// return money gain for each player in session,
// map[playerId]winMoney,
// betInfo is map after balanced,

func CalcResult(
	hostPlayerId int64, // 0 tức là không có ai làm host
	hostAcceptEvenOrOdd string, // SELECTION_ or ""
	hostMoneyOnEvenOrOdd int64,
	betInfo map[int64]map[string]int64, // map[playerId](map[selection]soTienCuoc), not include host
	acceptedMap map[string]int64, // id của người cân 4 cửa ăn to, map[selection]playerId
	outcome string,
) map[int64]int64 {
	result := make(map[int64]int64)
	// trả tiền thắng cho cửa chẵn và cửa lẻ
	var evenOrOddWin string
	if CheckIsIn(outcome, mapSelectionToOutcomes[SELECTION_EVEN]) {
		evenOrOddWin = SELECTION_EVEN
	} else {
		evenOrOddWin = SELECTION_ODD
	}
	for playerId, _ := range betInfo {
		result[playerId] += int64(2) * betInfo[playerId][evenOrOddWin]
	}
	if (hostPlayerId != 0) && (hostAcceptEvenOrOdd != evenOrOddWin) {
		result[hostPlayerId] = int64(2) * hostMoneyOnEvenOrOdd
	}
	// trả tiền cho các cửa còn lại
	for _, selection := range []string{SELECTION_0_RED, SELECTION_4_RED} {
		if acceptedMap[selection] != 0 { // có người cân cửa
			if CheckIsIn(outcome, mapSelectionToOutcomes[selection]) {
				for playerId, _ := range betInfo {
					result[playerId] += int64(16) * betInfo[playerId][selection]
				}
			} else {
				result[acceptedMap[selection]] += int64(16) * GetAllBetOn1Selection(betInfo, selection)
			}
		}
	}
	for _, selection := range []string{SELECTION_1_RED, SELECTION_3_RED} {
		if acceptedMap[selection] != 0 { // có người cân cửa
			if CheckIsIn(outcome, mapSelectionToOutcomes[selection]) {
				for playerId, _ := range betInfo {
					result[playerId] += int64(4) * betInfo[playerId][selection]
				}
			} else {
				result[acceptedMap[selection]] += int64(4) * GetAllBetOn1Selection(betInfo, selection)
			}
		}
	}

	return result
}
