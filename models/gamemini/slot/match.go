package slot

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
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
)

const (
	ACTION_STOP_GAME = "ACTION_STOP_GAME"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION"

	ACTION_CHOOSE_MONEY_PER_LINE = "ACTION_CHOOSE_MONEY_PER_LINE"
	ACTION_CHOOSE_PAYLINES       = "ACTION_CHOOSE_PAYLINES"
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

	SlotResult                [][]string
	MapPaylineIndexToWonMoney map[int]int64
	MapPaylineIndexToIsWin    map[int]bool
	SumWonMoney               int64
	MatchWonType              string // MATCH_WON_TYPE_..
}

func (result1p *ResultOnePlayer) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"id":           result1p.Id,
		"username":     result1p.Username,
		"startedTime":  result1p.StartedTime.Format(time.RFC3339),
		"matchId":      result1p.MatchId,
		"moneyPerLine": result1p.MoneyPerLine,

		"slotResult":                result1p.SlotResult,
		"sumWonMoney":               result1p.SumWonMoney,
		"mapPaylineIndexToWonMoney": result1p.MapPaylineIndexToWonMoney,
		"mapPaylineIndexToIsWin":    result1p.MapPaylineIndexToIsWin,
		"matchWonType":              result1p.MatchWonType,
		"change":                    result1p.ChangedMoney,
	}
	return result
}

// for table match_record
func (result1p *ResultOnePlayer) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"id":       result1p.Id,
		"username": result1p.Username,

		"change": result1p.ChangedMoney,
	}
	return result
}

func (result1p *ResultOnePlayer) String() string {
	bytes, _ := json.Marshal(result1p.Serialize())
	return string(bytes)
}

type SlotMatch struct {
	game          *SlotGame
	player        *player.Player
	startedTime   time.Time
	matchId       string
	tax           int64
	moneyPerLine  int64
	payLineIndexs []int

	slotResult   [][]string
	playerResult *ResultOnePlayer

	mutex sync.RWMutex
}

type Action struct {
	actionName string
	playerId   int64

	data         map[string]interface{}
	chanResponse chan *ActionResponse
}

func (action *Action) ToMap() map[string]interface{} {
	if action != nil {
		result := map[string]interface{}{
			"actionTime": time.Now(),
			"actionName": action.actionName,
			"playerId":   action.playerId,
			"data":       action.data,
		}
		return result
	} else {
		return map[string]interface{}{}
	}
}

type ActionResponse struct {
	err  error
	data map[string]interface{}
}

func NewSlotMatch(
	slotG *SlotGame, createdPlayer *player.Player, matchCounter int64,
	moneyPerLine int64, payLineIndexs []int,
) *SlotMatch {
	match := &SlotMatch{
		game:          slotG,
		player:        createdPlayer,
		startedTime:   time.Now(),
		matchId:       fmt.Sprintf("%v_%v_%v", slotG.GetGameCode(), matchCounter, time.Now().Unix()),
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
func Start(match *SlotMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	match.slotResult = RandomSpin()
	//	 test hit jackpot
	isTesting := false
	if isTesting {
		goodHands := [][][]string{
			[][]string{
				[]string{"7", "1", "2"},
				[]string{"6", "1", "3"},
				[]string{"5", "1", "4"},
				[]string{"4", "1", "5"},
				[]string{"3", "1", "6"},
			},
			[][]string{
				[]string{"3", "1", "5"},
				[]string{"3", "4", "6"},
				[]string{"3", "4", "5"},
				[]string{"3", "4", "5"},
				[]string{"3", "1", "6"},
			},
			[][]string{
				[]string{"1", "4", "5"},
				[]string{"1", "2", "5"},
				[]string{"1", "2", "5"},
				[]string{"1", "2", "5"},
				[]string{"1", "4", "6"},
			},
			[][]string{
				[]string{"7", "3", "2"},
				[]string{"6", "4", "2"},
				[]string{"5", "5", "2"},
				[]string{"4", "6", "2"},
				[]string{"3", "7", "2"},
			},
			[][]string{
				[]string{"4", "2", "2"},
				[]string{"4", "1", "2"},
				[]string{"4", "3", "2"},
				[]string{"4", "1", "2"},
				[]string{"3", "4", "6"},
			},
			[][]string{
				[]string{"7", "1", "5"},
				[]string{"7", "6", "3"},
				[]string{"7", "6", "5"},
				[]string{"4", "6", "5"},
				[]string{"3", "1", "6"},
			},
			[][]string{
				[]string{"7", "5", "2"},
				[]string{"7", "5", "2"},
				[]string{"7", "1", "2"},
				[]string{"4", "5", "2"},
				[]string{"3", "5", "2"},
			},
			[][]string{
				[]string{"7", "1", "2"},
				[]string{"6", "1", "3"},
				[]string{"5", "1", "4"},
				[]string{"4", "1", "5"},
				[]string{"3", "1", "6"},
			},
			[][]string{
				[]string{"7", "3", "5"},
				[]string{"2", "3", "5"},
				[]string{"2", "3", "4"},
				[]string{"2", "1", "5"},
				[]string{"3", "3", "5"},
			},
			[][]string{
				[]string{"3", "2", "1"},
				[]string{"3", "2", "1"},
				[]string{"3", "2", "1"},
				[]string{"3", "2", "1"},
				[]string{"3", "2", "1"},
			},
		}
		if rand.Intn(3) < 1 {
			match.slotResult = goodHands[rand.Intn(len(goodHands))]
		}
	}

	// quay lại nếu trúng to :v
	for {
		var t1 map[int]int64
		t1, _, _ = CalcWonMoneys(
			match.slotResult, match.payLineIndexs, match.moneyPerLine)
		if 1500000 <= CalcSumPay(t1) && rand.Intn(2) < 1 {
			match.slotResult = RandomSpin()
			t1, _, _ = CalcWonMoneys(
				match.slotResult, match.payLineIndexs, match.moneyPerLine)
		}
		if CalcSumPay(t1) < 15000000 {
			break
		}
	}

	match.playerResult.SlotResult = match.slotResult
	// fmt.Println(SlotResult(match.slotResult))
	match.playerResult.MapPaylineIndexToWonMoney,
		match.playerResult.MapPaylineIndexToIsWin,
		match.playerResult.MatchWonType = CalcWonMoneys(
		match.slotResult, match.payLineIndexs, match.moneyPerLine)
	sumMoneyIncludeFreeSpin := CalcSumPay(match.playerResult.MapPaylineIndexToWonMoney)

	var jackpotObj *jackpot.Jackpot
	var jacpotHitRate float64
	if match.moneyPerLine == 0 {
		//
	} else if match.moneyPerLine <= 100 {
		jackpotObj = match.game.jackpot100
		jacpotHitRate = float64(match.moneyPerLine) / 100
	} else if match.moneyPerLine <= 1000 {
		jackpotObj = match.game.jackpot1000
		jacpotHitRate = float64(match.moneyPerLine) / 1000
	} else if match.moneyPerLine <= 10000 {
		jackpotObj = match.game.jackpot10000
		jacpotHitRate = float64(match.moneyPerLine) / 10000
	} else {
	}

	if jackpotObj != nil {
		if match.playerResult.MatchWonType == MATCH_WON_TYPE_JACKPOT {
			amount := int64(float64(jackpotObj.Value()) * jacpotHitRate)
			sumMoneyIncludeFreeSpin += amount
			jackpotObj.AddMoney(-amount +
				int64(jacpotHitRate*10000*float64(match.moneyPerLine)))
			jackpotObj.NotifySomeoneHitJackpot(
				match.GameCode(),
				amount,
				match.player.Id(),
				match.player.Name(),
			)
		}
		//
		if IS_FAKE_JACKPOT {

		} else {
			temp := match.moneyPerLine * int64(len(match.payLineIndexs))
			temp = int64(0.01 * float64(temp))
			jackpotObj.AddMoney(temp)
		}
	}

	match.playerResult.SumWonMoney = sumMoneyIncludeFreeSpin
	match.playerResult.ChangedMoney = match.playerResult.SumWonMoney - match.moneyPerLine*int64(len(match.payLineIndexs))
	match.game.SendDataToPlayerId("SlotResult", match.SerializedData(), match.player.Id())
	// _________________________________________________________________________
	// nếu moneyPerLine > 0
	// lưu trận đấu vào database
	// ...
	//
	time.Sleep(3 * time.Second)
	if match.playerResult.SumWonMoney > 0 {
		match.player.ChangeMoneyAndLog(
			match.playerResult.SumWonMoney, match.CurrencyType(), false, "",
			ACTION_FINISH_SESSION, match.game.GetGameCode(), match.matchId)
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
	// xoá mapPlayerIdToMatch[pid] ngay sau khi hết trận
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
		match.game.GetGameCode(), match.game.GetCurrencyType(), match.moneyPerLine, 0,
		humanWon, humanLost, botWon, botLost,
		match.matchId, playerIpAdds,
		playerResults)
}

//
func (match *SlotMatch) GameCode() string {
	return match.game.GetGameCode()
}

func (match *SlotMatch) CurrencyType() string {
	return match.game.GetCurrencyType()
}

// json obj represent general match info
func (match *SlotMatch) SerializedData() map[string]interface{} {
	result := match.playerResult.Serialize()
	result["gameCode"] = match.game.gameCode
	result["currencyType"] = match.game.currencyType
	return result
}

// unique data for specific player
func (match *SlotMatch) SerializedDataForPlayer(playerId int64) map[string]interface{} {
	return map[string]interface{}{}
}
