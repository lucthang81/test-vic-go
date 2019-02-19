package datacenter

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"math/rand"
	"strconv"
)

const layout = "2006-01-02T15:04:05Z"
const SqlTimeNow string = "now() at time zone 'utc'"

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type DataCenter struct {
	db         DBInterface
	redisCache *redisCache
	// conversations      map[int64]*Conversation
	// gameAccounts       map[int64]*GameAccount
	// gameAccountsOnline map[int64]*GameAccount

	hiddenDB DBInterface
}

func NewDataCenter(
	username string, password string,
	postgresAddress string, databaseName string,
	redisAddress string,
) (dataCenter *DataCenter) {
	fmt.Printf("Initializing dbPool")
	dataSource := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		username, password,
		postgresAddress, databaseName)
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.LogSerious(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(40)
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.LogSerious(err.Error())
	}
	fmt.Printf("Initialized dbPool %+v\n", dataSource)

	dataCenter = &DataCenter{
		db: db,
	}
	if redisAddress != "" {
		fmt.Print(redisAddress)
		redisCache := startRedis(redisAddress)
		dataCenter.redisCache = redisCache
	}
	return dataCenter
}

func (dataCenter *DataCenter) Db() (db DBInterface) {
	return dataCenter.db
}

func (dataCenter *DataCenter) HideDBForTesting() {
	if dataCenter.db != nil {
		dataCenter.hiddenDB = dataCenter.db
		dataCenter.db = &TestDB{}
	}
}

func (dataCenter *DataCenter) StopHideDBForTesting() {
	if dataCenter.hiddenDB != nil {
		dataCenter.db = dataCenter.hiddenDB
	}
}

func (dataCenter *DataCenter) SaveKeyValueToCache(group string, key string, value string) (err error) {
	return dataCenter.redisCache.saveGroupKeyValue(group, key, value)
}

func (dataCenter *DataCenter) LoadAllKeysFromCache(group string) (keys []string, err error) {
	return dataCenter.redisCache.loadAllKeysInGroup(group)
}

func (dataCenter *DataCenter) LoadValueFromCache(group string, key string) (value string, err error) {
	return dataCenter.redisCache.loadGroupKey(group, key)
}

func (dataCenter *DataCenter) RemoveKeyValueFromCache(group string, key string) (err error) {
	return dataCenter.redisCache.removeGroupKey(group, key)
}

func (dataCenter *DataCenter) SaveObject(dataObject DataObject, keys []string, values []interface{}, shouldCache bool) (err error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("UPDATE %s SET ", dataObject.DatabaseTableName()))
	for index, key := range keys {
		buffer.WriteString(fmt.Sprintf("%s = $%d", key, index+1))
		if index != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(fmt.Sprintf(" WHERE id = $%d", len(keys)+1))
	values = append(values, dataObject.Id())
	_, err = dataCenter.db.Exec(buffer.String(), values...)
	if err != nil {
		return err
	}
	// cache
	if shouldCache {
		err = dataCenter.redisCache.saveMany(dataObject, keys, utils.GetStringSliceFromInterfaceSlice(values))
	}
	return err
}

func (dataCenter *DataCenter) InsertObject(dataObject DataObject, keys []string, values []interface{}, shouldCache bool) (id int64, err error) {
	var keyBuffer bytes.Buffer
	var valueBuffer bytes.Buffer
	keyBuffer.WriteString(fmt.Sprintf("INSERT INTO %s (", dataObject.DatabaseTableName()))
	valueBuffer.WriteString(") VALUES (")
	for index, key := range keys {
		keyBuffer.WriteString(fmt.Sprintf("%s", key))
		valueBuffer.WriteString(fmt.Sprintf("$%d", index+1))
		if index != len(keys)-1 {
			keyBuffer.WriteString(", ")
			valueBuffer.WriteString(", ")
		}
	}
	valueBuffer.WriteString(") RETURNING id;")
	keyBuffer.WriteString(valueBuffer.String())
	err = dataCenter.db.QueryRow(keyBuffer.String(), values...).Scan(&id)
	if err != nil {
		return 0, err
	}
	dataObject.SetId(id)
	// cache
	if shouldCache {
		err = dataCenter.redisCache.saveMany(dataObject, keys, utils.GetStringSliceFromInterfaceSlice(values))
	}
	return id, err
}
func (dataCenter *DataCenter) InsertObject2(dataObject DataObject, keys []string, values []interface{}, shouldCache bool) (id int64, err error) {
	var keyBuffer bytes.Buffer
	var valueBuffer bytes.Buffer
	keyBuffer.WriteString(fmt.Sprintf("INSERT INTO %s (", dataObject.DatabaseTableName()))
	valueBuffer.WriteString(") VALUES (")
	for index, key := range keys {
		keyBuffer.WriteString(fmt.Sprintf("%s", key))
		valueBuffer.WriteString(fmt.Sprintf("$%d", index+1))
		if index != len(keys)-1 {
			keyBuffer.WriteString(", ")
			valueBuffer.WriteString(", ")
		}
	}
	valueBuffer.WriteString(") RETURNING id;")
	keyBuffer.WriteString(valueBuffer.String())
	err = dataCenter.db.QueryRow(keyBuffer.String(), values...).Scan(&id)
	if err != nil {
		return 0, err
	}
	dataObject.SetId(id)
	// cache
	if shouldCache {
		err = dataCenter.redisCache.saveMany(dataObject, keys, utils.GetStringSliceFromInterfaceSlice(values))
	}
	return id, err
}

func (dataCenter *DataCenter) RemoveObject(dataObject DataObject) (err error) {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE id = $1", dataObject.DatabaseTableName())
	_, err = dataCenter.db.Exec(queryString, dataObject.Id())
	return err
}

func (dataCenter *DataCenter) getObjectFromDbWithSelectedKeys(dataObject DataObject, keys []string, value []interface{}) (err error) {
	var buffer bytes.Buffer
	for index, key := range keys {
		buffer.WriteString(key)
		if index != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1;", buffer.String(), dataObject.DatabaseTableName())
	row := dataCenter.db.QueryRow(query, dataObject.Id())
	return row.Scan(value...)
}

func (dataCenter *DataCenter) getStringFromDbWithKey(dataObject DataObject, key string) (value string, err error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1;", key, dataObject.DatabaseTableName())
	row := dataCenter.db.QueryRow(query, dataObject.Id())
	var sqlValue sql.NullString
	err = row.Scan(&sqlValue)
	if err != nil {
		return "", err
	}
	return sqlValue.String, nil
}

func (dataCenter *DataCenter) getInt64FromDbWithKey(dataObject DataObject, key string) (value int64, err error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1;", key, dataObject.DatabaseTableName())
	row := dataCenter.db.QueryRow(query, dataObject.Id())
	var sqlValue sql.NullInt64
	err = row.Scan(&sqlValue)
	if err != nil {
		return 0, err
	}
	return sqlValue.Int64, nil
}

func (dataCenter *DataCenter) GetStringFieldForObject(dataObject DataObject, key string, canHitCache bool) (value string, hitCache bool, err error) {
	if canHitCache {
		value, err = dataCenter.redisCache.get(dataObject, key)
		if err != nil {
			return "", true, err
		}
	}
	if value != "" {
		return value, true, nil
	}

	// hit database
	value, err = dataCenter.getStringFromDbWithKey(dataObject, key)
	if err != nil {
		return "", false, err
	}

	// save to cache
	if value != "" {
		dataCenter.redisCache.save(dataObject, key, value)
	}
	return value, false, nil
}

func (dataCenter *DataCenter) GetInt64FieldForObject(dataObject DataObject, key string, canHitCache bool) (value int64, hitCache bool, err error) {
	var cacheValue string
	if canHitCache {
		cacheValue, err = dataCenter.redisCache.get(dataObject, key)
		if err != nil {
			return 0, true, err
		}
	}
	if cacheValue != "" {
		value, err = strconv.ParseInt(cacheValue, 10, 64)
		if err != nil {
			return 0, true, err
		}
		return value, true, nil
	}

	// hit database
	value, err = dataCenter.getInt64FromDbWithKey(dataObject, key)
	if err != nil {
		return 0, false, err
	}

	// save to cache
	dataCenter.redisCache.save(dataObject, key, fmt.Sprintf("%d", value))
	return value, false, nil
}

func (dataCenter *DataCenter) GetInt64FromQuery(queryString string, a ...interface{}) int64 {
	row := dataCenter.Db().QueryRow(queryString, a...)
	var value sql.NullInt64
	row.Scan(&value)
	return value.Int64
}

func (dataCenter *DataCenter) IsObjectExist(dataObject DataObject, keys []string, values []interface{}, canHitCache bool) (exist bool, hitCache bool, err error) {
	if canHitCache {
		cacheValues, err := dataCenter.redisCache.getMany(dataObject, keys)
		if err != nil {
			return false, true, err
		}
		same := true
		for index, interfaceValue := range values {
			strValue := utils.GetStringFromInterface(interfaceValue)
			if strValue != cacheValues[index] {
				same = false
				break
			}
		}
		if same == true {
			return true, true, nil
		}
	}

	// hit db
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("SELECT id FROM %s WHERE ", dataObject.DatabaseTableName()))
	for index, key := range keys {
		buffer.WriteString(fmt.Sprintf("%s = $%d", key, index+1))
		if index != len(keys)-1 {
			buffer.WriteString("AND ")
		}
	}
	buffer.WriteString(" LIMIT 1;")
	var id int64
	err = dataCenter.db.QueryRow(buffer.String(), values...).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}
	dataObject.SetId(id)
	return true, false, nil
}

func (dataCenter *DataCenter) FlushCache() {
	dataCenter.redisCache.do("FLUSHDB")
}

func (dataCenter *DataCenter) GetCardsFix(key string) (value []byte, err error) {
	reply, err := dataCenter.redisCache.do("GET", key)
	if err != nil {
		return []byte{}, err
	}
	resultBytes, isOk := reply.([]byte)
	if isOk {
		return resultBytes, nil
	} else {
		return []byte{}, errors.New("result type is not bytes")
	}
}

func (dataCenter *DataCenter) GetOtpPhone() (value string, err error) {
	//random 1 so phone de gui otp
	/*
		reply, err := redisCache.do("GET", "TestKey")
		c.Assert(err, IsNil)
		c.Assert(string(reply.([]byte)), Equals, "TestKey11")
	*/
	//redisCache1 := startRedis("127.0.0.1:6379")
	reply1, err1 := dataCenter.redisCache.do("GET", "otpPhones1")
	if err1 != nil {
		return "", err1
	}
	data := []string{}
	err = json.Unmarshal(reply1.([]byte), &data)

	if err != nil || len(data) == 0 {
		return "", err
	}
	n := rand.Int() % len(data)
	// fmt.Printf("truong %v", data[n])
	return data[n], nil
}
