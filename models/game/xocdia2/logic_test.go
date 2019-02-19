package xocdia2

import (
	"fmt"
	"testing"
)

func init() {
	fmt.Print("")
}

func TestRandom(t *testing.T) {
	counter := make(map[string]int)
	for i := 0; i < 16000; i++ {
		counter[RandomShake()] += 1
	}
	//fmt.Printf("%#v", counter)
}

func TestHaha(t *testing.T) {
	acceptedMap := map[string]int64{
		"SELECTION_0_RED": 0,
		"SELECTION_1_RED": 2,
		"SELECTION_3_RED": 0,
		"SELECTION_4_RED": 0,
	}
	betInfo := map[int64]map[string]int64{
		3: {
			"SELECTION_0_RED": 0,
			"SELECTION_1_RED": 10000,
			"SELECTION_3_RED": 0,
			"SELECTION_4_RED": 0,
			"SELECTION_EVEN":  0,
			"SELECTION_ODD":   0,
		},
	}
	result := CalcResult(2, "", 0, betInfo, acceptedMap, "OUTCOME_1_RED")
	fmt.Println(result)
}
