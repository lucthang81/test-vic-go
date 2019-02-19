package utils

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"math/rand"
	"time"

	// "os/exec"
	"testing"
	// "database/sql"
	// "github.com/vic/vic_go/log"
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

func (s *TestSuite) TestTimeOut(c *C) {
	timeOut := NewTimeOut(3 * time.Second)
	c.Assert(timeOut.Start(), Equals, true)

	// cancel a timeout mid way
	timeOut = NewTimeOut(3 * time.Second)
	go cancelThisTimeOut(timeOut)
	c.Assert(timeOut.Start(), Equals, false)

	timeOut = NewTimeOut(5 * time.Second)
	anotherTimeOut := NewTimeOut(3 * time.Second)
	go cancelTimeOut1AfterTimeOut2End(timeOut, anotherTimeOut)
	c.Assert(timeOut.Start(), Equals, false)
}

func cancelTimeOut1AfterTimeOut2End(timeOut1 *TimeOut, timeOut2 *TimeOut) {
	if timeOut2.Start() {
		timeOut1.SetShouldHandle(false)
	}
}

func cancelThisTimeOut(timeOut *TimeOut) {
	timeOut.SetShouldHandle(false)
}

func (s *TestSuite) TestFormatWithComma(c *C) {
	result := FormatWithComma(23494500 - 3064500)
	c.Assert(result, Equals, "20.430.000")
}

func (s *TestSuite) TestLevelFormular(c *C) {
	var level, exp int64
	exp = 5000000
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(99))
	exp = 4900000
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(98))
	exp = 796435
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(48))
	exp = 53
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(2))
	exp = 54
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(2))
	exp = 290
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(2))
	exp = 300
	level = LevelFromExp(exp)
	c.Assert(level, Equals, int64(3))
}

func (s *TestSuite) TestGetStuffFromData(c *C) {
	var data map[string]interface{}
	data = map[string]interface{}{
		"fields": []interface{}{"room"},
	}
	c.Assert(ContainsByString(GetStringSliceAtPath(data, "fields"), "room"), Equals, true)
}

func (s *TestSuite) TestCompareVersion(c *C) {
	c.Assert(IsVersion1BiggerThanVersion2("1.0", "2.3"), Equals, false)
	c.Assert(IsVersion1BiggerThanVersion2("1.0.3", "2.3"), Equals, false)
	c.Assert(IsVersion1BiggerThanVersion2("1.0", "1.3"), Equals, false)
	c.Assert(IsVersion1BiggerThanVersion2("3.0", "2.3"), Equals, true)
	c.Assert(IsVersion1BiggerThanVersion2("3.0.1", "3.0"), Equals, true)
}

func (s *TestSuite) TestGetMapWithStringSliceOut(c *C) {
	rawMapData := map[string]interface{}{
		"data": map[string][]string{
			"1": []string{"a", "b", "c"},
			"2": []string{"e", "f", "g"},
			"3": []string{"h", "i", "j"},
			"4": []string{"k", "l", "m"},
		},
	}
	payload, err := json.Marshal(rawMapData)
	c.Assert(err, IsNil)
	c.Assert(len(payload) != 0, Equals, true)

	var data map[string]interface{}
	err = json.Unmarshal(payload, &data)
	c.Assert(err, IsNil)
	resultMap := GetMapAtPath(data, "data")
	c.Assert(len(resultMap), Equals, 4)

	for _, value := range resultMap {
		// keyInt, _ := strconv.ParseInt(key, 10, 16)
		stringSlice := GetStringSliceFromScanResult(value)
		c.Assert(len(stringSlice), Equals, 3)
	}
}

func (s *TestSuite) TestGetMapWithMapSliceOut(c *C) {
	rawMapData := map[string]interface{}{
		"data": []map[string]interface{}{
			map[string]interface{}{"a": "b", "c": "d"},
			map[string]interface{}{"a": "b", "c": "d"},
			map[string]interface{}{"a": "b", "c": "d"},
			map[string]interface{}{"a": "b", "c": "d"},
		},
	}
	payload, err := json.Marshal(rawMapData)
	c.Assert(err, IsNil)
	c.Assert(len(payload) != 0, Equals, true)

	var data map[string]interface{}
	err = json.Unmarshal(payload, &data)
	c.Assert(err, IsNil)
	resultMap := GetMapSliceAtPath(data, "data")
	c.Assert(len(resultMap), Equals, 4)

	for _, value := range resultMap {
		// keyInt, _ := strconv.ParseInt(key, 10, 16)
		c.Assert(value["a"], Equals, "b")
	}
}

func (s *TestSuite) TestBoolScan(c *C) {
	rawData := map[string]interface{}{
		"a": false,
		"b": true,
		"c": true,
		"d": false,
	}

	rawConverted := ConvertData(rawData)
	for key, boolData := range rawConverted {
		fmt.Println("in this test")
		if key == "a" {
			c.Assert(GetBoolFromScanResult(boolData), Equals, false)
			fmt.Println("did compare")
		} else if key == "b" {
			c.Assert(GetBoolFromScanResult(boolData), Equals, true)
			fmt.Println("did compare")
		} else if key == "c" {
			c.Assert(GetBoolFromScanResult(boolData), Equals, true)
			fmt.Println("did compare")
		} else {
			c.Assert(GetBoolFromScanResult(boolData), Equals, false)
			fmt.Println("did compare")
		}
	}
}

func (s *TestSuite) TestGetTimeDurationUntils(c *C) {

	currentTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", "Thu, 10 Sep 2015 15:04:05 -0700")
	durationUntilsEndOfDay := TimeDurationUntilEndOfDay(currentTime)

	c.Assert(durationUntilsEndOfDay.Seconds(), Equals, float64(32154)) // 54s + 55m*60 + 8h*3600

	durationUntilsEndOfWeek := TimeDurationUntilEndOfWeek(currentTime)
	c.Assert(durationUntilsEndOfWeek.Seconds(), Equals, float64(291354)) // 54s + 55m*60 + 8h*3600 + 3d*86400

	currentTime, _ = time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", "Sun, 13 Sep 2015 20:04:05 -0700")
	durationUntilsEndOfWeek = TimeDurationUntilEndOfWeek(currentTime)
	c.Assert(durationUntilsEndOfWeek.Seconds(), Equals, float64(14154)) // 54s + 55m*60 + 3h*3600
}

func (s *TestSuite) TestStartEndWeekTime(c *C) {

	currentTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", "Thu, 10 Sep 2015 15:04:05 -0700")

	startDate := StartOfWeekFromTime(currentTime)
	endDate := EndOfWeekFromTime(currentTime)

	durationUntilsEndOfWeek := endDate.Sub(currentTime)
	c.Assert(durationUntilsEndOfWeek.Seconds(), Equals, float64(291354)) // 54s + 55m*60 + 8h*3600 + 3d*86400

	durationUntilsStartOfWeek := currentTime.Sub(startDate)
	c.Assert(durationUntilsStartOfWeek.Seconds(), Equals, float64(313445)) // 5s + 04m*60 + 15h*3600 + 3d*86400

	// =======
	currentTime, _ = time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", "Mon, 12 Dec 2016 3:24:05 -0700")
	fmt.Println(currentTime)
	startDate = StartOfWeekFromTime(currentTime)
	endDate = EndOfWeekFromTime(currentTime)

	durationUntilsEndOfWeek = endDate.Sub(currentTime)
	c.Assert(durationUntilsEndOfWeek.Seconds(), Equals, float64(592554)) // 54s + 35m*60 + 20h*3600 + 6d*86400

	durationUntilsStartOfWeek = currentTime.Sub(startDate)
	c.Assert(durationUntilsStartOfWeek.Seconds(), Equals, float64(12245)) // 5s + 24m*60 + 3h*3600 + 0d*86400

	// =======
	currentTime, _ = time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", "Sun, 18 Dec 2016 3:24:05 -0700")
	fmt.Println(currentTime)
	startDate = StartOfWeekFromTime(currentTime)
	endDate = EndOfWeekFromTime(currentTime)

	durationUntilsEndOfWeek = endDate.Sub(currentTime)
	c.Assert(durationUntilsEndOfWeek.Seconds(), Equals, float64(74154)) // 54s + 35m*60 + 20h*3600 + 0d*86400

	durationUntilsStartOfWeek = currentTime.Sub(startDate)
	c.Assert(durationUntilsStartOfWeek.Seconds(), Equals, float64(530645)) // 5s + 24m*60 + 3h*3600 + 6d*86400
}

func (s *TestSuite) TestParse(c *C) {
	body := `
	{"status":true,"error_code":0,"message":"Successfully!","data":{"transaction_id":"C11569DFEE6D49D0","type":"CARD","amount":"20000","currency":"VND","country_code":"VN","target":false,"state":false,"time":"16:16:23 01\/19\/2016","sandbox":0,"revenue":20000},"data_signature":"LLbJkvWcrU1tEnuBGMgtXhQ1T\/hozZDopbY\/o5F9ODqX4MSElIniUp1\/5IjzZufBBsUYmnMxdtFtp\/vFmXIQSD1sHJh5YRPPjTb+G4hqN0P4+V41vYNwLcHJBo3OktwkJ8aTz1705VN1pupe4SPgerd6diUUb28rTu\/kHw+uO9k="}
	`
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		// handle error
		log.LogSerious("error parsing appvn response %v", err)
		// return "", "", errors.New(l.Get(l.M0003))
	}
	fmt.Println("response", string(body))
	errorCode := GetInt64AtPath(data, "error_code")

	if errorCode != 0 {
		// return "", "", errors.New("err:wrong_card_data")
	}

	transactionId := GetStringAtPath(data, "data/transaction_id")
	cardValue := GetInt64OrStringAsInt64AtPath(data, "data/amount")
	// return transactionId, cardValue, nil
	fmt.Println(transactionId, cardValue)
}

func (s *TestSuite) TestPhoneNumber(c *C) {
	var phoneNumber string
	phoneNumber = "098432573"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8498432573")

	phoneNumber = "+8498432573"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8498432573")

	phoneNumber = "0123457463"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "84123457463")

	phoneNumber = "8497634587"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8497634587")

	phoneNumber = "+84 97 634 587"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8497634587")

	phoneNumber = " +84 97 634 587"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8497634587")

	phoneNumber = " + 84 97 634 587"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "8497634587")

	phoneNumber = "4182kdfsajs"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "4182kdfsajs")

	phoneNumber = "fkdaj slfkjas ldfja"
	c.Assert(NormalizePhoneNumber(phoneNumber), Equals, "fkdajslfkjasldfja")
}

func (s *TestSuite) TestHideString(c *C) {
	testString := "12345678"
	c.Assert(HideString(testString, 3, true), Equals, "xxx45678")
	c.Assert(HideString(testString, 3, false), Equals, "12345xxx")
	c.Assert(HideString(testString, 5, true), Equals, "xxxxx678")
	c.Assert(HideString(testString, 5, false), Equals, "123xxxxx")
}

// func (s *TestSuite) TestEncrypt(c *C) {
// 	password := "{K^ltmv^j/KFHfQ.MW(&:_Lt\":{5"

// 	err = bcrypt.CompareHashAndPassword(passwordFromDb, []byte(oldPassword))
// 	if err != nil {
// 		return errors.New("Wrong password")

// 	}

// 	err = bcrypt.CompareHashAndPassword(passwordActionFromDb, []byte(oldPasswordAction))
// 	if err != nil {
// 		return errors.New("Wrong password (action)")

// 	}

// 	passwordByte := []byte(password)
// 	passwordActionByte := []byte(passwordAction)
// 	// Hashing the password with the default cost of 10
// 	hashedPassword, err := bcrypt.GenerateFromPassword(passwordByte, bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}

// 	hashedPasswordAction, err := bcrypt.GenerateFromPassword(passwordActionByte, bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
// }
