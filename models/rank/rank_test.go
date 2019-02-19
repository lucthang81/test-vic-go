package rank

import (
	"fmt"
	//	"math/rand"
	"math"
	"testing"
	//	"github.com/vic/vic_go/datacenter"
	//	"github.com/vic/vic_go/record"
	//	"github.com/vic/vic_go/zconfig"
)

func Test01(t *testing.T) {
	_ = fmt.Println
}

func Test02(t *testing.T) {
	e := Reset(RANK_TEST1)
	if e != nil {
		t.Error(e)
	}
	e1 := ChangeKey(RANK_TEST1, 1, 53)
	e2 := ChangeKey(RANK_TEST1, 2, 57)
	e3 := ChangeKey(RANK_TEST1, 3, 41)
	e4 := ChangeKey(RANK_TEST1, 4, 21)
	e5 := ChangeKey(RANK_TEST1, 5, 78)
	if (e1 != nil) || (e2 != nil) || (e3 != nil) || (e4 != nil) || (e5 != nil) {
		t.Error(e1, e2, e3, e4, e5)
	}
	// Reset(RANK_TEST2)
	//	for i := int64(1); i < 100000; i++ {
	//		e = ChangeKey(RANK_TEST2, i, rand.Float64())
	//	}
	leaderboard, e := LoadLeaderboard(RANK_TEST1)
	if e != nil || len(leaderboard) == 0 {
		t.Error(e)
	}
	rkey, pos, e := LoadKeyAndPosition(RANK_TEST1, 3)
	if e != nil {
		t.Error(e)
	}
	epsilon := 0.001
	if math.Abs(rkey-41) > epsilon || pos != 4 {
		t.Error(rkey, pos)
	}

}
