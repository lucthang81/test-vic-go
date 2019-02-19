package slotacp

import (
	"fmt"
	//	"time"
	"math/rand"
	"testing"

	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/utils"
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
	_ = rand.Int()
	utils.GetInt64AtPath(map[string]interface{}{}, "")
}

func TestHaha(t *testing.T) {
	CalcPrizeInfo()
	in := int64(0)
	out := int64(0)
	mapPieces := map[int]int{}
	paylineIndexs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	moneyPerLine := int64(100)

	nHit := 0
	nAll := 0
	for i := 0; i < 10000; i++ {
		r := RandomSpin()
		r1, _, r3 := CalcWonMoneys(r, paylineIndexs, moneyPerLine)
		wonMoney := CalcSumPay(r1)
		if r3 == consts.MATCH_WON_TYPE_AG {
			mapPieces[rand.Intn(8)] += 1
			nHit += 1
		}
		nAll += 1
		//
		in += 20 * 100
		out += wonMoney
	}
	nPics := 111111111
	for _, v := range mapPieces {
		if nPics > v {
			nPics = v
		}
	}
	out += int64(nPics) * 300 * moneyPerLine
	fmt.Println("out, in, nPics, out/in", out, in, nPics, float64(out)/float64(in))
	fmt.Println("nHit, nAll", nHit, nAll)
}
