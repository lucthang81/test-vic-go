package sql

import (
	"io/ioutil"
	"math/rand"
	// "os/exec"
	"testing"
	"time"
	// "database/sql"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	// "github.com/vic/vic_go/log"
	"github.com/vic/vic_go/test"
	. "gopkg.in/check.v1"
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

func (s *TestSuite) TestAllTestSchemaDatabase(c *C) {
	test.DropTestDatabase("sql_test")
	files, _ := ioutil.ReadDir("test_schema")
	for _, f := range files {
		path := fmt.Sprintf("test_schema/%s", f.Name())
		err := test.CreateTestDatabase("sql_test")
		c.Assert(err, IsNil)
		dataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", "sql_test", ":63791")

		err = ReadSqlToDb(dataCenter, "init_schema.sql")
		c.Assert(err, IsNil)
		err = ReadSqlToDb(dataCenter, path)
		c.Assert(err, IsNil)
		dataCenter.Db().Close()
		test.DropTestDatabase("sql_test")
	}
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestBackup(c *C) {
	// BackupTable("casino_x1_db", []string{"player", "card"})
	// utils.DelayInDuration(10 * time.Second)
}
