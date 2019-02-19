package models

import (
	"errors"
	"fmt"

	"strconv"
	//"strings"

	"github.com/garyburd/redigo/redis"

	"github.com/vic/vic_go/models/player"

	"github.com/vic/vic_go/utils"
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
			_, err = c.Do("SELECT", 5)
			if err != nil {
				c.Close()
				fmt.Println(err)
				return nil, err
			}
			return c, err
		},
	}
}

/*
	"get_u23_game":                   get_u23_game,
	"bet_u23_game":                   bet_u23_game,
*/
func get_u23_game(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	conn := RedisPool.Get()
	var reply interface{}

	var thang, hoa, thua, tongsonguoi int64

	reply, err = conn.Do("MGET", "u23THANG", "u23HOA", "u23THUA", "u23TOTAL")
	if err != nil {
		return nil, err
	}

	vis, isOk := reply.([]interface{})
	if !isOk {
		return nil, errors.New("Khong ro loi")
	}
	thang = 0
	if vis[0] != nil {
		replyB, isOk := vis[0].([]byte)
		if !isOk {
			return nil, errors.New("Khong ro loi")
		}
		thang, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			return nil, err
		}
	}
	hoa = 0
	if vis[1] != nil {
		replyB, isOk := vis[1].([]byte)
		if !isOk {
			return nil, errors.New("Khong ro loi")
		}
		hoa, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			return nil, err
		}
	}
	thua = 0
	if vis[2] != nil {
		replyB, isOk := vis[2].([]byte)
		if !isOk {
			return nil, errors.New("Khong ro loi")
		}
		thua, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			return nil, err
		}
	}
	tongsonguoi = 0
	if vis[3] != nil {
		replyB, isOk := vis[3].([]byte)
		if !isOk {
			return nil, errors.New("Khong ro loi")
		}
		tongsonguoi, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			return nil, err
		}
	}
	responseData = make(map[string]interface{})
	responseData["thang"] = thang
	responseData["hoa"] = hoa
	responseData["thua"] = thua
	responseData["tongsonguoi"] = tongsonguoi
	return responseData, nil
}
func bet_u23_game(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	return nil, errors.New("Hết thời gian dự đoán! Kết quả sẽ được chúng tôi cập nhật sớm nhất qua hòm thư trong ứng dụng Chơi Lớn! Cảm ơn!")
	door := utils.GetStringAtPath(data, "door")
	money := utils.GetInt64AtPath(data, "money")
	if money < 1000 {
		return nil, errors.New("Bạn phải cược tối thiểu 1000 KIM!")
	}
	if door != "THANG" && door != "THUA" && door != "HOA" {
		return nil, errors.New("Khong ro loi")
	}

	conn := RedisPool.Get()
	var reply interface{}

	var k string

	playObj, err := player.GetPlayer(playerId)
	if err != nil {
		fmt.Println("hihi 1")
		return nil, err
	}
	err = playObj.ChangeMoneyAndLog(-money, "money", false, "", "U23_GAME", "U23_GAME", "")
	if err != nil {
		fmt.Println("hihi 2")
		return nil, err
	}

	v := fmt.Sprintf("%v_%v", playerId, money)
	k = fmt.Sprintf("mylist_u23%v", door)
	reply, err = conn.Do("RPUSH", k, v)
	if err != nil {
		fmt.Println("hihi 4")
		return nil, err
	}
	k = fmt.Sprintf("u23%v", door)
	reply, err = conn.Do("GET", k)
	returnValue := int64(0)
	if reply != nil {
		replyB, isOk := reply.([]byte)
		if !isOk {
			fmt.Println("hihi 6")
			return nil, err
		}
		returnValue, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			fmt.Println("hihi 7")
			return nil, err
		}
	}
	v = fmt.Sprintf("%v", returnValue+money)
	reply, err = conn.Do("SET", k, v)
	if err != nil {
		fmt.Println("hihi 9")
		return nil, err
	}
	returnValue = int64(0)

	reply, err = conn.Do("GET", "u23TOTAL")
	if reply != nil {
		replyB, isOk := reply.([]byte)
		if !isOk {
			fmt.Println("hihi 10")
			return nil, err
		}
		returnValue, err = strconv.ParseInt(string(replyB), 10, 64)
		if err != nil {
			fmt.Println("hihi 11")
			return nil, err
		}
		returnValue += 1
		reply, err = conn.Do("INCR", "u23TOTAL")
		if err != nil {
			fmt.Println("hihi 3")
			return nil, err
		}
	} else {
		returnValue = 1
		_, _ = conn.Do("SET", "u23TOTAL", "1")
	}
	responseData = make(map[string]interface{})
	responseData[door] = v
	responseData["tongsonguoi"] = returnValue
	return responseData, nil
}
