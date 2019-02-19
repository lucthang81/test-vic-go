package wheel2

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/currency"
	//"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/cardgame"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	//	"github.com/vic/vic_go/zconfig"
)

const (
	ACTION_STOP_GAME = "ACTION_STOP_GAME"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION"

	ACTION_RECEIVE_FREE_SPIN = "ACTION_RECEIVE_FREE_SPIN"
	ACTION_SPIN              = "ACTION_SPIN"
	ACTION_GET_HISTORY       = "ACTION_GET_HISTORY"

	DELAY_CHANGING_MONEY = 5 * time.Second
)

func init() {
	fmt.Print("")
	_ = jackpot.Jackpot{}
	_ = errors.New("")
	_, _ = json.Marshal([]int{})
}

type ResultOnePlayer struct {
	// playerId
	Id           int64
	Username     string
	ChangedMoney int64

	MatchId     string
	StartedTime time.Time

	WheelResult       []string
	WinningMoney      int64
	WinningTestMoney  int64
	WinningWheel2Spin int64
}

func (result1p *ResultOnePlayer) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"playerId":    result1p.Id,
		"username":    result1p.Username,
		"startedTime": result1p.StartedTime.Format(time.RFC3339),
		"matchId":     result1p.MatchId,

		"wheelResult":       result1p.WheelResult,
		"winningMoney":      result1p.WinningMoney,
		"winningTestMoney":  result1p.WinningTestMoney,
		"winningWheel2Spin": result1p.WinningWheel2Spin,
	}
	return result
}

func (result1p *ResultOnePlayer) String() string {
	bytes, _ := json.Marshal(result1p.Serialize())
	return string(bytes)
}

type WheelMatch struct {
	game        *WheelGame
	player      *player.Player
	startedTime time.Time
	matchId     string
	tax         int64

	wheelResult  []string
	playerResult *ResultOnePlayer

	mutex sync.RWMutex
}

type Action struct {
	actionName string
	playerId   int64

	data         map[string]interface{}
	chanResponse chan *ActionResponse
}

type ActionResponse struct {
	err  error
	data map[string]interface{}
}

func NewWheelMatch(
	wheelG *WheelGame, createdPlayer *player.Player, matchCounter int64) *WheelMatch {
	match := &WheelMatch{
		game:         wheelG,
		player:       createdPlayer,
		startedTime:  time.Now(),
		matchId:      fmt.Sprintf("%v_%v", wheelG.GameCode(), matchCounter),
		playerResult: &ResultOnePlayer{},
	}
	// init vars code in match here
	match.playerResult.Id = match.player.Id()
	match.playerResult.Username = match.player.Name()
	match.playerResult.MatchId = match.matchId
	match.playerResult.StartedTime = match.startedTime
	//
	go Start(match)
	return match
}

// match main flow
func Start(match *WheelMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	// khoảng cách giữa 2 lần quay
	time.Sleep(500 * time.Millisecond)
	//
	isNewUser := 0 <= match.player.GetMoney(currency.SlotSpin100) &&
		match.player.GetMoney(currency.SlotSpin100) <= 2
	isRichUser := false
	top.GlobalMutex.Lock()
	event := top.MapEvents[top.EVENT_CHARGING_MONEY]
	top.GlobalMutex.Unlock()
	if event != nil {
		_, chargeInWeek := event.GetPosAndValue(match.player.Id())
		weekday := int64(time.Now().Weekday()) // Sunday = 0
		if weekday == 0 {
			weekday = 7
		}
		//		averageChargeInDay := chargeInWeek / weekday
		if chargeInWeek >= 50000 {
			isRichUser = true
		}
	}

	if isRichUser {
		if match.game.balance > -match.game.yesterday10PercentAllProfit/24 {
			match.wheelResult = LimitedSpin("c", "l")
		} else {
			match.wheelResult = RandomSpin()
		}
	} else if isNewUser {
		match.wheelResult = LimitedSpin("a", "b")
	} else {
		match.wheelResult = LimitedSpin("a", "b")
	}
	match.playerResult.WheelResult = match.wheelResult
	match.game.SendDataToPlayerId("Wheel2Result", match.SerializedData(), match.player.Id())
	// _________________________________________________________________________
	// lưu trận đấu vào database
	// ...
	//
	w0symbol, w1symbol := match.wheelResult[0], match.wheelResult[1]
	var changedCurrency string
	var changedAmount int64
	if w0symbol == "1" ||
		w0symbol == "2" {
		changedCurrency = currency.TestMoney
		changedAmount = 500
	} else if w0symbol == "3" ||
		w0symbol == "4" {
		changedCurrency = currency.TestMoney
		changedAmount = 5000
	} else if w0symbol == "5" ||
		w0symbol == "6" {
		changedCurrency = currency.TestMoney
		changedAmount = 10000
	} else if w0symbol == "7" ||
		w0symbol == "8" {
		changedCurrency = currency.TestMoney
		changedAmount = 20000
	} else if w0symbol == "9" {
		changedCurrency = currency.TestMoney
		changedAmount = 50000
	} else if w0symbol == "10" {
		changedCurrency = currency.TestMoney
		changedAmount = 500000
	} else if w0symbol == "11" {
		changedCurrency = currency.Wheel2Spin
		changedAmount = 1
	} else {
		changedCurrency = currency.Wheel2Spin
		changedAmount = 2
	}
	if changedCurrency == currency.TestMoney {
		match.playerResult.WinningTestMoney = changedAmount
	} else {
		match.playerResult.WinningWheel2Spin = changedAmount
	}

	go func(changedCurrency string, changedAmount int64) {
		time.Sleep(DELAY_CHANGING_MONEY)
		if changedCurrency == currency.TestMoney {
			match.player.ChangeMoneyAndLog(
				changedAmount, changedCurrency, false, "",
				ACTION_FINISH_SESSION, match.game.GameCode(), match.matchId)
		}
		if changedCurrency == currency.Wheel2Spin {
			match.player.ChangeMoneyAndLog(
				changedAmount, changedCurrency, false, "",
				ACTION_FINISH_SESSION, match.game.GameCode(), match.matchId)
		}
	}(changedCurrency, changedAmount)

	//
	if w1symbol == "a" {
		changedCurrency = currency.Money
		changedAmount = 20
	} else if w1symbol == "b" {
		changedCurrency = currency.Money
		changedAmount = 50
	} else if w1symbol == "c" {
		changedCurrency = currency.Money
		changedAmount = 1000
	} else if w1symbol == "d" {
		changedCurrency = currency.Money
		changedAmount = 2500
	} else if w1symbol == "e" {
		changedCurrency = currency.Money
		changedAmount = 5000
	} else if w1symbol == "f" {
		changedCurrency = currency.Money
		changedAmount = 10000
	} else if w1symbol == "g" {
		changedCurrency = currency.Money
		changedAmount = 50000
	} else if w1symbol == "h" {
		changedCurrency = currency.Money
		changedAmount = 100000
	} else if w1symbol == "i" {
		changedCurrency = currency.Money
		changedAmount = 200000
	} else if w1symbol == "j" {
		changedCurrency = currency.Money
		changedAmount = 500000
	} else if w1symbol == "k" {
		changedCurrency = currency.Money
		changedAmount = 1000000
	} else if w1symbol == "l" {
		changedCurrency = currency.Money
		changedAmount = 1500000
	} else {
		changedCurrency = currency.Money
		changedAmount = 0
	}
	match.playerResult.WinningMoney = changedAmount
	go func(changedCurrency string, changedAmount int64) {
		time.Sleep(DELAY_CHANGING_MONEY)
		if (changedCurrency == currency.Money) && (changedAmount > 0) {
			match.player.ChangeMoneyAndLog(
				changedAmount, changedCurrency, false, "",
				ACTION_FINISH_SESSION, match.game.GameCode(), match.matchId)
			match.game.balance -= changedAmount
		}
	}(changedCurrency, changedAmount)
	//
	match.playerResult.WheelResult = match.wheelResult
	match.game.SendDataToPlayerId("Wheel2Result", match.SerializedData(), match.player.Id())
	// LogMatchRecord2
	playerIpAdds := map[int64]string{match.player.Id(): match.player.IpAddress()}
	playerObj := match.player
	playerIpAdds[playerObj.Id()] = playerObj.IpAddress()
	playerResults := make([]map[string]interface{}, 0)
	record.LogMatchRecord2(
		match.game.GameCode(), currency.Money, 0, 0,
		changedAmount, 0, 0, 0,
		match.matchId, playerIpAdds,
		playerResults)
	// cập nhật lịch sửa 10 ván chơi gần nhất
	match.game.mutex.Lock()
	if _, isIn := match.game.mapPlayerIdToHistory[match.player.Id()]; !isIn {
		temp := cardgame.NewSizedList(20)
		match.game.mapPlayerIdToHistory[match.player.Id()] = &temp
	}
	match.game.mapPlayerIdToHistory[match.player.Id()].Append(
		match.playerResult.String())
	match.game.mutex.Unlock()
	// cập nhật danh sách thắng lớn
	if changedAmount >= 50000 {
		match.game.mutex.Lock()
		match.game.bigWinList.Append(match.playerResult.String())
		match.game.mutex.Unlock()
	}

	match.game.mutex.Lock()
	delete(match.game.mapPlayerIdToMatch, match.player.Id())
	match.game.mutex.Unlock()
}

//
func (match *WheelMatch) GameCode() string {
	return match.game.GameCode()
}

func (match *WheelMatch) CurrencyType() string {
	return match.game.CurrencyType()
}

// json obj represent general match info
func (match *WheelMatch) SerializedData() map[string]interface{} {
	result := match.playerResult.Serialize()
	result["gameCode"] = match.game.gameCode
	result["currencyType"] = match.game.currencyType
	return result
}

// unique data for specific player
func (match *WheelMatch) SerializedDataForPlayer(playerId int64) map[string]interface{} {
	return map[string]interface{}{}
}
