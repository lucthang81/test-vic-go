package notification

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/test"
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
	dbName: "casino_notification_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	feature.UnlockAllFeature()
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../sql/init_schema.sql"})
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

func (s *TestSuite) TestFetchSavePNData(c *C) {
	// fetchPNData()
	// c.Assert(packageApnsType, Equals, "sandbox")
	// c.Assert(packageGcmApiKey, Equals, "")
	// c.Assert(packageApnsCerFileContent, Equals, "")
	// c.Assert(packageApnsKeyFileContent, Equals, "")

	// UpdatePNData("tada", "keycontent", "cercontent", "gcmkey")
	// c.Assert(packageApnsType, Equals, "tada")
	// c.Assert(packageApnsKeyFileContent, Equals, "keycontent")
	// c.Assert(packageApnsCerFileContent, Equals, "cercontent")
	// c.Assert(packageGcmApiKey, Equals, "gcmkey")

	// fetchPNData()
	// c.Assert(packageApnsType, Equals, "tada")
	// c.Assert(packageApnsKeyFileContent, Equals, "keycontent")
	// c.Assert(packageApnsCerFileContent, Equals, "cercontent")
	// c.Assert(packageGcmApiKey, Equals, "gcmkey")

}
