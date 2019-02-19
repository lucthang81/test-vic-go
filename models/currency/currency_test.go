package currency

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/test"
	"github.com/vic/vic_go/utils"
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

var _ = Suite(&TestSuite{
	dbName: "casino_currency_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	feature.UnlockAllFeature()
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../sql/init_schema.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	s.dataCenter.FlushCache()
	RegisterDataCenter(s.dataCenter)
}

func (s *TestSuite) TearDownSuite(c *C) {
	test.DropTestDatabase(s.dbName)
}

func (s *TestSuite) SetUpTest(c *C) {
	// Use s.dir to prepare some data.

	fmt.Printf("start test %s \n", c.TestName())
}

func (s *TestSuite) TearDownTest(c *C) {
	s.dataCenter.FlushCache()
}

/*



THE ACTUAL TESTS



*/

func (s *TestSuite) TestRunnable(c *C) {
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestAddRemoveFreeze(c *C) {
	playerId1 := s.newPlayer()
	currencyGroup := NewCurrencyGroup(playerId1)
	c.Assert(currencyGroup, NotNil)

	currencyType := Money

	currencyGroup.IncreaseMoney(10000, currencyType, true)
	c.Assert(currencyGroup.GetValue(currencyType), Equals, int64(10000))

	currencyGroup.DecreaseMoney(2000, currencyType, true)
	c.Assert(currencyGroup.GetValue(currencyType), Equals, int64(8000))

	currencyGroup.DecreaseMoney(9000, currencyType, true)
	c.Assert(currencyGroup.GetValue(currencyType), Equals, int64(0))

	currencyGroup.IncreaseMoney(10000, currencyType, true)
	c.Assert(currencyGroup.GetValue(currencyType), Equals, int64(10000))

	err := currencyGroup.FreezeValue(currencyType, "test1", 20000, true)
	c.Assert(err, NotNil)

	err = currencyGroup.FreezeValue(currencyType, "test1", 8000, true)
	c.Assert(err, IsNil)

	// now: money 8000, freeze 8000

	currencyGroup.DecreaseMoney(3000, currencyType, true)
	c.Assert(currencyGroup.GetValue(currencyType), Equals, int64(8000))
	c.Assert(currencyGroup.GetFreezeValue(currencyType, "test1"), Equals, int64(8000))
	c.Assert(currencyGroup.GetFreezeValue(currencyType, "test2"), Equals, int64(0))

	// now: money 8000, freeze 8000

	err = currencyGroup.FreezeValue(currencyType, "test3", 2000, true)
	c.Assert(err, NotNil)

	// now: money 8000, freeze 8000

	currencyGroup.IncreaseMoney(15000, currencyType, true)

	// now: money 23000, freeze 8000
	err = currencyGroup.FreezeValue(currencyType, "test3", 2000, true)
	c.Assert(err, IsNil)
	// now: money 23000, freeze 8000+2000=10000
	c.Assert(currencyGroup.TotalFreezeValue(currencyType), Equals, int64(10000))
	c.Assert(currencyGroup.TotalAvailableValue(currencyType), Equals, int64(13000))

	currencyGroup.DecreaseMoney(3000, currencyType, true)
	// now: money 20000, freeze 8000+2000=10000
	c.Assert(currencyGroup.TotalFreezeValue(currencyType), Equals, int64(10000))
	c.Assert(currencyGroup.TotalAvailableValue(currencyType), Equals, int64(10000))

	currencyGroup.IncreaseFreezeValue(2000, currencyType, "test1", true)
	// now: money 20000, freeze 10000+2000=12000
	c.Assert(currencyGroup.TotalFreezeValue(currencyType), Equals, int64(12000))
	c.Assert(currencyGroup.TotalAvailableValue(currencyType), Equals, int64(8000))

	currencyGroup.DecreaseFromFreezeValue(5000, currencyType, "test1", true)
	// now: money 15000, freeze 5000+2000=7000
	c.Assert(currencyGroup.TotalFreezeValue(currencyType), Equals, int64(7000))
	c.Assert(currencyGroup.TotalAvailableValue(currencyType), Equals, int64(8000))

	// unfree
	currencyGroup.FreezeValue(currencyType, "test1", 0, true)
	// now: money 15000, freeze 2000=2000
	c.Assert(currencyGroup.TotalFreezeValue(currencyType), Equals, int64(2000))
	c.Assert(currencyGroup.TotalAvailableValue(currencyType), Equals, int64(13000))
}

/*
helper
*/

func (s *TestSuite) newPlayer() int64 {
	username := utils.RandSeq(20)
	row := s.dataCenter.Db().QueryRow("INSERT INTO player (username, player_type,identifier) VALUES ($1,$2,$3) RETURNING id",
		username, "", utils.RandSeq(15))
	var id int64
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("err create new player", err)
	}

	return id
}
