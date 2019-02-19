package maubinh

import (
	"errors"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/record"
	"reflect"
	"sync"

	// "github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/test"
	"github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

const waitTimeForTurn time.Duration = 100 * time.Millisecond

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter      *datacenter.DataCenter
	server          *TestServer
	dbName          string
	playerIdCounter int64
}

var _ = Suite(&TestSuite{
	dbName: "casino_maubinh_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../../sql/init_schema.sql", "../../../sql/test_schema/player_test.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	s.server = NewTestServer()
	RegisterDataCenter(s.dataCenter)
	currency.RegisterDataCenter(s.dataCenter)
	game.RegisterDataCenter(s.dataCenter)
	game.RegisterServer(s.server)
	record.RegisterDataCenter(s.dataCenter)
}

func (s *TestSuite) TearDownSuite(c *C) {
	s.dataCenter.Db().Close()
	test.DropTestDatabase(s.dbName)

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

func (s *TestSuite) TestLogic(c *C) {
	gameInstance := NewMauBinhGame(currency.Money)
	// c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "c 2", "c 4", "c 3", "c 5"})), Equals, TypeStraightFlushBottom)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"h 10", "h j", "h q", "h k", "h a"})), Equals, TypeStraightFlushTop)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"h 5", "h 6", "h 7", "h 8", "h 9"})), Equals, TypeStraightFlush)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "d a", "c 3", "c 4", "c 5"})), Equals, TypePair)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "d a", "c 3", "d 3", "h 3"})), Equals, TypeFullHouse)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "s 3", "c 3", "d 3", "h 3"})), Equals, TypeFourOfAKind)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "s a", "d a", "h a", "h 10"})), Equals, TypeFourAces)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"h 2", "h 5", "h 9", "h 10", "h k"})), Equals, TypeFlush)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "h 2", "h 3", "h 4", "h 5"})), Equals, TypeStraight)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "h 2", "h 3"})), Equals, TypeHighCard)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"c a", "h 2", "h 3"})), Equals, TypeHighCard)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"h q", "h k", "h a"})), Equals, TypeHighCard)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h j", "h q", "h k", "h a"})), Equals, TypeStraight)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h 10", "c 10", "h k", "h a"})), Equals, TypeThreeOfAKind)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h 10", "c 10", "d 10"})), Equals, TypeFourOfAKind)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h 10", "c 10"})), Equals, TypeThreeOfAKind)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h 10", "c j", "h j", "s a"})), Equals, TypeTwoPair)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 10", "h 10", "c j", "h q", "s k"})), Equals, TypePair)
	c.Assert(gameInstance.getTypeOfCards(gameInstance.sortCards([]string{"s 8", "h 10", "c j", "h q", "s k"})), Equals, TypeHighCard)

	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"})), Equals, WhiteWinTypeDragonStraight)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "d 3", "d 4", "d 5", "d 6", "d 7", "d 8", "d 9", "d 10", "d j", "d q", "d k"})), Equals, WhiteWinTypeDragonRollingStraight)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 2", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})), Equals, WhiteWinTypeSixPair)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c k", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})), Equals, WhiteWinTypeSixPair)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})), Equals, WhiteWinTypeFivePairOneThreeOfAKind)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c 9", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})), Equals, WhiteWinTypeFivePairOneThreeOfAKind)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"h 3", "h 3", "h 10", "h 10", "h 7", "c j", "c 3", "c 8", "c 8", "c 9", "s 9", "s j", "s 3"})), Equals, WhiteWinTypeThreeFlush)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"d 3", "d 3", "d 10", "s 10", "s 7", "s j", "s 3", "s 8", "h 8", "h 9", "h 9", "h j", "h 3"})), Equals, WhiteWinTypeThreeFlush)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "c 3", "d 3", "s 4", "s 5", "s 6", "d 7", "h j", "h q", "h k", "h a", "h 10"})), Equals, WhiteWinTypeThreeStraight)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"d 3", "d 4", "d 5", "h 3", "s 4", "s 5", "s 6", "s 7", "h 8", "h 9", "h 10", "h j", "h q"})), Equals, WhiteWinTypeThreeStraight)
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})), Equals, "")
	c.Assert(gameInstance.getTypeOfWhiteWin(gameInstance.sortCards([]string{"s 3", "h 3", "c 3", "s a", "c 2", "d 2", "c 6", "h 7", "h 8", "c 9", "c j", "c q", "c k"})), Equals, "")

	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c a", "d a", "h a"}, []string{"c 3", "d 4", "d 5"}, TopPart), Equals, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAces])
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c 5", "d 5", "h 5"}, []string{"c 3", "d 4", "d 5"}, TopPart), Equals, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAKind])
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c a", "d a", "h a", "c 3", "s 4"}, []string{"s 10", "h 10", "c 10", "h 10", "h q"}, MiddlePart), Equals, -gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeFourOfAKind])
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c a", "d a", "h a", "c 3", "s 4"}, []string{"s a", "h a", "c a", "h a", "h q"}, BottomPart), Equals, -gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeFourAces])
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c 2", "s 3", "c 3"}, []string{"s 2", "d 2", "c 4"}, TopPart), Equals, float64(1))
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c 2", "d 2", "h 2", "s 10", "c 10"}, []string{"s 4", "h 4", "s 5", "d 5", "c 5"}, BottomPart), Equals, float64(-1))
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c 2", "d 2", "h 2", "s 3", "c 3"}, []string{"s 4", "h 4", "s 5", "d 5", "c 5"}, BottomPart), Equals, float64(-1))
	c.Assert(gameInstance.getMultiplierBetweenCards([]string{"c 2", "d 2", "h 2", "s 3", "c 3"}, []string{"s 4", "h 4", "s 5", "d 5", "c 5"}, BottomPart), Equals, float64(-1))
	c.Assert(gameInstance.getMultiplierBetweenCards(gameInstance.sortCards([]string{"h 10", "c j", "c q", "c k", "c a"}), gameInstance.sortCards([]string{"c a", "s 2", "c 3", "c 4", "c 5"}), BottomPart), Equals, float64(1))
	c.Assert(gameInstance.getMultiplierBetweenCards(gameInstance.sortCards([]string{"s 3", "h 3", "d 10"}), gameInstance.sortCards([]string{"c 7", "h j", "h k"}), TopPart), Equals, float64(1))
	c.Assert(gameInstance.getMultiplierBetweenCards(gameInstance.sortCards([]string{"s 3", "h 3", "d 10", "h 10", "s 4"}), gameInstance.sortCards([]string{"c 3", "d 3", "c 10", "s 10", "c 7"}), TopPart), Equals, float64(-1))
	c.Assert(gameInstance.getMultiplierBetweenCards(gameInstance.sortCards([]string{"s 3", "h 3", "s 10"}), gameInstance.sortCards([]string{"c 3", "d 3", "c 7"}), TopPart), Equals, float64(1))
	c.Assert(gameInstance.getMultiplierBetweenCards(gameInstance.sortCards([]string{"s 3", "h 3", "s 10"}), gameInstance.sortCards([]string{"c 3", "d 3", "d 10"}), TopPart), Equals, float64(0))

	var cardsData map[string][]string
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = map[string][]string{
		TopPart:    []string{"c a", "d a", "h a"},
		MiddlePart: []string{"d 10", "h 10", "c 4", "s 5", "h 6"}, // cannot be smaller than toppart
		BottomPart: []string{"c 3", "d 2", "h 5", "c 6", "h 7"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = map[string][]string{
		TopPart:    []string{"c 4", "s 5", "h 6"},
		MiddlePart: []string{"d 10", "h 10", "c a", "s a", "h a"}, // cannot be bigger than bottom
		BottomPart: []string{"c 3", "d 2", "h 5", "c 6", "h 7"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = map[string][]string{
		TopPart:    []string{"c 4", "s 5", "h 6"},
		MiddlePart: []string{"c 3", "d 2", "h 5", "c 6", "h 7"},
		BottomPart: []string{"d 10", "h 10", "c a", "s a", "h a"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, true)

	cardsData = map[string][]string{
		TopPart:    []string{"c 4", "s 5", "h 6"},
		MiddlePart: []string{"c 3", "d 2", "h 5", "c 6", "h 7", "s 10"}, // too many cards
		BottomPart: []string{"d 10", "h 10", "c a", "s a", "h a"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = map[string][]string{
		TopPart:    []string{"c 4", "s 5", "h 6", "s 10"}, // too many cards
		MiddlePart: []string{"c 3", "d 2", "h 5", "c 6", "h 7"},
		BottomPart: []string{"d 10", "h 10", "c a", "s a", "h a"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = map[string][]string{
		TopPart:    []string{"c 3", "h 4", "s k"},
		MiddlePart: []string{"d 5", "h 8", "h 10", "d q", "h a"},
		BottomPart: []string{"c 6", "d 6", "s j", "d j", "h j"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, true)

	cardsData = map[string][]string{
		TopPart:    []string{"c 3", "h 4", "s k"}, // too many cards
		MiddlePart: []string{"d 5", "h 8", "h 10", "d q", "h a"},
		BottomPart: []string{"c 6", "d 6", "s j", "d j", "h j"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, true)

	cardsData = map[string][]string{
		TopPart:    []string{"c 3", "h 4", "s k"}, // too many cards
		MiddlePart: []string{"d a", "h 2", "h 3", "d 4", "h 5"},
		BottomPart: []string{"c 6", "d 7", "s 8", "d 9", "h 10"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, true)

	cardsData = map[string][]string{
		TopPart:    []string{"c 3", "h 4", "s k"}, // too many cards
		MiddlePart: []string{"c 6", "d 7", "s 8", "d 9", "h 10"},
		BottomPart: []string{"d a", "h 2", "h 3", "d 4", "h 5"},
	}
	c.Assert(gameInstance.isCardsDataValid(cardsData), Equals, false)

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"}), WhiteWinTypeDragonRollingStraight)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"d 2", "h 3", "d 4"})
	c.Assert(cardsData[MiddlePart], DeepEquals, []string{"h 5", "d 6", "d 7", "d 8", "h 9"})
	c.Assert(cardsData[BottomPart], DeepEquals, []string{"h 10", "h j", "d q", "d k", "d a"})

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"}), WhiteWinTypeSixPair)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"d 2", "h 3", "d 4"})
	c.Assert(cardsData[MiddlePart], DeepEquals, []string{"h 5", "d 6", "d 7", "d 8", "h 9"})
	c.Assert(cardsData[BottomPart], DeepEquals, []string{"h 10", "h j", "d q", "d k", "d a"})

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 2", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"}), WhiteWinTypeSixPair)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"h 2", "s 6", "c 6"})
	c.Assert(cardsData[MiddlePart], DeepEquals, []string{"c 7", "d 7", "c 9", "h 9", "c 10"})
	c.Assert(cardsData[BottomPart], DeepEquals, []string{"d 10", "s j", "d j", "c a", "c a"})

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"}), WhiteWinTypeFivePairOneThreeOfAKind)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"s 6", "c 6", "h 6"})
	c.Assert(cardsData[MiddlePart], DeepEquals, []string{"c 7", "d 7", "c 9", "h 9", "c 10"})
	c.Assert(cardsData[BottomPart], DeepEquals, []string{"d 10", "s j", "d j", "c a", "c a"})

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"h 3", "h 3", "h 10", "h 10", "h 7", "c j", "c 3", "c 8", "c 8", "c 9", "s 9", "s j", "s 3"}), WhiteWinTypeThreeFlush)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"s 3", "s 9", "s j"})
	if reflect.DeepEqual(cardsData[MiddlePart], []string{"c 3", "c 8", "c 8", "c 9", "c j"}) &&
		reflect.DeepEqual(cardsData[BottomPart], []string{"h 3", "h 3", "h 7", "h 10", "h 10"}) {
		c.Assert(true, Equals, true)
	} else if reflect.DeepEqual(cardsData[BottomPart], []string{"c 3", "c 8", "c 8", "c 9", "c j"}) &&
		reflect.DeepEqual(cardsData[MiddlePart], []string{"h 3", "h 3", "h 7", "h 10", "h 10"}) {
		c.Assert(true, Equals, true)
	} else {
		c.Assert(false, Equals, false)
	}

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d 3", "d 3", "d 10", "s 10", "s 7", "s j", "s 3", "s 8", "h 8", "h 9", "h 9", "h j", "h 3"}), WhiteWinTypeThreeFlush)
	fmt.Println(cardsData)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"d 3", "d 3", "d 10"})
	if reflect.DeepEqual(cardsData[MiddlePart], []string{"s 3", "s 7", "s 8", "s 10", "s j"}) &&
		reflect.DeepEqual(cardsData[BottomPart], []string{"h 3", "h 8", "h 9", "h 9", "h j"}) {
		c.Assert(true, Equals, true)
	} else if reflect.DeepEqual(cardsData[BottomPart], []string{"s 3", "s 7", "s 8", "s 10", "s j"}) &&
		reflect.DeepEqual(cardsData[MiddlePart], []string{"h 3", "h 8", "h 9", "h 9", "h j"}) {
		c.Assert(true, Equals, true)
	} else {
		c.Assert(false, Equals, false)
	}

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "c 3", "d 3", "s 4", "s 5", "s 6", "d 7", "h j", "h q", "h k", "h a", "h 10"}), WhiteWinTypeThreeStraight)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"h a", "d 2", "c 3"})
	if reflect.DeepEqual(cardsData[MiddlePart], []string{"d 3", "s 4", "s 5", "s 6", "d 7"}) &&
		reflect.DeepEqual(cardsData[BottomPart], []string{"h j", "h q", "h k", "d a", "h 10"}) {
		c.Assert(true, Equals, true)
	} else if reflect.DeepEqual(cardsData[BottomPart], []string{"d 3", "s 4", "s 5", "s 6", "d 7"}) &&
		reflect.DeepEqual(cardsData[MiddlePart], []string{"h j", "h q", "h k", "d a", "h 10"}) {
		c.Assert(true, Equals, true)
	} else {
		c.Assert(false, Equals, false)
	}

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d 3", "d 4", "d 5", "h 3", "s 4", "s 5", "s 6", "s 7", "h 8", "h 9", "h 10", "h j", "h q"}), WhiteWinTypeThreeStraight)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"h 3", "s 4", "s 5"})
	if reflect.DeepEqual(cardsData[MiddlePart], []string{"d 3", "d 4", "d 5", "s 6", "s 7"}) &&
		reflect.DeepEqual(cardsData[BottomPart], []string{"h 8", "h 9", "h 10", "h j", "h q"}) {
		c.Assert(true, Equals, true)
	} else if reflect.DeepEqual(cardsData[BottomPart], []string{"d 3", "d 4", "d 5", "s 6", "s 7"}) &&
		reflect.DeepEqual(cardsData[MiddlePart], []string{"h 8", "h 9", "h 10", "h j", "h q"}) {
		c.Assert(true, Equals, true)
	} else {
		c.Assert(false, Equals, false)
	}

	cardsData = gameInstance.organizeCardsForWhiteWin(gameInstance.sortCards([]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"}), WhiteWinTypeDragonStraight)
	fmt.Println(cardsData)
	c.Assert(cardsData[TopPart], DeepEquals, []string{"d 2", "h 3", "d 4"})
	c.Assert(cardsData[MiddlePart], DeepEquals, []string{"h 5", "d 6", "d 7", "d 8", "h 9"})
	c.Assert(cardsData[BottomPart], DeepEquals, []string{"h 10", "h j", "d q", "d k", "d a"})
}

func (s *TestSuite) TestWhiteWin(c *C) {
	currencyType := currency.Money

	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(10)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(10, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end

	cards[player1.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "c 7", "c 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon rolling win
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c a", "s 5", "c 5", "d 7", "c 10", "s 10", "h 9", "d 10", "c j", "c q", "d k"}) //
	cards[player3.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "h 6", "c 7", "c 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon win
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})   //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		response = s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
		cards := utils.GetStringSliceAtPath(response, "data/cards")
		if player.Id() == player1.Id() {
			// check card
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "c 3")

			cardsDataTop := utils.GetStringSliceAtPath(response, "data/cards_data/top")
			cardsDataMid := utils.GetStringSliceAtPath(response, "data/cards_data/middle")
			cardsDataBot := utils.GetStringSliceAtPath(response, "data/cards_data/bottom")

			c.Assert(cardsDataTop, DeepEquals, []string{"c 2", "c 3", "c 4"})
			c.Assert(cardsDataMid, DeepEquals, []string{"c 5", "c 6", "c 7", "c 8", "c 9"})
			c.Assert(cardsDataBot, DeepEquals, []string{"c 10", "c j", "c q", "c k", "c a"})
		} else if player.Id() == player2.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 3")
		} else if player.Id() == player3.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "c 3")

			cardsDataTop := utils.GetStringSliceAtPath(response, "data/cards_data/top")
			cardsDataMid := utils.GetStringSliceAtPath(response, "data/cards_data/middle")
			cardsDataBot := utils.GetStringSliceAtPath(response, "data/cards_data/bottom")
			fmt.Println(response)
			c.Assert(cardsDataTop, DeepEquals, []string{"c 2", "c 3", "c 4"})
			c.Assert(cardsDataMid, DeepEquals, []string{"c 5", "h 6", "c 7", "c 8", "c 9"})
			c.Assert(cardsDataBot, DeepEquals, []string{"c 10", "c j", "c q", "c k", "c a"})

		} else if player.Id() == player4.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "d 2")
		}
	}

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                 // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"}, // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err.Error(), Equals, "Bạn đã đạt mậu binh, không cần làm gì nữa cả")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "s 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err.Error(), Equals, "Bạn đã đạt mậu binh, không cần làm gì nữa cả")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c a"},                 // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"}, // two pair
		BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(maubinhSession.finished, Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
	}

	// cardsData2 = map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c a"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"},  // two pair
	// 	BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"}, // straight
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	changeData := make(map[int64]int64)
	// bottom part player4 > player2
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("bottom", changeData)

	// middle part player2 > player4
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	changeData[player4.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("middle", changeData)

	// top part player4 > player2
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	// then game end
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		player1WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonRollingStraight]
		player3WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonStraight]
		fmt.Println("whitewin", moneyAfterApplyMultiplier(bet, player1WinMultiplier),
			moneyAfterApplyMultiplier(bet, player3WinMultiplier))

		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			cards := utils.GetStringSliceAtPath(resultData, "cards")
			result := utils.GetStringAtPath(resultData, "result")
			change := utils.GetInt64AtPath(resultData, "change")
			aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
			aceChange := utils.GetInt64AtPath(resultData, "ace_change")
			c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "white_win")
			c.Assert(aceChange, Equals, int64(0))
			c.Assert(aceMultiplier, Equals, int64(0))
			if playerId == player1.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "c 2")
				c.Assert(cards[1], Equals, "c 3")
				c.Assert(result, Equals, "win")
				c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player1WinMultiplier), maubinhSession.betEntry)*3)
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))

			} else if playerId == player2.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "s 2")
				c.Assert(cards[1], Equals, "h 3")
				c.Assert(result, Equals, "lose")
				c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player1WinMultiplier)-
					moneyAfterApplyMultiplier(bet, player3WinMultiplier)+
					changeData[player2.id])
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
			} else if playerId == player3.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "c 2")
				c.Assert(cards[1], Equals, "c 3")
				c.Assert(result, Equals, "win")
				c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player3WinMultiplier), maubinhSession.betEntry)*2-
					moneyAfterApplyMultiplier(bet, player1WinMultiplier))
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
			} else if playerId == player4.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "s 2")
				c.Assert(cards[1], Equals, "d 2")
				c.Assert(result, Equals, "lose")
				c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player1WinMultiplier)-
					moneyAfterApplyMultiplier(bet, player3WinMultiplier)+
					changeData[player4.id])
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
			}
		}
	}
}

func (s *TestSuite) TestWhiteWinOrganizeCards(c *C) {
	currencyType := currency.Money

	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(10)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(10, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end

	cards[player1.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "c 7", "c 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon rolling win
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c a", "s 5", "c 5", "d 7", "c 10", "s 10", "h 9", "d 10", "c j", "c q", "d k"}) //
	cards[player3.Id()] = gameInstance.sortCards([]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"})   // dragon win
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})   //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		response = s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
		cards := utils.GetStringSliceAtPath(response, "data/cards")
		if player.Id() == player1.Id() {
			// check card
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "c 3")

			cardsDataTop := utils.GetStringSliceAtPath(response, "data/cards_data/top")
			cardsDataMid := utils.GetStringSliceAtPath(response, "data/cards_data/middle")
			cardsDataBot := utils.GetStringSliceAtPath(response, "data/cards_data/bottom")

			c.Assert(cardsDataTop, DeepEquals, []string{"c 2", "c 3", "c 4"})
			c.Assert(cardsDataMid, DeepEquals, []string{"c 5", "c 6", "c 7", "c 8", "c 9"})
			c.Assert(cardsDataBot, DeepEquals, []string{"c 10", "c j", "c q", "c k", "c a"})
		} else if player.Id() == player2.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 3")
		} else if player.Id() == player3.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "d 2")
			c.Assert(cards[1], Equals, "h 3")

			cardsDataTop := utils.GetStringSliceAtPath(response, "data/cards_data/top")
			cardsDataMid := utils.GetStringSliceAtPath(response, "data/cards_data/middle")
			cardsDataBot := utils.GetStringSliceAtPath(response, "data/cards_data/bottom")
			fmt.Println(response)
			c.Assert(cardsDataTop, DeepEquals, []string{"d 2", "h 3", "d 4"})
			c.Assert(cardsDataMid, DeepEquals, []string{"h 5", "d 6", "d 7", "d 8", "h 9"})
			c.Assert(cardsDataBot, DeepEquals, []string{"h 10", "h j", "d q", "d k", "d a"})

		} else if player.Id() == player4.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "d 2")
		}
	}

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                 // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"}, // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err.Error(), Equals, "Bạn đã đạt mậu binh, không cần làm gì nữa cả")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "s 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err.Error(), Equals, "Bạn đã đạt mậu binh, không cần làm gì nữa cả")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c a"},                 // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"}, // two pair
		BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(maubinhSession.finished, Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
	}

	// cardsData2 = map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c a"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"},  // two pair
	// 	BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"}, // straight
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	changeData := make(map[int64]int64)
	// bottom part player4 > player2
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("bottom", changeData)

	// middle part player2 > player4
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	changeData[player4.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("middle", changeData)

	// top part player4 > player2
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	// then game end
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		player1WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonRollingStraight]
		player3WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonStraight]
		fmt.Println("whitewin", moneyAfterApplyMultiplier(bet, player1WinMultiplier),
			moneyAfterApplyMultiplier(bet, player3WinMultiplier))

		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			cards := utils.GetStringSliceAtPath(resultData, "cards")
			result := utils.GetStringAtPath(resultData, "result")
			change := utils.GetInt64AtPath(resultData, "change")
			aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
			aceChange := utils.GetInt64AtPath(resultData, "ace_change")
			c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "white_win")
			c.Assert(aceChange, Equals, int64(0))
			c.Assert(aceMultiplier, Equals, int64(0))
			if playerId == player1.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "c 2")
				c.Assert(cards[1], Equals, "c 3")
				c.Assert(result, Equals, "win")
				c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player1WinMultiplier), maubinhSession.betEntry)*3)
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))

			} else if playerId == player2.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "s 2")
				c.Assert(cards[1], Equals, "h 3")
				c.Assert(result, Equals, "lose")
				c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player1WinMultiplier)-
					moneyAfterApplyMultiplier(bet, player3WinMultiplier)+
					changeData[player2.id])
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
			} else if playerId == player3.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "d 2")
				c.Assert(cards[1], Equals, "h 3")
				c.Assert(result, Equals, "win")
				c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player3WinMultiplier), maubinhSession.betEntry)*2-
					moneyAfterApplyMultiplier(bet, player1WinMultiplier))
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
			} else if playerId == player4.Id() {
				c.Assert(len(cards), Equals, 13)
				c.Assert(cards[0], Equals, "s 2")
				c.Assert(cards[1], Equals, "d 2")
				c.Assert(result, Equals, "lose")
				c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player1WinMultiplier)-
					moneyAfterApplyMultiplier(bet, player3WinMultiplier)+
					changeData[player4.id])
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
			}
		}
	}
}

func (s *TestSuite) TestWhiteWinAndEnd(c *C) {
	currencyType := currency.Money

	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(10)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(10, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      bet,
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c a", "s 5", "c 5", "d 7", "c 10", "s 10", "h 9", "d 10", "c j", "c q", "d k"}) //
	cards[player2.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "d 5", "d 6", "d 7", "d 10", "d j", "s q", "s k", "s a"}                           // 3 flush
	cards[player3.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "h 6", "c 7", "e 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon win
	cards[player4.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "c 7", "c 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon rolling win

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(maubinhSession.finished, Equals, true)

	fmt.Println(player1.GetMoney(currencyType), player2.GetMoney(currencyType), player3.GetMoney(currencyType), player4.GetMoney(currencyType))

	// then game end
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {

		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 4)

				player4WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonRollingStraight]
				player3WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonStraight]
				player2WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeThreeFlush]

				for _, resultData := range resultsData {
					fmt.Println(resultData)
					playerId := utils.GetInt64AtPath(resultData, "id")
					cards := utils.GetStringSliceAtPath(resultData, "cards")
					result := utils.GetStringAtPath(resultData, "result")
					change := utils.GetInt64AtPath(resultData, "change")
					aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					aceChange := utils.GetInt64AtPath(resultData, "ace_change")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "white_win")
					c.Assert(aceChange, Equals, int64(0))
					c.Assert(aceMultiplier, Equals, int64(0))
					if playerId == player4.Id() {
						c.Assert(result, Equals, "win")
						c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player4WinMultiplier), maubinhSession.betEntry)*3)
						c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
					} else if playerId == player3.Id() {
						c.Assert(result, Equals, "win")
						c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player3WinMultiplier), maubinhSession.betEntry)*2-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier))
						c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))

					} else if playerId == player2.Id() {
						c.Assert(result, Equals, "lose")
						c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player2WinMultiplier), maubinhSession.betEntry)-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier)-
							moneyAfterApplyMultiplier(bet, player3WinMultiplier))
						c.Assert(player4OldMoney+change, Equals, player2.GetMoney(currencyType))
					} else if playerId == player1.Id() {
						c.Assert(len(cards), Equals, 13)
						c.Assert(result, Equals, "lose")
						c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player2WinMultiplier)-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier)-
							moneyAfterApplyMultiplier(bet, player3WinMultiplier))
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
					}
				}

			}

		}
	}
}

func (s *TestSuite) TestWhiteWinSameTime(c *C) {
	currencyType := currency.Money

	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(10)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(10, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      bet,
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c a", "s 5", "c 5", "d 7", "c 10", "s 10", "h 9", "d 10", "c j", "c q", "d k"}) //
	cards[player2.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "d 5", "d 6", "d 7", "d 10", "d j", "s q", "s k", "s a"}                           // 3 flush
	cards[player3.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})   //
	cards[player4.Id()] = []string{"c 2", "c 3", "c 4", "c 5", "c 6", "c 7", "c 8", "c 9", "c 10", "c j", "c q", "c k", "c a"}                           // dragon rolling win

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn * 3)

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c a"},                 // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"}, // two pair
		BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(maubinhSession.finished, Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
	}

	// cardsData1 = map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c a"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 10", "s 10"},  // two pair
	// 	BottomPart: []string{"h 9", "d 10", "c j", "c q", "d k"}, // straight
	// }

	// cardsData3 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	changeData := make(map[int64]int64)
	// bottom part player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player3.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("bottom", changeData)

	// middle part player1 > player3
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("middle", changeData)

	// top part player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player3.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	fmt.Println(player1.GetMoney(currencyType), player2.GetMoney(currencyType), player3.GetMoney(currencyType), player4.GetMoney(currencyType))

	// then game end
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {

		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 4)

				player4WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeDragonRollingStraight]
				player2WinMultiplier := gameInstance.logicInstance.WhiteWinMultiplier()[WhiteWinTypeThreeFlush]

				for _, resultData := range resultsData {
					fmt.Println(resultData)
					playerId := utils.GetInt64AtPath(resultData, "id")
					cards := utils.GetStringSliceAtPath(resultData, "cards")
					result := utils.GetStringAtPath(resultData, "result")
					change := utils.GetInt64AtPath(resultData, "change")
					aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					aceChange := utils.GetInt64AtPath(resultData, "ace_change")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "white_win")
					c.Assert(aceChange, Equals, int64(0))
					c.Assert(aceMultiplier, Equals, int64(0))
					if playerId == player4.Id() {
						c.Assert(result, Equals, "win")
						c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player4WinMultiplier), maubinhSession.betEntry)*3)
						c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
					} else if playerId == player3.Id() {
						c.Assert(result, Equals, "lose")
						c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player2WinMultiplier)-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier)+
							changeData[player3.id])
						c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))

					} else if playerId == player2.Id() {
						c.Assert(result, Equals, "win")
						c.Assert(change, Equals, game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, player2WinMultiplier), maubinhSession.betEntry)*2-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier))
						c.Assert(player4OldMoney+change, Equals, player2.GetMoney(currencyType))
					} else if playerId == player1.Id() {
						c.Assert(len(cards), Equals, 13)
						c.Assert(result, Equals, "lose")
						c.Assert(change, Equals, -moneyAfterApplyMultiplier(bet, player2WinMultiplier)-
							moneyAfterApplyMultiplier(bet, player4WinMultiplier)+
							changeData[player1.id])
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
					}
				}

			}

		}
	}
}

func (s *TestSuite) TestOrganizedCards(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	testGame.turnTimeInSeconds = 3 * time.Second
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 3", "c 5", "d 5", "c 7", "c 8", "c 9", "d 9", "h 9", "c q", "c k", "c a"})   //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "s 9", "h 10", "c j", "c q", "c k", "c a"})  //
	cards[player3.Id()] = gameInstance.sortCards([]string{"s 2", "c 2", "d 2", "c 4", "c 5", "h 7", "c 7", "d 8", "c 10", "s 10", "h j", "h q", "c a"}) //
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})  //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)
	c.Assert(maubinhSession.isEveryoneFinishOrganizedCards(), Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
		cards := utils.GetStringSliceAtPath(response, "data/cards")
		if player.Id() == player1.Id() {
			// check card
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "s 3")
		} else if player.Id() == player2.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 3")
		} else if player.Id() == player3.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "c 2")
		} else if player.Id() == player4.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "d 2")
		}
	}

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			fmt.Println(playerData)
			c.Assert(utils.GetIntAtPath(playerData, "turn_time") <= 3, Equals, true) // just round this
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"},
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}

		if player.Id() == player1.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)

	err = testGame.StartOrganizeCardsAgain(maubinhSession, player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, false)
	c.Assert(len(maubinhSession.organizedCardsData[player1.Id()]), Equals, 3)
	c.Assert(maubinhSession.organizedCardsData[player1.Id()][TopPart][0], Equals, "c 2")

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 7", "c 3", "s 3"},               // pair, change c 7 to test
		MiddlePart: []string{"c 5", "d 5", "c 2", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)
	c.Assert(maubinhSession.organizedCardsData[player1.Id()][TopPart][2], Equals, "c 7")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "h 5", "d 7", "c 9", "s 9"},  // there is no h 5 in player2 cards
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	s.server.cleanupAllResponse()
	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "s 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player2.Id()), Equals, true)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		playerId := player.Id()
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player2.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}

		if playerId == player1.Id() || playerId == player2.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	// will not organize for player3
	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
		BottomPart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair, smaller than straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		playerId := player.Id()
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player2.Id() || playerId == player4.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}

		if playerId == player1.Id() || playerId == player2.Id() || playerId == player4.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	utils.DelayInDuration(3 * time.Second)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 4)

	// cardsData1 := map[string]interface{}{
	// 	TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
	// 	MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
	// 	BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	// }

	// cardsData2 := map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// player3 will not collapse because he did not have valid organize
	var collapseMultiplier float64
	collapseMultiplier = 1
	changeData := make(map[int64]int64)
	// player 3 collapse
	// bottom part player2 = player4 > player1 > player 3
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1*collapseMultiplier), maubinhSession.betEntry) - bet*2 // 2 from player3
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*collapseMultiplier), maubinhSession.betEntry)       // 2 from player3
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, collapseMultiplier*3)                                                     // collapse with 3 people
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*collapseMultiplier), maubinhSession.betEntry)       // 2 from player3
	fmt.Println("bottom", changeData)

	// middle part player2 = player4 > player1 > player 3
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1*collapseMultiplier), maubinhSession.betEntry) - bet*2 // 2 from player3
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*collapseMultiplier), maubinhSession.betEntry)       // 2 from player3
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, collapseMultiplier*3)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*collapseMultiplier), maubinhSession.betEntry) // 2 from player3
	fmt.Println("middle", changeData)

	// top part player1 > player4 > player2 > player3
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 2+1*collapseMultiplier), maubinhSession.betEntry)       // 2 from player3
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1*collapseMultiplier), maubinhSession.betEntry) - bet*2 // 2 from player3
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, 1*collapseMultiplier*3)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*collapseMultiplier), maubinhSession.betEntry) - bet // 2 from player3
	fmt.Println("top", changeData)

	fmt.Println(changeData)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		fmt.Println(resultsData)
		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			change := utils.GetInt64AtPath(resultData, "change")
			c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
			if playerId == player1.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, float64(-1))

			} else if playerId == player2.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, float64(0))
			} else if playerId == player3.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, -1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, -1*collapseMultiplier)
			} else if playerId == player4.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, 1*collapseMultiplier)
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, 1*collapseMultiplier)
			}
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)
}

func (s *TestSuite) TestOrganizedCardsWithCollapsing(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	testGame.turnTimeInSeconds = 3 * time.Second
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 3", "c 5", "d 5", "c 7", "c 8", "c 9", "d 9", "h 9", "c q", "c k", "c a"})   //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "s 9", "h 10", "c j", "c q", "c k", "c a"})  //
	cards[player3.Id()] = gameInstance.sortCards([]string{"s 2", "c 2", "d 2", "c 4", "c 5", "h 7", "c 7", "d 8", "c 10", "s 10", "h j", "h q", "c a"}) //
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "h 9", "h 10", "c j", "c q", "c k", "c a"})  //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)
	c.Assert(maubinhSession.isEveryoneFinishOrganizedCards(), Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
		cards := utils.GetStringSliceAtPath(response, "data/cards")
		if player.Id() == player1.Id() {
			// check card
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "s 3")
		} else if player.Id() == player2.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 3")
		} else if player.Id() == player3.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "c 2")
		} else if player.Id() == player4.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "d 2")
		}
	}

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			fmt.Println(playerData)
			c.Assert(utils.GetIntAtPath(playerData, "turn_time") <= 3, Equals, true) // just round this
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"},
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}

		if player.Id() == player1.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)

	err = testGame.StartOrganizeCardsAgain(maubinhSession, player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, false)
	c.Assert(len(maubinhSession.organizedCardsData[player1.Id()]), Equals, 3)
	c.Assert(maubinhSession.organizedCardsData[player1.Id()][TopPart][0], Equals, "c 2")

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 7", "c 3", "s 3"},               // pair, change c 7 to test
		MiddlePart: []string{"c 5", "d 5", "c 2", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player1.Id()), Equals, true)
	c.Assert(maubinhSession.organizedCardsData[player1.Id()][TopPart][2], Equals, "c 7")

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "h 5", "d 7", "c 9", "s 9"},  // there is no h 5 in player2 cards
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	s.server.cleanupAllResponse()
	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "s 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(maubinhSession.isPlayerFinishOrganizedCards(player2.Id()), Equals, true)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		playerId := player.Id()
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player2.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}

		if playerId == player1.Id() || playerId == player2.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	cardsData = map[string]interface{}{
		TopPart:    []string{"h 7", "c 5", "c 4"},                 // high card
		MiddlePart: []string{"c 7", "h j", "c 2", "s 2", "d 8"},   // pair
		BottomPart: []string{"c 10", "s 10", "d 2", "h q", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
		BottomPart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair, smaller than straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err.Error(), Equals, l.Get(l.M0022))

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "h 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(maubinhSession.finished, Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		playerId := player.Id()
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
		}

		if playerId == player1.Id() || playerId == player2.Id() || playerId == player4.Id() {
			c.Assert(len(utils.GetMapAtPath(response, "data/cards_data")), Equals, 3)
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 1)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 1)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 1)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 1)

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 4)

	// cardsData1 := map[string]interface{}{
	// 	TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
	// 	MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
	// 	BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	// }

	// cardsData2 := map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// cardsData3 = map[string]interface{}{
	// 	TopPart:    []string{"h 7", "c 5", "c 4"},                 // high card
	// 	MiddlePart: []string{"c 7", "h j", "c 2", "s 2", "d 8"},   // pair
	// 	BottomPart: []string{"c 10", "s 10", "d 2", "h q", "c a"}, // pair
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// player3 collapse with player4

	changeData := make(map[int64]int64)
	// bottom part player2 = player4 > player3 > player 1
	changeData[player1.Id()] += -bet * 3                                                                                                                                                             // 2 from player3
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 2), maubinhSession.betEntry)                                                                                       // 2 from player3
	changeData[player3.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, 1+gameInstance.logicInstance.CollapseMultiplier()*1) // collapse with 3 people
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry)                                     // 2 from player3
	fmt.Println("bottom", changeData)

	// middle part player2 = player4 > player1 > player 3
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry) - bet*2
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1), maubinhSession.betEntry)
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, 2+1*gameInstance.logicInstance.CollapseMultiplier())
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player1 > player4 > player3 > player2
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 3), maubinhSession.betEntry)
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 3)
	changeData[player3.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, 1+1*gameInstance.logicInstance.CollapseMultiplier())
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1+1*gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry) - bet
	fmt.Println("top", changeData)

	fmt.Println(changeData)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		fmt.Println(resultsData)
		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			change := utils.GetInt64AtPath(resultData, "change")
			c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
			if playerId == player1.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, float64(-1))

			} else if playerId == player2.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, float64(-1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, float64(0))
			} else if playerId == player3.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, -float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player4.Id())), Equals, -float64(1*gameInstance.logicInstance.CollapseMultiplier()))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, -float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, -float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player4.Id())), Equals, -float64(1*gameInstance.logicInstance.CollapseMultiplier()))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, -float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player4.Id())), Equals, -float64(1*gameInstance.logicInstance.CollapseMultiplier()))
			} else if playerId == player4.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player1.Id())), Equals, -float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player2.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", TopPart, player3.Id())), Equals, float64(1*gameInstance.logicInstance.CollapseMultiplier()))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player2.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", MiddlePart, player3.Id())), Equals, float64(1*gameInstance.logicInstance.CollapseMultiplier()))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player1.Id())), Equals, float64(1))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player2.Id())), Equals, float64(0))
				c.Assert(utils.GetFloat64AtPath(resultData, fmt.Sprintf("compare_data/%s/%d/multiplier", BottomPart, player3.Id())), Equals, float64(1*gameInstance.logicInstance.CollapseMultiplier()))
			}
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)
}

func (s *TestSuite) Test2Players(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney)

	moneyOnTable := testGame.MoneyOnTable(100, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 4", "c 4", "d 4", "c a", "c 8", "c 9", "d k", "h k", "c q", "d q", "h q"}) //
	cards[player2.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s k", "d 5", "c 5", "d 7", "c 8", "c 9", "d 9", "h 9", "c q", "c k", "s 5"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 4"},               // high card
		MiddlePart: []string{"c 4", "d 4", "c a", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "d q", "h q"}, // full house
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s k"},               // high card
		MiddlePart: []string{"c 5", "d 5", "d 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "s 5"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	utils.DelayInDuration(waitTimeForTurn)

	utils.DelayInDuration(3 * time.Second)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	changeData := make(map[int64]int64)
	// bottom part player1 > player2
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("bottom", changeData)

	// middle part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 2 aces
	acesMap[player1.Id()] = 1
	acesChangeMap[player1.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)

	// player2 0 aces
	acesMap[player2.Id()] = -1
	acesChangeMap[player2.Id()] = moneyAfterApplyMultiplier(bet, -1)

	for _, player := range []game.GamePlayer{player1, player2} {
		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {
				fmt.Println("in")
				c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
				playersData := utils.GetMapSliceAtPath(response, "data/players_data")
				// check no cards in player data
				for _, playerData := range playersData {
					c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
				}
				c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 2)

				for _, resultData := range resultsData {
					playerId := utils.GetInt64AtPath(resultData, "id")
					change := utils.GetInt64AtPath(resultData, "change")
					acesMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					acesChange := utils.GetInt64AtPath(resultData, "ace_change")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
					c.Assert(acesMultiplier, Equals, acesMap[playerId])
					c.Assert(acesChange, Equals, acesChangeMap[playerId])
					if playerId == player1.Id() {
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))

					} else if playerId == player2.Id() {
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
					}
				}
			}

		}

	}
}

func (s *TestSuite) Test2PlayersRecord(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()

	player1.playerType = "bot"
	player2.playerType = "normal"

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney)

	moneyOnTable := testGame.MoneyOnTable(100, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 4", "c 4", "d 4", "c a", "c 8", "c 9", "d k", "h k", "c q", "d q", "h q"}) //
	cards[player2.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s k", "d 5", "c 5", "d 7", "c 8", "c 9", "d 9", "h 9", "c q", "c k", "s 5"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 4"},               // high card
		MiddlePart: []string{"c 4", "d 4", "c a", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "d q", "h q"}, // full house
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s k"},               // high card
		MiddlePart: []string{"c 5", "d 5", "d 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "s 5"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	utils.DelayInDuration(waitTimeForTurn)

	utils.DelayInDuration(3 * time.Second)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	changeData := make(map[int64]int64)
	// bottom part player1 > player2
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("bottom", changeData)

	// middle part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 1)
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 2 aces
	acesMap[player1.Id()] = 1
	acesChangeMap[player1.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 1), maubinhSession.betEntry)

	// player2 0 aces
	acesMap[player2.Id()] = -1
	acesChangeMap[player2.Id()] = moneyAfterApplyMultiplier(bet, -1)

	fmt.Println("winlose", maubinhSession.win, maubinhSession.lose, maubinhSession.botWin, maubinhSession.botLose)
	for _, player := range []game.GamePlayer{player1, player2} {
		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {
				fmt.Println("in")
				c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
				playersData := utils.GetMapSliceAtPath(response, "data/players_data")
				// check no cards in player data
				for _, playerData := range playersData {
					c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
				}
				c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 2)

				for _, resultData := range resultsData {
					playerId := utils.GetInt64AtPath(resultData, "id")
					change := utils.GetInt64AtPath(resultData, "change")
					acesMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					acesChange := utils.GetInt64AtPath(resultData, "ace_change")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
					c.Assert(acesMultiplier, Equals, acesMap[playerId])
					c.Assert(acesChange, Equals, acesChangeMap[playerId])
					if playerId == player1.Id() {
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
						c.Assert(maubinhSession.botWin-maubinhSession.botLose, Equals, change)

					} else if playerId == player2.Id() {
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
						c.Assert(maubinhSession.win-maubinhSession.lose, Equals, change)
					}
				}
			}
		}
	}
}

func (s *TestSuite) Test2PlayersCollapse(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney)

	moneyOnTable := testGame.MoneyOnTable(100, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 4", "c 4", "d 4", "c a", "c 8", "c 9", "d k", "h k", "c q", "d q", "h a"}) //
	cards[player2.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s k", "c 5", "d 5", "d 7", "c 8", "c 9", "d k", "h k", "c q", "c k", "s q"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 4"},               // high card
		MiddlePart: []string{"c 4", "d 4", "c a", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "d q", "h a"}, // 2 pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s k"},               // high card
		MiddlePart: []string{"c 5", "d 5", "d 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "c k", "s q"}, // full house
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	utils.DelayInDuration(waitTimeForTurn)

	utils.DelayInDuration(3 * time.Second)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	// collapse
	changeData := make(map[int64]int64)
	// bottom part player1 < player2
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier())
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry)
	fmt.Println("bottom", changeData)

	// middle part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier())
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player2 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier())
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry)
	fmt.Println("top", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 2 aces
	acesMap[player1.Id()] = 2
	acesChangeMap[player1.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 2), maubinhSession.betEntry)

	// player2 0 aces
	acesMap[player2.Id()] = -2
	acesChangeMap[player2.Id()] = moneyAfterApplyMultiplier(bet, -2)

	for _, player := range []game.GamePlayer{player1, player2} {
		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {
				fmt.Println("in")
				c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
				playersData := utils.GetMapSliceAtPath(response, "data/players_data")
				// check no cards in player data
				for _, playerData := range playersData {
					c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
				}
				c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 2)
				fmt.Println(resultsData)

				for _, resultData := range resultsData {
					playerId := utils.GetInt64AtPath(resultData, "id")
					change := utils.GetInt64AtPath(resultData, "change")
					multiplier := utils.GetInt64AtPath(resultData, "multiplier")
					acesMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					acesChange := utils.GetInt64AtPath(resultData, "ace_change")
					isCollapse := utils.GetBoolAtPath(resultData, "is_collapsing")
					winCollapseAll := utils.GetBoolAtPath(resultData, "win_collapse_all")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
					c.Assert(acesMultiplier, Equals, acesMap[playerId])
					c.Assert(acesChange, Equals, acesChangeMap[playerId])

					if playerId == player1.Id() {
						c.Assert(multiplier, Equals, int64(-4))
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
						c.Assert(isCollapse, Equals, true)
						c.Assert(winCollapseAll, Equals, false)

					} else if playerId == player2.Id() {
						c.Assert(multiplier, Equals, int64(4))
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
						c.Assert(isCollapse, Equals, false)
						c.Assert(winCollapseAll, Equals, false)
					}
				}
			}

		}

	}
}

func (s *TestSuite) Test3PlayersCollapse(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney)

	moneyOnTable := testGame.MoneyOnTable(100, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 4", "c 4", "d 4", "c a", "c 8", "c 9", "d k", "h k", "c q", "d q", "h a"})   //
	cards[player2.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s k", "c 10", "d 10", "d 7", "c 8", "c 9", "d k", "h k", "c q", "c k", "s q"}) //
	cards[player3.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 5", "s 7", "d 7", "c a", "c 8", "c 9", "d 10", "h 10", "c q", "d q", "h a"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)

	var cardsData map[string]interface{}
	var err error
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 4"},               // high card
		MiddlePart: []string{"c 4", "d 4", "c a", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "d q", "h a"}, // 2 pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s k"},                 // high card
		MiddlePart: []string{"c 10", "d 10", "d 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d k", "h k", "c q", "c k", "s q"},   // full house
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 5"},                 // high card
		MiddlePart: []string{"d 7", "s 7", "c a", "c 8", "c 9"},   // pair
		BottomPart: []string{"d 10", "h 10", "c q", "d q", "h a"}, // 2 pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)

	utils.DelayInDuration(waitTimeForTurn)

	utils.DelayInDuration(3 * time.Second)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	// collapse
	changeData := make(map[int64]int64)
	// bottom part player3 < player1 < player2
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) + game.MoneyAfterTax(bet, maubinhSession.betEntry)
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry) * 2
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) - bet
	fmt.Println("bottom", changeData)

	// middle part player2 > player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) - bet
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry) * 2
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) + game.MoneyAfterTax(bet, maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player2 > player1 > player3
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) - bet
	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry) * 2
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) + game.MoneyAfterTax(bet, maubinhSession.betEntry)
	fmt.Println("top", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 2 aces
	acesMap[player1.Id()] = 2

	acesChangeMap[player1.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 2), maubinhSession.betEntry)

	// player2 0 aces
	acesMap[player2.Id()] = -4
	acesChangeMap[player2.Id()] = moneyAfterApplyMultiplier(bet, -4)

	// player3 0 aces
	acesMap[player3.Id()] = 2
	acesChangeMap[player3.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 2), maubinhSession.betEntry)

	for _, player := range []game.GamePlayer{player1, player2} {
		numResponse := s.server.getNumberOfResponse(player.Id())
		for i := 0; i < numResponse; i++ {
			response := s.server.getAndRemoveResponse(player.Id())
			method := utils.GetStringAtPath(response, "method")
			if method == "maubinh_finish_game_session" {
				fmt.Println("in")
				c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
				playersData := utils.GetMapSliceAtPath(response, "data/players_data")
				// check no cards in player data
				for _, playerData := range playersData {
					c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
				}
				c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

				resultsData := utils.GetMapSliceAtPath(response, "data/results")
				c.Assert(len(resultsData), Equals, 3)
				fmt.Println(resultsData)

				for _, resultData := range resultsData {
					playerId := utils.GetInt64AtPath(resultData, "id")
					fmt.Println(playerId)
					change := utils.GetInt64AtPath(resultData, "change")
					multiplier := utils.GetInt64AtPath(resultData, "multiplier")
					acesMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
					acesChange := utils.GetInt64AtPath(resultData, "ace_change")
					isCollapse := utils.GetBoolAtPath(resultData, "is_collapsing")
					winCollapseAll := utils.GetBoolAtPath(resultData, "win_collapse_all")
					c.Assert(utils.GetStringAtPath(resultData, "result_type"), Equals, "")
					c.Assert(acesMultiplier, Equals, acesMap[playerId])
					c.Assert(acesChange, Equals, acesChangeMap[playerId])

					if playerId == player1.Id() {
						c.Assert(multiplier, Equals, int64(-5))
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
						c.Assert(isCollapse, Equals, true)
						c.Assert(winCollapseAll, Equals, false)

					} else if playerId == player2.Id() {
						c.Assert(multiplier, Equals, int64(8))
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
						c.Assert(isCollapse, Equals, false)
						c.Assert(winCollapseAll, Equals, true)
					} else if playerId == player3.Id() {
						c.Assert(multiplier, Equals, int64(-3))
						c.Assert(change, Equals, changeData[playerId]+acesChange)
						c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
						c.Assert(isCollapse, Equals, true)
						c.Assert(winCollapseAll, Equals, false)
					}
				}
			}

		}

	}
}

func (s *TestSuite) TestOrganizedCardsAndEndWhenEveryoneFinish(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 3", "c 5", "d 5", "c 7", "c 8", "c 9", "d 9", "h 9", "c q", "c k", "c a"})   //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "d 9", "h 10", "c j", "c q", "c k", "c a"})  //
	cards[player3.Id()] = gameInstance.sortCards([]string{"s 2", "h 2", "c 3", "c 4", "c 5", "h 5", "c 7", "d 8", "c 10", "s 10", "h j", "h q", "c a"}) //
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 2", "s 2", "c 4", "s 5", "c 5", "d 7", "c 9", "d 9", "h 10", "c j", "c q", "c k", "c a"})  //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
		cards := utils.GetStringSliceAtPath(response, "data/cards")
		if player.Id() == player1.Id() {
			// check card
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "c 2")
			c.Assert(cards[1], Equals, "s 3")
		} else if player.Id() == player2.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 3")
		} else if player.Id() == player3.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "h 2")
		} else if player.Id() == player4.Id() {
			c.Assert(len(cards), Equals, 13)
			c.Assert(cards[0], Equals, "s 2")
			c.Assert(cards[1], Equals, "d 2")
		}
	}

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "d 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player2.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {

				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}
	}
	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "d 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player2.Id() || playerId == player4.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			} else {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			}
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	// player 1 want to organize again
	err = testGame.StartOrganizeCardsAgain(maubinhSession, player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() || playerId == player3.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			} else {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			}
		}
	}

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 7", "h j", "h q"},                 // high card
		MiddlePart: []string{"h 5", "d 8", "c 10", "s 10", "h 2"}, // 1 pair
		BottomPart: []string{"c a", "s 2", "c 3", "c 4", "c 5"},   // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			playerId := utils.GetInt64AtPath(playerData, "id")
			if playerId == player1.Id() {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, false)
			} else {
				c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
			}
		}
	}

	c.Assert(s.server.getNumberOfResponse(player1.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player2.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player3.Id()), Equals, 0)
	c.Assert(s.server.getNumberOfResponse(player4.Id()), Equals, 0)

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false) // player 1 is still organize cards

	// player 1 finish
	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
		BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_change_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(utils.GetBoolAtPath(playerData, "finish_organizing_cards"), Equals, true)
		}
	}

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 4)

	// cardsData1 := map[string]interface{}{
	// 	TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
	// 	MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
	// 	BottomPart: []string{"d 9", "h 9", "c q", "c k", "c a"}, // pair
	// }

	// cardsData2 := map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// cardsData3 := map[string]interface{}{
	// 	TopPart:    []string{"c 7", "h j", "h q"},                 // high card
	// 	MiddlePart: []string{"h 5", "d 8", "c 10", "s 10", "h 2"}, // 1 pair
	// 	BottomPart: []string{"c a", "s 2", "c 3", "c 4", "c 5"},   // straight
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 2", "s 2", "c 4"},                // pair
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// player3 collapse with player4

	changeData := make(map[int64]int64)
	// bottom part player2 = player4 > player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 3)
	changeData[player2.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 2
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1), maubinhSession.betEntry)
	fmt.Println("bottom1", changeData)

	// middle part player2 = player4 > player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 3)
	changeData[player2.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 2
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1), maubinhSession.betEntry)
	fmt.Println("middle", changeData)

	// top part player1 > player4 > player3 > player2
	changeData[player1.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 3
	changeData[player2.Id()] += -bet * 3
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1)
	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1), maubinhSession.betEntry) - bet
	fmt.Println("top1", changeData)

	fmt.Println(changeData)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			change := utils.GetInt64AtPath(resultData, "change")
			if playerId == player1.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))

			} else if playerId == player2.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
			} else if playerId == player3.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
			} else if playerId == player4.Id() {
				c.Assert(change, Equals, changeData[playerId])
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
			}
		}
	}
}

func (s *TestSuite) TestCountAces(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 3", "c 5", "d 5", "c 7", "c 8", "c 9", "h 10", "d 9", "c q", "c k", "s k"})  //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "d 9", "h 10", "c j", "c q", "c k", "c a"})  //
	cards[player3.Id()] = gameInstance.sortCards([]string{"s 2", "h 2", "c 3", "c 4", "c 5", "h 5", "c 7", "d 8", "c 10", "s 10", "h j", "h a", "c a"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},                // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"},  // pair
		BottomPart: []string{"d 9", "h 10", "c q", "c k", "s k"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "d 9"},  // two pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 7", "h j", "h a"},                 // high card
		MiddlePart: []string{"h 5", "d 8", "c 10", "s 10", "h 2"}, // 1 pair
		BottomPart: []string{"c a", "s 2", "c 3", "c 4", "c 5"},   // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 3)

	// cardsData1 := map[string]interface{}{
	// 	TopPart:    []string{"c 2", "c 3", "s 3"},               // pair
	// 	MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"}, // pair
	// 	BottomPart: []string{"d 9", "h 10", "c q", "c k", "s k"}, // pair
	// }

	// cardsData2 := map[string]interface{}{
	// 	TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
	// 	MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "c 9"},  // two pair
	// 	BottomPart: []string{"h 10", "c j", "c q", "c k", "c a"}, // straight
	// }

	// cardsData3 := map[string]interface{}{
	// 	TopPart:    []string{"c 7", "h j", "h a"},                 // high card
	// 	MiddlePart: []string{"h 5", "d 8", "c 10", "s 10", "h 2"}, // 1 pair
	// 	BottomPart: []string{"c a", "s 2", "c 3", "c 4", "c 5"},   // straight
	// }

	changeData := make(map[int64]int64)
	// bottom part player2 > player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 2)
	changeData[player2.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 2
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("bottom1", changeData)

	// middle part player2 > player3 > player1
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet, 2)
	changeData[player2.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 2
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("middle", changeData)

	// top part player1 > player3 > player2
	changeData[player1.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) * 2
	changeData[player2.Id()] += -bet * 2
	changeData[player3.Id()] += game.MoneyAfterTax(bet, maubinhSession.betEntry) - moneyAfterApplyMultiplier(bet, 1)
	fmt.Println("top1", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 0 aces
	acesMap[player1.Id()] = -3
	acesChangeMap[player1.Id()] = moneyAfterApplyMultiplier(bet, -3)

	// player2 1 aces
	acesMap[player2.Id()] = 0
	acesChangeMap[player2.Id()] = 0

	// player3 2 aces
	acesMap[player3.Id()] = 3
	acesChangeMap[player3.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 3), maubinhSession.betEntry)

	fmt.Println(changeData)
	for _, player := range []game.GamePlayer{player2, player3, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 3)

		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			change := utils.GetInt64AtPath(resultData, "change")
			aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
			aceChange := utils.GetInt64AtPath(resultData, "ace_change")
			c.Assert(acesMap[playerId], Equals, aceMultiplier)
			c.Assert(acesChangeMap[playerId], Equals, aceChange)
			if playerId == player1.Id() {
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
			} else if playerId == player2.Id() {
				// player2OldMoney += 1
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
			} else if playerId == player3.Id() {
				// player3OldMoney += 1
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
			}
		}
	}
}

func (s *TestSuite) TestNotCountAces(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "c 3", "s 3", "c 5", "d 5", "c 7", "c 8", "c 9", "h 10", "d 9", "c q", "c k", "s k"}) //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 2", "h 3", "c 4", "s 5", "c 5", "d 7", "c 9", "d 9", "h 10", "c j", "c q", "c k", "d k"}) //
	cards[player3.Id()] = gameInstance.sortCards([]string{"s 2", "h 2", "c 3", "c 4", "c 5", "h 5", "c 7", "d 8", "c 10", "d a", "c a", "h a", "s a"}) //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "c 3", "s 3"},                // pair
		MiddlePart: []string{"c 5", "d 5", "c 7", "c 8", "c 9"},  // pair
		BottomPart: []string{"d 9", "h 10", "c q", "c k", "s k"}, // pair
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 3", "c 4"},                // high cards
		MiddlePart: []string{"s 5", "c 5", "d 7", "c 9", "d k"},  // 1 pair
		BottomPart: []string{"h 10", "c j", "c q", "c k", "d 9"}, // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 2", "h 2", "c 4"},                // high card
		MiddlePart: []string{"c 5", "h 5", "c 7", "d 8", "c 10"}, // 1 pair
		BottomPart: []string{"c a", "s a", "d a", "h a", "c 3"},  // straight
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 3)

	for _, player := range []game.GamePlayer{player2, player3, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 3)

		for _, resultData := range resultsData {
			aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
			aceChange := utils.GetInt64AtPath(resultData, "ace_change")
			c.Assert(aceChange, Equals, int64(0))
			c.Assert(aceMultiplier, Equals, int64(0))
		}
	}
}

func (s *TestSuite) TestSpecialMultiplierWhenCompareDuringEndGame(c *C) {
	currencyType := currency.Money
	testGame := NewMauBinhGame(currencyType)
	gameInstance := testGame
	c.Assert(testGame.gameCode, Equals, "maubinh")
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	bet := int64(100)

	player1OldMoney := player1.GetMoney(currencyType)
	player2OldMoney := player2.GetMoney(currencyType)
	player3OldMoney := player3.GetMoney(currencyType)
	player4OldMoney := player4.GetMoney(currencyType)
	fmt.Println(player1OldMoney, player2OldMoney, player3OldMoney, player4OldMoney)

	moneyOnTable := testGame.MoneyOnTable(bet, 4, 4)
	fmt.Println("table", moneyOnTable)

	player1.setMoney(player1.GetMoney(currencyType)-moneyOnTable, currencyType)
	player2.setMoney(player2.GetMoney(currencyType)-moneyOnTable, currencyType)
	player3.setMoney(player3.GetMoney(currencyType)-moneyOnTable, currencyType)
	player4.setMoney(player4.GetMoney(currencyType)-moneyOnTable, currencyType)

	// player1OldExp := player1.exp
	// player2OldExp := player2.exp
	// player3OldExp := player3.exp
	// player4OldExp := player4.exp

	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(moneyOnTable)
	s.server.cleanupAllResponse()
	c.Assert(testFinishCallback.didFinish, Equals, false)

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)

	players := []game.GamePlayer{player1, player2, player3, player4}
	moneysOnTable := map[int64]int64{
		player1.Id(): bet,
		player2.Id(): bet,
		player3.Id(): bet,
		player4.Id(): bet,
	}
	for index, player := range players {
		playerData := &PlayerData{
			id:       player.Id(),
			order:    index,
			turnTime: 0,
			money:    player.GetMoney(currencyType),
			bet:      moneysOnTable[player.Id()],
		}
		playersData = append(playersData, playerData)
	}

	// hard code cards to test white win
	// this mean owner win and end
	cards[player1.Id()] = gameInstance.sortCards([]string{"c 2", "h 2", "s 2", "c 5", "d 5", "s 5", "h 5", "c 10", "h 9", "h 10", "h j", "h q", "h k"})  //
	cards[player2.Id()] = gameInstance.sortCards([]string{"s 3", "h 3", "d 10", "s 5", "s 6", "s 7", "s 8", "s 9", "c 10", "c j", "c q", "c k", "c a"})  //
	cards[player3.Id()] = gameInstance.sortCards([]string{"c 7", "h j", "h k", "h 5", "d 5", "c 10", "s 10", "h 10", "c a", "c 2", "c 3", "c 4", "c 5"}) //
	cards[player4.Id()] = gameInstance.sortCards([]string{"d 5", "h 8", "h q", "s a", "s 2", "s 3", "s 4", "s 5", "c 10", "c j", "c q", "c k", "c a"})   //

	maubinhSession := NewMauBinhSession(testGame, currencyType, player1, players)
	maubinhSession.playersData = playersData
	maubinhSession.cards = cards
	maubinhSession.betEntry = testGame.BetData().GetEntry(bet)
	maubinhSession.sessionCallback = testFinishCallback
	maubinhSession.start()

	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(testFinishCallback.didFinish, Equals, false)
	c.Assert(maubinhSession.finished, Equals, false)

	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_start_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)
	}

	var cardsData map[string]interface{}
	var err error

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 2", "h 2", "s 2"},                // three of a kind
		MiddlePart: []string{"c 5", "d 5", "s 5", "h 5", "c 10"}, // four of a kind
		BottomPart: []string{"h 9", "h 10", "h j", "h q", "h k"}, // straight flush
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player1, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"s 3", "h 3", "d 10"},               // pair
		MiddlePart: []string{"s 5", "s 6", "s 7", "s 8", "s 9"},  // straight flush
		BottomPart: []string{"c 10", "c j", "c q", "c k", "c a"}, // straight flush top
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player2, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"c 7", "h j", "h k"},                  // high card
		MiddlePart: []string{"h 5", "d 5", "c 10", "s 10", "h 10"}, // full house
		BottomPart: []string{"c a", "c 2", "c 3", "c 4", "c 5"},    // straight flush bottom
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player3, cardsData)
	c.Assert(err, IsNil)

	cardsData = map[string]interface{}{
		TopPart:    []string{"d 5", "h 8", "h q"},                // high cards
		MiddlePart: []string{"s a", "s 2", "s 3", "s 4", "s 5"},  // straight flush bottom
		BottomPart: []string{"c 10", "c j", "c q", "c k", "c a"}, // straight flush top
	}
	err = testGame.FinishOrganizeCards(maubinhSession, player4, cardsData)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	utils.DelayInDuration(waitTimeForTurn * 3)
	c.Assert(testFinishCallback.didFinish, Equals, true)
	c.Assert(maubinhSession.finished, Equals, true)

	session := maubinhSession
	c.Assert(len(session.organizedCardsData), Equals, 4)

	// cardsData1 := map[string]interface{}{
	// 	TopPart:    []string{"c 2", "c 2", "s 2"},                // three of a kind
	// 	MiddlePart: []string{"c 5", "d 5", "s 5", "h 5", "c 10"}, // four of a kind
	// 	BottomPart: []string{"d 9", "h 10", "h j", "h q", "h k"}, // straight flush
	// }

	// cardsData2 := map[string]interface{}{
	// 	TopPart:    []string{"s 3", "h 3", "c 10"},               // pair
	// 	MiddlePart: []string{"s 5", "s 6", "s 7", "s 8", "s 9"},  // straight flush
	// 	BottomPart: []string{"c 10", "c j", "c q", "c k", "c a"}, // straight flush top
	// }

	// cardsData3 := map[string]interface{}{
	// 	TopPart:    []string{"c 7", "h j", "h k"},                  // high card
	// 	MiddlePart: []string{"h 5", "d 5", "c 10", "s 10", "h 10"}, // full house
	// 	BottomPart: []string{"c a", "s 2", "c 3", "c 4", "c 5"},    // straight flush bottom
	// }

	// cardsData4 := map[string]interface{}{
	// 	TopPart:    []string{"d 5", "s 8", "c q"},                // high cards
	// 	MiddlePart: []string{"s a", "s 2", "s 3", "s 4", "s 5"},  // straight flush bottom
	// 	BottomPart: []string{"c 10", "c j", "c q", "c k", "c a"}, // straight flush top
	// }

	// player3 collapse with player2

	changeData := make(map[int64]int64)
	// bottom part player2 = player4  > player3 > player1, all special so x2
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet,
		(gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop]+ // player2
			gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop]+ // player4
			gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushBottom])* // player3
			gameInstance.logicInstance.SpecialCompareMultiplier()) // special x2

	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet,
		(gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop]*gameInstance.logicInstance.CollapseMultiplier()+ //player3
			gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop])* // player1
			gameInstance.logicInstance.SpecialCompareMultiplier()), // special x2
		maubinhSession.betEntry)

	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet,
		(gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop]*gameInstance.logicInstance.CollapseMultiplier()+ // player2,collapse
			gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop])* // player4
			gameInstance.logicInstance.SpecialCompareMultiplier()) + // special x2
		game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushBottom]*gameInstance.logicInstance.SpecialCompareMultiplier()), maubinhSession.betEntry) // player1

	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet,
		(gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop]+ //player3
			gameInstance.logicInstance.WinMultiplierForBottomPart()[TypeStraightFlushTop])* // player1
			gameInstance.logicInstance.SpecialCompareMultiplier()), // special x2
		maubinhSession.betEntry)
	fmt.Println("bottom", changeData)

	// middle part   player4 > player2 > player1 > player3
	changeData[player1.Id()] += -moneyAfterApplyMultiplier(bet,
		gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlush]+ // player2
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]) + // player4
		game.MoneyAfterTax(moneyAfterApplyMultiplier(bet,
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeFourOfAKind]), // player3
			maubinhSession.betEntry)

	changeData[player2.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet,
		gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlush]+ // player1
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlush]*gameInstance.logicInstance.CollapseMultiplier()), maubinhSession.betEntry) - // player3, collapse
		moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]*gameInstance.logicInstance.SpecialCompareMultiplier()) // player4

	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet,
		gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlush]*gameInstance.logicInstance.CollapseMultiplier()+ // player2
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]+ //player4
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeFourOfAKind]) //player1

	changeData[player4.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet,
		gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]*gameInstance.logicInstance.SpecialCompareMultiplier()+ // player2
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]+ //player1
			gameInstance.logicInstance.WinMultiplierForMiddlePart()[TypeStraightFlushBottom]), maubinhSession.betEntry) // player3
	fmt.Println("middle", changeData)

	// top part player1 > player2 > player3 > player4
	changeData[player1.Id()] += game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAKind]), maubinhSession.betEntry) * 3
	changeData[player2.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAKind]) + // player1
		game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()+1), maubinhSession.betEntry) // player3 and player 4
	changeData[player3.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAKind]) - // player1
		moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.CollapseMultiplier()) + // player2
		game.MoneyAfterTax(bet, maubinhSession.betEntry) // player4
	changeData[player4.Id()] += -moneyAfterApplyMultiplier(bet, gameInstance.logicInstance.WinMultiplierForTopPart()[TypeThreeOfAKind]) - bet*2
	fmt.Println("top", changeData)

	acesMap := make(map[int64]int64)
	acesChangeMap := make(map[int64]int64)

	// player1 0 aces
	acesMap[player1.Id()] = -4
	acesChangeMap[player1.Id()] = moneyAfterApplyMultiplier(bet, -4)

	// player2 1 aces
	acesMap[player2.Id()] = 0
	acesChangeMap[player2.Id()] = 0

	// player3 1 aces
	acesMap[player3.Id()] = 0
	acesChangeMap[player3.Id()] = 0

	// player4 2 aces
	acesMap[player4.Id()] = 4
	acesChangeMap[player4.Id()] = game.MoneyAfterTax(moneyAfterApplyMultiplier(bet, 4), maubinhSession.betEntry)

	fmt.Println(changeData)
	for _, player := range []game.GamePlayer{player2, player3, player4, player1} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "maubinh_finish_game_session")
		c.Assert(utils.GetStringAtPath(response, "data/game_code"), Equals, testGame.gameCode)
		playersData := utils.GetMapSliceAtPath(response, "data/players_data")
		// check no cards in player data
		for _, playerData := range playersData {
			c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
		}
		c.Assert(len(utils.GetMapSliceAtPath(response, "data/players_cards")), Equals, 0)

		resultsData := utils.GetMapSliceAtPath(response, "data/results")
		c.Assert(len(resultsData), Equals, 4)

		for _, resultData := range resultsData {
			playerId := utils.GetInt64AtPath(resultData, "id")
			change := utils.GetInt64AtPath(resultData, "change")
			aceMultiplier := utils.GetInt64AtPath(resultData, "ace_multiplier")
			aceChange := utils.GetInt64AtPath(resultData, "ace_change")
			c.Assert(acesMap[playerId], Equals, aceMultiplier)
			c.Assert(acesChangeMap[playerId], Equals, aceChange)
			if playerId == player1.Id() {
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player1OldMoney+change, Equals, player1.GetMoney(currencyType))
			} else if playerId == player2.Id() {
				// player2OldMoney += 1
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player2OldMoney+change, Equals, player2.GetMoney(currencyType))
			} else if playerId == player3.Id() {
				// player3OldMoney += 1
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player3OldMoney+change, Equals, player3.GetMoney(currencyType))
			} else if playerId == player4.Id() {
				c.Assert(change, Equals, changeData[playerId]+aceChange)
				c.Assert(player4OldMoney+change, Equals, player4.GetMoney(currencyType))
			}
		}
	}
}

func (s *TestSuite) TestCheatCode(c *C) {

}

/*
helper
*/

type TestFinishSessionCallback struct {
	server       *TestServer
	didFinish    bool
	players      []game.GamePlayer
	owner        game.GamePlayer
	currencyType string

	moneysOnTable map[int64]int64
}

func NewTestFinishCallback(server *TestServer, currencyType string, players []game.GamePlayer, owner game.GamePlayer) *TestFinishSessionCallback {
	return &TestFinishSessionCallback{
		didFinish:    false,
		server:       server,
		players:      players,
		owner:        owner,
		currencyType: currencyType,
	}
}

func (callback *TestFinishSessionCallback) setMoneysOnTable(money int64) {
	callback.moneysOnTable = make(map[int64]int64)
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.moneysOnTable[playerId] = money
	}
}

func (callback *TestFinishSessionCallback) DidStartGame(session game.GameSessionInterface) {
	method := fmt.Sprintf("%s_start_game_session", "maubinh")
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.server.SendRequest(method, session.SerializedDataForPlayer(session.GetPlayer(playerId)), playerId)
	}

}

func (callback *TestFinishSessionCallback) DidChangeGameState(session game.GameSessionInterface) {
	method := fmt.Sprintf("%s_change_game_session", "maubinh")
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.server.SendRequest(method, session.SerializedDataForPlayer(session.GetPlayer(playerId)), playerId)
	}
}

func (callback *TestFinishSessionCallback) DidEndGame(results map[string]interface{}, delaySeconds int) {
	method := fmt.Sprintf("%s_finish_game_session", "maubinh")
	callback.server.SendRequests(method, results, getIdFromPlayersMap(callback.players))

	callback.didFinish = true
	for _, player := range callback.players {
		player.IncreaseMoney(callback.moneysOnTable[player.Id()], callback.currencyType, true)
	}
	callback.moneysOnTable = make(map[int64]int64)
}

func (callback *TestFinishSessionCallback) MoneyDidChange(session game.GameSessionInterface, playerId int64, change int64, reason string, additionalData map[string]interface{}) {

}
func (callback *TestFinishSessionCallback) SendNotifyMoneyChange(playerId int64, change int64, reason string, additionalData map[string]interface{}) {
}
func (callback *TestFinishSessionCallback) SetSubLog(format string, a ...interface{}) {

}

func (callback *TestFinishSessionCallback) SendMessageToPlayer(session game.GameSessionInterface, playerId int64, method string, data map[string]interface{}) {

}

func (callback *TestFinishSessionCallback) Owner() game.GamePlayer {
	return callback.owner
}

func (callback *TestFinishSessionCallback) IncreaseMoney(playerInstance game.GamePlayer, amount int64, shouldBlock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}

	// normal card game, money add decrease will first use money on table
	// increase
	callback.moneysOnTable[playerInstance.Id()] += amount
	return nil
}

func (callback *TestFinishSessionCallback) SetMoneyOnTable(playerId int64, value int64, shouldLock bool) (err error) {
	callback.moneysOnTable[playerId] = value
	return nil
}

func (callback *TestFinishSessionCallback) IncreaseAndFreezeThoseMoney(playerInstance game.GamePlayer, amount int64, shouldLock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}
	callback.moneysOnTable[playerInstance.Id()] += amount
	return err
}

func (callback *TestFinishSessionCallback) DecreaseMoney(playerInstance game.GamePlayer, amount int64, shouldBlock bool) (err error) {
	if amount < 0 {
		return errors.New("err:decrease_negative")
	}
	// decrease
	callback.moneysOnTable[playerInstance.Id()] -= amount
	if callback.moneysOnTable[playerInstance.Id()] < 0 {
		decreaseFromPlayer := -callback.moneysOnTable[playerInstance.Id()]
		callback.moneysOnTable[playerInstance.Id()] = 0
		_, err = playerInstance.DecreaseMoney(decreaseFromPlayer, callback.currencyType, shouldBlock)
		return err
	}
	return nil
}

func (callback *TestFinishSessionCallback) MoveMoneyFromPlayerToTable(playerInstance game.GamePlayer, amount int64, shouldNotify bool, shouldBlock bool) {
	if amount > 0 {
		callback.moneysOnTable[playerInstance.Id()] += amount
	}
}

func (callback *TestFinishSessionCallback) MoveMoneyFromTableToPlayer(playerInstance game.GamePlayer, amount int64, shouldNotify bool, shouldBlock bool) {
	if amount > 0 {
		callback.moneysOnTable[playerInstance.Id()] -= amount
		if callback.moneysOnTable[playerInstance.Id()] < 0 {
			playerInstance.DecreaseMoney(-callback.moneysOnTable[playerInstance.Id()], callback.currencyType, shouldBlock)
			callback.moneysOnTable[playerInstance.Id()] = 0
		}
	}
}

func (callback *TestFinishSessionCallback) GetTotalPlayerMoney(playerId int64) int64 {
	playerInstance := callback.getPlayer(playerId)
	if playerInstance == nil {
		return 0
	}
	return playerInstance.GetMoney(callback.currencyType) + callback.moneysOnTable[playerId]
}

func (callback *TestFinishSessionCallback) AssignOwner(playerInstance game.GamePlayer) error {
	return nil
}

func (callback *TestFinishSessionCallback) RemoveOwner() error {
	return nil
}

func (callback *TestFinishSessionCallback) GetMoneyOnTable(playerId int64) int64 {
	return callback.moneysOnTable[playerId]
}

func (callback *TestFinishSessionCallback) getPlayer(playerId int64) (player game.GamePlayer) {
	for _, player := range callback.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (callback *TestFinishSessionCallback) GetPlayersDataForDisplay(currentPlayer game.GamePlayer) (playersData map[string]map[string]interface{}) {
	return nil
}

func (callback *TestFinishSessionCallback) GetPlayerAtIndex(index int) game.GamePlayer {
	for indexInList, player := range callback.players {
		if indexInList == index {
			return player
		}
	}
	return nil
}

type TestPlayer struct {
	currencyGroup *currency.CurrencyGroup
	id            int64
	name          string
	room          *game.Room
	isOnline      bool
	exp           int64

	recentResult   string
	recentChange   int64
	recentGameCode string

	gameCount  int
	bet        int64
	playerType string

	moneyMutex sync.Mutex
}

func (s *TestSuite) newPlayer() *TestPlayer {
	username := utils.RandSeq(20)
	row := s.dataCenter.Db().QueryRow("INSERT INTO player (username, player_type,identifier) VALUES ($1,$2,$3) RETURNING id",
		username, "", utils.RandSeq(15))
	var id int64
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("err create new player", err)
	}
	testPlayer := &TestPlayer{
		currencyGroup: currency.NewCurrencyGroup(id),
		id:            id,
		name:          username,
		playerType:    "bot",
		isOnline:      true,
	}
	testPlayer.IncreaseMoney(100000, currency.Money, true)
	testPlayer.IncreaseMoney(100000, currency.TestMoney, true)
	return testPlayer
}

func (player *TestPlayer) LockMoney(currencyType string) {
	player.currencyGroup.Lock(currencyType)
}

func (player *TestPlayer) UnlockMoney(currencyType string) {
	player.currencyGroup.Unlock(currencyType)
}

func (player *TestPlayer) IncreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	return player.currencyGroup.IncreaseMoney(money, currencyType, shouldLock)
}

func (player *TestPlayer) DecreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	return player.currencyGroup.DecreaseMoney(money, currencyType, shouldLock)
}

func (player *TestPlayer) FreezeMoney(money int64, currencyType string, reasonString string, shouldLock bool) (err error) {
	return player.currencyGroup.FreezeValue(currencyType, reasonString, money, shouldLock)
}

func (player *TestPlayer) IncreaseFreezeMoney(increaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	return player.currencyGroup.IncreaseFreezeValue(increaseAmount, currencyType, reasonString, shouldLock)
}

func (player *TestPlayer) DecreaseFromFreezeMoney(decreaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	return player.currencyGroup.DecreaseFromFreezeValue(decreaseAmount, currencyType, reasonString, shouldLock)
}

func (player *TestPlayer) GetFreezeValue(currencyType string) int64 {
	return player.currencyGroup.TotalFreezeValue(currencyType)
}

func (player *TestPlayer) GetFreezeValueForReason(currencyType string, reasonString string) int64 {
	return player.currencyGroup.GetFreezeValue(currencyType, reasonString)
}

func (player *TestPlayer) GetAvailableMoney(currencyType string) int64 {
	return player.currencyGroup.TotalAvailableValue(currencyType)
}

func (player *TestPlayer) GetMoney(currencyType string) int64 {
	return player.currencyGroup.GetValue(currencyType)
}

func (player *TestPlayer) setMoney(money int64, currencyType string) {
	amount := money - player.GetMoney(currencyType)
	if amount > 0 {
		player.currencyGroup.IncreaseMoney(amount, currencyType, true)
	} else {
		player.currencyGroup.DecreaseMoney(-amount, currencyType, true)
	}
}
func (player *TestPlayer) Id() int64 {
	return player.id
}
func (player *TestPlayer) Name() string {
	return player.name
}

func (player *TestPlayer) Room() *game.Room {
	return player.room
}

func (player *TestPlayer) SetRoom(room *game.Room) {
	player.room = room
}

func (player *TestPlayer) IpAddress() string {
	return ""
}

func (player *TestPlayer) IsOnline() bool {
	return player.isOnline
}

func (player *TestPlayer) SetIsOnline(isOnline bool) {
	player.isOnline = isOnline
}
func (player *TestPlayer) IncreaseBet(bet int64) {
	player.bet += bet
}

func (player *TestPlayer) SerializedData() map[string]interface{} {
	return map[string]interface{}{
		"id": player.id,
	}
}

func (player *TestPlayer) SerializedDataMinimal() map[string]interface{} {
	return map[string]interface{}{
		"id": player.id,
	}
}

func (player *TestPlayer) IncreaseExp(exp int64) (newExp int64, err error) {
	player.exp = player.exp + exp
	return player.exp, nil
}

func (player *TestPlayer) RecordGameResult(gameCode string, result string, change int64, currencyType string) (err error) {
	player.recentResult = result
	player.recentChange = change
	player.recentGameCode = gameCode
	player.gameCount += 1
	return nil
}

func (player *TestPlayer) IncreaseVipPointForMatch(bet int64, matchId int64, gameCode string) {
}

func (player *TestPlayer) PlayerType() string {
	return player.playerType
}

type TestServer struct {
	receiveDataMap map[int64][]map[string]interface{}
}

func NewTestServer() *TestServer {
	server := &TestServer{
		receiveDataMap: make(map[int64][]map[string]interface{}),
	}
	return server
}

func (server *TestServer) getNumberOfResponse(playerId int64) int {
	return len(server.receiveDataMap[playerId])
}

func (server *TestServer) getAndRemoveResponse(toPlayerId int64) map[string]interface{} {
	fullData := server.receiveDataMap[toPlayerId][0]
	server.receiveDataMap[toPlayerId] = server.receiveDataMap[toPlayerId][1:]
	return fullData
}

func (server *TestServer) cleanupAllResponse() {
	server.receiveDataMap = make(map[int64][]map[string]interface{})
}

func (server *TestServer) SendRequest(requestType string, data map[string]interface{}, toPlayerId int64) {
	responseList := server.receiveDataMap[toPlayerId]
	if responseList == nil {
		responseList = make([]map[string]interface{}, 0)
	}
	fullData := make(map[string]interface{})
	fullData["method"] = requestType
	fullData["data"] = data
	server.receiveDataMap[toPlayerId] = append(responseList, utils.ConvertData(fullData))
}

func (server *TestServer) SendHotFixRequest(requestType string, data map[string]interface{}, currencyType string, toPlayerId int64) {
	responseList := server.receiveDataMap[toPlayerId]
	if responseList == nil {
		responseList = make([]map[string]interface{}, 0)
	}
	fullData := make(map[string]interface{})
	fullData["method"] = requestType
	fullData["data"] = data
	server.receiveDataMap[toPlayerId] = append(responseList, utils.ConvertData(fullData))
}
func (server *TestServer) SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64) {
	for _, toPlayerId := range toPlayerIds {
		responseList := server.receiveDataMap[toPlayerId]
		if responseList == nil {
			responseList = make([]map[string]interface{}, 0)
		}
		fullData := make(map[string]interface{})
		fullData["method"] = requestType
		fullData["data"] = data
		server.receiveDataMap[toPlayerId] = append(responseList, utils.ConvertData(fullData))
	}
}

func (server *TestServer) SendRequestsToAll(requestType string, data map[string]interface{}) {

}

func (s *TestSuite) genPlayerId() int64 {
	s.playerIdCounter = s.playerIdCounter + 1
	return s.playerIdCounter
}

// func (s *TestSuite) justCreateRoom(game GameInterface) {
// 	player := &TestPlayer{
// 		id:    s.genPlayerId(),
// 		name:  "test2",
// 		money: 10000,
// 	}
// 	CreateRoom(game, player, 10, 4)
// }
