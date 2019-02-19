package slotax1to5

import (
	"fmt"
	//	"time"
	"testing"

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
	utils.GetInt64AtPath(map[string]interface{}{}, "")
}

func TestHaha(t *testing.T) {
	CalcPrizeInfo()

	//	for j := 0; j < 10; j++ {
	//		s := int64(0)
	//		for i := 0; i < 100; i++ {
	//			paylineIndexs := []int{0, 1, 2}
	//			moneyPerLine := int64(1000)
	//			slotResult := RandomSpin()
	//
	//			mapPaylineIndexToWonMoney, _, _ := CalcWonMoneys(slotResult, paylineIndexs, moneyPerLine)
	//			s1 := CalcSumPay(mapPaylineIndexToWonMoney)
	//			s += s1
	//		}
	//		fmt.Println("s", s)
	//	}

}
