package cardgame

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
}

func TestCalcNOC(t *testing.T) {
	var noc int
	//
	noc = CalcNOCombs(4, 2)
	if noc != 6 {
		t.Error()
	}
	//
	noc = CalcNOCombs(10, 2)
	if noc != 45 {
		t.Error()
	}
	//
	noc = CalcNOCombs(9, 4)
	if noc != 126 {
		t.Error()
	}

}

func TestBisectRight(t *testing.T) {
	var a []int64
	var x int64
	var i int
	var err error
	_ = []interface{}{a, x, i, err}

	//a = []int64{1, 1, 2, 3, 4, 5, 6, 6, 8, 9, 10}
	a = []int64{10, 9, 8, 6, 6, 5, 4, 3, 2, 1, 1}
	x = 7
	i, err = BisectRight(a, x)
	if i != 3 {
		t.Error()
	}
	//	fmt.Println(x, i)
	x = 6
	i, err = BisectRight(a, x)
	if i != 5 {
		t.Error()
	}
	//	fmt.Println(x, i)
	x = -5
	i, err = BisectRight(a, x)
	if i != 11 {
		t.Error()
	}
	//	fmt.Println(x, i)
	x = 15
	i, err = BisectRight(a, x)
	if i != 0 {
		t.Error()
	}
	//	fmt.Println(x, i)
}

func TestHumanFormatNumber(t *testing.T) {
	var n interface{}
	n = float64(167678.123)
	r := HumanFormatNumber(n)
	if r != "167,678" {
		t.Error()
	}
}

func TestHoho(t *testing.T) {
	// test list all combs
	list := Range(100)
	//	allCombs := GetCombinations(list, 2)
	allCombs := GetCombinations2(list, 2, time.Now().Add(5000*time.Microsecond))
	fmt.Println(len(allCombs))
	//	fmt.Println(allCombs)
}

func TestHihi(t *testing.T) {
	//	is := []int{0, 1, 2, 3, 4, 5, 6}
	//	for k := 0; k < 10; k++ {
	//		temp := make([]int, len(is))
	//		copy(temp, is)
	//		ShuffleInts(temp)
	//		fmt.Println(temp)
	//	}
	//	fmt.Println("is", is)
}
