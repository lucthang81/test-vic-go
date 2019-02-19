package utils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vic/vic_go/log"
	"golang.org/x/crypto/bcrypt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MEDIA_URL = "https://snap-media-bucket.s3.amazonaws.com/media/"
const layout = "2006-01-02T15:04:05Z"

func test() int64 {
	return 0
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var lettersLowercase = []rune("abcdefghijklmnopqrstuvwxyz")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandSeqLowercase(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = lettersLowercase[rand.Intn(len(lettersLowercase))]
	}
	return string(b)
}

func GenerateMediaUrl(relativePath string) (absolutePath string) {
	if len(relativePath) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%s%s", MEDIA_URL, relativePath)
	}

}

func AbsInt64(x int64) int64 {
	if x >= 0 {
		return x
	} else {
		return -x
	}
}

func MinInt64(x int64, y int64) int64 {
	if x > y {
		return y
	} else {
		return x
	}
}

func MaxInt64(x int64, y int64) int64 {
	if x < y {
		return y
	} else {
		return x
	}
}

func MinInt(x int, y int) int {
	if x > y {
		return y
	} else {
		return x
	}
}

func MaxInt(x int, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}

func GetInt64AtPath(data map[string]interface{}, path string) int64 {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(float64); ok {
			// fmt.Println("realValue", realValue)
			if (realValue < -9223372036854775808) || (realValue > 9223372036854775808) {
				realValue = 0
			}
			return int64(realValue)
		} else if realValue, ok := value.(int64); ok {
			return realValue
		} else if realValue, ok := value.(int); ok {
			return int64(realValue)
		} else {
			return 0
		}
	}
	return 0
}

func GetInt64OrStringAsInt64AtPath(data map[string]interface{}, path string) int64 {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(float64); ok {
			return int64(realValue)
		} else if realValue, ok := value.(int64); ok {
			return realValue
		} else if realValue, ok := value.(int); ok {
			return int64(realValue)
		} else if realValue, ok := value.(string); ok {
			intValue, _ := strconv.ParseInt(realValue, 10, 64)
			return intValue
		} else {
			return 0
		}
	}
	return 0
}

func GetFloat64AtPath(data map[string]interface{}, path string) float64 {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(float64); ok {
			return realValue
		} else if realValue, ok := value.(int64); ok {
			return float64(realValue)
		} else {
			log.LogSerious("cannot handle get float64 at path for %v, path %s", data, path)
			return 0
		}
	}
	return 0
}

func GetIntAtPath(data map[string]interface{}, path string) int {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(float64); ok {
			return int(realValue)
		} else if realValue, ok := value.(int64); ok {
			return int(realValue)
		} else if realValue, ok := value.(int); ok {
			return realValue
		} else {
			return 0
		}

	}
	return 0
}

func GetBoolAtPath(data map[string]interface{}, path string) bool {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(float64); ok {
			return realValue == 1
		} else if realValue, ok := value.(int64); ok {
			return realValue == 1
		} else if realValue, ok := value.(int); ok {
			return realValue == 1
		} else if realValue, ok := value.(bool); ok {
			return realValue
		} else if realValue, ok := value.(string); ok {
			return realValue == "true"
		} else {
			return false
		}

	}
	return false
}

func GetStringAtPath(data map[string]interface{}, path string) string {
	value := GetDataAtPath(data, path)
	if value != nil {
		if realValue, ok := value.(string); ok {
			return realValue
		}
	}
	return ""
}

func GetInt64SliceAtPath(data map[string]interface{}, path string) []int64 {
	value := GetDataAtPath(data, path)
	if value != nil {
		dataSlices, ok := value.([]interface{})
		if ok {
			result := make([]int64, 0, len(dataSlices))
			for _, interfaceValue := range dataSlices {
				if interfaceValue != nil {
					result = append(result, int64(interfaceValue.(float64)))
				}
			}
			return result
		} else {
			int64Slices := value.([]int64)
			return int64Slices
		}
	}
	return nil
}

func GetIntSliceAtPath(data map[string]interface{}, path string) []int {
	value := GetDataAtPath(data, path)
	if value != nil {
		dataSlices, ok := value.([]interface{})
		if ok {
			result := make([]int, 0, len(dataSlices))
			for _, interfaceValue := range dataSlices {
				if interfaceValue != nil {
					result = append(result, int(interfaceValue.(float64)))
				}
			}
			return result
		} else {
			int64Slices := value.([]int)
			return int64Slices
		}
	}
	return nil
}

func ConvertRawStringToInt64Slice(rawString string) []int64 {
	stringSlice := strings.Split(rawString, ",")
	int64Slice := make([]int64, 0)
	for _, stringValue := range stringSlice {
		int64Value, _ := strconv.ParseInt(stringValue, 10, 64)
		int64Slice = append(int64Slice, int64Value)
	}
	return int64Slice
}

func ConvertInt64SliceToRawString(slice []int64) string {
	stringSlice := make([]string, 0)
	for _, element := range slice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", element))
	}
	value := strings.Join(stringSlice, ",")
	return value
}

func ConvertIntSliceToInt64Slice(slice []int) []int64 {
	temp := make([]int64, 0)
	for _, value := range slice {
		temp = append(temp, int64(value))
	}
	return temp
}

func ConvertInt64SliceToIntSlice(slice []int64) []int {
	temp := make([]int, 0)
	for _, value := range slice {
		temp = append(temp, int(value))
	}
	return temp
}

func GetStringSliceAtPath(data map[string]interface{}, path string) []string {
	value := GetDataAtPath(data, path)
	if value != nil {
		if val, ok := value.([]string); ok {
			return val
		} else if val, ok := value.([]interface{}); ok {
			result := make([]string, 0, len(val))
			for _, interfaceValue := range val {
				if interfaceValue != nil {
					if interfaceVal, ok := interfaceValue.(string); ok {
						result = append(result, interfaceVal)
					}
				}
			}
			return result
		}
	}
	return nil
}

func GetStringSliceSliceAtPath(data map[string]interface{}, path string) [][]string {
	value := GetDataAtPath(data, path)
	if value != nil {
		if val, ok := value.([][]string); ok {
			return val
		} else if val, ok := value.([]interface{}); ok {
			result := make([][]string, 0, len(val))
			for _, interfaceValue := range val {
				if interfaceValue != nil {
					if interfaceVal, ok := interfaceValue.([]string); ok {
						result = append(result, interfaceVal)
					}
				}
			}
			return result
		}
	}
	return nil
}

func GetBoolSliceAtPath(data map[string]interface{}, path string) []bool {
	value := GetDataAtPath(data, path)
	if value != nil {
		if val, ok := value.([]bool); ok {
			return val
		} else if dataSlices, ok := value.([]interface{}); ok {
			result := make([]bool, 0, len(dataSlices))
			for _, interfaceValue := range dataSlices {
				if interfaceValue != nil {
					result = append(result, interfaceValue.(bool))
				}
			}
			return result
		}
	}
	return nil
}

func GetMapInt64Int64AtPath(data map[string]interface{}, path string) map[int64]int64 {
	value := GetDataAtPath(data, path)
	if value != nil {
		return value.(map[int64]int64)
	}
	return nil
}

func GetMapAtPath(data map[string]interface{}, path string) map[string]interface{} {
	value := GetDataAtPath(data, path)
	if value != nil {
		return value.(map[string]interface{})
	}
	return nil
}

func GetMapSliceAtPath(data map[string]interface{}, path string) []map[string]interface{} {
	value := GetDataAtPath(data, path)
	if value != nil {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(value)

			result := make([]map[string]interface{}, 0, s.Len())
			for i := 0; i < s.Len(); i++ {
				result = append(result, s.Index(i).Interface().(map[string]interface{}))
			}
			return result
		}

	}
	return nil
}

func GetMapSliceSliceAtPath(data map[string]interface{}, path string) [][]map[string]interface{} {
	value := GetDataAtPath(data, path)
	if value != nil {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(value)

			result := make([][]map[string]interface{}, 0, s.Len())
			for i := 0; i < s.Len(); i++ {
				subValue := s.Index(i).Interface()
				if subValue != nil {
					switch reflect.TypeOf(subValue).Kind() {
					case reflect.Slice:
						subS := reflect.ValueOf(subValue)
						subResult := make([]map[string]interface{}, 0, subS.Len())
						for j := 0; j < subS.Len(); j++ {
							subResult = append(subResult, subS.Index(j).Interface().(map[string]interface{}))
						}
						result = append(result, subResult)
					}
				}
			}
			return result
		}

	}
	return nil
}

func GetStringFromScanResult(interfaceValue interface{}) string {
	if value, ok := interfaceValue.(*sql.NullString); ok {
		return (*value).String
	}
	return ""
}

func GetBoolFromScanResult(interfaceValue interface{}) bool {
	if value, ok := interfaceValue.(bool); ok {
		return value
	}
	return false
}

func GetInt64FromScanResult(value interface{}) int64 {
	if realValue, ok := value.(float64); ok {
		return int64(realValue)
	} else if realValue, ok := value.(int64); ok {
		return realValue
	} else {
		return 0
	}
	return 0
}

func GetStringSliceFromScanResult(value interface{}) []string {
	if val, ok := value.([]string); ok {
		return val
	} else if val, ok := value.([]interface{}); ok {
		return GetStringSliceFromInterfaceSlice(val)
	}
	return nil
}

func GetMapFromScanResult(value interface{}) map[string]interface{} {
	if val, ok := value.(map[string]interface{}); ok {
		return val
	}
	return nil
}

func GetStringFromInterface(interfaceValue interface{}) string {
	if value, ok := interfaceValue.(int); ok {
		return fmt.Sprintf("%d", value)
	} else if value, ok := interfaceValue.(int64); ok {
		return fmt.Sprintf("%d", value)
	} else if value, ok := interfaceValue.(string); ok {
		return value
	} else if value, ok := interfaceValue.(sql.NullString); ok {
		return value.String
	}
	return ""
}

func GetStringSliceFromInterfaceSlice(origin []interface{}) (result []string) {
	result = make([]string, len(origin))
	for index, interfaceValue := range origin {
		result[index] = GetStringFromInterface(interfaceValue)
	}
	return result
}

func ConvertData(data map[string]interface{}) map[string]interface{} {
	payload, _ := json.Marshal(data)
	var convertedData map[string]interface{}
	json.Unmarshal(payload, &convertedData)
	return convertedData
}

// data/component:1/abc
func GetDataAtPath(data map[string]interface{}, path string) interface{} {
	pathComponents := strings.Split(path, "/")
	subData := data
	lastIndex := len(pathComponents) - 1
	for index, pathComponent := range pathComponents {
		pathComponentAndIndex := strings.Split(pathComponent, ":")
		var indexPath int
		var finalPathComponent string
		if len(pathComponentAndIndex) == 1 {
			finalPathComponent = pathComponentAndIndex[0]
			if subData[finalPathComponent] != nil {
				if index != lastIndex {
					subData = subData[finalPathComponent].(map[string]interface{})
				} else {
					return subData[finalPathComponent]
				}
			} else {
				return nil
			}
		} else {
			finalPathComponent = pathComponentAndIndex[0]
			indexPath, _ = strconv.Atoi(pathComponentAndIndex[1])

			if val, ok := subData[finalPathComponent].([]interface{}); ok {
				subData = val[indexPath].(map[string]interface{})
			} else if val, ok := subData[finalPathComponent].([]map[string]interface{}); ok {
				subData = val[indexPath]
			}
			if index == lastIndex {
				return subData
			}
		}

	}
	return nil
}

func CheckEnoughParameters(data map[string]interface{}, parametersName ...string) error {
	for _, parameterName := range parametersName {
		if _, ok := data[parameterName]; !ok {
			return errors.New(fmt.Sprintf("err:missing_parameter_%s", parameterName))
		}
	}
	return nil
}

func GenerateArrayForSqlArray(array []int64) string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	for index, intData := range array {
		buffer.WriteString(fmt.Sprintf("%d", intData))
		if index != len(array)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

func GetBlockedIdsFromBlockIdsString(blockIds string) []int64 {
	idsStr := strings.Split(blockIds, ",")
	keys := make([]int64, len(idsStr))
	i := 0
	for _, k := range idsStr {
		int64Data, _ := strconv.ParseInt(k, 10, 64)
		keys[i] = int64Data
		i += 1
	}
	return keys
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsByString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsByInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsByInt64(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ModInt(x, y int) int {
	return int(math.Mod(float64(x), float64(y)))
}

func LevelFromExp(exp int64) int64 {
	return int64(math.Pow(float64(exp)/5000000, 1/2.5)*98 + 1)
}

// input is nSeconds
func Delay(seconds int) {
	timeout := time.After(time.Duration(seconds) * time.Second)
	<-timeout
}

func DelayInDuration(duration time.Duration) {
	timeout := time.After(duration)
	<-timeout
}

func IsVersion1BiggerThanVersion2(version1 string, version2 string) bool {
	tokens1 := strings.Split(version1, ".")
	tokens2 := strings.Split(version2, ".")
	for i := 0; i < MinInt(len(tokens1), len(tokens2)); i++ {
		value1, _ := strconv.Atoi(tokens1[i])
		value2, _ := strconv.Atoi(tokens2[i])
		if value1 != value2 {
			return value1 > value2
		}
	}
	// still same
	return len(tokens1) > len(tokens2)
}

func IsVersion1BiggerThanOrEqualsVersion2(version1 string, version2 string) bool {
	if version1 == version2 {
		return true
	}
	return IsVersion1BiggerThanVersion2(version1, version2)
}

func FormatTime(theTime time.Time) string {
	return theTime.UTC().Format(layout)
}

func FormatTimeToVietnamTime(theTime time.Time) (dateString string, timeString string) {
	vnLayout := "2-1-2006 15:04:05 -0700"
	dateTimeString := TranslateTimeToVNTime(theTime).Format(vnLayout)
	tokens := strings.Split(dateTimeString, " ")
	if len(tokens) > 1 {
		return tokens[0], tokens[1]
	}
	return "", ""
}

func FormatTimeToVietnamTimeMonthYear(theTime time.Time) (dateString string) {
	vnLayout := "1-2006"
	dateTimeString := TranslateTimeToVNTime(theTime).Format(vnLayout)
	return dateTimeString
}

func FormatTimeToVietnamDateTimeString(theTime time.Time) (dateTimeString string) {
	vnLayout := "2-1-2006 15:04:05 -0700"
	dateTimeString = TranslateTimeToVNTime(theTime).Format(vnLayout)
	tokens := strings.Split(dateTimeString, " ")
	if len(tokens) > 1 {
		return fmt.Sprintf("%s %s", tokens[0], tokens[1])
	}
	return ""
}

func ParseTime(dateTimeString string) time.Time {
	timeObject, _ := time.Parse(layout, dateTimeString)
	return timeObject
}

func TimeFromVietnameseTimeString(dateString string, timeString string) time.Time {
	dateTimeString := fmt.Sprintf("%s %s +0700", dateString, timeString)

	//Mon Jan 2 15:04:05 MST 2006
	vnLayout := "2-1-2006 15:04:05 -0700"
	dateTimeObject, _ := time.Parse(vnLayout, dateTimeString)
	return dateTimeObject
}

func TimeFromVietnameseDateString(dateString string) time.Time {
	dateTimeString := fmt.Sprintf("%s %s +0700", dateString, "00:00:00")

	//Mon Jan 2 15:04:05 MST 2006
	vnLayout := "2-1-2006 15:04:05 -0700"
	dateTimeObject, _ := time.Parse(vnLayout, dateTimeString)
	return dateTimeObject
}

func StartOfWeekFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	dayAgoMonday := 0
	if currentTime.Weekday() != 1 { // not monday currently
		dayAgoMonday = int(currentTime.Weekday()) - 1
		if dayAgoMonday == -1 {
			dayAgoMonday = 6
		}
	}
	if dayAgoMonday == 0 {
		calculatedTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	} else {
		calculatedTime = currentTime.Add(-time.Duration(dayAgoMonday*24) * time.Hour)
		calculatedTime = time.Date(calculatedTime.Year(), calculatedTime.Month(), calculatedTime.Day(), 0, 0, 0, 0, calculatedTime.Location())
	}

	return calculatedTime
}

func EndOfWeekFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	dayTillSunday := 0
	if currentTime.Weekday() != 0 { // not sunday currently
		dayTillSunday = 7 - int(currentTime.Weekday())
	}
	if dayTillSunday == 0 {
		calculatedTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
	} else {
		calculatedTime = currentTime.Add(time.Duration(dayTillSunday*24) * time.Hour)
		calculatedTime = time.Date(calculatedTime.Year(), calculatedTime.Month(), calculatedTime.Day(), 23, 59, 59, 0, calculatedTime.Location())
	}

	return calculatedTime
}

func EndOfDayFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	calculatedTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
	return calculatedTime
}
func StartOfDayFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	calculatedTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	return calculatedTime
}

func StartOfMonthFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	calculatedTime = time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
	return calculatedTime
}

func EndOfMonthFromTime(fromTime time.Time) time.Time {
	currentTime := fromTime
	var calculatedTime time.Time
	calculatedTime = time.Date(currentTime.Year(), currentTime.Month()+1, 1, 23, 59, 59, 0, currentTime.Location())
	calculatedTime = calculatedTime.Add(-24 * time.Hour)
	return calculatedTime
}

func NextTimeFromTimeOnly(timeOnly time.Time) time.Time {
	var calculatedTime time.Time
	timeOnly = TranslateTimeToVNTime(timeOnly)
	timeNow := CurrentTimeInVN()
	calculatedTime = time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), timeOnly.Nanosecond(), timeNow.Location())
	if calculatedTime.Before(timeNow) {
		calculatedTime = calculatedTime.Add(24 * time.Hour)
	}
	return calculatedTime
}

func TimeDurationUntilEndOfWeek(fromTime time.Time) time.Duration {
	return EndOfWeekFromTime(fromTime).Sub(fromTime)
}

func TimeDurationUntilEndOfDay(fromTime time.Time) time.Duration {
	return EndOfDayFromTime(fromTime).Sub(fromTime)
}

func CurrentTimeInVN() time.Time {
	location := time.FixedZone("ICT", 25200)
	return time.Now().In(location)
}

func TranslateTimeToVNTime(timeObject time.Time) time.Time {
	location := time.FixedZone("ICT", 25200)
	return timeObject.In(location)
}

func RoundDurationToSeconds(duration time.Duration) time.Duration {
	newDuration := time.Duration(duration.Seconds()) * time.Second
	return newDuration
}

func HashPassword(password string) (hash string) {
	bytePassword := []byte(password)
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}

func CompareHashedPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func FormatWithComma(number int64) string {
	var isPositive bool
	if number > 0 {
		isPositive = true
	}
	absNumber := int64(math.Abs(float64(number)))
	if absNumber/1000 == 0 {
		return fmt.Sprintf("%d", number)
	} else {
		firstPart := absNumber / 1000
		secondPart := absNumber % 1000
		secondPartString := fmt.Sprintf("%d", secondPart+1000)
		finalString := fmt.Sprintf("%s.%s", FormatWithComma(firstPart), secondPartString[1:])
		if !isPositive {
			finalString = fmt.Sprintf("-%s", finalString)
		}
		return finalString
	}
}

func ShortNumberStringToNumber(shortNumberString string) int64 {
	shortNumberString = strings.Replace(shortNumberString, "m", "000000", 1)
	shortNumberString = strings.Replace(shortNumberString, "k", "000", 1)
	number, _ := strconv.ParseInt(shortNumberString, 10, 64)
	return number

}

func Round(number float64) (result float64) {
	roundOn := 0.5
	places := 0
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * number
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	result = round / pow
	return
}

func Int64AfterApplyFloat64Multiplier(int64Value int64, multiplier float64) int64 {
	return int64(Round((float64(int64Value) * multiplier)))
}

func SecondsBetweenTime(time1 time.Time, time2 time.Time) float64 {
	seconds, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", time2.Sub(time1).Seconds()), 10)
	return seconds
}

func NormalizePhoneNumber(phoneNumber string) string {
	phoneNumber = strings.Replace(phoneNumber, " ", "", -1)
	if len(phoneNumber) == 0 {
		return ""
	}
	if phoneNumber[0] == '0' {
		phoneNumber = fmt.Sprintf("84%s", phoneNumber[1:len(phoneNumber)])
	}

	if phoneNumber[0] == '+' {
		phoneNumber = phoneNumber[1:len(phoneNumber)]
	}

	return phoneNumber
}

func HideString(originString string, length int, hideFromHead bool) string {
	xString := ""
	for i := 0; i < length; i++ {
		xString += "x"
	}
	if hideFromHead {
		return fmt.Sprintf("%s%s", xString, originString[length:len(originString)])
	} else {
		return fmt.Sprintf("%s%s", originString[:len(originString)-length], xString)
	}
}

type DBInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func GetInt64FromQuery(db DBInterface, queryString string, a ...interface{}) int64 {
	row := db.QueryRow(queryString, a...)
	var value sql.NullInt64
	err := row.Scan(&value)
	if err != nil {
		log.LogSeriousWithStack("Error fetch  data %v", err)
	}
	return value.Int64
}

// pretty json indent 4
func PFormat(data map[string]interface{}) string {
	ps, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err.Error()
	} else {
		return string(ps[:])
	}
}

//
type ByValue []int64

func (a ByValue) Len() int           { return len(a) }
func (a ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByValue) Less(i, j int) bool { return a[i] < a[j] }

func SortedInt64(list []int64) []int64 {
	result := make([]int64, 0)
	for _, e := range list {
		result = append(result, e)
	}
	sort.Sort(ByValue(result))
	return result
}
