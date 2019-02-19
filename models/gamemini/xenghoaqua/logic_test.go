package xeng

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	fmt.Println("hihi")
	CalcPrizeInfo()
	if 1 < 2 {
		t.Error()
	}
}
