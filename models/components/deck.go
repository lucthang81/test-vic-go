package components

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/utils"
)

type CardGameDeck struct {
	cards map[string]bool
}

func (deck *CardGameDeck) Cards() map[string]bool {
	return deck.cards
}

var suits = []string{"c", "d", "h", "s"}
var values = []string{"a", "2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k"}

type ByRandom []string

func (a ByRandom) Len() int      { return len(a) }
func (a ByRandom) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRandom) Less(i, j int) bool {
	return rand.Intn(2) == 0
}

func NewCardGameDeck() *CardGameDeck {
	deck := &CardGameDeck{
		cards: make(map[string]bool),
	}
	for _, suit := range suits {
		for _, value := range values {
			deck.cards[CardString(suit, value)] = true
		}
	}
	return deck
}

func NewBacayGameDeck() *CardGameDeck {
	deck := &CardGameDeck{
		cards: make(map[string]bool),
	}
	for _, suit := range suits {
		for _, value := range values {
			if !utils.ContainsByString([]string{"10", "j", "q", "k"}, value) {
				deck.cards[CardString(suit, value)] = true
			}
		}
	}
	return deck
}

func NewPokerSlotGameDeck() *CardGameDeck {
	deck := &CardGameDeck{
		cards: make(map[string]bool),
	}

	for _, suit := range suits {
		for _, value := range values {
			deck.cards[CardString(suit, value)] = true
		}
	}
	deck.cards["joker"] = true
	return deck
}

func NewCardGameDeckWithCards(cards []string) *CardGameDeck {
	deck := &CardGameDeck{
		cards: make(map[string]bool),
	}
	for _, cardString := range cards {
		deck.cards[cardString] = true
	}
	return deck
}

func NewCardGameDeckWithData(data map[string]interface{}) *CardGameDeck {
	deck := &CardGameDeck{
		cards: make(map[string]bool),
	}
	for cardString, cardStringBoolInterface := range data {
		deck.cards[cardString] = cardStringBoolInterface.(bool)
	}
	return deck
}

func (deck *CardGameDeck) DrawCard(card string) bool {
	if deck.cards[card] {
		delete(deck.cards, card)
		return true
	}
	return false
}

func (deck *CardGameDeck) DrawRandomCard() string {
	if deck.NumberOfCardsLeft() == 0 {
		return ""
	}

	cardsLeft := make([]string, 0)
	for cardString, stillThere := range deck.cards {
		if stillThere {
			cardsLeft = append(cardsLeft, cardString)
		}
	}
	sort.Sort(ByRandom(cardsLeft))
	card := cardsLeft[0]
	deck.DrawCard(card)
	return card
}

func (deck *CardGameDeck) DrawRandomCards(quantity int) []string {

	cardsLeft := make([]string, 0)
	for cardString, stillThere := range deck.cards {
		if stillThere {
			cardsLeft = append(cardsLeft, cardString)
		}
	}
	sort.Sort(ByRandom(cardsLeft))
	cards := make([]string, 0)
	for index, card := range cardsLeft {
		if index < quantity {
			cards = append(cards, card)
		}
	}

	deck.DrawSpecificCards(cards)
	return cards
}

func (deck *CardGameDeck) DrawSpecificCards(cards []string) {
	for _, removedCard := range cards {
		deck.DrawCard(removedCard)
	}
}

func (deck *CardGameDeck) PutCardBack(cards []string) {
	for _, removedCard := range cards {
		deck.cards[removedCard] = true
	}
}

func (deck *CardGameDeck) NumberOfCardsLeft() int {
	return len(deck.cards)
}

func (deck *CardGameDeck) Contain(cardString string) bool {
	return deck.cards[cardString]
}

func (deck *CardGameDeck) ContainCards(cards []string) bool {
	for _, card := range cards {
		if !deck.Contain(card) {
			return false
		}
	}
	return true
}

func (deck *CardGameDeck) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	for cardString, boolValue := range deck.cards {
		data[cardString] = boolValue
	}
	return data
}

func ContainCards(cardString string, cards []string) bool {
	for _, cardInCards := range cards {
		if cardInCards == cardString {
			return true
		}
	}
	return false
}

func SuitAndValueFromCard(card string) (suit string, value string) {
	if card == "joker" {
		return card, card
	}
	tokens := strings.Split(card, " ")
	if len(tokens) >= 2 {
		return tokens[0], tokens[1]
	} else {
		return "", ""
	}
}

func IsCardValueEqual(card1 string, card2 string) bool {
	_, value1 := SuitAndValueFromCard(card1)
	_, value2 := SuitAndValueFromCard(card2)
	return value1 == value2
}

func IsAllSpecialCards(cards []string) bool {
	for _, card := range cards {
		_, value := SuitAndValueFromCard(card)
		if !utils.ContainsByString([]string{"j", "q", "k"}, value) {
			return false
		}
	}
	return true
}

func TotalValueOfCards(cards []string) (totalValue int) {
	totalValue = 0
	for _, card := range cards {
		_, value := SuitAndValueFromCard(card)
		valueInt := 0
		for index, valueToCompare := range values {
			if valueToCompare == value {
				valueInt = utils.MinInt(index+1, 10)
				break
			}
		}
		totalValue = totalValue + valueInt
	}
	return totalValue
}

func CardString(suit string, value string) string {
	return fmt.Sprintf("%s %s", suit, value)
}

func ConvertOldStringsToMinahCards(ss []string) []z.Card {
	// oldRank "a", "2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k"
	// oldSuit "c", "d", "h", "s"
	// newRank "A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"
	// newSuit "s", "c", "d", "h"
	result := make([]z.Card, 0)
	for _, s := range ss {
		su, r := SuitAndValueFromCard(s)
		if r == "a" {
			r = "A"
		} else if r == "10" {
			r = "T"
		} else if r == "j" {
			r = "J"
		} else if r == "q" {
			r = "Q"
		} else if r == "k" {
			r = "K"
		}
		cardStr := fmt.Sprintf("%v%v", r, su)
		result = append(result, z.FNewCardFS(cardStr))
	}
	return result
}

func ConvertMinahCardsToOldStrings(cards []z.Card) []string {
	result := make([]string, 0)
	for _, card := range cards {
		r := card.Rank
		if r == "A" {
			r = "a"
		} else if r == "T" {
			r = "10"
		} else if r == "J" {
			r = "j"
		} else if r == "Q" {
			r = "q"
		} else if r == "K" {
			r = "k"
		}
		oldStr := fmt.Sprintf("%v %v", card.Suit, r)
		result = append(result, oldStr)
	}
	return result
}
