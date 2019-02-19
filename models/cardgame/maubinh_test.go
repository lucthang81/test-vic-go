package cardgame

import (
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
}

func TestMaubinh(t *testing.T) {
	//	hand, _ := ToCardsFromStrings([]string{
	//		"Qh", "As", "Ac", "6d", "7c", "4s", "4h", "8c", "9c", "Td", "5s", "5d", "5h"})
	//	fmt.Println("hand", hand)
	//	st := time.Now()
	//	ways := MaubinhArrangeCards(hand)
	//	fmt.Println(time.Now().Sub(st))
	//	for _, way := range ways {
	//		fmt.Println(way)
	//	}
}

func TestMaubinh2(t *testing.T) {
	//	for i := 0; i < 5; i++ {
	//		deck := NewDeck()
	//		Shuffle(deck)
	//		hand, _ := DealCards(&deck, 13)
	//		fmt.Println("hand", hand)
	//		st := time.Now()
	//		ways := MaubinhArrangeCards(hand)
	//		fmt.Println("calc dur", time.Now().Sub(st))
	//		for _, way := range ways {
	//			fmt.Println(way)
	//		}
	//	}
}
