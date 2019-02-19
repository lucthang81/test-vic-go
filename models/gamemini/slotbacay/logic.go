package slotbacay

import (
	"errors"
	"fmt"
	// "math"
	"encoding/json"
	"math/rand"
	"sort"
	"time"

	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/game/phom"
	// "github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/models/gamemini/consts"
)

const (
	LINE_TYPE_8              = "LINE_TYPE_8"
	LINE_TYPE_9              = "LINE_TYPE_9"
	LINE_TYPE_10             = "LINE_TYPE_10"
	LINE_TYPE_10AD           = "LINE_TYPE_10AD"
	LINE_TYPE_TRIPS          = "LINE_TYPE_TRIPS"
	LINE_TYPE_STRAIGHT_FLUSH = "LINE_TYPE_STRAIGHT_FLUSH"
	LINE_TYPE_TRIPS_A        = "LINE_TYPE_TRIPS_A"
	LINE_TYPE_NOTHING        = "LINE_TYPE_NOTHING"
)

// map symbol to number of this symbol in a reel
var SYMBOLS map[string]int
var PAYLINES [][]int
var MONEYS_PER_LINE []int64

var NUMBER_OF_SYMBOLS int // include weight
var SYMBOLS_ORDER []string
var MAP_SYMBOL_TO_RANGE map[string][]int
var MAP_LINE_TYPE_TO_PRIZE_RATE map[string]int64

func init() {
	// for not used import error
	_ = fmt.Printf
	_ = errors.New("")
	rand.Seed(time.Now().Unix())
	_ = cardgame.Card{}
	_, _ = json.MarshalIndent(map[string]int{}, "", "    ")
	// sylbols, map symbol to rate
	SYMBOLS = map[string]int{
		"As": 1, "Ac": 1, "Ad": 1, "Ah": 1,
		"2s": 1, "2c": 1, "2d": 1, "2h": 1,
		"3s": 1, "3c": 1, "3d": 1, "3h": 1,
		"4s": 1, "4c": 1, "4d": 1, "4h": 1,
		"5s": 1, "5c": 1, "5d": 1, "5h": 1,
		"6s": 1, "6c": 1, "6d": 1, "6h": 1,
		"7s": 1, "7c": 1, "7d": 1, "7h": 1,
		"8s": 1, "8c": 1, "8d": 1, "8h": 1,
		"9s": 1, "9c": 1, "9d": 1, "9h": 1,
	}
	SYMBOLS_ORDER = []string{}
	for k, _ := range SYMBOLS {
		SYMBOLS_ORDER = append(SYMBOLS_ORDER, k)
	}
	sort.Strings(SYMBOLS_ORDER)

	// []rowIndex for each column
	PAYLINES = [][]int{
		[]int{0, 0, 0},
	}
	MONEYS_PER_LINE = []int64{0, 100, 1000, 10000}
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
	// fmt.Println("MAP_SYMBOL_TO_RANGE", MAP_SYMBOL_TO_RANGE)
	MAP_LINE_TYPE_TO_PRIZE_RATE = map[string]int64{
		LINE_TYPE_TRIPS_A:        0, // 100
		LINE_TYPE_STRAIGHT_FLUSH: 15,
		LINE_TYPE_TRIPS:          15,
		LINE_TYPE_10AD:           7,
		LINE_TYPE_10:             3,
		LINE_TYPE_9:              2,
		LINE_TYPE_8:              2,
		LINE_TYPE_NOTHING:        0,
	}
}

const (
	MATCH_WON_TYPE_JACKPOT = "MATCH_WON_TYPE_JACKPOT"
	MATCH_WON_TYPE_BIG     = "MATCH_WON_TYPE_BIG"
	MATCH_WON_TYPE_NORMAL  = "MATCH_WON_TYPE_NORMAL"
)

type LineType struct {
	isWin    bool
	lineType string // MAP_LINE_TYPE_TO_PRIZE_RATE keys
}

type SlotbacayResult [][]string

func (slotbacayResult SlotbacayResult) String() string {
	result := ""
	for ri := 0; ri < len(slotbacayResult[0]); ri++ {
		rowString := ""
		for ci := 0; ci < len(slotbacayResult); ci++ {
			rowString += slotbacayResult[ci][ri] + " "
		}
		result += rowString + "\n"
	}
	result = result[:len(result)-1]
	return result
}

// for test
func Dummy(n int) int {
	return n
}

// return slotbacayResult
// 3 columns, each column is a []int(1)
func RandomSpin() [][]string {
	// kiểm tra lá bài đã xuất hiện chưa;
	// nên random bằng shufle, nhưng do copy từ slot nên làm như này
	resultSymbolsSet := map[string]bool{}
	result := [][]string{}
	for ci := 0; ci < 3; ci++ {
		column := []string{}
		for ri := 0; ri < 1; ri++ {
			symbol := SYMBOLS_ORDER[len(SYMBOLS_ORDER)-1] // nhỡ có gì sai thì trả về symbol phổ biến nhất
			for {
				randomInt := rand.Intn(NUMBER_OF_SYMBOLS) + 1
				for symb, symbRange := range MAP_SYMBOL_TO_RANGE {
					if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
						symbol = symb
						break
					}
				}
				if _, isIn := resultSymbolsSet[symbol]; !isIn {
					resultSymbolsSet[symbol] = true
					break
				} else {
					// pick other symbol
				}
			}
			column = append(column, symbol)
		}
		result = append(result, column)
	}
	return result
}

// get symbols on payline of slotbacayResult
func GetPayline(payline []int, slotbacayResult [][]string) []string {
	result := []string{}
	for colIndex, rowIndex := range payline {
		result = append(result, slotbacayResult[colIndex][rowIndex])
	}
	return result
}

// try RandomSpin
func tryRandomSpin() {
	typeCounter := map[string]int{}
	for t, _ := range MAP_LINE_TYPE_TO_PRIZE_RATE {
		typeCounter[t] = 0
	}
	for i := 0; i < 71400; i++ {
		rs := RandomSpin()
		typeRs := CalcLineType(GetPayline(PAYLINES[0], rs))
		// fmt.Println(SlotbacayResult(rs).String(), typeRs)
		typeCounter[typeRs.lineType] += 1
	}
	for k, v := range typeCounter {
		fmt.Println(k, v)
	}
	//LINE_TYPE_10 90643
	//LINE_TYPE_8 100273
	//LINE_TYPE_TRIPS 4806
	//LINE_TYPE_10AD 8248
	//LINE_TYPE_9 97813
	//LINE_TYPE_NOTHING 693533
	//LINE_TYPE_TRIPS_A 513
	//LINE_TYPE_STRAIGHT_FLUSH 4168
}

// just print, for setup MAP_LINE_TYPE_TO_PRIZE_RATE
func calcPrizeInfo() {
	s := float64(0)
	mapLineTypeToProbability := map[string]float64{}
	//
	mapLineTypeToCounter := map[string]int{}
	all3cardStrs := cardgame.GetCombinationsForStrings(SYMBOLS_ORDER, 3)
	n36C3 := len(all3cardStrs)
	fmt.Println("n36C3", n36C3)
	for _, comb := range all3cardStrs {
		lineType := CalcLineType(comb)
		mapLineTypeToCounter[lineType.lineType] += 1
	}
	for k, v := range mapLineTypeToCounter {
		mapLineTypeToProbability[k] = float64(v) / float64(n36C3)
	}
	//
	for k, p := range mapLineTypeToProbability {
		if k != LINE_TYPE_NOTHING {
			s += p
		}
	}
	fmt.Println("hit prob: ", s)
	mapLineTypeToPrizePart := map[string]float64{
		LINE_TYPE_TRIPS_A:        0.11,
		LINE_TYPE_STRAIGHT_FLUSH: 0.06,
		LINE_TYPE_TRIPS:          0.07,
		LINE_TYPE_10AD:           0.07,
		LINE_TYPE_10:             0.28,
		LINE_TYPE_9:              0.20,
		LINE_TYPE_8:              0.21,
	}
	for _, lineType := range []string{LINE_TYPE_TRIPS_A, LINE_TYPE_STRAIGHT_FLUSH, LINE_TYPE_TRIPS, LINE_TYPE_10AD, LINE_TYPE_10, LINE_TYPE_9, LINE_TYPE_8} {
		prob := mapLineTypeToProbability[lineType]
		fmt.Printf("%30.30s | %10.2f | %.6f | %.2f \n",
			lineType, 1/prob, prob,
			mapLineTypeToPrizePart[lineType]/prob)
	}

	sumPayToUser := float64(0)
	for lineType, rate := range MAP_LINE_TYPE_TO_PRIZE_RATE {
		sumPayToUser += mapLineTypeToProbability[lineType] * float64(rate)

	}
	fmt.Println("sumPayToUser", sumPayToUser)
}

// line = []symbols(5)
func CalcLineType(line []string) LineType {
	cards := []cardgame.Card{}
	for _, cardStr := range line {
		cards = append(cards, cardgame.FNewCardFS(cardStr))
	}
	cards = phom.SortedByRank(cards)
	c0 := cards[0]
	c1 := cards[1]
	c2 := cards[2]
	score := phom.MapRankToInt[c0.Rank] + phom.MapRankToInt[c1.Rank] + phom.MapRankToInt[c2.Rank]
	score = score % 10
	if score == 0 {
		score = 10
	}
	if c0.Rank == "A" && c1.Rank == "A" && c2.Rank == "A" {
		return LineType{isWin: true, lineType: LINE_TYPE_TRIPS_A}
	} else if cr, _ := phom.CheckIsCombo(cards); cr {
		if c0.Suit == c1.Suit {
			return LineType{isWin: true, lineType: LINE_TYPE_STRAIGHT_FLUSH}
		} else {
			return LineType{isWin: true, lineType: LINE_TYPE_TRIPS}
		}
	} else if score == 10 &&
		cardgame.FindCardInSlice(cardgame.FNewCardFS("Ad"), cards) != -1 {
		return LineType{isWin: true, lineType: LINE_TYPE_10AD}
	} else if score == 10 {
		return LineType{isWin: true, lineType: LINE_TYPE_10}
	} else if score == 9 {
		return LineType{isWin: true, lineType: LINE_TYPE_9}
	} else if score == 8 {
		return LineType{isWin: true, lineType: LINE_TYPE_8}
	} else {
		return LineType{isWin: false, lineType: LINE_TYPE_NOTHING}
	}
}

func CalcWonMoneys(slotbacayResult [][]string, paylineIndexs []int, moneyPerLine int64) (
	mapLineIndexToMoney map[int]int64, mapLineIndexToIsWin map[int]bool, matchWonType string) {
	//
	mapLineIndexToMoney = map[int]int64{}
	mapLineIndexToIsWin = map[int]bool{}

	isHitJackpot := false
	//
	if moneyPerLine == 0 {
		moneyPerLine = 3 // để tính thắng lớn cho quay thử
	}
	for _, paylineIndex := range paylineIndexs {
		payline := PAYLINES[paylineIndex]            // payline = []RowIndexOnColumn
		line := GetPayline(payline, slotbacayResult) // line = []symbols
		winType := CalcLineType(line)
		mapLineIndexToMoney[paylineIndex] = MAP_LINE_TYPE_TO_PRIZE_RATE[winType.lineType] * moneyPerLine
		mapLineIndexToIsWin[paylineIndex] = winType.isWin
		if winType.lineType == LINE_TYPE_TRIPS_A {
			isHitJackpot = true
		}
	}
	sumPay := CalcSumPay(mapLineIndexToMoney)
	if isHitJackpot {
		matchWonType = MATCH_WON_TYPE_JACKPOT
	} else if sumPay >= MAP_LINE_TYPE_TO_PRIZE_RATE[LINE_TYPE_TRIPS]*moneyPerLine &&
		sumPay >= consts.BIG_WIN_ABS_LOWWER_BOUND {
		matchWonType = MATCH_WON_TYPE_BIG
	} else {
		matchWonType = MATCH_WON_TYPE_NORMAL
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
