package event_player

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var RedisPool *redis.Pool

func init() {
	//
	_, _ = strconv.ParseInt("15", 10, 64)
	_, _ = time.Parse(time.RFC3339, "2017-01-20T00:00:00+07:00")
	//
	RedisPool = &redis.Pool{
		MaxIdle:   2000,
		MaxActive: 4000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			_, err = c.Do("SELECT", 4)
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
func SaveEvent(event *EventCollectingPieces) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	k = fmt.Sprintf("%v:%v", event.EventName, "StartingTime")
	reply, err = conn.Do("SET", k, event.StartingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "FinishingTime")
	reply, err = conn.Do("SET", k, event.FinishingTime.Format(time.RFC3339))

	k = fmt.Sprintf("%v:%v", event.EventName, "nPiecesToComplete")
	reply, err = conn.Do("SET", k, event.nPiecesToComplete)

	k = fmt.Sprintf("%v:%v", event.EventName, "NLimitPrizes")
	reply, err = conn.Do("SET", k, event.NLimitPrizes)

	k = fmt.Sprintf("%v:%v", event.EventName, "nRarePieces")
	reply, err = conn.Do("SET", k, event.nRarePieces)

	k = fmt.Sprintf("%v:%v", event.EventName, "TotalPrize")
	reply, err = conn.Do("SET", k, event.TotalPrize)

	k = fmt.Sprintf("%v:%v", event.EventName, "ChanceToDropRarePiece")
	reply, err = conn.Do("SET", k, event.ChanceToDropRarePiece)

	conn.Close()

	_ = fmt.Sprintf("%v %v", reply, err)
}

// only call when the program start
func LoadEvent(eventName string) *EventCollectingPieces {
	conn := RedisPool.Get()
	var reply interface{}
	var err error
	var k string

	event := &EventCollectingPieces{
		EventName:         eventName,
		MapPidToMapPieces: map[int64]map[int]int{},
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
		fmt.Println("checkpoint 1")
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
		fmt.Println("checkpoint 2")
		return nil
	}
	event.FinishingTime = finishingTime
	//
	k = fmt.Sprintf("%v:%v", event.EventName, "nPiecesToComplete")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	temp, _ := strconv.Atoi(string(replyB))
	if temp < 2 {
		fmt.Println("checkpoint 3")
		return nil
	}
	event.nPiecesToComplete = temp
	//
	k = fmt.Sprintf("%v:%v", event.EventName, "NLimitPrizes")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		return nil
	}
	temp1, _ := strconv.Atoi(string(replyB))
	event.NLimitPrizes = temp1
	//
	k = fmt.Sprintf("%v:%v", event.EventName, "nRarePieces")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		fmt.Printf("checkpoint 4 %v %T", reply, reply)
		return nil
	}
	temp2, _ := strconv.Atoi(string(replyB))
	event.nRarePieces = temp2
	//
	k = fmt.Sprintf("%v:%v", event.EventName, "TotalPrize")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		fmt.Printf("checkpoint 4 %v %T", reply, reply)
		return nil
	}
	temp3, _ := strconv.ParseInt(string(replyB), 10, 64)
	event.TotalPrize = temp3
	//
	k = fmt.Sprintf("%v:%v", event.EventName, "ChanceToDropRarePiece")
	reply, err = conn.Do("GET", k)
	replyB, isOk = reply.([]byte)
	if !isOk {
		fmt.Printf("checkpoint 4 %v %T", reply, reply)
		return nil
	}
	temp4, _ := strconv.ParseFloat(string(replyB), 64)
	event.ChanceToDropRarePiece = temp4
	//
	event.MapPidToMapPieces = map[int64]map[int]int{}
	pattern := fmt.Sprintf("%v:%v:*", eventName, "MapPidToMapPieces")
	reply, err = conn.Do("KEYS", pattern)
	replySI, isOk := reply.([]interface{})
	if !isOk {
		return nil
	} else {
		// key = eventName:MapPidToMapPieces:pid
		keys := make([]string, 0)
		for _, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {
				return nil
			} else {
				key := string(eB)
				keys = append(keys, key)
			}
		}
		for _, k := range keys {
			temp := strings.LastIndex(k, ":")
			pidS := k[temp+1:]
			pid, _ := strconv.ParseInt(pidS, 10, 64)
			event.MapPidToMapPieces[pid] = map[int]int{}

			reply, err = conn.Do("HGETALL", k)
			replySI, isOk := reply.([]interface{})
			if !isOk {
				return nil
			} else {
				var piece, nPiece int
				for i, e := range replySI {
					eB, isOk := e.([]byte)
					if !isOk {
						return nil
					} else {
						if i%2 == 0 {
							piece, _ = strconv.Atoi(string(eB))
						} else {
							nPiece, _ = strconv.Atoi(string(eB))
							event.MapPidToMapPieces[pid][piece] = nPiece
						}
					}
				}
			}
		}
	}

	conn.Close()
	_ = fmt.Sprintf("%v %v", reply, err)

	return event
}

// TODO
// free memory redis
func ClearEvent(eventName string) {
	conn := RedisPool.Get()
	var reply interface{}
	var err error

	pattern := fmt.Sprintf("%v:*", eventName)
	reply, err = conn.Do("KEYS", pattern)
	replySI, isOk := reply.([]interface{})
	if !isOk {

	} else {
		keys := make([]string, 0)
		for _, e := range replySI {
			eB, isOk := e.([]byte)
			if !isOk {

			} else {
				key := string(eB)
				keys = append(keys, key)
			}
		}
		for _, k := range keys {
			reply, err = conn.Do("DEL", k)
		}
	}

	_ = fmt.Sprintf("%v %v", reply, err)

	conn.Close()
}

// save pieces for player,
// call when system drop a piece,
// already run in a goroutine
func Persist(eventName string, pid int64, piece int, nPiece int) {
	go func() {
		conn := RedisPool.Get()
		var reply interface{}
		var err error
		var k string

		k = fmt.Sprintf("%v:%v:%v", eventName, "MapPidToMapPieces", pid)
		reply, err = conn.Do("HSET", k, piece, nPiece)
		if err != nil {
			fmt.Println("ERROR Persist 4", reply, err)
		}

		conn.Close()
	}()
}

// call when system drop a rare piece
func Persist2(eventName string, nRarePieces int) {
	go func() {
		conn := RedisPool.Get()
		var reply interface{}
		var err error
		var k string

		k = fmt.Sprintf("%v:%v", eventName, "nRarePieces")
		reply, err = conn.Do("SET", k, nRarePieces)
		if err != nil {
			fmt.Println("ERROR Persist 5", reply, err)
		}

		conn.Close()
	}()
}
