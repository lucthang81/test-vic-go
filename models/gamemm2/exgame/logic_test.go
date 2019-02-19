package exgame

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
	fmt.Println("Hihihi")
}
