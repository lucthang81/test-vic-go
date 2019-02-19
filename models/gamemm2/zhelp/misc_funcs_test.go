package zhelp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
	_, _ = json.Marshal([]int{})
	_ = z.NewDeck()
}

func TestHihi(t *testing.T) {
	PlayersOrder := []int64{10, 11, 12, 13}
	var r int64
	r = GetNextPlayer(10, PlayersOrder)
	if r != 11 {
		t.Error()
	}
	r = GetNextPlayer(13, PlayersOrder)
	if r != 10 {
		t.Error()
	}
	r = GetPrevPlayer(10, PlayersOrder)
	if r != 13 {
		t.Error()
	}
	r = GetPrevPlayer(13, PlayersOrder)
	if r != 12 {
		t.Error()
	}
}
