package taixiu

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	ccu := GetSNHuman()
	_ = ccu
	_ = fmt.Println
	//	fmt.Println("ccu", ccu)
}

func Test2(t *testing.T) {
	top_date := "2018-09-12"
	pid := int64(1)
	var e error
	e = TopChangeKey(top_date, pid, 5, 10, 15)
	if e != nil {
		t.Error(e)
	}
	TopChangeKey(top_date, pid, 7, 15, 22)
	TopChangeKey(top_date, pid, 10, 22, 32)
	TopChangeKey(top_date, pid, -4, 32, 28)
	TopChangeKey(top_date, pid, -6, 28, 22)
	TopChangeKey(top_date, pid, 5, 22, 27)
	TopChangeKey(top_date, pid, 3, 27, 30)
	//
	TopChangeKey(top_date, 2, 50, 100, 150)
	TopChangeKey(top_date, 3, -50, 150, 100)
	TopChangeKey(top_date, 4, 40, 30, 70)
	TopChangeKey(top_date, 5, 30, 200, 230)
	TopChangeKey(top_date, 6, 70, 0, 70)
	TopChangeKey(top_date, 7, -5, 100, 95)
	TopChangeKey(top_date, 8, 0, 0, 0)
	//
	r, e := TopLoadLeaderboard(top_date, false)
	fmt.Println(r)
	r, e = TopLoadLeaderboard(top_date, true)
	fmt.Println(r)
}
