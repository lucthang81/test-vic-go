package wheel2

import (
	"fmt"
	//	"time"
	//	"math/rand"
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

	fmt.Println(LimitedSpin("b", "b"))

	//	CalcPrizeInfo()
}
