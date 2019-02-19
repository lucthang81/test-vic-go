// Common functions for card games.
package cardgame

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

var RANKS []string
var SUITS []string

func init() {
	fmt.Print("")
	RANKS = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"}
	SUITS = []string{"s", "c", "d", "h"}
	rand.Seed(time.Now().Unix())
}

// Represent a card in game
type Card struct {
	Rank string
	Suit string
}

// Get string represent a card obj, ex: Ad, 7s, Kd,..
func (card Card) String() string {
	return card.Rank + card.Suit
}

// Chuyen tu bai viet dinh dang string mau binh sang card cua Tung beo
func ToCard(value string) Card {
	chars := strings.Split(value, " ")
	if len(chars) != 2 {
		return Card{}
	}
	ret := Card{}
	ret.Suit = chars[0]
	if chars[1] == "10" {
		ret.Rank = "T"
	} else {
		ret.Rank = strings.ToUpper(chars[1])
	}
	return ret

}

// for json data
func ToSliceString(cards []Card) []string {
	result := make([]string, len(cards))
	for i, card := range cards {
		result[i] = card.String()
	}
	return result
}

// for json data
func ToStringss(cardss [][]Card) [][]string {
	result := make([][]string, 0)
	for _, cards := range cardss {
		result = append(result, ToSliceString(cards))
	}
	return result
}

// for read data from client
func ToCardssFromStringss(strss [][]string) ([][]Card, error) {
	result := [][]Card{}
	for _, strs := range strss {
		cards := []Card{}
		for _, cardStr := range strs {
			card, err := NewCardFS(cardStr)
			if err != nil {
				return nil, err
			} else {
				cards = append(cards, card)
			}
		}
		result = append(result, cards)
	}
	return result, nil
}

// for read data from client
func ToCardsFromStrings(strs []string) ([]Card, error) {
	result := []Card{}
	for _, cardStr := range strs {
		card, err := NewCardFS(cardStr)
		if err != nil {
			return nil, err
		} else {
			result = append(result, card)
		}
	}
	return result, nil
}

// dont notify error
func ToCardsFromStrings2(strs []string) []Card {
	result := []Card{}
	for _, cardStr := range strs {
		card, err := NewCardFS(cardStr)
		if err != nil {
			return nil
		} else {
			result = append(result, card)
		}
	}
	return result
}

// for json data
func ToStringsss(cardsss [][][]Card) [][][]string {
	result := make([][][]string, 0)
	for _, cardss := range cardsss {
		result = append(result, ToStringss(cardss))
	}
	return result
}

// for filter combos
func ToString(cards []Card) string {
	temp := ToSliceString(cards)
	bytes, err := json.Marshal(temp)
	if err != nil {
		return ""
	} else {
		return string(bytes)
	}
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindCardInSlice(sub Card, list []Card) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// Make new obj card from string, ex: Ad, 7s, Kd,..
func NewCardFS(cardStr string) (Card, error) {
	chars := strings.Split(cardStr, "")
	if len(chars) == 0 {
		return Card{}, nil
	} else {
		if len(chars) != 2 {
			return Card{}, errors.New("Wrong card string len")
		}
		if FindStringInSlice(chars[0], RANKS) == -1 {
			return Card{}, errors.New("Wrong card rank")
		}
		if FindStringInSlice(chars[1], SUITS) == -1 {
			return Card{}, errors.New("Wrong card suit")
		}
		return Card{chars[0], chars[1]}, nil
	}
}

// Force make new obj card from string, ex: Ad, 7s, Kd,.. Dont notify error
func FNewCardFS(cardStr string) Card {
	card, _ := NewCardFS(cardStr)
	return card
}

// full 52 cards,
// need to shuffle after call this func
func NewDeck() []Card {
	result := make([]Card, 0)
	for _, Rank := range RANKS {
		for _, Suit := range SUITS {
			result = append(result, Card{Rank, Suit})
		}
	}
	return result
}

func NewBacayDeck() []Card {
	result := make([]Card, 0)
	for _, Rank := range RANKS {
		if Rank != "T" && Rank != "J" && Rank != "Q" && Rank != "K" {
			for _, Suit := range SUITS {
				result = append(result, Card{Rank, Suit})
			}
		}
	}
	return result
}

func Shuffle(cards []Card) {
	n := len(cards)
	for i := n - 1; i >= 1; i-- {
		j := rand.Intn(i + 1) // 0 <= j <=i
		temp := cards[i]
		cards[i] = cards[j]
		cards[j] = temp
	}
}

// remove some cards from input deck, at the end of the slice
// return dealt cards
// return error if not enough cards
func DealCards(deckP *[]Card, quantity int) ([]Card, error) {
	deck := *deckP
	if len(deck) < quantity {
		return nil, errors.New("Not enough cards in deck")
	}
	dealtCards := make([]Card, quantity)
	for i, card := range deck[len(deck)-quantity:] {
		dealtCards[quantity-1-i] = card
	}
	*deckP = deck[:len(deck)-quantity]
	return dealtCards, nil
}

// đánh bài, thay đổi hand
// trả  về lỗi nếu lá định đánh không có trên tay
func PopCards(handP *[]Card, cardsToPop []Card) error {
	possToPop := make(map[int]bool)
	for _, card := range cardsToPop {
		temp := FindCardInSlice(card, *handP)
		if temp == -1 {
			return errors.New("Card to pop is not in hand")
		} else {
			possToPop[temp] = true
		}
	}
	newHand := make([]Card, 0)
	for i, card := range *handP {
		if possToPop[i] == false {
			newHand = append(newHand, card)
		}
	}
	*handP = newHand
	return nil
}

/// sub two slice
func Subtracted(fullSet []Card, subSet []Card) []Card {
	result := make([]Card, 0, len(fullSet))
	for _, card := range fullSet {
		if FindCardInSlice(card, subSet) == -1 {
			result = append(result, card)
		}
	}
	return result
}

//
func GetCombinationsForCards(cards []Card, k int) [][]Card {
	result := make([][]Card, 0)
	iCombs := GetCombinations(Range(len(cards)), k)
	for _, iComb := range iCombs {
		comb := make([]Card, k)
		for i, ci := range iComb {
			comb[i] = cards[ci]
		}
		result = append(result, comb)
	}
	return result
}

func GetCombinationsForCards2(cards []Card, k int, deadline time.Time) [][]Card {
	result := make([][]Card, 0)
	iCombs := GetCombinations2(Range(len(cards)), k, deadline)
	for _, iComb := range iCombs {
		comb := make([]Card, k)
		for i, ci := range iComb {
			comb[i] = cards[ci]
		}
		result = append(result, comb)
	}
	return result
}

// sort các hand theo thứ tự nhất định
type ByMinAh [][]Card

func (a ByMinAh) Len() int      { return len(a) }
func (a ByMinAh) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByMinAh) Less(i, j int) bool {
	return ToString(a[i]) >= ToString(a[j])
}

func SortedCardss(cardss [][]Card) [][]Card {
	result := make([][]Card, len(cardss))
	copy(result, cardss)
	sort.Sort(ByMinAh(result))
	return result
}
