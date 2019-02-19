package event

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var RedisPool *redis.Pool

func init() {
	RedisPool = &redis.Pool{
		MaxIdle:   2000,
		MaxActive: 4000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			_, err = c.Do("SELECT", 3)
			if err != nil {
				c.Close()
				fmt.Println(err)
				return nil, err
			}
			return c, err
		},
	}
}

// save event infos to redis,
// only call on init event
func SaveEvent(event *EventTop) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	k = fmt.Sprintf("%v:%v", event.EventName, "EventName")
	reply, err = conn.Do("SET", k, event.EventName)

	k = fmt.Sprintf("%v:%v", event.EventName, "StartingTime")
	reply, err = conn.Do("SET", k, event.StartingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "FinishingTime")
	reply, err = conn.Do("SET", k, event.FinishingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "MapPositionToPrize")
	for f, v := range event.MapPositionToPrize {
		reply, err = conn.Do("HSET", k, f, v)
	}

	conn.Close()

	_ = fmt.Sprintf("%v %v", reply, err)
}

// only call when the program start
func LoadEvent(eventName string) *EventTop {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	event := &EventTop{
		EventName:                 eventName,
		MapPositionToPrize:        make(map[int]int64),
		MapPlayerIdToValue:        make(map[int64]int64),
		MapPlayerIdToCreatedTime:  make(map[int64]time.Time),
		MapPlayerIdToCurrentValue: make(map[int64]int64),
	}
	//
	k = fmt.Sprintf("%v:%v", eventName, "StartingTime")
	reply, err = conn.Do("GET", k)
	replyB, isOk := reply.([]byte)
	if !isOk {
		return nil
	}
	startingTime, err := time.Parse(time.RFC3339, string(replyB))
	if err != nil {
		return nil
	}
	event.StartingTime = startingTime
	//
	k = fmt.Sprintf("%v:%v", eventName, "FinishingTime")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	finishingTime, err := time.Parse(time.RFC3339, string(replyB))
	if err != nil {
		return nil
	}
	event.FinishingTime = finishingTime
	//
	k = fmt.Sprintf("%v:%v", eventName, "MapPositionToPrize")
	reply, err = conn.Do("HGETALL", k)
	replySI, isOk := reply.([]interface{})
	if !isOk {
		return nil
	} else {
		var pos int
		var prize int64
		for i, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				if i%2 == 0 {
					pos, _ = strconv.Atoi(string(eB))
				} else {
					prize, _ = strconv.ParseInt(string(eB), 10, 64)
					event.MapPositionToPrize[pos] = prize
				}
			}
		}
	}
	//
	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
	reply, err = conn.Do("HGETALL", k)
	replySI, isOk = reply.([]interface{})
	if !isOk {
		return nil
	} else {
		var pid int64
		var value int64
		for i, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				if i%2 == 0 {
					pid, _ = strconv.ParseInt(string(eB), 10, 64)
				} else {
					value, _ = strconv.ParseInt(string(eB), 10, 64)
					event.MapPlayerIdToValue[pid] = value
				}
			}
		}
	}
	//
	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCreatedTime")
	reply, err = conn.Do("HGETALL", k)
	replySI, isOk = reply.([]interface{})
	if !isOk {
		return nil
	} else {
		var pid int64
		var createdTime time.Time
		for i, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				if i%2 == 0 {
					pid, _ = strconv.ParseInt(string(eB), 10, 64)
				} else {
					createdTime, _ = time.Parse(time.RFC3339, string(eB))
					event.MapPlayerIdToCreatedTime[pid] = createdTime
				}
			}
		}
	}
	//
	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCurrentValue")
	reply, err = conn.Do("HGETALL", k)
	replySI, isOk = reply.([]interface{})
	if !isOk {
		return nil
	} else {
		var pid int64
		var value int64
		for i, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				if i%2 == 0 {
					pid, _ = strconv.ParseInt(string(eB), 10, 64)
				} else {
					value, _ = strconv.ParseInt(string(eB), 10, 64)
					event.MapPlayerIdToCurrentValue[pid] = value
				}
			}
		}
	}
	//
	conn.Close()
	_ = fmt.Sprintf("%v %v", reply, err)

	return event
}

// free memory redis
func ClearEvent(eventName string) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	k = fmt.Sprintf("%v:%v", eventName, "EventName")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "StartingTime")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "FinishingTime")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "MapPositionToPrize")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCreatedTime")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCurrentValue")
	reply, err = conn.Do("DEL", k)

	_ = fmt.Sprintf("%v %v", reply, err)

	conn.Close()
}

// save value for player in event top,
// already run in a goroutine
func Persist(
	eventName string, pid int64,
	value int64, createdTime time.Time, currentValue int64) {
	go func() {
		conn := RedisPool.Get()
		var reply interface{}
		var err error
		var k string

		k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
		reply, err = conn.Do("HSET", k, pid, value)
		if err != nil {
			fmt.Println("ERROR Persist", reply, err)
		}

		k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCreatedTime")
		reply, err = conn.Do("HSET", k, pid, createdTime.Format(time.RFC3339))
		if err != nil {
			fmt.Println("ERROR Persist 1", reply, err)
		}

		k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToCurrentValue")
		reply, err = conn.Do("HSET", k, pid, currentValue)
		if err != nil {
			fmt.Println("ERROR Persist 2", reply, err)
		}

		conn.Close()
	}()
}

// save event infos to redis,
// only call on init event
func SaveEventSC(event *EventSC) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	k = fmt.Sprintf("%v:%v", event.EventName, "EventName")
	reply, err = conn.Do("SET", k, event.EventName)

	k = fmt.Sprintf("%v:%v", event.EventName, "StartingTime")
	reply, err = conn.Do("SET", k, event.StartingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "FinishingTime")
	reply, err = conn.Do("SET", k, event.FinishingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "TimeUnit")
	reply, err = conn.Do("SET", k, int64(event.TimeUnit))

	k = fmt.Sprintf("%v:%v", event.EventName, "LimitNOBonus")
	reply, err = conn.Do("SET", k, event.LimitNOBonus)

	conn.Close()

	_ = fmt.Sprintf("%v %v", reply, err)
}

//
func LoadEventSC(eventName string) *EventSC {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	event := &EventSC{
		EventName:          eventName,
		MapPlayerIdToValue: make(map[int64]int64),
	}
	//
	k = fmt.Sprintf("%v:%v", eventName, "StartingTime")
	reply, err = conn.Do("GET", k)
	replyB, isOk := reply.([]byte)
	if !isOk {
		return nil
	}
	startingTime, err := time.Parse(time.RFC3339, string(replyB))
	if err != nil {
		return nil
	}
	event.StartingTime = startingTime
	//
	k = fmt.Sprintf("%v:%v", eventName, "FinishingTime")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	finishingTime, err := time.Parse(time.RFC3339, string(replyB))
	if err != nil {
		return nil
	}
	event.FinishingTime = finishingTime
	//
	k = fmt.Sprintf("%v:%v", eventName, "TimeUnit")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	timeUnit, err := strconv.ParseInt(string(replyB), 10, 64)
	if err != nil {
		return nil
	}
	event.TimeUnit = time.Duration(timeUnit)
	//
	k = fmt.Sprintf("%v:%v", eventName, "LimitNOBonus")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	limitBonus, err := strconv.ParseInt(string(replyB), 10, 64)
	if err != nil {
		return nil
	}
	event.LimitNOBonus = limitBonus
	//
	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
	reply, err = conn.Do("HGETALL", k)
	replySI, isOk := reply.([]interface{})
	if !isOk {
		return nil
	} else {
		var pid int64
		var value int64
		for i, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				if i%2 == 0 {
					pid, _ = strconv.ParseInt(string(eB), 10, 64)
				} else {
					value, _ = strconv.ParseInt(string(eB), 10, 64)
					event.MapPlayerIdToValue[pid] = value
				}
			}
		}
	}

	//
	conn.Close()
	_ = fmt.Sprintf("%v %v", reply, err)

	return event
}

// free memory redis
func ClearEventSC(eventName string) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	k = fmt.Sprintf("%v:%v", eventName, "EventName")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "StartingTime")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "FinishingTime")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "TimeUnit")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "LimitNOBonus")
	reply, err = conn.Do("DEL", k)

	k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
	reply, err = conn.Do("DEL", k)

	_ = fmt.Sprintf("%v %v", reply, err)

	conn.Close()
}

// mark user hit event condition,
// already run in a goroutine
func PersistSC(
	eventName string, pid int64,
	value int64) {
	go func() {
		conn := RedisPool.Get()
		var reply interface{}
		var err error
		var k string

		k = fmt.Sprintf("%v:%v", eventName, "MapPlayerIdToValue")
		reply, err = conn.Do("HSET", k, pid, value)
		if err != nil {
			fmt.Println("ERROR Persist 3", reply, err)
		}

		conn.Close()
	}()
}
