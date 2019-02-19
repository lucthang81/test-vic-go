package gamemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
	_, _ = json.Marshal([]int{})
}

func TestCalcDrawsForACard(t *testing.T) {
	mapPotIndexToPrize, moneyToOpenPots := createAgGoldMinerPots(1000)
	for i := 0; i < len(mapPotIndexToPrize); i++ {
		fmt.Print(fmt.Sprintf("%6d", mapPotIndexToPrize[i]))
	}
	fmt.Println("")
	for i := 0; i < len(moneyToOpenPots); i++ {
		fmt.Print(fmt.Sprintf("%6d", moneyToOpenPots[i]))
	}
	fmt.Println("")
}
