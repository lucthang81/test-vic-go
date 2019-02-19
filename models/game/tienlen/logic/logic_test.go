package logic

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter *datacenter.DataCenter
	dbName     string
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

func (s *TestSuite) TestAddRemoveFreeze(c *C) {

}

func (s *TestSuite) TestGameLogicIndo(c *C) {
	var logic TLLogic
	logic = NewIndoLogic()
	// contain cards
	c.Assert(ContainCards([]string{"c 1", "d 3", "d j"}, []string{"c 1", "d 3"}), Equals, true)
	c.Assert(ContainCards([]string{"c 1", "d 3", "d j"}, []string{"c 1"}), Equals, true)
	c.Assert(ContainCards([]string{"c 1", "d 3", "d j"}, []string{"c 1", "d 3", "c 10"}), Equals, false)
	c.Assert(ContainCards([]string{"c 1", "d 3", "d j", "h 7"}, []string{"c 1", "d 3"}), Equals, true)

	// value as int, suit as int
	c.Assert(valueAsInt(logic, "a"), Equals, 12)
	c.Assert(valueAsInt(logic, "2"), Equals, 0)
	c.Assert(valueAsInt(logic, "k"), Equals, 11)
	c.Assert(valueAsInt(logic, "3"), Equals, 1)
	c.Assert(valueAsInt(logic, "5"), Equals, 3)

	// var cardSuitOrder = []string{"d", "c", "h", "s"}
	c.Assert(suitAsInt(logic, "c"), Equals, 1)
	c.Assert(suitAsInt(logic, "s"), Equals, 3)
	c.Assert(suitAsInt(logic, "d"), Equals, 0)
	c.Assert(suitAsInt(logic, "h"), Equals, 2)

	// is in order
	c.Assert(isStreak(logic, sortCards(logic, []string{"c a", "c 2", "c 9"})), Equals, false)
	c.Assert(isStreak(logic, sortCards(logic, []string{"c a", "c 2", "d 3"})), Equals, true)
	c.Assert(isStreak(logic, sortCards(logic, []string{"c a", "c 2", "d 3", "c 4", "h 5"})), Equals, true)
	c.Assert(isStreak(logic, sortCards(logic, []string{"h a", "c j", "s q", "d k", "h 10"})), Equals, true)
	c.Assert(isStreak(logic, sortCards(logic, []string{"h a", "c j", "s q", "d k", "h 2"})), Equals, false)
	c.Assert(isStreak(logic, sortCards(logic, []string{"c 3", "c 7", "d 4", "h 5", "d 6"})), Equals, true)
	c.Assert(isStreak(logic, sortCards(logic, []string{"d 2", "c 3", "c 7", "d 4", "h 5", "d 6"})), Equals, true)

	// is couple card and in order
	c.Assert(isDupCardsInOrderAndIncreaseByOne(logic, sortCards(logic, []string{"c a", "c a", "c 9"})), Equals, false)
	c.Assert(isDupCardsInOrderAndIncreaseByOne(logic, sortCards(logic, []string{"c a", "c a", "d 3", "d 4"})), Equals, false)
	c.Assert(isDupCardsInOrderAndIncreaseByOne(logic, sortCards(logic, []string{"c 3", "c 3", "d 5", "h 5", "d 7", "h 7"})), Equals, false)
	c.Assert(isDupCardsInOrderAndIncreaseByOne(logic, sortCards(logic, []string{"c 3", "c 3", "d 5", "h 5", "d 4", "h 4"})), Equals, true)
	c.Assert(isDupCardsInOrderAndIncreaseByOne(logic, sortCards(logic, []string{"c 3", "c 3", "d 5", "h 5", "d 4", "h 4", "d 6", "h 6"})), Equals, true)

	logic = NewVNLogic()
	// test instant win
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c 2", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q", "c k"})), Equals, true)  // is 12 cards in order and increase by 1
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})), Equals, true) // is 6 dup cards group
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c 2", "c 2", "d a", "h k", "s k", "c q", "d q", "c j", "c j", "h 10", "d 10", "s q"})), Equals, true) // is 5 dup cards in order
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c 3", "c 3", "d 3", "h 4", "s 4", "c 4", "d a", "c a", "c 9", "h 9", "d 9", "s q"})), Equals, true)   // has 4 triple cards
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c 2", "c 2", "d 3", "h 4", "s 2", "c 2", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})), Equals, true)  // has 4 12 cards
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 3", "c 3", "c 7", "c 8", "s 9", "s 10", "s j", "s q"})), Equals, true)  // same color
	c.Assert(isInstantWin(logic, sortCards(logic, []string{"c 4", "c 5", "c 6", "s 7", "s 8", "s 9", "c 10", "c j", "c q", "c k", "s a", "h a", "s 2"})), Equals, false) // miss 12 streak

	c.Assert(isInstantWin(logic, sortCards(logic, []string{"d 2", "h 2"})), Equals, false)
}

func isInstantWin(logic TLLogic, cards []string) bool {
	return len(logic.GetInstantWinType(cards)) > 0
}

func (s *TestSuite) TestLoseMultiplier(c *C) {
	var logic TLLogic
	logic = NewIndoLogic()
	var cards []string
	multiplier, _ := logic.LoseMultiplierByCardLeft([]string{"c 3", "h 3", "c 6", "d 6", "c 7", "h 7", "s 8", "c 8"}, true)
	c.Assert(multiplier, Equals, float64(8))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 6", "c 7", "c 8", "s 9", "s 10", "s j", "s q"}) // full 13
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(39))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 6", "c 7", "c 8", "s 9", "s 10", "s j"}) // nothing
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(24))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 5", "c 7", "c 8", "s 9", "s 10", "s j"}) // 3 double card
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(24))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 5", "c 7", "c 8", "s 9", "s 10", "s j", " s q"}) // 3 double cards full 13
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(39))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 5", "c 6", "c 6", "s 9", "s 10", "s j", " s q"}) // 4 double cards full 13
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(39))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 5", "c 6", "c 6", "s 9", "s 10", "d 2", "h 2"}) // 4 double cards full 13, d 2 and h 2
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64((13+logic.D2LoseMultiplier()+logic.H2LoseMultiplier())*3))

	cards = sortCards(logic, []string{"c a", "c 3", "c 3", "s 4", "s 4", "s 5", "c 5", "c 6", "c 6", "s j", "c j", "d j", "h j"}) // 4 double cards full 13, 1 quadruple
	multiplier, _ = logic.LoseMultiplierByCardLeft(cards, true)
	c.Assert(multiplier, Equals, float64(39))
}

func (s *TestSuite) TestContainCards(c *C) {
	c.Assert(ContainCards([]string{"s 4", "c 5", "d 6", "c 8", "h 8", "d 9", "d 10", "h q", "h k"}, []string{"c 8", "d 9", "d 10"}), Equals, true)
}

func playCardsOverCards(logic TLLogic, cardsOnTable []string, cards []string) (isValidMove bool) {
	// TODO: write code for vn slash
	if logic.IsCards1BiggerThanCards2(cards, cardsOnTable) {
		return true
	} else {
		return false
	}
}

func (s *TestSuite) TestCompareCardsAndPlayCardsIndo(c *C) {
	var logic TLLogic
	logic = NewIndoLogic()
	// check type
	c.Assert(logic.GetMoveType([]string{"c 7", "s 8", "c 9", "h 10", "d j"}), Equals, MoveType5OrderCards)

	// compare
	c.Assert(logic.IsCards1BiggerThanCards2([]string{"d 3"}, []string{"c 3"}), Equals, false)
	c.Assert(logic.IsCards1BiggerThanCards2([]string{"s 2"}, []string{"s 3"}), Equals, true)
	c.Assert(logic.IsCards1BiggerThanCards2(sortCards(logic, []string{"c 3", "d 3"}), sortCards(logic, []string{"s 3", "h 3"})), Equals, false)
	// no 4 streak cards
	c.Assert(logic.IsCards1BiggerThanCards2(sortCards(logic, []string{"c j", "c q", "s k", "h a"}), sortCards(logic, []string{"c j", "c q", "s k", "c a"})), Equals, false)
	c.Assert(logic.IsCards1BiggerThanCards2(sortCards(logic, []string{"c 2", "h 2"}), sortCards(logic, []string{"s 2", "d 2"})), Equals, false)

	// play cards
	var isValid bool
	// remember that we have to sort this first
	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 2"}), sortCards(logic, []string{"c 3"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 3"}), sortCards(logic, []string{"c 2"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 3"}), sortCards(logic, []string{"s 4", "h 4"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 3", "d 3"}), sortCards(logic, []string{"c 4", "h 4"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 3", "d 3"}), sortCards(logic, []string{"c 2", "h 2"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 2", "d 2"}), sortCards(logic, []string{"c 5", "h 5"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 5", "d 5"}), sortCards(logic, []string{"c 4", "h 4"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 5", "d 6", "h 7"}), sortCards(logic, []string{"d 5", "h 6", "d 7"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 5", "d 6", "h 7"}), sortCards(logic, []string{"d 5", "d 6", "d 7"}))
	c.Assert(isValid, Equals, false) // indo can't play 3 streak cards

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"d 5", "d 6", "d 7"}), sortCards(logic, []string{"c 5", "d 6", "h 7"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"d 5", "d 6", "d 7"}), sortCards(logic, []string{"c 6", "d 7", "h 8"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"d 8", "s 9", "d 10"}), sortCards(logic, []string{"h 6", "h 7", "h 8"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 5", "d 6", "h 7"}), sortCards(logic, []string{"d 6", "h 7", "d 8"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 4", "d 4"}), sortCards(logic, []string{"c a", "h a", "s a", "d a"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "d 6"}), sortCards(logic, []string{"c j", "d j", "c q", "d q", "c k", "d k"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "d 6"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "h 6"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "h 6"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "d 6"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "h 7", "s 8", "d 8", "s 9", "d 9", "c 10", "s 10"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "d 6"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "s 6", "d 6", "c 7", "s 7"}), sortCards(logic, []string{"c 7", "h 7", "s 8", "d 8", "s 9", "d 9", "c 10", "s 10"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}), sortCards(logic, []string{"c 7", "s 8", "s 9", "c 10", "s j"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "s 8", "s 9", "c 10", "s j"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}), sortCards(logic, []string{"c 7", "d 7", "h 7"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "d 7", "h 7"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "c 5", "c 6", "c 9", "c 2"}), sortCards(logic, []string{"c 7", "s 8", "s 9", "c 10", "s j"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "s 8", "s 9", "c 10", "s j"}), sortCards(logic, []string{"c 4", "c 5", "c 6", "c 9", "c 2"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}), sortCards(logic, []string{"c 7", "c 8", "c 9", "c 10", "c j"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "c 8", "c 9", "c 10", "c j"}), sortCards(logic, []string{"c 4", "h 4", "s 5", "d 5", "h 5"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "s 7", "h 7", "c 2", "d 2"}), sortCards(logic, []string{"c 4", "h 4", "s 10", "d 10", "h 10"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "c 8", "c 9", "c 10", "c j"}), sortCards(logic, []string{"c 8", "c 9", "c 10", "c j", "c q"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 7", "s 7", "c 9", "h 9", "d 9"}), sortCards(logic, []string{"c 8", "h 8", "c 2", "h 2", "d 2"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 8", "h 8", "c 2", "h 2", "d 2"}), sortCards(logic, []string{"c 7", "s 7", "c 9", "h 9", "d 9"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 8", "h 8", "c q", "h q", "d q"}), sortCards(logic, []string{"c 7", "s 7", "c a", "h a", "d a"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c a", "h 2", "c 3", "h 4", "d 5"}), sortCards(logic, []string{"c 7", "s 8", "c 9", "h 10", "d j"}))
	c.Assert(isValid, Equals, true)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 10", "h j", "c q", "h k", "d a"}), sortCards(logic, []string{"c 7", "s 8", "c 9", "h 10", "d j"}))
	c.Assert(isValid, Equals, false)

	isValid = playCardsOverCards(logic, sortCards(logic, []string{"c 10", "h j", "c q", "h k", "d a"}), sortCards(logic, []string{"c a", "s 2", "c 3", "h 4", "d 5"}))
	c.Assert(isValid, Equals, false)
}
