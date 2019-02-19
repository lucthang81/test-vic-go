package record

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/vic/vic_go/datacenter"
)

const (
	PARTNER_IAP_ANDROID = "iapAndroid"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION" // common

	ACTION_MANDATORY_BET  = "ACTION_MANDATORY_BET"  // bacay2
	ACTION_JOIN_GROUP_BET = "ACTION_JOIN_GROUP_BET" // bacay2
	ACTION_JOIN_PAIR_BET  = "ACTION_JOIN_PAIR_BET"  // bacay2

	ACTION_EAT_CARD = "ACTION_EAT_CARD" // phom

	ACTION_SPIN = "ACTION_SPIN" // slot2, wheel

	ACTION_ADD_BET     = "ACTION_ADD_BET"     // taixiu
	ACTION_BET_AS_LAST = "ACTION_BET_AS_LAST" // taixiu

	ACTION_BUY_IN      = "ACTION_BUY_IN"      // poker
	ACTION_LEAVE_LOBBY = "ACTION_LEAVE_LOBBY" // poker

	ACTION_SLOTACP_COMPLETE_PICTURE = "ACTION_SLOTACP_COMPLETE_PICTURE" // slotacp

	ACTION_ADMIN_CHANGE  = "ACTION_ADMIN_CHANGE"
	ACTION_SMS_CHARGE    = "ACTION_SMS_CHARGE"
	ACTION_SEND_MONEY    = "ACTION_SEND_MONEY"
	ACTION_RECEIVE_MONEY = "ACTION_RECEIVE_MONEY"

	ACTION_PROMOTE                = "ACTION_PROMOTE"
	ACTION_PROMOTE_AGENCY         = "ACTION_PROMOTE_AGENCY"
	ACTION_PROMOTE_AGENCY_WEEKLY  = "ACTION_PROMOTE_AGENCY_WEEKLY"
	ACTION_PROMOTE_AGENCY_MONTHLY = "ACTION_PROMOTE_AGENCY_MONTHLY"
	ACTION_GIFT_LOGINS            = "ACTION_GIFT_LOGINS"
	ACTION_DAILY_FREE_SPIN        = "ACTION_DAILY_FREE_SPIN"

	ACTION_EVENT_PAY_LUCKY_NUMBER = "ACTION_EVENT_PAY_LUCKY_NUMBER"
	ACTION_EVENT_PAY_TOPS         = "ACTION_EVENT_PAY_TOPS"

	ACTION_CUSTOM_ROOM_PERIODIC_TAX = "ACTION_CUSTOM_ROOM_PERIODIC_TAX"
	ACTION_CUSTOM_ROOM_SET_MONEY    = "ACTION_CUSTOM_ROOM_SET_MONEY"

	ACTION_GIVE_STORE_TESTER = "GIVE_STORE_TESTER"
	ACTION_BUY_SHOP_ITEM     = "ACTION_BUY_SHOP_ITEM"

	ACTION_PAYTRUST88_CHARGING = "ACTION_PAYTRUST88_CHARGING"
)

var dataCenter *datacenter.DataCenter
var DataCenter *datacenter.DataCenter
var RedisPool *redis.Pool

func init() {
	RedisPool = &redis.Pool{
		MaxIdle:   20,
		MaxActive: 40, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			_, err = c.Do("SELECT", 6)
			if err != nil {
				c.Close()
				fmt.Println(err)
				return nil, err
			}
			return c, err
		},
	}
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	DataCenter = dataCenter
}

func LogStartActiveRecord(playerId int64, deviceCode string, deviceType string, ipAddress string) {
	logStartActiveRecord(playerId, deviceCode, deviceType, ipAddress)
}

func LogEndActiveRecord(playerId int64) {
	logEndActiveRecord(playerId)
}

func LogMatchRecord(gameCode string,
	currencyType string,
	requirement int64,
	bet int64,
	tax int64,
	win int64,
	lose int64,
	botWin int64,
	botLose int64,
	matchData map[string]interface{}) int64 {
	return logMatchRecord(gameCode, currencyType, requirement, bet, tax, win, lose, botWin, botLose, matchData)
}

func LogRefererIdForPurchase(playerId int64, purchaseType string, cardCode string, cardSerial string) int64 {
	return logRefererIdForPurchase(playerId, purchaseType, cardCode, cardSerial)
}

func LogTransactionIdRefererId(refererId int64, transactionId string) {
	logTransactionIdRefererId(refererId, transactionId)
}

func LogCurrencyRecord(playerId int64, action string, gameCode string, additionalData map[string]interface{}, currencyType string, valueBefore int64, valueAfter int64, change int64) {
	logCurrencyRecord(playerId, action, gameCode, additionalData, currencyType, valueBefore, valueAfter, change)
}

func LogCCU(totalOnline int, totalNormalOnline int, totalBotOnline int, onlineGameData map[string]interface{}) {
	logCCU(totalOnline, totalNormalOnline, totalBotOnline, onlineGameData)
}

func LogBankRecord(matchId int64, gameCode string, currencyType string, moneyBefore int64, moneyAfter int64) {
	logBankRecord(matchId, gameCode, currencyType, moneyBefore, moneyAfter)
}

func LogBankRecordByBot(playerId int64, gameCode string, currencyType string, moneyBefore int64, moneyAfter int64) {
	logBankRecordByBot(playerId, gameCode, currencyType, moneyBefore, moneyAfter)
}

func LogAdminActivity(id int64, possibleIps string) {
	logAdminActivity(id, possibleIps)
}

func RedisSaveFloat64(key string, value float64) {
	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("SET", key, value)
	if err != nil {
		fmt.Println("ERROR RedisSaveFloat64", err, key, value)
	}
	_ = reply
}

func RedisLoadFloat64(key string) float64 {
	var result float64

	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("GET", key)
	if err != nil {
		fmt.Println("ERROR RedisLoadFloat64", err, key)
		return result
	}
	replyB, _ := reply.([]byte)
	result, err = strconv.ParseFloat(string(replyB), 64)
	return result
}

func RedisSaveString(key string, value string) {
	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("SET", key, value)
	if err != nil {
		fmt.Println("ERROR RedisSaveString", err, key, value)
	}
	_ = reply
}

// timeout measures by seconds
func RedisSaveStringExpire(key string, value string, timeout int) {
	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("SET", key, value)
	if err != nil {
		fmt.Println("ERROR RedisSaveString", err, key, value)
	}
	_ = reply
	conn.Do("EXPIRE", key, timeout)
}

// timeout measures by seconds
func RedisDeleteKey(key string) {
	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("DEL", key)
	_, _ = reply, err
}

func RedisLoadString(key string) string {
	var result string

	conn := RedisPool.Get()
	defer conn.Close()
	var reply interface{}
	var err error

	reply, err = conn.Do("GET", key)
	if err != nil {
		fmt.Println("ERROR RedisLoadString", err, key)
		return result
	}
	replyB, _ := reply.([]byte)
	result = string(replyB)
	return result
}

func PsqlSaveString(key string, value string) {
	dataCenter.Db().Exec(
		"INSERT INTO zkey_value (zkey, zvalue, last_modified) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (zkey) DO UPDATE "+
			"  SET zvalue = EXCLUDED.zvalue, last_modified=EXCLUDED.last_modified ",
		key, value, time.Now().UTC())
}

func PsqlLoadString(key string) string {
	var result string
	row := dataCenter.Db().QueryRow(
		"SELECT zvalue FROM zkey_value WHERE zkey = $1",
		key)
	_ = row.Scan(&result)
	return result
}

func PsqlSaveGlobal(key string, value string) {
	dataCenter.Db().Exec(
		"INSERT INTO zglobal_var (zkey, zvalue, last_modified) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (zkey) DO UPDATE "+
			"  SET zvalue = EXCLUDED.zvalue, last_modified=EXCLUDED.last_modified ",
		key, value, time.Now().UTC())
}

func PsqlLoadGlobal(key string) string {
	var result string
	// fmt.Println("dataCenter", dataCenter)
	row := dataCenter.Db().QueryRow(
		"SELECT zvalue FROM zglobal_var WHERE zkey = $1",
		key)
	_ = row.Scan(&result)
	return result
}
