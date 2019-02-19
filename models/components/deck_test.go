package components

import (
	"fmt"
	"github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (s *TestSuite) TearDownSuite(c *C) {
}

func (s *TestSuite) SetUpTest(c *C) {
	// Use s.dir to prepare some data.

	fmt.Printf("start test %s \n", c.TestName())
}

func (s *TestSuite) TearDownTest(c *C) {
}

/*



THE ACTUAL TESTS




*/

func (s *TestSuite) TestRunnable(c *C) {
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestDeck(c *C) {
	deck := NewCardGameDeck()
	drawCards := make([]string, 0)
	for i := 0; i < 52; i++ {
		card := deck.DrawRandomCard()
		c.Assert(card, NotNil)
		c.Assert(utils.ContainsByString(drawCards, card), Equals, false)
		drawCards = append(drawCards, card)
	}
	c.Assert(deck.DrawRandomCard(), Equals, "")

	deck = NewCardGameDeck()
	drawCards = make([]string, 0)
	for {
		cards := deck.DrawRandomCards(3)
		if len(cards) < 3 {
			c.Assert(len(cards), Equals, 1)
			break
		} else {
			c.Assert(len(cards), Equals, 3)
			for _, card := range cards {
				c.Assert(card, NotNil)
				c.Assert(utils.ContainsByString(drawCards, card), Equals, false)
				drawCards = append(drawCards, card)
			}
		}
	}
}

func (s *TestSuite) TestSpecialFunction(c *C) {
	cards := []string{"c j", "d k", "d q"}
	c.Assert(IsAllSpecialCards(cards), Equals, true)

	cards = []string{"c j", "h 6", "d 7"}
	c.Assert(IsAllSpecialCards(cards), Equals, false)

	cards = []string{"c j", "h 6", "d 7"}
	c.Assert(TotalValueOfCards(cards), Equals, 23)

	cards = []string{"c j", "h k", "d a"}
	c.Assert(TotalValueOfCards(cards), Equals, 21)

	cards = []string{"c 2", "d 7", "c q"}
	c.Assert(TotalValueOfCards(cards), Equals, 19)

	cards = []string{"c 1", "d 1"}
	c.Assert(IsCardValueEqual(cards[0], cards[1]), Equals, true)

	cards = []string{"c 1", "c 1"}
	c.Assert(IsCardValueEqual(cards[0], cards[1]), Equals, true)

	cards = []string{"c 1", "c j"}
	c.Assert(IsCardValueEqual(cards[0], cards[1]), Equals, false)

	cards = []string{"d j", "c j"}
	c.Assert(IsCardValueEqual(cards[0], cards[1]), Equals, true)
}

func (s *TestSuite) TestDrawCards(c *C) {
	cardsToDraw := []string{"c j", "d 4", "h 7"}
	deck := NewCardGameDeck()
	deck.DrawSpecificCards(cardsToDraw)

	c.Assert(deck.Contain("c j"), Equals, false)
	c.Assert(deck.Contain("d 4"), Equals, false)
	c.Assert(deck.Contain("h 7"), Equals, false)
	c.Assert(len(deck.cards), Equals, 49)
}

func (s *TestSuite) TestDrawTiles(c *C) {
	tilesToDraw := []string{"0 0", "1 1", "2 3"}
	deck := NewDominoesDeck()
	deck.DrawSpecificTiles(tilesToDraw)

	c.Assert(deck.Contain("0 0"), Equals, false)
	c.Assert(deck.Contain("1 1"), Equals, false)
	c.Assert(deck.Contain("2 3"), Equals, false)
	c.Assert(len(deck.tiles), Equals, 25)

	randomTile := deck.DrawRandomTile()
	c.Assert(deck.Contain(randomTile), Equals, false)
	c.Assert(len(deck.tiles), Equals, 24)

	// add back
	deck.PutTilesBack([]string{"0 0"})
	c.Assert(deck.Contain("0 0"), Equals, true)
	c.Assert(deck.Contain("1 1"), Equals, false)
	c.Assert(deck.Contain("2 3"), Equals, false)
	c.Assert(deck.Contain(randomTile), Equals, false)
	c.Assert(len(deck.tiles), Equals, 25)
}

func (s *TestSuite) TestSaveLoadDeck(c *C) {
	deck := NewCardGameDeck()
	data := deck.SerializedData()
	deck1 := NewCardGameDeckWithData(data)

	for cardString, boolValue := range deck.cards {
		c.Assert(deck1.cards[cardString], Equals, boolValue)
	}
}

/*
helper
*/
