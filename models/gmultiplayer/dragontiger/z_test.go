package dragontiger

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	//		z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/server"
	"github.com/vic/vic_go/zconfig"
)

func Test01(t *testing.T) {
	choices := []string{}
	for choice, _ := range MapChoiceToRate {
		choices = append(choices, choice)
	}
	sort.Strings(choices)
	bs, _ := json.Marshal(choices)
	fmt.Println(string(bs))
}

func Test02(t *testing.T) {
	dataCenterInstance := datacenter.NewDataCenter(
		zconfig.PostgresUsername, zconfig.PostgresPassword,
		zconfig.PostgresAddress, zconfig.PostgresDatabaseName,
		zconfig.RedisAddress)
	//	fmt.Println("cp 0")
	record.RegisterDataCenter(dataCenterInstance)
	player.RegisterDataCenter(dataCenterInstance)
	currency.RegisterDataCenter(dataCenterInstance)
	serverObj := server.NewServer()
	gamemini.RegisterServer(serverObj)
	player.RegisterServer(serverObj)

	uid1, uid2, uid3, uid4, uid5 :=
		int64(1973), int64(1974), int64(1975), int64(1976), int64(1977)
	_ = []int64{uid1, uid2, uid3, uid4, uid5}
	game := &CarGame{}
	game.Init(GAME_CODE, currency.Money, 0)
	match := &CarMatch{}
	game.InitMatch(match)
	game.SharedMatch = match

	time.Sleep(100 * time.Millisecond)
	var err error
	err = match.SendMove(map[string]interface{}{
		"UserId": uid1, "Choice": C_TIGER, "BetValue": 1000})
	fmt.Println("err1", err)
	err = match.SendMove(map[string]interface{}{
		"UserId": uid1, "Choice": C_DRAGON, "BetValue": 3000})
	fmt.Println("err2", err)
	err = match.SendMove(map[string]interface{}{
		"UserId": uid2, "Choice": C_DRAGON, "BetValue": 5000})
	fmt.Println("err3", err)
	select {}
}
