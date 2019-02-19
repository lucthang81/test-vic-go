package wheel

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
var MAP_SYMBOL_TO_RANGE map[string][]int
var MAP_SYMBOL1_TO_RANGE map[string][]int

func init() {
	// for not used import error
	_ = fmt.Printf
	_ = errors.New("")
	rand.Seed(time.Now().Unix())
	_ = cardgame.Card{}
	// sylbols, map symbol to rate
	SYMBOLS = map[string]int{
		"1":  300, // 50k
		"2":  300, // 50k
		"3":  50,  // 100k
		"4":  50,  // 100k
		"5":  35,  // 200k
		"6":  35,  // 200k
		"7":  25,  // 300k
		"8":  25,  // 300k
		"9":  30,  // 500k
		"10": 50,  // Thêm 1 lượt
		"11": 50,  // Thêm 1 lượt
		"12": 50,  // Thêm 2 lượt
	}
	SYMBOLS1 = map[string]int{
		"a": 5,     // 100k
		"b": 10,    // 50k
		"c": 50,    // 10k
		"d": 100,   // 5k
		"e": 250,   // 2k
		"f": 500,   // 1k
		"g": 1000,  // 0.5k
		"h": 98085, // miss
	}
	SYMBOLS_ORDER = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	SYMBOLS1_ORDER = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
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
}

func Dummy(n int) int {
	return n
}

// return wheelResult, []string(2),
// wheelResult[0] là kq vòng SYMBOLS 12 , wheelResult[1] là kq vòng SYMBOLS1 8
func RandomSpin() []string {
	w0symbol := SYMBOLS_ORDER[len(SYMBOLS_ORDER)-1]
	randomInt := rand.Intn(NUMBER_OF_SYMBOLS) + 1
	for symb, symbRange := range MAP_SYMBOL_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w0symbol = symb
			break
		}
	}
	w1symbol := SYMBOLS1_ORDER[len(SYMBOLS1_ORDER)-1]
	randomInt = rand.Intn(NUMBER_OF_SYMBOLS1) + 1
	for symb, symbRange := range MAP_SYMBOL1_TO_RANGE {
		if (symbRange[0] <= randomInt) && (randomInt <= symbRange[1]) {
			w1symbol = symb
			break
		}
	}
	return []string{w0symbol, w1symbol}
}
