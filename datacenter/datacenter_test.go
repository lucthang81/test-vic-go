package datacenter

import (
	"math/rand"
	"os/exec"
	"testing"
	"time"
	// "database/sql"
	"fmt"
	. "gopkg.in/check.v1"
	// "log"
	"database/sql"
	"github.com/vic/vic_go/utils"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter *DataCenter
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {
	rand.Seed(time.Now().UTC().UnixNano())
	CloneSchemaToTestDatabase("datacenter_test", []string{"../sql/testobject_schema.sql"})
	s.dataCenter = NewDataCenter("vic_user", "9ate328di4rese7dra", "datacenter_test", ":63791")
}

func (s *TestSuite) TearDownSuite(c *C) {
	dropTestDatabase("datacenter_test")
	s.dataCenter.redisCache.do("FLUSHDB")
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

func (s *TestSuite) TestCreateDatabase(c *C) {
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestSimpleSetFunction(c *C) {
	redisCache := startRedis(":63791")
	_, err := redisCache.do("SET", "TestKey", "TestKey11")
	c.Assert(err, IsNil)
	reply, err := redisCache.do("GET", "TestKey")
	c.Assert(err, IsNil)
	c.Assert(string(reply.([]byte)), Equals, "TestKey11")
}

func (s *TestSuite) TestSaveObject(c *C) {
	redisCache := startRedis(":63791")
	var err error
	var value string
	var testObject *TestObject

	testObject = &TestObject{identifier: "identifierabcd", token: "token"}
	testObject.SetId(10983)

	err = redisCache.save(testObject, "SampleKey", "TestValue")
	c.Assert(err, IsNil)

	value, err = redisCache.get(testObject, "SampleKey")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "TestValue")
	value, err = redisCache.get(testObject, "SampleKey1notfound")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "")

	err = redisCache.saveMany(testObject, []string{"SampleKey1a", "SampleKey2a"}, []string{"TestValue3", "TestValue4"})
	c.Assert(err, IsNil)
	value, err = redisCache.get(testObject, "SampleKey1a")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "TestValue3")
	value, err = redisCache.get(testObject, "SampleKey2a")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "TestValue4")
}

func (s *TestSuite) TestSaveObjectToDatabase(c *C) {
	// save a test testObject
	testObject := &TestObject{identifier: "identifierabcd", token: "token"}
	testObject.username = utils.RandSeq(25)
	_, err := s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username}, true)
	c.Assert(err, IsNil)

	var token, identifier, username sql.NullString
	var id sql.NullInt64
	keys := []string{"id", "token", "identifier", "username"}
	values := []interface{}{&id, &token, &identifier, &username}
	err = s.dataCenter.getObjectFromDbWithSelectedKeys(
		testObject,
		keys,
		values)
	c.Assert(err, IsNil)
	c.Assert(utils.GetStringFromScanResult(values[1]), Equals, "token")
	c.Assert(utils.GetStringFromScanResult(values[2]), Equals, "identifierabcd")
	c.Assert(len(utils.GetStringFromScanResult(values[3])), Equals, 25)

	keys = keys[1:]
	cacheValues, err := s.dataCenter.redisCache.getMany(testObject, keys)
	c.Assert(cacheValues[0], Equals, "token")
	c.Assert(cacheValues[1], Equals, "identifierabcd")
	c.Assert(len(cacheValues[2]), Equals, 25)
}

func (s *TestSuite) TestRemoveObjectFromDatabase(c *C) {
	// save a test testObject
	testObject := &TestObject{identifier: "identifierabcd", token: "token"}
	testObject.username = utils.RandSeq(25)
	_, err := s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username}, false)
	c.Assert(err, IsNil)

	var token, identifier, username sql.NullString
	var id sql.NullInt64
	keys := []string{"id", "token", "identifier", "username"}
	values := []interface{}{&id, &token, &identifier, &username}
	err = s.dataCenter.getObjectFromDbWithSelectedKeys(
		testObject,
		keys,
		values)
	c.Assert(err, IsNil)

	s.dataCenter.RemoveObject(testObject)
	keys = []string{"id", "token", "identifier", "username"}
	values = []interface{}{&id, &token, &identifier, &username}
	err = s.dataCenter.getObjectFromDbWithSelectedKeys(
		testObject,
		keys,
		values)
	c.Assert(err, NotNil)

}

func (s *TestSuite) TestGetData(c *C) {
	s.dataCenter.redisCache.do("FLUSHDB")
	// save a test testObject
	var testObject *TestObject
	var err error
	var power int64
	var iden string
	var hitCache bool

	testObject = &TestObject{identifier: "identifierabcd", token: "token"}
	testObject.username = utils.RandSeq(25)
	testObject.power = 129
	_, err = s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username", "power"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username, testObject.power}, true)
	c.Assert(err, IsNil)

	power, hitCache, err = s.dataCenter.GetInt64FieldForObject(testObject, "power", true)
	c.Assert(err, IsNil)
	c.Assert(power, Equals, int64(129))
	c.Assert(hitCache, Equals, true)

	iden, hitCache, err = s.dataCenter.GetStringFieldForObject(testObject, "identifier", true)
	c.Assert(err, IsNil)
	c.Assert(iden, Equals, "identifierabcd")
	c.Assert(hitCache, Equals, true)

	// will not hit cache here
	testObject = &TestObject{identifier: "identdxxe", token: "tokssssen"}
	testObject.username = utils.RandSeq(25)
	testObject.power = 199
	_, err = s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username", "power"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username, testObject.power}, false)
	c.Assert(err, IsNil)

	power, hitCache, err = s.dataCenter.GetInt64FieldForObject(testObject, "power", true)
	c.Assert(err, IsNil)
	c.Assert(power, Equals, int64(199))
	c.Assert(hitCache, Equals, false)

	iden, hitCache, err = s.dataCenter.GetStringFieldForObject(testObject, "identifier", true)
	c.Assert(err, IsNil)
	c.Assert(iden, Equals, "identdxxe")
	c.Assert(hitCache, Equals, false)

}

func (s *TestSuite) TestCheckDataExist(c *C) {
	// save a test testObject
	var testObject *TestObject
	var err error
	var hitCache bool
	var exist bool

	testObject = &TestObject{identifier: "identifierabcd", token: "token"}
	testObject.username = utils.RandSeq(25)
	testObject.power = 129
	_, err = s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username", "power"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username, testObject.power}, true)
	c.Assert(err, IsNil)

	// check exist
	exist, hitCache, err = s.dataCenter.IsObjectExist(testObject, []string{"token", "identifier"}, []interface{}{"token", "identifierabcd"}, true)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)
	c.Assert(hitCache, Equals, true)

	// check exist
	exist, hitCache, err = s.dataCenter.IsObjectExist(testObject, []string{"token", "identifier"}, []interface{}{"token2222", "identifierabcd"}, true)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, false)
	c.Assert(hitCache, Equals, false)

	// will not hit cache here
	testObject = &TestObject{identifier: "identdxxe", token: "tokssssen"}
	testObject.username = utils.RandSeq(25)
	testObject.power = 199
	_, err = s.dataCenter.InsertObject(testObject,
		[]string{"token", "identifier", "username", "power"},
		[]interface{}{testObject.token, testObject.identifier, testObject.username, testObject.power}, false)
	c.Assert(err, IsNil)

	exist, hitCache, err = s.dataCenter.IsObjectExist(testObject, []string{"token", "identifier"}, []interface{}{"tokssssen", "identdxxe"}, true)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)
	c.Assert(hitCache, Equals, false)

	exist, hitCache, err = s.dataCenter.IsObjectExist(testObject, []string{"token", "identifier"}, []interface{}{"dd2", "33"}, true)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, false)
	c.Assert(hitCache, Equals, false)
}

func (s *TestSuite) TestCache(c *C) {
	s.dataCenter.SaveKeyValueToCache("room", "1", "room1hasA")
	s.dataCenter.SaveKeyValueToCache("room", "2", "room2hasA")
	s.dataCenter.SaveKeyValueToCache("room", "3", "room3hasB")
	s.dataCenter.SaveKeyValueToCache("room", "4", "room4hasC")
	s.dataCenter.SaveKeyValueToCache("room", "5", "room5hasD")

	keys, _ := s.dataCenter.LoadAllKeysFromCache("room1")
	c.Assert(len(keys), Equals, 0)

	value, _ := s.dataCenter.LoadValueFromCache("room", "1")
	c.Assert(value, Equals, "room1hasA")

	keys, _ = s.dataCenter.LoadAllKeysFromCache("room")
	c.Assert(len(keys), Equals, 5)

	s.dataCenter.RemoveKeyValueFromCache("room", "3")
	keys, _ = s.dataCenter.LoadAllKeysFromCache("room")
	c.Assert(len(keys), Equals, 4)

}

type TestObject struct {
	id         int64
	identifier string
	token      string
	username   string
	power      int64
}

const TestObjectCacheKey string = "testObject"
const TestObjectDatabaseTableName string = "avahardcore_testobject"
const TestObjectClassName string = "TestObject"

func (testObject *TestObject) CacheKey() string {
	return TestObjectCacheKey
}

func (testObject *TestObject) DatabaseTableName() string {
	return TestObjectDatabaseTableName
}

func (testObject *TestObject) ClassName() string {
	return TestObjectClassName
}

func (testObject *TestObject) Id() int64 {
	return testObject.id
}

func (testObject *TestObject) SetId(id int64) {
	testObject.id = id
}

func CloneSchemaToTestDatabase(dbName string, sqlPaths []string) (err error) {
	dataCenter := NewDataCenter("vic_user", "9ate328di4rese7dra", "casino_db", "")
	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
	dataCenter.Db().Close()

	// _, err = exec.Command("sh", "-c", "createdb --username=gia --owner=avahardcore_user --encoding=UTF8 test_avahardcore_db; pg_dump avahardcore_db -s | psql test_avahardcore_db").Output()
	_, err = exec.Command("sh", "-c", fmt.Sprintf("createdb --owner=vic_user --encoding=UTF8 %s;", dbName)).Output()
	if err != nil {
		fmt.Println("???")
		fmt.Println(err)
		return err
	}
	for _, sqlPath := range sqlPaths {
		_, err = exec.Command("sh", "-c", fmt.Sprintf("psql -d %s -f %s", dbName, sqlPath)).Output()
		if err != nil {
			fmt.Println("??? ??")
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func dropTestDatabase(dbName string) {
	dataCenter := NewDataCenter("vic_user", "9ate328di4rese7dra", "datacenter_test", "")
	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
}
