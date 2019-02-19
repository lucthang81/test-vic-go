package wheel

import (
	"fmt"
	//	"time"
	"testing"
)

var _, _ = fmt.Println("")

func TestDummy(t *testing.T) {
	var r int
	//
	r = Dummy(1)
	if r != 1 {
		t.Error()
	}
	//
}

func TestHaha(t *testing.T) {
	//
	/*
		for _, nDup := range []int{5, 4, 3} {
			for _, symbol := range SYMBOLS_ORDER {
				fmt.Printf("%v | nDup %v | %.2f\n", symbol, nDup, CalcRateNDupInALine(symbol, nDup))
			}
		}

		fmt.Printf("%v | nDup %v | %.2f\n", "oeoe", "haha", CalcRateNDupInALine2(2/15.0, 5))
	*/

	// test random spin
	fmt.Println("MAP_SYMBOL_TO_RANGE", MAP_SYMBOL_TO_RANGE)
	fmt.Println("MAP_SYMBOL1_TO_RANGE", MAP_SYMBOL1_TO_RANGE)
	for i := 0; i < 20; i++ {
		fmt.Println(RandomSpin())
	}
}
