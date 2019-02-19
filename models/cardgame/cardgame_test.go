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

func TestCardString(t *testing.T) {
	var cardStr string

	cardStr = Card{Rank: "A", Suit: "d"}.String()
	if cardStr != "Ad" {
		t.Error()
	}

	cardStr = Card{}.String()
	if cardStr != "" {
		t.Error()
	}

	cardStr = Card{Rank: "T", Suit: "s"}.String()
	if cardStr != "Ts" {
		t.Error()
	}
}

func TestNewCardFS(t *testing.T) {
	var card Card
	var err error

	card, err = NewCardFS("Kd")
	if (card != Card{"K", "d"}) ||
		(err != nil) {
		t.Error()
	}

	card, err = NewCardFS("Ld")
	if (card != Card{}) ||
		(err.Error() != errors.New("Wrong card rank").Error()) {
		t.Error()
	}

	card, err = NewCardFS("Th")
	if (card != Card{"T", "h"}) ||
		(err != nil) {
		fmt.Print(err)
		t.Error()
	}

	card, err = NewCardFS("")
	if (card != Card{"", ""}) ||
		(err != nil) {
		fmt.Print(err)
		t.Error()
	}

	card, err = NewCardFS("AKo")
	if (card != Card{"", ""}) ||
		(err.Error() != errors.New("Wrong card string len").Error()) {
		fmt.Print(err)
		t.Error()
	}
}

func TestPopCards(t *testing.T) {
	var hand []Card
	var err error

	hand = []Card{
		FNewCardFS("As"), FNewCardFS("Ad"), FNewCardFS("7h"),
		FNewCardFS("8h"), FNewCardFS("2d"), FNewCardFS("3c"),
		FNewCardFS("6h"), FNewCardFS("4d"), FNewCardFS("3s"),
	}
	err = PopCards(&hand, []Card{FNewCardFS("As"), FNewCardFS("Ad")})
	fmt.Println(hand, ToString(hand))
	if err != nil {
		t.Error()
	}

	hand = []Card{
		FNewCardFS("As"), FNewCardFS("Ad"), FNewCardFS("7h"),
		FNewCardFS("8h"), FNewCardFS("2d"), FNewCardFS("3c"),
		FNewCardFS("6h"), FNewCardFS("4d"), FNewCardFS("3s"),
	}
	err = PopCards(&hand, []Card{FNewCardFS("Ad"), FNewCardFS("Ah")})
	//fmt.Println(hand)
	if err == nil {
		t.Error()
	}
}

func TestHaha(t *testing.T) {

}
