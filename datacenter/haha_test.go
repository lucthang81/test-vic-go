package datacenter

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
}

func TestHaha(t *testing.T) {
	d := NewDataCenter("vic_user", "123qwe", "casino_vic_db", "127.0.0.1:6379")
	d.GetCardsFix("anh truong")
}
