package slotbacay

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	//"github.com/vic/vic_go/models/currency"
	//"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/models/cardgame"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
)

const (
	ACTION_STOP_GAME = "ACTION_STOP_GAME"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION"

	ACTION_CHOOSE_MONEY_PER_LINE = "ACTION_CHOOSE_MONEY_PER_LINE"
	ACTION_GET_HISTORY           = "ACTION_GET_HISTORY"
	ACTION_SPIN                  = "ACTION_SPIN"
)

func init() {
	fmt.Print("")
	_ = jackpot.Jackpot{}
	_ = errors.New("")
	_, _ = json.Marshal([]int{})
	_ = rand.Intn(10)
}

type ResultOnePlayer struct {
	// playerId
	Id           int64
	Username     string
	ChangedMoney int64

	MatchId      string
	StartedTime  time.Time
	MoneyPerLine int64

	SlotbacayResult           [][]string
	MapPaylineIndexToWonMoney map[int]int64
	MapPaylineIndexToIsWin    map[int]bool
	SumWonMoney               int64
	MatchWonType              string // MATCH_WON_TYPE_..
}

func (result1p *ResultOnePlayer) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"playerId": result1p.Id,
		"username": result1p.Username,

		"startedTime":  result1p.StartedTime.Format(time.RFC3339),
		"matchId":      result1p.MatchId,
		"moneyPerLine": result1p.MoneyPerLine,

		"slotbacayResult":           result1p.SlotbacayResult,
		"sumWonMoney":               result1p.SumWonMoney,
		"mapPaylineIndexToWonMoney": result1p.MapPaylineIndexToWonMoney,
		"mapPaylineIndexToIsWin":    result1p.MapPaylineIndexToIsWin,
		"matchWonType":              result1p.MatchWonType,
		"changedMoney":              result1p.ChangedMoney,
	}
	return result
}

// for table match_record
func (result1p *ResultOnePlayer) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"id":       result1p.Id,
		"username": result1p.Username,
		"change":   result1p.ChangedMoney,
	}
	return result
}

func (result1p *ResultOnePlayer) String() string {
	bytes, _ := json.Marshal(result1p.Serialize())
	return string(bytes)
}

type SlotbacayMatch struct {
	game          *SlotbacayGame
	player        *player.Player
	startedTime   time.Time
	matchId       string
	tax           int64
	moneyPerLine  int64
	payLineIndexs []int

	slotbacayResult [][]string
	playerResult    *ResultOnePlayer

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

func NewSlotbacayMatch(
	slotbacayG *SlotbacayGame, createdPlayer *player.Player, matchCounter int64,
	moneyPerLine int64, payLineIndexs []int,
) *SlotbacayMatch {
	match := &SlotbacayMatch{
		game:          slotbacayG,
		player:        createdPlayer,
		startedTime:   time.Now(),
		matchId:       fmt.Sprintf("%v_%v_%v", slotbacayG.GameCode(), matchCounter, time.Now().Unix()),
		playerResult:  &ResultOnePlayer{},
		moneyPerLine:  moneyPerLine,
		payLineIndexs: payLineIndexs,
	}
	// init vars code in match here
	match.playerResult.Id = match.player.Id()
	match.playerResult.Username = match.player.Name()
	match.playerResult.MatchId = match.matchId
	match.playerResult.StartedTime = match.startedTime
	match.playerResult.MoneyPerLine = match.moneyPerLine
	//
	go Start(match)
	return match
}

// match main flow
func Start(match *SlotbacayMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	match.slotbacayResult = RandomSpin()

	//	 test hit jackpot
	isTesting := false
	if isTesting {
		goodHands := [][][]string{
			[][]string{[]string{"Ac"}, []string{"As"}, []string{"Ah"}},
			[][]string{[]string{"Ad"}, []string{"Ac"}, []string{"As"}},
			[][]string{[]string{"Ah"}, []string{"Ad"}, []string{"Ac"}},
			[][]string{[]string{"8d"}, []string{"7d"}, []string{"9d"}},
			[][]string{[]string{"4h"}, []string{"6h"}, []string{"5h"}},
			[][]string{[]string{"3s"}, []string{"As"}, []string{"2s"}},
			[][]string{[]string{"8s"}, []string{"8c"}, []string{"8d"}},
			[][]string{[]string{"6s"}, []string{"6c"}, []string{"6d"}},
			[][]string{[]string{"9s"}, []string{"9c"}, []string{"9d"}},
			[][]string{[]string{"Ad"}, []string{"4c"}, []string{"5d"}},
			[][]string{[]string{"As"}, []string{"Ad"}, []string{"8c"}},
			[][]string{[]string{"7h"}, []string{"2c"}, []string{"Ad"}},
		}
		if rand.Intn(1) < 2 {
			match.slotbacayResult = goodHands[rand.Intn(len(goodHands))]
		}
	}

	//
	match.playerResult.SlotbacayResult = match.slotbacayResult
	match.playerResult.MapPaylineIndexToWonMoney, match.playerResult.MapPaylineIndexToIsWin, match.playerResult.MatchWonType = CalcWonMoneys(
		match.slotbacayResult, match.payLineIndexs, match.moneyPerLine)
	sumMoneyIncludeFreeSpin := CalcSumPay(match.playerResult.MapPaylineIndexToWonMoney)

	var jackpotObj *jackpot.Jackpot
	if match.moneyPerLine == MONEYS_PER_LINE[1] {
		jackpotObj = match.game.jackpot100
	} else if match.moneyPerLine == MONEYS_PER_LINE[2] {
		jackpotObj = match.game.jackpot1000
	} else if match.moneyPerLine == MONEYS_PER_LINE[3] {
		jackpotObj = match.game.jackpot10000
	} else {

	}
	if jackpotObj != nil {
		if match.playerResult.MatchWonType == MATCH_WON_TYPE_JACKPOT {
			temp := jackpotObj.Value()
			sumMoneyIncludeFreeSpin += temp
			jackpotObj.AddMoney(-temp + 100*match.moneyPerLine)
			jackpotObj.NotifySomeoneHitJackpot(
				match.GameCode(),
				temp,
				match.player.Id(),
				match.player.Name(),
			)
		}
		//
		temp := match.moneyPerLine * int64(len(match.payLineIndexs))
		temp = int64(0.01 * float64(temp))
		jackpotObj.AddMoney(temp)
	}

	match.playerResult.SumWonMoney = sumMoneyIncludeFreeSpin
	match.playerResult.ChangedMoney = match.playerResult.SumWonMoney - match.moneyPerLine
	match.game.SendDataToPlayerId("SlotbacayResult", match.SerializedData(), match.player.Id())

	// event number of tens
	top.GlobalMutex.Lock()
	event := top.MapEvents[top.EVENT_SLOTBACAY_TEN_POINT]
	top.GlobalMutex.Unlock()
	if event != nil {
		if match.playerResult.SumWonMoney == MAP_LINE_TYPE_TO_PRIZE_RATE[LINE_TYPE_10]*match.moneyPerLine ||
			match.playerResult.SumWonMoney == MAP_LINE_TYPE_TO_PRIZE_RATE[LINE_TYPE_10AD]*match.moneyPerLine {
			event.ChangeValue(match.player.Id(), 1)
		}
	}

	// _________________________________________________________________________
	// nếu moneyPerLine > 0
	// lưu trận đấu vào database
	// ...
	//
	time.Sleep(3 * time.Second)
	if match.playerResult.SumWonMoney > 0 {
		match.player.ChangeMoneyAndLog(
			match.playerResult.SumWonMoney, match.CurrencyType(), false, "",
			ACTION_FINISH_SESSION, match.game.GameCode(), match.matchId)
	}
	if match.playerResult.SumWonMoney >= zmisc.GLOBAL_TEXT_LOWER_BOUND {
		zmisc.InsertNewGlobalText(map[string]interface{}{
			"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
			"username": match.player.DisplayName(),
			"wonMoney": match.playerResult.SumWonMoney,
			"gamecode": match.GameCode(),
		})
	}

	// cập nhật lịch sửa 10 ván chơi gần nhất
	match.game.mutex.Lock()
	if _, isIn := match.game.mapPlayerIdToHistory[match.player.Id()]; !isIn {
		temp := cardgame.NewSizedList(10)
		match.game.mapPlayerIdToHistory[match.player.Id()] = &temp
	}
	match.game.mapPlayerIdToHistory[match.player.Id()].Append(
		match.playerResult.String())
	match.game.mutex.Unlock()
	// cập nhật danh sách thắng lớn
	if match.playerResult.MatchWonType != MATCH_WON_TYPE_NORMAL {
		match.game.mutex.Lock()
		match.game.bigWinList.Append(match.playerResult.String())
		match.game.mutex.Unlock()
	}
	//
	match.game.mutex.Lock()
	delete(match.game.mapPlayerIdToMatch, match.player.Id())
	match.game.mutex.Unlock()

	// LogMatchRecord2
	var humanWon, humanLost, botWon, botLost int64
	humanWon = match.playerResult.SumWonMoney
	humanLost = match.moneyPerLine * int64(len(match.payLineIndexs))
	if humanWon > humanLost {
		rank.ChangeKey(rank.RANK_NUMBER_OF_WINS, match.playerResult.Id, 1)
	}

	playerIpAdds := map[int64]string{}
	playerObj := match.player
	playerIpAdds[playerObj.Id()] = playerObj.IpAddress()

	playerResults := make([]map[string]interface{}, 0)
	r1p := match.playerResult
	playerResults = append(playerResults, r1p.ToMap())

	record.LogMatchRecord2(
		match.game.GameCode(), match.game.CurrencyType(), match.moneyPerLine, 0,
		humanWon, humanLost, botWon, botLost,
		match.matchId, playerIpAdds,
		playerResults)
}

//
func (match *SlotbacayMatch) GameCode() string {
	return match.game.GameCode()
}

func (match *SlotbacayMatch) CurrencyType() string {
	return match.game.CurrencyType()
}

// json obj represent general match info
func (match *SlotbacayMatch) SerializedData() map[string]interface{} {
	result := match.playerResult.Serialize()
	result["gameCode"] = match.game.gameCode
	result["currencyType"] = match.game.currencyType
	return result
}

// unique data for specific player
func (match *SlotbacayMatch) SerializedDataForPlayer(playerId int64) map[string]interface{} {
	return map[string]interface{}{}
}
