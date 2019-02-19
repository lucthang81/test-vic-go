package slotxxx

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
	"github.com/vic/vic_go/models/gamemini/consts"
	// "github.com/vic/vic_go/utils"
)

const (
	LINE_TYPE_TRIPS_A_FLUSH = "LINE_TYPE_TRIPS_A_FLUSH"
	LINE_TYPE_TRIPS_X_FLUSH = "LINE_TYPE_TRIPS_X_FLUSH"
	LINE_TYPE_TRIPS         = "LINE_TYPE_TRIPS"
	LINE_TYPE_FLUSH         = "LINE_TYPE_FLUSH"
	LINE_TYPE_PAIR          = "LINE_TYPE_PAIR"
	LINE_TYPE_9_OR_10       = "LINE_TYPE_9_OR_10"
	LINE_TYPE_NOTHING       = "LINE_TYPE_NOTHING"
)

// map symbol to number of this symbol in a reel
var SYMBOLS map[string]int
var PAYLINES [][]int
var MONEYS_PER_LINE []int64

var NUMBER_OF_SYMBOLS int // include weight
var SYMBOLS_ORDER []string
var MAP_SYMBOL_TO_RANGE map[string][]int
var MAP_LINE_TYPE_TO_PRIZE_RATE map[string]float64

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
		"Ts": 1, "Tc": 1, "Td": 1, "Th": 1,
		"Js": 1, "Jc": 1, "Jd": 1, "Jh": 1,
		"Qs": 1, "Qc": 1, "Qd": 1, "Qh": 1,
		"Ks": 1, "Kc": 1, "Kd": 1, "Kh": 1,
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
	MAP_LINE_TYPE_TO_PRIZE_RATE = map[string]float64{
		LINE_TYPE_TRIPS_A_FLUSH: 0, // jackpot
		LINE_TYPE_TRIPS_X_FLUSH: 30,
		LINE_TYPE_TRIPS:         25,
		LINE_TYPE_FLUSH:         3,
		LINE_TYPE_PAIR:          1.5,
		LINE_TYPE_9_OR_10:       1,
		LINE_TYPE_NOTHING:       0,
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

type SlotxxxResult [][]string

func (slotxxxResult SlotxxxResult) String() string {
	result := ""
	for ri := 0; ri < len(slotxxxResult[0]); ri++ {
		rowString := ""
		for ci := 0; ci < len(slotxxxResult); ci++ {
			rowString += slotxxxResult[ci][ri] + " "
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

// return slotxxxResult
// 3 columns, each column is a []string(1)
func RandomSpin() [][]string {
	result := [][]string{}
	for ci := 0; ci < 3; ci++ {
		column := []string{}
		for ri := 0; ri < 1; ri++ {
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

// đã bao gồm hư cấu,
// i là level x2
func Random1Card(userChoice string, i int) string {
	smallCards := []string{
		"2c", "2d", "2h", "2s", "3c", "3d", "3h", "3s",
		"4c", "4d", "4h", "4s", "5c", "5d", "5h", "5s",
		"6c", "6d", "6h", "6s", "Ac", "Ad", "Ah", "As"}
	bigCards := []string{
		"8c", "8d", "8h", "8s", "9c", "9d", "9h", "9s",
		"Jc", "Jd", "Jh", "Js", "Kc", "Kd", "Kh", "Ks",
		"Qc", "Qd", "Qh", "Qs", "Tc", "Td", "Th", "Ts"}
	var goodCards, badCards []string
	if userChoice == ACTION_SELECT_SMALL {
		goodCards = smallCards
		badCards = bigCards
	} else {
		goodCards = bigCards
		badCards = smallCards
	}
	var result string
	r := rand.Intn(13)
	if r < 6 {
		result = goodCards[rand.Intn(len(goodCards))]
	} else {
		result = badCards[rand.Intn(len(badCards))]
	}

	if (i >= MAX_XXX_LEVEL-1) && (rand.Intn(100) < 70) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-2) && (rand.Intn(100) < 60) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-3) && (rand.Intn(100) < 50) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-4) && (rand.Intn(100) < 40) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-5) && (rand.Intn(100) < 30) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-6) && (rand.Intn(100) < 20) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= MAX_XXX_LEVEL-7) && (rand.Intn(100) < 10) {
		result = badCards[rand.Intn(len(badCards))]
	}
	return result
}

// get symbols on payline of slotxxxResult
func GetPayline(payline []int, slotxxxResult [][]string) []string {
	result := []string{}
	for colIndex, rowIndex := range payline {
		result = append(result, slotxxxResult[colIndex][rowIndex])
	}
	return result
}

// try RandomSpin
func tryRandomSpin() {
	typeCounter := map[string]int{}
	for t, _ := range MAP_LINE_TYPE_TO_PRIZE_RATE {
		typeCounter[t] = 0
	}
	for i := 0; i < 140608; i++ {
		rs := RandomSpin()
		typeRs := CalcLineType(GetPayline(PAYLINES[0], rs))
		// fmt.Println(SlotxxxResult(rs).String(), typeRs)
		typeCounter[typeRs.lineType] += 1
	}
	for k, v := range typeCounter {
		fmt.Println(k, v)
	}
}

// just print, for setup MAP_LINE_TYPE_TO_PRIZE_RATE
func calcPrizeInfo() {
	s := float64(0)
	mapLineTypeToProbability := map[string]float64{}
	//
	mapLineTypeToCounter := map[string]int{}
	all3cardStrs := make([][]string, 0)
	for _, c1 := range SYMBOLS_ORDER {
		for _, c2 := range SYMBOLS_ORDER {
			for _, c3 := range SYMBOLS_ORDER {
				all3cardStrs = append(all3cardStrs, []string{c1, c2, c3})
			}
		}
	}
	n52xx3 := len(all3cardStrs)
	fmt.Println("n52xx3", n52xx3)
	for _, comb := range all3cardStrs {
		lineType := CalcLineType(comb)
		mapLineTypeToCounter[lineType.lineType] += 1
	}
	for k, v := range mapLineTypeToCounter {
		mapLineTypeToProbability[k] = float64(v) / float64(n52xx3)
	}
	//
	for k, p := range mapLineTypeToProbability {
		if k != LINE_TYPE_NOTHING {
			s += p
		}
	}
	fmt.Println("hit prob: ", s)
	mapLineTypeToPrizePart := map[string]float64{
		LINE_TYPE_TRIPS_A_FLUSH: 0.05,
		LINE_TYPE_TRIPS_X_FLUSH: 0.01,
		LINE_TYPE_TRIPS:         0.14,
		LINE_TYPE_FLUSH:         0.2,
		LINE_TYPE_PAIR:          0.3,
		LINE_TYPE_9_OR_10:       0.3,
	}
	for _, lineType := range []string{LINE_TYPE_TRIPS_A_FLUSH, LINE_TYPE_TRIPS_X_FLUSH, LINE_TYPE_TRIPS, LINE_TYPE_FLUSH, LINE_TYPE_PAIR, LINE_TYPE_9_OR_10} {
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

	if c0.Rank == "A" && c1.Rank == "A" && c2.Rank == "A" &&
		c0.Suit == c1.Suit && c0.Suit == c2.Suit {
		return LineType{isWin: true, lineType: LINE_TYPE_TRIPS_A_FLUSH}
	} else if c0.Rank == c1.Rank && c0.Rank == c2.Rank &&
		c0.Suit == c1.Suit && c0.Suit == c2.Suit {
		return LineType{isWin: true, lineType: LINE_TYPE_TRIPS_X_FLUSH}
	} else if c0.Rank == c1.Rank && c0.Rank == c2.Rank {
		return LineType{isWin: true, lineType: LINE_TYPE_TRIPS}
	} else if c0.Suit == c1.Suit && c0.Suit == c2.Suit {
		return LineType{isWin: true, lineType: LINE_TYPE_FLUSH}
	} else if (c0.Rank == c1.Rank) || (c1.Rank == c2.Rank) {
		return LineType{isWin: true, lineType: LINE_TYPE_PAIR}
	} else if score >= 7 {
		return LineType{isWin: true, lineType: LINE_TYPE_9_OR_10}
	} else {
		return LineType{isWin: false, lineType: LINE_TYPE_NOTHING}
	}
}

func CalcWonMoneys(slotxxxResult [][]string, paylineIndexs []int, moneyPerLine int64) (
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
		payline := PAYLINES[paylineIndex]          // payline = []RowIndexOnColumn
		line := GetPayline(payline, slotxxxResult) // line = []symbols
		winType := CalcLineType(line)
		mapLineIndexToMoney[paylineIndex] = int64(MAP_LINE_TYPE_TO_PRIZE_RATE[winType.lineType] * float64(moneyPerLine))
		mapLineIndexToIsWin[paylineIndex] = winType.isWin
		if winType.lineType == LINE_TYPE_TRIPS_A_FLUSH {
			isHitJackpot = true
		}
	}
	sumPay := CalcSumPay(mapLineIndexToMoney)
	if isHitJackpot {
		matchWonType = MATCH_WON_TYPE_JACKPOT
	} else if sumPay >= int64(MAP_LINE_TYPE_TO_PRIZE_RATE[LINE_TYPE_TRIPS]*float64(moneyPerLine)) &&
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

// for test
func XXX() {
	nTry := 30000

	MAX_XXX_LEVEL := 8
	rateLineToJackpot := 0.025
	rateX2ToJackpot := 1.0 / 52
	firstTryRequiredRate := (2.0 / 3.0)
	fullX2WinJackpotRate := 0.05
	spinWinJackpotRate := 0.5

	moneyPerLine := int64(1000)
	jackpot := 500 * moneyPerLine
	userMoney := int64(0)

	for j := 0; j < nTry; j++ {
		userMoney -= moneyPerLine
		slotxxxResult := RandomSpin()
		MapPaylineIndexToWonMoney, _, MatchWonType := CalcWonMoneys(slotxxxResult, []int{0}, moneyPerLine)
		sumMoneyAfterSpin := CalcSumPay(MapPaylineIndexToWonMoney)
		is1stTryFailed := make([]bool, MAX_XXX_LEVEL)

		temp := int64(rateLineToJackpot * float64(moneyPerLine))
		jackpot += temp

		winningMoneyIfStop := int64(0)
		if MatchWonType == MATCH_WON_TYPE_JACKPOT {
			amount := int64(float64(jackpot) * spinWinJackpotRate)
			winningMoneyIfStop = amount
			jackpot += -amount
		} else if sumMoneyAfterSpin > 0 {
			winningMoneyIfStop = sumMoneyAfterSpin
			currentXxxMoney := sumMoneyAfterSpin
			requiredMoneyToGoOn := int64(0)
			// loop x2 game, i is level counter
			i := 0
			for i < 0 {
				isRightPhase3 := false
				phase3result := Random1Card(ACTION_SELECT_SMALL, i)
				var cardRank string
				if len(phase3result) > 0 {
					cardRank = string(phase3result[0])
				}
				if cardRank == "A" || cardRank == "2" || cardRank == "3" ||
					cardRank == "4" || cardRank == "5" || cardRank == "6" {
					isRightPhase3 = true
				} else {
					isRightPhase3 = false
				}
				isFirstTry := (is1stTryFailed[i] == false)
				if isRightPhase3 {
					requiredMoneyToGoOn = 0
					currentXxxMoney = 2 * currentXxxMoney
					winningMoneyIfStop = currentXxxMoney
					i += 1
				} else {
					is1stTryFailed[i] = true
					winningMoneyIfStop = currentXxxMoney
					if i == 0 && isFirstTry {
						requiredMoneyToGoOn = int64(float64(currentXxxMoney) * firstTryRequiredRate)
					} else {
						requiredMoneyToGoOn = currentXxxMoney
					}
					if requiredMoneyToGoOn > 0 {
						userMoney -= requiredMoneyToGoOn
						jackpot += int64(float64(requiredMoneyToGoOn) * rateX2ToJackpot)
					}
				}
			} // end loop x2 game
			if i == MAX_XXX_LEVEL {
				amount := int64(float64(jackpot) * fullX2WinJackpotRate)
				winningMoneyIfStop += amount
				jackpot -= amount
			}
		}
		SumWonMoney := winningMoneyIfStop
		userMoney += SumWonMoney
	}

	fmt.Println("jackpot", jackpot)
	fmt.Println("userMoney", userMoney)
}
