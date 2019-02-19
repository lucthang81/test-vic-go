package xeng

import (
	"encoding/json"
	"fmt"
)

const (
	APPLE      = "APPLE"
	ORANGE     = "ORANGE"
	LEMON      = "LEMON"
	BELL       = "BELL"
	WATERMELON = "WATERMELON"
	STAR       = "STAR"
	SEVEN      = "SEVEN"
	BAR        = "BAR"

	OE01_3_APPLE       = "OE01_3_APPLE"
	OE02_3_ORANGE      = "OE02_3_ORANGE"
	OE03_3_LEMON       = "OE03_3_LEMON"
	OE04_3_BELL        = "OE04_3_BELL"
	OE05_3_WATERMELON  = "OE05_3_WATERMELON"
	OE06_3_STAR        = "OE06_3_STAR"
	OE07_3_SEVEN       = "OE07_3_SEVEN"
	OE08_5_APPLE       = "OE08_5_APPLE"
	OE09_10_ORANGE     = "OE09_10_ORANGE"
	OE10_15_LEMON      = "OE10_15_LEMON"
	OE11_20_BELL       = "OE11_20_BELL"
	OE12_20_WATERMELON = "OE12_20_WATERMELON"
	OE13_30_STAR       = "OE13_30_STAR"
	OE14_40_SEVEN      = "OE14_40_SEVEN"
	OE15_50_BAR        = "OE15_50_BAR"
	OE16_100_BAR       = "OE16_100_BAR"
	OE17_LOST          = "OE17_LOST"

	OUTCOME01_3                     = "OUTCOME01_3"
	OUTCOME02_3                     = "OUTCOME02_3"
	OUTCOME03_3                     = "OUTCOME03_3"
	OUTCOME04_3                     = "OUTCOME04_3"
	OUTCOME05_3                     = "OUTCOME05_3"
	OUTCOME06_3                     = "OUTCOME06_3"
	OUTCOME07_3                     = "OUTCOME07_3"
	OUTCOME08_5                     = "OUTCOME08_5"
	OUTCOME09_10                    = "OUTCOME09_10"
	OUTCOME10_15                    = "OUTCOME10_15"
	OUTCOME11_20                    = "OUTCOME11_20"
	OUTCOME12_20                    = "OUTCOME12_20"
	OUTCOME13_30                    = "OUTCOME13_30"
	OUTCOME14_40                    = "OUTCOME14_40"
	OUTCOME15_50                    = "OUTCOME15_50"
	OUTCOME16_100                   = "OUTCOME16_100"
	OUTCOME17_3_3_3_3               = "OUTCOME17_3_3_3_3"
	OUTCOME18_3_3_3                 = "OUTCOME18_3_3_3"
	OUTCOME19_5_10_15_20            = "OUTCOME19_5_10_15_20"
	OUTCOME20_20_30_40_50_100       = "OUTCOME20_20_30_40_50_100"
	OUTCOME21_3_3_3_3_5_10_15_20    = "OUTCOME21_3_3_3_3_5_10_15_20"
	OUTCOME22_3_3_3_20_30_40_50_100 = "OUTCOME22_3_3_3_20_30_40_50_100"
	OUTCOME23_ALL                   = "OUTCOME23_ALL"
	OUTCOME24_LOST                  = "OUTCOME24_LOST"
	OUTCOME25_                      = "OUTCOME25_"
	OUTCOME26_                      = "OUTCOME26_"
	OUTCOME27_                      = "OUTCOME27_"
	OUTCOME28_                      = "OUTCOME28_"
	OUTCOME29_                      = "OUTCOME29_"
)

var MAP_OE_TO_RATE map[string]int64
var MAP_OUTCOME map[string][]string
var MAP_OUTCOME_TO_PROBABILITY map[string]float64

func init() {
	MAP_OE_TO_RATE = map[string]int64{
		OE01_3_APPLE:       3,
		OE02_3_ORANGE:      3,
		OE03_3_LEMON:       3,
		OE04_3_BELL:        3,
		OE05_3_WATERMELON:  3,
		OE06_3_STAR:        3,
		OE07_3_SEVEN:       3,
		OE08_5_APPLE:       5,
		OE09_10_ORANGE:     10,
		OE10_15_LEMON:      15,
		OE11_20_BELL:       20,
		OE12_20_WATERMELON: 20,
		OE13_30_STAR:       30,
		OE14_40_SEVEN:      40,
		OE15_50_BAR:        50,
		OE16_100_BAR:       100,
		OE17_LOST:          0,
	}
	MAP_OUTCOME = map[string][]string{
		OUTCOME01_3:   []string{OE01_3_APPLE},
		OUTCOME02_3:   []string{OE02_3_ORANGE},
		OUTCOME03_3:   []string{OE03_3_LEMON},
		OUTCOME04_3:   []string{OE04_3_BELL},
		OUTCOME05_3:   []string{OE05_3_WATERMELON},
		OUTCOME06_3:   []string{OE06_3_STAR},
		OUTCOME07_3:   []string{OE07_3_SEVEN},
		OUTCOME08_5:   []string{OE08_5_APPLE},
		OUTCOME09_10:  []string{OE09_10_ORANGE},
		OUTCOME10_15:  []string{OE10_15_LEMON},
		OUTCOME11_20:  []string{OE11_20_BELL},
		OUTCOME12_20:  []string{OE12_20_WATERMELON},
		OUTCOME13_30:  []string{OE13_30_STAR},
		OUTCOME14_40:  []string{OE14_40_SEVEN},
		OUTCOME15_50:  []string{OE15_50_BAR},
		OUTCOME16_100: []string{OE16_100_BAR},
		OUTCOME17_3_3_3_3: []string{
			OE01_3_APPLE, OE02_3_ORANGE, OE03_3_LEMON, OE04_3_BELL},
		OUTCOME18_3_3_3: []string{
			OE05_3_WATERMELON, OE06_3_STAR, OE07_3_SEVEN},
		OUTCOME19_5_10_15_20: []string{
			OE08_5_APPLE, OE09_10_ORANGE, OE10_15_LEMON, OE11_20_BELL},
		OUTCOME20_20_30_40_50_100: []string{
			OE12_20_WATERMELON, OE13_30_STAR, OE14_40_SEVEN,
			OE15_50_BAR, OE16_100_BAR},
		OUTCOME21_3_3_3_3_5_10_15_20: []string{
			OE01_3_APPLE, OE02_3_ORANGE, OE03_3_LEMON, OE04_3_BELL,
			OE08_5_APPLE, OE09_10_ORANGE, OE10_15_LEMON, OE11_20_BELL},
		OUTCOME22_3_3_3_20_30_40_50_100: []string{
			OE05_3_WATERMELON, OE06_3_STAR, OE07_3_SEVEN,
			OE12_20_WATERMELON, OE13_30_STAR, OE14_40_SEVEN,
			OE15_50_BAR, OE16_100_BAR},
		OUTCOME23_ALL: []string{
			OE01_3_APPLE, OE02_3_ORANGE, OE03_3_LEMON, OE04_3_BELL,
			OE05_3_WATERMELON, OE06_3_STAR, OE07_3_SEVEN,
			OE08_5_APPLE, OE09_10_ORANGE, OE10_15_LEMON, OE11_20_BELL,
			OE12_20_WATERMELON, OE13_30_STAR, OE14_40_SEVEN,
			OE15_50_BAR, OE16_100_BAR},
		OUTCOME24_LOST: []string{OE17_LOST},
	}
	MAP_OUTCOME_TO_PROBABILITY = map[string]float64{}
}

func CalcPrizeInfo() {
	MapOutcomeToPrize := map[string]int64{}
	for outcome, listOe := range MAP_OUTCOME {
		for _, oe := range listOe {
			MapOutcomeToPrize[outcome] += MAP_OE_TO_RATE[oe]
		}
	}
	bs, _ := json.MarshalIndent(MapOutcomeToPrize, "", "    ")
	fmt.Println(string(bs))
}
