package wheel2

import (
	"errors"
	"fmt"
	// "math"
	"math/rand"
	"time"

	"github.com/vic/vic_go/models/cardgame"
)

// map symbol to number of this symbol in a reel
var SYMBOLS map[string]int
var SYMBOLS1 map[string]int
var NUMBER_OF_SYMBOLS int // include weight
var NUMBER_OF_SYMBOLS1 int
var SYMBOLS_ORDER []string
var SYMBOLS1_ORDER []string

// range from 1 to len
var MAP_SYMBOL_TO_RANGE map[string][]int

// range from 1 to len
var MAP_SYMBOL1_TO_RANGE map[string][]int

func init() {
	// for not used import error
	_ = fmt.Printf
	_ = errors.New("")
	rand.Seed(time.Now().Unix())
	_ = cardgame.Card{}
	// sylbols, map symbol to rate
	SYMBOLS = map[string]int{
		"1":  5,   // 500
		"3":  525, // 5000
		"5":  290, // 10000
		"7":  100, // 20000
		"9":  50,  // 50000
		"10": 5,   // 500000
		"11": 20,  // Thêm 1 lượt
		"12": 5,   // Thêm 2 lượt
	}
	SYMBOLS1 = map[string]int{
		"a": 200000, // 200
		"b": 200000, // 500
		"c": 20000,  // 1000
		"d": 400000, // 2500
		"e": 50000,  // 5000
		"f": 15000,  // 10000
		"g": 1000,   // 50000
		"h": 20,     // 100000
		"i": 10,     // 200000
		"j": 5,      // 500000
		"k": 2,      // 1000000
		"l": 1,      // 1500000
	}
	SYMBOLS_ORDER = []string{"1", "3", "5", "7", "9", "10", "11", "12"}
	SYMBOLS1_ORDER = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	NUMBER_OF_SYMBOLS = 0
	for _, counter := range SYMBOLS {
		NUMBER_OF_SYMBOLS += counter
	}
	NUMBER_OF_SYMBOLS1 = 0
	for _, counter := range SYMBOLS1 {
		NUMBER_OF_SYMBOLS1 += counter
	}
	MAP_SYMBOL_TO_RANGE = map[string][]int{}
	rangeLowerB := 0
	for _, symb := range SYMBOLS_ORDER {
		noSymb := SYMBOLS[symb]
		symbRange := []int{rangeLowerB + 1, rangeLowerB + noSymb}
		rangeLowerB = rangeLowerB + noSymb
		MAP_SYMBOL_TO_RANGE[symb] = symbRange
	}
	MAP_SYMBOL1_TO_RANGE = map[string][]int{}
	rangeLowerB = 0
	for _, symb := range SYMBOLS1_ORDER {
		noSymb := SYMBOLS1[symb]
		symbRange := []int{rangeLowerB + 1, rangeLowerB + noSymb}
		rangeLowerB = rangeLowerB + noSymb
		MAP_SYMBOL1_TO_RANGE[symb] = symbRange
	}
	//	fmt.Println("MAP_SYMBOL_TO_RANGE", MAP_SYMBOL_TO_RANGE["12"])
	//	fmt.Println("MAP_SYMBOL1_TO_RANGE", MAP_SYMBOL1_TO_RANGE["c"])
}

func Dummy(n int) int {
	return n
}

// return wheelResult, []string(2),
// wheelResult[0] là kq vòng SYMBOLS  , wheelResult[1] là kq vòng SYMBOLS1
func RandomSpin() []string {
	w0symbol := SYMBOLS_ORDER[0]
	randomInt := rand.Intn(NUMBER_OF_SYMBOLS) + 1
	for symb, symbRange := range MAP_SYMBOL_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w0symbol = symb
			break
		}
	}
	w1symbol := SYMBOLS1_ORDER[0]
	randomInt = rand.Intn(NUMBER_OF_SYMBOLS1) + 1
	for symb, symbRange := range MAP_SYMBOL1_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w1symbol = symb
			break
		}
	}
	return []string{w0symbol, w1symbol}
}

// only return w1symbol in [w1lb .. w1up]
func LimitedSpin(w1lb string, w1up string) []string {
	w0symbol := SYMBOLS_ORDER[0]
	randomInt := rand.Intn(NUMBER_OF_SYMBOLS) + 1
	for symb, symbRange := range MAP_SYMBOL_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w0symbol = symb
			break
		}
	}
	w1symbol := SYMBOLS1_ORDER[0]
	lowerRange := MAP_SYMBOL1_TO_RANGE[w1lb][0]
	upperRange := MAP_SYMBOL1_TO_RANGE[w1up][1]
	randomInt = lowerRange + rand.Intn(upperRange+1-lowerRange)
	for symb, symbRange := range MAP_SYMBOL1_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w1symbol = symb
			break
		}
	}
	return []string{w0symbol, w1symbol}
}

func CalcPrizeInfo() {
	rateSpinToMoney := float64(1700)
	prize := map[string]float64{
		"a": 200,
		"b": 500,
		"c": 1000,
		"d": 2500,
		"e": 5000,
		"f": 10000,
		"g": 50000,
		"h": 100000,
		"i": 200000,
		"j": 500000,
		"k": 1000000,
		"l": 1500000,
	}
	prizePart := map[string]float64{
		"a": 0.05,
		"b": 0.05,
		"c": 0.05,
		"d": 0.05,
		"e": 0.1,
		"f": 0.1,
		"g": 0.1,
		"h": 0.1,
		"i": 0.1,
		"j": 0.1,
		"k": 0.1,
		"l": 0.1,
	}
	_, _, _ = rateSpinToMoney, prize, prizePart

	// estimate
	//	sumP := float64(0)
	//	for _, k := range SYMBOLS1_ORDER {
	//		p := prizePart[k] * rateSpinToMoney / prize[k]
	//		sumP += p
	//		fmt.Println("p k", p*10000, k)
	//	}
	//	fmt.Println("sumP", sumP)

	//
	fmt.Println("NUMBER_OF_SYMBOLS1", NUMBER_OF_SYMBOLS1)
	sumM := float64(0)
	for _, k := range SYMBOLS1_ORDER {
		p := float64(SYMBOLS1[k]) / float64(NUMBER_OF_SYMBOLS1)
		M := p * prize[k]
		fmt.Println(k, M, p)
		sumM += M
	}
	fmt.Println("sumM", sumM)
}
