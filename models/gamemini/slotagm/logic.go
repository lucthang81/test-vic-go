package slotagoldminer

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/gamemini/consts"
)

// map symbol to number of this symbol in a reel
var SYMBOLS map[string]int
var PAYLINES [][]int
var MONEYS_PER_LINE []int64

var NUMBER_OF_SYMBOLS int // include weight
var SYMBOLS_ORDER []string
var MAP_SYMBOL_TO_RANGE map[string][]int
var MAP_SYLBOL_NDUP_TO_PRIZE_RATE map[string]map[int]float64
var MAP_SYLBOL_NDUP_TO_FREE_SPIN map[string]map[int]bool

func init() {
	// for not used import error
	_ = errors.New("")
	rand.Seed(time.Now().Unix())
	// sylbols, map symbol to rate
	SYMBOLS = map[string]int{
		"1": 15,
		"2": 20,
		"3": 25,
		"4": 28,
		"5": 32,
		"6": 60,
		"7": 70,
	}
	SYMBOLS_ORDER = []string{"1", "2", "3", "4", "5", "6", "7"}
	// []rowIndex for each column
	PAYLINES = [][]int{
		[]int{1, 1, 1, 1, 1},
		[]int{0, 0, 0, 0, 0},
		[]int{2, 2, 2, 2, 2},
		[]int{1, 1, 0, 1, 1},
		[]int{1, 1, 2, 1, 1},
		// 5
		[]int{0, 0, 1, 0, 0},
		[]int{2, 2, 1, 2, 2},
		[]int{0, 2, 0, 2, 0},
		[]int{2, 0, 2, 0, 2},
		[]int{1, 0, 2, 0, 1},
		// 10
		[]int{2, 1, 0, 1, 2},
		[]int{0, 1, 2, 1, 0},
		[]int{1, 2, 1, 0, 1},
		[]int{1, 0, 1, 2, 1},
		[]int{2, 1, 1, 1, 2},
		// 15
		[]int{0, 1, 1, 1, 0},
		[]int{1, 2, 2, 2, 1},
		[]int{1, 0, 0, 0, 1},
		[]int{2, 2, 1, 0, 0},
		[]int{0, 0, 1, 2, 2},
		// 20
	}
	MONEYS_PER_LINE = []int64{
		//		0,
		1, 50, 100,
		250, 500, 1000,
		2500, 5000, 10000,
	}
	NUMBER_OF_SYMBOLS = 0
	for _, counter := range SYMBOLS {
		NUMBER_OF_SYMBOLS += counter
	}
	MAP_SYMBOL_TO_RANGE = map[string][]int{}
	rangeLowerB := 0
	for _, symb := range SYMBOLS_ORDER {
		noSymb := SYMBOLS[symb]
		symbRange := []int{rangeLowerB + 1, rangeLowerB + noSymb}
		rangeLowerB = rangeLowerB + noSymb
		MAP_SYMBOL_TO_RANGE[symb] = symbRange
	}
	MAP_SYLBOL_NDUP_TO_PRIZE_RATE = map[string]map[int]float64{
		"1": map[int]float64{
			5: 0, // 10000
			4: 400,
			3: 6,
		},
		"2": map[int]float64{
			5: 900,
			4: 150,
			3: 3,
		},
		"3": map[int]float64{
			5: 700,
			4: 30, // MATCH_WON_TYPE_AG, ave * 3.4
			3: 2.5,
		},
		"4": map[int]float64{
			5: 600,
			4: 72, // MATCH_WON_TYPE_AG, ave * 3.4
			3: 1.5,
		},
		"5": map[int]float64{
			5: 115, // MATCH_WON_TYPE_AG, ave * 3.4
			4: 58,  // MATCH_WON_TYPE_AG, ave * 3.4
			3: 1,
		},
		"6": map[int]float64{
			5: 29, // MATCH_WON_TYPE_AG, ave * 3.4
			4: 1.5,
			3: 0.5,
		},
		"7": map[int]float64{
			5: 55,
			4: 1,
			3: 0.5,
		},
	}
	// dont use anymore
	MAP_SYLBOL_NDUP_TO_FREE_SPIN = map[string]map[int]bool{
		"6": map[int]bool{
			4: true,
		},
		"7": map[int]bool{
			5: true,
		},
	}
}

type LineType struct {
	isWin  bool
	sylbol string
	nDup   int
}

type SlotResult [][]string

func (slotResult SlotResult) String() string {
	result := ""
	for ri := 0; ri < len(slotResult[0]); ri++ {
		rowString := ""
		for ci := 0; ci < len(slotResult); ci++ {
			rowString += slotResult[ci][ri] + " "
		}
		result += rowString + "\n"
	}
	return result
}

// for test
func Dummy(n int) int {
	return n
}

// x = probability have exact nDup symbol in a line
// return 1/x
// slot machine: 1 line have 5 symbols
func CalcProbabilityNDupInALine(symbol string, nDup int) float64 {
	p := float64(SYMBOLS[symbol]) / float64(NUMBER_OF_SYMBOLS)
	return CalcProbabilityNDupInALineVer2(p, nDup)
}

// p = SYMBOLS[symbol] / NUMBER_OF_SYMBOLS
// x = probability have exact nDup symbol in a line
// slot machine: 1 line have 5 symbols
func CalcProbabilityNDupInALineVer2(p float64, nDup int) float64 {
	pDup := (float64(cardgame.CalcNOCombs(5, nDup)) *
		math.Pow(p, float64(nDup)) *
		math.Pow(1-p, float64(5-nDup)))
	return pDup
}

// return slotResult
// 5 columns, each column is a []int(3)
func RandomSpin() [][]string {
	result := [][]string{}
	for ci := 0; ci < 5; ci++ {
		column := []string{}
		for ri := 0; ri < 3; ri++ {
			symbol := SYMBOLS_ORDER[len(SYMBOLS_ORDER)-1] // nhỡ có gì sai thì trả về symbol phổ biến nhất
			randomInt := rand.Intn(NUMBER_OF_SYMBOLS) + 1
			for symb, symbRange := range MAP_SYMBOL_TO_RANGE {
				if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
					symbol = symb
					break
				}
			}
			column = append(column, symbol)
		}
		result = append(result, column)
	}
	return result
}

// get symbols on payline of slotResult
func GetPayline(payline []int, slotResult [][]string) []string {
	result := []string{}
	for colIndex, rowIndex := range payline {
		result = append(result, slotResult[colIndex][rowIndex])
	}
	return result
}

// just print, for setup MAP_SYLBOL_NDUP_TO_PRIZE_RATE
func CalcPrizeInfo() {
	s := float64(0)
	for _, nDup := range []int{5, 4, 3} {
		for _, symbol := range SYMBOLS_ORDER {
			prob := CalcProbabilityNDupInALine(symbol, nDup)
			s += prob
		}
	}
	fmt.Println("NUMBER_OF_SYMBOLS: ", NUMBER_OF_SYMBOLS)
	fmt.Println("hit prob: ", s)

	mapSymbolToPrizePart := map[string]float64{
		"1": 0.04,
		"2": 0.05,
		"3": 0.07,
		"4": 0.20,
		"5": 0.26,
		"6": 0.14,
		"7": 0.17,
	}
	mapNDupToPrizePart := map[int]float64{
		3: 0.18,
		4: 0.55,
		5: 0.20,
	}
	for _, symbol := range SYMBOLS_ORDER {
		for _, nDup := range []int{5, 4, 3} {
			prob := CalcProbabilityNDupInALine(symbol, nDup)
			rate := MAP_SYLBOL_NDUP_TO_PRIZE_RATE[symbol][nDup]
			fmt.Printf("%v | nDup %v | %10.2f | %6.5f | %8.2f | %7.1f | %6.5f \n",
				symbol, nDup, 1/prob, prob,
				mapSymbolToPrizePart[symbol]*mapNDupToPrizePart[nDup]/prob,
				rate, prob*rate)
		}
	}

	sumPayToUser := float64(0)
	for symbol, mapNdupToRate := range MAP_SYLBOL_NDUP_TO_PRIZE_RATE {
		for nDup, rate := range mapNdupToRate {
			if true {
				sumPayToUser += CalcProbabilityNDupInALine(symbol, nDup) * float64(rate)
			}
		}
	}
	fmt.Println("sumPayToUser", sumPayToUser)
}

// line = []symbols(5)
func CalcLineType(line []string) LineType {
	for symbol, _ := range SYMBOLS {
		noSymb := 0
		for i, _ := range line {
			if line[i] == symbol {
				noSymb += 1
			}
		}
		if noSymb >= 3 {
			return LineType{
				isWin:  true,
				nDup:   noSymb,
				sylbol: symbol,
			}
		}
	}
	return LineType{isWin: false}
}

//
func CalcWonMoneys(slotResult [][]string, paylineIndexs []int, moneyPerLine int64) (
	mapLineIndexToMoney map[int]int64, mapLineIndexToIsWin map[int]bool, matchWonType string) {
	//
	mapLineIndexToMoney = map[int]int64{}
	mapLineIndexToIsWin = map[int]bool{}
	//
	if moneyPerLine == 0 {
		moneyPerLine = 3 // để tính thắng lớn cho quay thử
	}
	//
	isHitJackpot := false
	isAg := false
	for _, paylineIndex := range paylineIndexs {
		payline := PAYLINES[paylineIndex]       // payline = []RowIndexOnColumn
		line := GetPayline(payline, slotResult) // line = []symbols
		winType := CalcLineType(line)
		if winType.isWin == false {
			mapLineIndexToMoney[paylineIndex] = 0
			mapLineIndexToIsWin[paylineIndex] = false
		} else {
			mapLineIndexToMoney[paylineIndex] =
				int64(MAP_SYLBOL_NDUP_TO_PRIZE_RATE[winType.sylbol][winType.nDup] *
					float64(moneyPerLine))
			mapLineIndexToIsWin[paylineIndex] = true
			if winType.sylbol == SYMBOLS_ORDER[0] && winType.nDup == 5 {
				isHitJackpot = true
			}
			if (winType.sylbol == "3" && winType.nDup == 4) ||
				(winType.sylbol == "4" && winType.nDup == 4) ||
				(winType.sylbol == "5" && winType.nDup == 5) ||
				(winType.sylbol == "5" && winType.nDup == 4) ||
				(winType.sylbol == "6" && winType.nDup == 5) {
				isAg = true
			}
		}
	}
	sumPay := CalcSumPay(mapLineIndexToMoney)
	if isHitJackpot {
		matchWonType = consts.MATCH_WON_TYPE_JACKPOT
	} else if isAg {
		matchWonType = consts.MATCH_WON_TYPE_AG
	} else if sumPay >= 100*moneyPerLine &&
		sumPay >= consts.BIG_WIN_ABS_LOWWER_BOUND {
		matchWonType = consts.MATCH_WON_TYPE_BIG
	} else {
		matchWonType = consts.MATCH_WON_TYPE_NORMAL
	}
	//
	if moneyPerLine == 3 {
		for li, _ := range mapLineIndexToMoney {
			mapLineIndexToMoney[li] = 0
		}
	}
	//	fmt.Println("paylineIndexs", paylineIndexs)
	//	fmt.Println("moneyPerLine", moneyPerLine)
	//	fmt.Println("CalcMoneyToPayresult", mapLineIndexToMoney)
	return
}

//
func CalcSumPay(mapPaylineIndexToWonMoney map[int]int64) int64 {
	r := int64(0)
	for _, wonMoney := range mapPaylineIndexToWonMoney {
		r += wonMoney
	}
	return r
}
