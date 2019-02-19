package slotxxx

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
	MAX_XXX_LEVEL = 8

	ACTION_STOP_GAME = "ACTION_STOP_GAME"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION"

	ACTION_CHOOSE_MONEY_PER_LINE = "ACTION_CHOOSE_MONEY_PER_LINE"
	ACTION_GET_HISTORY           = "ACTION_GET_HISTORY"
	ACTION_SPIN                  = "ACTION_SPIN"

	ACTION_GET_MATCH_INFO = "ACTION_GET_MATCH_INFO"

	ACTION_STOP_PLAYING = "ACTION_STOP_PLAYING"
	ACTION_SELECT_SMALL = "ACTION_SELECT_SMALL"
	ACTION_SELECT_BIG   = "ACTION_SELECT_BIG"

	PHASE_1_SPIN         = "PHASE_1_SPIN"
	PHASE_3_CHOOSE_GO_ON = "PHASE_2_CHOOSE_GO_ON"
	PHASE_4_RESULT       = "PHASE_4_RESULT"

	DURATION_PHASE_1_SPIN         = time.Duration(4 * time.Second)
	DURATION_PHASE_3_CHOOSE_GO_ON = time.Duration(600 * time.Second)
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

	SlotxxxResult             [][]string
	MapPaylineIndexToWonMoney map[int]int64
	MapPaylineIndexToIsWin    map[int]bool
	SumWonMoney               int64
	// SumLostMoney is negative value
	SumLostMoney int64
	MatchWonType string // MATCH_WON_TYPE_..
}

func (result1p *ResultOnePlayer) Serialize() map[string]interface{} {
	result := map[string]interface{}{
		"playerId": result1p.Id,
		"username": result1p.Username,

		"startedTime":  result1p.StartedTime.Format(time.RFC3339),
		"matchId":      result1p.MatchId,
		"moneyPerLine": result1p.MoneyPerLine,

		"slotxxxResult":             result1p.SlotxxxResult,
		"sumWonMoney":               result1p.SumWonMoney,
		"sumLostMoney":              result1p.SumLostMoney,
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

type SlotxxxMatch struct {
	game          *SlotxxxGame
	player        *player.Player
	startedTime   time.Time
	matchId       string
	tax           int64
	moneyPerLine  int64
	payLineIndexs []int

	slotxxxResult [][]string
	// from 0 to 6
	currentXxxLevel int
	// money to bet at current level,
	// if choose the right phase3 and stop, the user receives x2 this amount
	currentXxxMoney     int64
	requiredMoneyToGoOn int64
	winningMoneyIfStop  int64
	// choose the right small or big
	isRightPhase3 bool

	// đoán một lần là đúng nhỏ hay lớn,
	// cả 7 lần đoán một phát là trúng thì được thưởng jackpot
	is1stTryFailed []bool // đoán một lần là đúng nhỏ hay lớn,

	playerResult *ResultOnePlayer

	phase string
	// 1 card, A <= card.rank <= 6 is small, 8 <= card.rank <= K is big
	phase3result string

	// choose stop, small or big
	ChanPhase3         chan string
	ChanActionReceiver chan *Action

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

func NewSlotxxxMatch(
	slotxxxG *SlotxxxGame, createdPlayer *player.Player, matchCounter int64,
	moneyPerLine int64, payLineIndexs []int,
) *SlotxxxMatch {
	match := &SlotxxxMatch{
		game:          slotxxxG,
		player:        createdPlayer,
		startedTime:   time.Now(),
		matchId:       fmt.Sprintf("%v_%v_%v", slotxxxG.GameCode(), matchCounter, time.Now().Unix()),
		playerResult:  &ResultOnePlayer{},
		moneyPerLine:  moneyPerLine,
		payLineIndexs: payLineIndexs,

		is1stTryFailed: make([]bool, MAX_XXX_LEVEL),

		phase: "PHASE_0_INITING",

		ChanPhase3:         make(chan string),
		ChanActionReceiver: make(chan *Action),
	}
	// init vars code in match here
	match.playerResult.Id = match.player.Id()
	match.playerResult.Username = match.player.Name()
	match.playerResult.MatchId = match.matchId
	match.playerResult.StartedTime = match.startedTime
	match.playerResult.MoneyPerLine = match.moneyPerLine
	match.playerResult.SumLostMoney = -match.moneyPerLine
	//
	go Start(match)
	go InMatchLoopReceiveActions(match)
	return match
}

// match main flow
func Start(match *SlotxxxMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	defer func() {
		match.game.mutex.Lock()
		delete(match.game.mapPlayerIdToMatch, match.player.Id())
		match.game.mutex.Unlock()
	}()
	// _________________________________________________________________________
	// _________________________________________________________________________
	match.mutex.Lock()
	match.phase = PHASE_1_SPIN
	match.slotxxxResult = RandomSpin()
	//	 test hit jackpot
	//	goodHands := [][][]string{
	//		[][]string{[]string{"Ac"}, []string{"As"}, []string{"Ah"}},
	//		[][]string{[]string{"Ad"}, []string{"Ac"}, []string{"As"}},
	//		[][]string{[]string{"Ah"}, []string{"Ad"}, []string{"Ac"}},
	//		[][]string{[]string{"8d"}, []string{"7d"}, []string{"9d"}},
	//		[][]string{[]string{"4h"}, []string{"6h"}, []string{"5h"}},
	//		[][]string{[]string{"3s"}, []string{"As"}, []string{"2s"}},
	//		[][]string{[]string{"8s"}, []string{"8c"}, []string{"8d"}},
	//		[][]string{[]string{"6s"}, []string{"6c"}, []string{"6d"}},
	//		[][]string{[]string{"9s"}, []string{"9c"}, []string{"9d"}},
	//		[][]string{[]string{"Ad"}, []string{"4c"}, []string{"5d"}},
	//		[][]string{[]string{"As"}, []string{"Ad"}, []string{"8c"}},
	//		[][]string{[]string{"7h"}, []string{"2c"}, []string{"Ad"}},
	//	}
	//	if rand.Intn(1) < 1 {
	//		match.slotxxxResult = goodHands[rand.Intn(len(goodHands))]
	//	}

	//
	match.playerResult.SlotxxxResult = match.slotxxxResult
	match.playerResult.MapPaylineIndexToWonMoney, match.playerResult.MapPaylineIndexToIsWin, match.playerResult.MatchWonType = CalcWonMoneys(
		match.slotxxxResult, match.payLineIndexs, match.moneyPerLine)
	sumMoneyAfterSpin := CalcSumPay(match.playerResult.MapPaylineIndexToWonMoney)
	match.mutex.Unlock()
	match.updateMatchStatus()
	time.Sleep(DURATION_PHASE_1_SPIN)
	// add % to jackpot
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
		temp := match.moneyPerLine * int64(len(match.payLineIndexs))
		temp = int64(0.025 * float64(temp)) // repay to users 95%
		jackpotObj.AddMoney(temp)

		if match.playerResult.MatchWonType == MATCH_WON_TYPE_JACKPOT {
			amount := int64(float64(jackpotObj.Value()) * 0.5)
			match.winningMoneyIfStop = amount
			jackpotObj.AddMoney(-amount)
			jackpotObj.NotifySomeoneHitJackpot(
				match.GameCode(),
				amount,
				match.player.Id(),
				match.player.Name(),
			)
		} else if sumMoneyAfterSpin > 0 {
			match.winningMoneyIfStop = sumMoneyAfterSpin
			match.currentXxxMoney = sumMoneyAfterSpin
			match.requiredMoneyToGoOn = 0
			// loop x2 game, i is level counter
			i := 0
			match.phase = PHASE_3_CHOOSE_GO_ON
			for i < MAX_XXX_LEVEL {
				if match.player.GetAvailableMoney(match.CurrencyType()) < match.currentXxxMoney {
					break
				}
				match.currentXxxLevel = i
				match.updateMatchStatus()
				timer := time.After(DURATION_PHASE_3_CHOOSE_GO_ON)
				var phase3choice string
				select {
				case <-timer:
					phase3choice = ACTION_STOP_PLAYING
				case phase3choice = <-match.ChanPhase3:
					// receive user action to phase3choice
				}
				if phase3choice == ACTION_STOP_PLAYING {
					break
				} else {
					match.phase3result = Random1Card(phase3choice, i)

					match.isRightPhase3 = false
					var cardRank string
					if len(match.phase3result) > 0 {
						cardRank = string(match.phase3result[0])
					}
					if cardRank == "A" || cardRank == "2" || cardRank == "3" ||
						cardRank == "4" || cardRank == "5" || cardRank == "6" {
						if phase3choice == ACTION_SELECT_SMALL {
							match.isRightPhase3 = true
						}
					} else if cardRank == "7" {

					} else {
						if phase3choice == ACTION_SELECT_BIG {
							match.isRightPhase3 = true
						}
					}

					isFirstTry := (match.is1stTryFailed[i] == false)
					if match.isRightPhase3 {
						match.requiredMoneyToGoOn = 0
						match.currentXxxMoney = 2 * match.currentXxxMoney
						match.winningMoneyIfStop = match.currentXxxMoney
						i += 1
					} else { // chọn sai, trừ tiền
						match.is1stTryFailed[i] = true
						match.winningMoneyIfStop = match.currentXxxMoney
						if i == 0 && isFirstTry {
							// match.requiredMoneyToGoOn = int64(float64(match.currentXxxMoney) * (2.0 / 3.0))
							match.requiredMoneyToGoOn = match.currentXxxMoney
						} else {
							match.requiredMoneyToGoOn = match.currentXxxMoney
						}

						if match.requiredMoneyToGoOn > 0 {
							match.player.ChangeMoneyAndLog(
								-match.requiredMoneyToGoOn, match.CurrencyType(), false, "",
								phase3choice, match.GameCode(), match.matchId)
							//
							match.playerResult.SumLostMoney -= match.requiredMoneyToGoOn
							// add half money to jackpot
							jackpotObj.AddMoney(match.requiredMoneyToGoOn / 52)
						}
					}
				}
			} // end loop x2 game
			match.currentXxxLevel = i
			match.updateMatchStatus()
			time.Sleep(200 * time.Millisecond)

			if i == MAX_XXX_LEVEL {
				amount := int64(float64(jackpotObj.Value()) * 0.05)
				match.winningMoneyIfStop += amount
				jackpotObj.AddMoney(-amount)
				jackpotObj.NotifySomeoneHitJackpot(
					match.GameCode(),
					amount,
					match.player.Id(),
					match.player.Name(),
				)
			}
		}
	}
	// _________________________________________________________________________
	// end the match
	// _________________________________________________________________________
	action := Action{
		actionName:   ACTION_FINISH_SESSION,
		chanResponse: make(chan *ActionResponse),
	}
	match.ChanActionReceiver <- &action
	<-action.chanResponse

	match.phase = PHASE_4_RESULT
	match.playerResult.SumWonMoney = match.winningMoneyIfStop
	match.updateMatchStatus()

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
	if match.playerResult.SumWonMoney >= 10*match.moneyPerLine {
		match.game.mutex.Lock()
		match.game.bigWinList.Append(match.playerResult.String())
		match.game.mutex.Unlock()
	}

	// LogMatchRecord2
	var humanWon, humanLost, botWon, botLost int64
	humanWon = match.playerResult.SumWonMoney
	humanLost = -match.playerResult.SumLostMoney
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
func (match *SlotxxxMatch) GameCode() string {
	return match.game.GameCode()
}

func (match *SlotxxxMatch) CurrencyType() string {
	return match.game.CurrencyType()
}

// json obj represent general match info
func (match *SlotxxxMatch) SerializedData() map[string]interface{} {
	data := match.playerResult.Serialize()
	data["phase"] = match.phase
	data["currentXxxLevel"] = match.currentXxxLevel
	data["currentXxxMoney"] = match.currentXxxMoney
	data["is1stTryFailed"] = match.is1stTryFailed
	data["phase3result"] = match.phase3result
	data["requiredMoneyToGoOn"] = match.requiredMoneyToGoOn
	return data
}

func (match *SlotxxxMatch) updateMatchStatus() {
	data := match.SerializedData()
	match.game.SendDataToPlayerId(
		"SlotxxxUpdateMatchStatus",
		data,
		match.player.Id(),
	)
}

func InMatchLoopReceiveActions(match *SlotxxxMatch) {
	for {
		action := <-match.ChanActionReceiver
		if action.actionName == ACTION_FINISH_SESSION {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(match *SlotxxxMatch, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.actionName == ACTION_GET_MATCH_INFO {
					action.chanResponse <- &ActionResponse{err: nil}
					match.updateMatchStatus()
				} else if action.actionName == ACTION_STOP_PLAYING ||
					action.actionName == ACTION_SELECT_SMALL ||
					action.actionName == ACTION_SELECT_BIG {
					timer := time.After(3 * time.Second)
					select {
					case match.ChanPhase3 <- action.actionName:
						action.chanResponse <- &ActionResponse{err: nil}
					case <-timer:
						action.chanResponse <- &ActionResponse{err: errors.New("timeout")}
					}
				} else {
					action.chanResponse <- &ActionResponse{err: errors.New("wrongActionName")}
				}
			}(match, action)
		}
	}
}
