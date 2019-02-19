package slotax1to5

// general funcs for slot games

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"

	//"github.com/vic/vic_go/models/currency"
	//"github.com/vic/vic_go/models/game"
	// "github.com/vic/vic_go/models/player"
	// "github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
)

const (
	DURATION_PHASE_1_SPIN = time.Duration(0 * time.Second)
)

func init() {
	fmt.Print("")
	_ = jackpot.Jackpot{}
	_ = errors.New("")
	_, _ = json.Marshal([]int{})
	_ = rand.Intn(10)
}

// match main flow
func StartMatchFunc(match *gamemini.SlotMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	defer func() {
		match.Game.Mutex.Lock()
		delete(match.Game.MapPlayerIdToMatch, match.Player.Id())
		match.Game.Mutex.Unlock()
	}()
	// _________________________________________________________________________
	// _________________________________________________________________________
	match.Mutex.Lock()
	match.Phase = consts.PHASE_1_SPIN
	match.SlotResult = RandomSpin()
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
			match.SlotResult = goodHands[rand.Intn(len(goodHands))]
		}
	}

	// quay lại nếu trúng to :v
	for {
		var t1 map[int]int64
		t1, _, _ = CalcWonMoneys(
			match.SlotResult, match.PayLineIndexs, match.MoneyPerLine)
		if 1500000 <= CalcSumPay(t1) && rand.Intn(2) < 1 {
			match.SlotResult = RandomSpin()
			t1, _, _ = CalcWonMoneys(
				match.SlotResult, match.PayLineIndexs, match.MoneyPerLine)
		}
		if CalcSumPay(t1) < 15000000 {
			break
		}
	}

	//
	match.PlayerResult.SlotResult = match.SlotResult
	match.PlayerResult.MapPaylineIndexToWonMoney,
		match.PlayerResult.MapPaylineIndexToIsWin,
		match.PlayerResult.MatchWonType = CalcWonMoneys(
		match.SlotResult, match.PayLineIndexs, match.MoneyPerLine)
	sumMoneyAfterSpin := CalcSumPay(match.PlayerResult.MapPaylineIndexToWonMoney)
	match.Mutex.Unlock()
	match.UpdateMatchStatus()
	time.Sleep(DURATION_PHASE_1_SPIN)
	//	 add % to jackpot
	var jackpotObj *jackpot.Jackpot
	var jacpotHitRate float64
	if match.MoneyPerLine == 0 {
		//
	} else if match.MoneyPerLine <= 100 {
		jackpotObj = match.Game.JackpotSmall
		jacpotHitRate = float64(match.MoneyPerLine) / 100
	} else if match.MoneyPerLine <= 1000 {
		jackpotObj = match.Game.JackpotMedium
		jacpotHitRate = float64(match.MoneyPerLine) / 1000
	} else if match.MoneyPerLine <= 10000 {
		jackpotObj = match.Game.JackpotBig
		jacpotHitRate = float64(match.MoneyPerLine) / 10000
	} else {
	}

	if jackpotObj != nil {
		temp := match.MoneyPerLine * int64(len(match.PayLineIndexs))
		temp = int64(0.01 * float64(temp))
		jackpotObj.AddMoney(temp)

		if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_JACKPOT {
			amount := int64(float64(jackpotObj.Value()) * jacpotHitRate)
			match.PlayerResult.SumWonMoney = amount
			jackpotObj.AddMoney(-amount)
			jackpotObj.NotifySomeoneHitJackpot(
				match.GetGameCode(),
				amount,
				match.Player.Id(),
				match.Player.Name(),
			)
		} else if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_AG {
			match.Phase = consts.PHASE_3_ADDITIONAL_GAME
			match.AdditionalGame = gamemini.NewAgRandomX1to5(
				match.Player, jackpotObj, match.PlayerResult,
				sumMoneyAfterSpin, match.GetCurrencyType())
			match.AdditionalGame.Start()
		} else {
			match.PlayerResult.SumWonMoney = sumMoneyAfterSpin
		}
	}
	// _________________________________________________________________________
	// end the match
	// _________________________________________________________________________
	action := gamemini.Action{
		ActionName:   consts.ACTION_FINISH_SESSION,
		ChanResponse: make(chan *gamemini.ActionResponse),
	}
	match.ChanActionReceiver <- &action
	<-action.ChanResponse

	match.Phase = consts.PHASE_4_RESULT
	match.UpdateMatchStatus()

	if match.PlayerResult.SumWonMoney > 0 {
		match.Player.ChangeMoneyAndLog(
			match.PlayerResult.SumWonMoney, match.GetCurrencyType(), false, "",
			consts.ACTION_FINISH_SESSION, match.Game.GetGameCode(), match.MatchId)
	}
	if match.PlayerResult.SumWonMoney >= zmisc.GLOBAL_TEXT_LOWER_BOUND {
		zmisc.InsertNewGlobalText(map[string]interface{}{
			"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
			"username": match.Player.DisplayName(),
			"wonMoney": match.PlayerResult.SumWonMoney,
			"gamecode": match.GetGameCode(),
		})
	}
	// cập nhật lịch sửa 10 ván chơi gần nhất
	match.Game.Mutex.Lock()
	if _, isIn := match.Game.MapPlayerIdToHistory[match.Player.Id()]; !isIn {
		temp := cardgame.NewSizedList(10)
		match.Game.MapPlayerIdToHistory[match.Player.Id()] = &temp
	}
	match.Game.MapPlayerIdToHistory[match.Player.Id()].Append(
		match.PlayerResult.String())
	match.Game.Mutex.Unlock()
	// cập nhật danh sách thắng lớn
	if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_BIG ||
		match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_JACKPOT {
		match.Game.Mutex.Lock()
		match.Game.BigWinList.Append(match.PlayerResult.String())
		match.Game.Mutex.Unlock()
	}

	// LogMatchRecord2
	var humanWon, humanLost, botWon, botLost int64
	humanWon = match.PlayerResult.SumWonMoney
	humanLost = -match.PlayerResult.SumLostMoney
	if humanWon > humanLost {
		rank.ChangeKey(rank.RANK_NUMBER_OF_WINS, match.PlayerResult.Id, 1)
	}

	playerIpAdds := map[int64]string{}
	playerObj := match.Player
	playerIpAdds[playerObj.Id()] = playerObj.IpAddress()

	playerResults := make([]map[string]interface{}, 0)
	r1p := match.PlayerResult
	playerResults = append(playerResults, r1p.ToMap())

	record.LogMatchRecord2(
		match.Game.GetGameCode(), match.Game.GetCurrencyType(), match.MoneyPerLine, 0,
		humanWon, humanLost, botWon, botLost,
		match.MatchId, playerIpAdds,
		playerResults)
}
