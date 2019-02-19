package log

import (
	"testing"
	"time"
	// "database/sql"
	"fmt"
	. "gopkg.in/check.v1"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {

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

func (s *TestSuite) TestLog(c *C) {
	str := Log("pleple")
	c.Assert(str, Equals, fmt.Sprintf("Info <%s>: pleple", time.Now().Format(time.ANSIC)))
	strFormat := Log("pleple %s", "uuu")
	c.Assert(strFormat, Equals, fmt.Sprintf("Info <%s>: pleple uuu", time.Now().Format(time.ANSIC)))
	strFormatSerious := LogSerious("pleple %s", "uuu")
	c.Assert(strFormatSerious, Equals, fmt.Sprintf("SERIOUS <%s>: pleple uuu", time.Now().Format(time.ANSIC)))

	strFormatSerious = LogSerious("pleple %s %s", "uuu", "aaaa")
	c.Assert(strFormatSerious, Equals, fmt.Sprintf("SERIOUS <%s>: pleple uuu aaaa", time.Now().Format(time.ANSIC)))

}
