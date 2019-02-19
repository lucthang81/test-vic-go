package components

import (
	"fmt"
	"sort"
)

func GetRandomSuits() []string {
	randomSuits := []string{"c", "s", "d", "h"}
	sort.Sort(ByRandom(randomSuits))
	return randomSuits
}
func GetRandomValues() []string {
	randomValues := []string{"a", "2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k"}
	sort.Sort(ByRandom(randomValues))
	return randomValues
}

func (deck *CardGameDeck) DrawCardsWithValue(value string, quantity int) []string {
	cards := make([]string, 0)
	randomSuits := GetRandomSuits()
	for _, suit := range randomSuits {
		card := fmt.Sprintf("%s %s", suit, value)
		available := deck.Contain(card)
		if available {
			cards = append(cards, card)
			if len(cards) == quantity {
				deck.DrawSpecificCards(cards)
				return cards
			}
		}
	}
	return cards
}

func (deck *CardGameDeck) DrawCardsWithSuit(suit string, quantity int) []string {
	cards := make([]string, 0)
	randomValues := GetRandomValues()
	for _, value := range randomValues {
		card := fmt.Sprintf("%s %s", suit, value)
		available := deck.Contain(card)
		if available {
			cards = append(cards, card)
			if len(cards) == quantity {
				deck.DrawSpecificCards(cards)
				return cards
			}
		}
	}
	return cards
}
