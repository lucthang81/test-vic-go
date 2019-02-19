package record

import (
	"fmt"
	"testing"
	"time"

	"github.com/vic/vic_go/datacenter"
)

func Test1(t *testing.T) {
	fmt.Println("")
	key := "a"
	RedisSaveFloat64(key, 5.6)
	lv := RedisLoadFloat64(key)
	if lv != 5.6 {
		t.Error(lv)
	}
}

func Test2(t *testing.T) {
	key := "b"
	value := "clgt"
	RedisSaveString(key, value)
	lv := RedisLoadString(key)
	if lv != value {
		t.Error(lv)
	}
}

func Test3(t *testing.T) {
	d := datacenter.NewDataCenter(
		"vic_user", "123qwe", ":5432", "casino_vic_db", ":6379")
	RegisterDataCenter(d)

	key := "b"
	value := "clgt"
	PsqlSaveString(key, value)
	lv := PsqlLoadString(key)
	if lv != value {
		fmt.Println("lv", lv)
		t.Error(lv)
	}
	RedisDeleteKey("b")
}

func Test4(t *testing.T) {
	key := "c"
	value := "1.7"
	RedisSaveStringExpire(key, value, 1)
	time.Sleep(500 * time.Millisecond)
	if RedisLoadFloat64(key) != float64(1.7) {
		t.Error()
	}
	time.Sleep(1000 * time.Millisecond)
	if RedisLoadString(key) != "" {
		t.Error()
	}
}
