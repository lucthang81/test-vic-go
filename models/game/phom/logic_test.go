package phom

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
}

func TestCheckIsDraw(t *testing.T) {
	var cr bool
	var err error
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("Kh"),
		z.FNewCardFS("Qh"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("Qh"),
		z.FNewCardFS("Kh"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("Qh"),
		z.FNewCardFS("Ks"),
	})
	if cr != false {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("9h"),
		z.FNewCardFS("9s"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("3s"),
		z.FNewCardFS("5s"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("2s"),
		z.FNewCardFS("5s"),
	})
	if cr != false {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("Jd"),
		z.FNewCardFS("9d"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, _ = CheckIsDraw([]z.Card{
		z.FNewCardFS("Kc"),
		z.FNewCardFS("Ac"),
	})
	if cr != false {
		t.Error()
	}
	//
	cr, err = CheckIsDraw([]z.Card{})
	if err == nil {
		t.Error()
	}
}

func TestCheckIsCombo(t *testing.T) {
	var cr bool
	var err error
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("2c"),
		z.FNewCardFS("3c"),
	})
	if err == nil {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("2c"),
		z.FNewCardFS("4c"),
		z.FNewCardFS("3c"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("5c"),
		z.FNewCardFS("2c"),
		z.FNewCardFS("4c"),
		z.FNewCardFS("3c"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("2c"),
		z.FNewCardFS("4s"),
		z.FNewCardFS("3c"),
	})
	if cr != false {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("4s"),
		z.FNewCardFS("4c"),
		z.FNewCardFS("4d"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("4s"),
		z.FNewCardFS("4c"),
		z.FNewCardFS("4d"),
		z.FNewCardFS("4h"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS("5c"),
		z.FNewCardFS("2c"),
		z.FNewCardFS("4c"),
		z.FNewCardFS("3c"),
		z.FNewCardFS("Ac"),
	})
	if cr != true {
		t.Error()
	}
	//
	cr, err = CheckIsCombo([]z.Card{
		z.FNewCardFS(""),
		z.FNewCardFS("2c"),
		z.FNewCardFS("Ac"),
	})
	if cr != false {
		t.Error()
	}
}

func TestCheckCanPopCard(t *testing.T) {
	var hand []z.Card
	var lockedCards []z.Card
	var cardToPop z.Card
	var cr bool
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("8s"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3s")}
	cardToPop = z.FNewCardFS("3d")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != false {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("8s"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3s")}
	cardToPop = z.FNewCardFS("8s")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("8s"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3s")}
	cardToPop = z.FNewCardFS("Kh")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != false {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("8s"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3d")}
	cardToPop = z.FNewCardFS("3s")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("8s"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3d"), z.FNewCardFS("4d")}
	cardToPop = z.FNewCardFS("3s")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != false {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("3h"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3d"), z.FNewCardFS("3s")}
	cardToPop = z.FNewCardFS("2d")
	cr = CheckCanPopCard(hand, lockedCards, cardToPop)
	if cr != false {
		t.Error()
	}
}

func TestCheckCanEatCard(t *testing.T) {
	var hand []z.Card
	var lockedCards []z.Card
	var newcard z.Card
	var cr bool
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("3s"),
	}
	lockedCards = []z.Card{}
	newcard = z.FNewCardFS("Ac")
	cr = CheckCanEatCard(hand, lockedCards, newcard)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("3s"),
	}
	lockedCards = []z.Card{}
	newcard = z.FNewCardFS("2s")
	cr = CheckCanEatCard(hand, lockedCards, newcard)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("3s"),
	}
	lockedCards = []z.Card{}
	newcard = z.FNewCardFS("2d")
	cr = CheckCanEatCard(hand, lockedCards, newcard)
	if cr != false {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("3h"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3d")}
	newcard = z.FNewCardFS("3h")
	cr = CheckCanEatCard(hand, lockedCards, newcard)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("As"), z.FNewCardFS("Ad"), z.FNewCardFS("2d"),
		z.FNewCardFS("3s"), z.FNewCardFS("3c"), z.FNewCardFS("3d"),
		z.FNewCardFS("4d"), z.FNewCardFS("3h"), z.FNewCardFS("8c"),
	}
	lockedCards = []z.Card{z.FNewCardFS("3s")}
	newcard = z.FNewCardFS("3h")
	cr = CheckCanEatCard(hand, lockedCards, newcard)
	if cr != false {
		t.Error()
	}
}

func TestShowCombos(t *testing.T) {
	var hand []z.Card
	var lockedCards []z.Card
	var ways [][][]z.Card
	var cr bool
	//
	hand = []z.Card{
		z.FNewCardFS("Kh"), z.FNewCardFS("Kd"), z.FNewCardFS("6s"),
		z.FNewCardFS("5h"), z.FNewCardFS("5c"), z.FNewCardFS("5s"),
		z.FNewCardFS("4s"), z.FNewCardFS("3s"), z.FNewCardFS("2s"),
	}
	lockedCards = []z.Card{}
	ways = ShowCombosMinPoint(hand, lockedCards)
	if fmt.Sprintf("%v", ways) != "[[[5h 5c 5s] [4s 3s 2s] [Kh Kd 6s]]]" {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("Kh"), z.FNewCardFS("Kc"), z.FNewCardFS("Ks"),
		z.FNewCardFS("7c"), z.FNewCardFS("5c"), z.FNewCardFS("6c"),
		z.FNewCardFS("6h"), z.FNewCardFS("6d"), z.FNewCardFS("6s"),
	}
	lockedCards = []z.Card{}
	cr, _ = CheckIsFullCombos(hand, lockedCards)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{}
	lockedCards = []z.Card{}
	cr, _ = CheckIsFullCombos(hand, lockedCards)
	if cr != true {
		t.Error()
	}
	//
	hand = []z.Card{
		z.FNewCardFS("2s"), z.FNewCardFS("2c"), z.FNewCardFS("2d"),
		z.FNewCardFS("2h"), z.FNewCardFS("3s"), z.FNewCardFS("4s"),
		z.FNewCardFS("5s"),
	}
	ways = SplitToCombos(hand)
	fmt.Println("\nTestShowCombos")
	for _, way := range ways {
		fmt.Println(z.ToStringss(way))
	}
}

func TestShowCombos2(t *testing.T) {
	var hand []z.Card
	var lockedCards []z.Card
	var ways [][][]z.Card
	hand = []z.Card{
		z.FNewCardFS("7h"), z.FNewCardFS("6h"), z.FNewCardFS("5h"),
		z.FNewCardFS("3h"), z.FNewCardFS("3c"), z.FNewCardFS("3s"),
		z.FNewCardFS("8s"), z.FNewCardFS("6s"), z.FNewCardFS("Ah"), z.FNewCardFS("Ad"),
	}
	lockedCards = []z.Card{}
	ways = ShowCombos(hand, lockedCards)
	for i, way := range ways {
		fmt.Println("TestShowCombos2", i, way)
	}
}

func TestCalcDrawsForACard(t *testing.T) {
	fmt.Println("\nTestCalcDrawsForACard")
	var card z.Card
	card = z.FNewCardFS("As")
	fmt.Println(card, z.ToStringss(CalcDrawsForACard(card)))
	card = z.FNewCardFS("Qc")
	fmt.Println(card, z.ToStringss(CalcDrawsForACard(card)))
	card = z.FNewCardFS("7d")
	fmt.Println(card, z.ToStringss(CalcDrawsForACard(card)))
	card = z.FNewCardFS("Kh")
	fmt.Println(card, z.ToStringss(CalcDrawsForACard(card)))
}
