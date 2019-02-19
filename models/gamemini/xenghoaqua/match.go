package xeng

//
//// general funcs for slot games
//
//import (
//	//	"encoding/json"
//	//	"errors"
//	"fmt"
//	//	"math/rand"
//	"runtime/debug"
//	"time"
//
//	//	"github.com/vic/vic_go/models/currency"
//	//"github.com/vic/vic_go/models/game"
//	"github.com/vic/vic_go/models/event_player"
//	"github.com/vic/vic_go/models/player"
//	// "github.com/vic/vic_go/utils"
//	"github.com/vic/vic_go/models/cardgame"
//	"github.com/vic/vic_go/models/game/jackpot"
//	"github.com/vic/vic_go/models/zmisc"
//	"github.com/vic/vic_go/record"
//)
//
//var MAP_OUTCOME_TO_RATE map[string]int64
//

//
//func init() {
//
//}
//
//type Match struct {
//	Game        *Game
//	Player      *player.Player
//	StartedTime time.Time
//	MatchId     string
//	MapBet      map[string]int64
//	SumLost     int64
//	SumWon      int64
//	Outcome     string
//}
//
//// match main flow
//func StartMatchFunc(match *Match) {
//	defer func() {
//		if r := recover(); r != nil {
//			bytes := debug.Stack()
//			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
//		}
//	}()
//
//	defer func() {
//		match.Game.Mutex.Lock()
//		delete(match.Game.MapPidToMatch, match.Player.Id())
//		match.Game.Mutex.Unlock()
//	}()
//	// _________________________________________________________________________
//	// _________________________________________________________________________
//	match.Mutex.Lock()
//	match.Phase = consts.PHASE_1_SPIN
//	match.SlotResult = RandomSpin()
//	// quay lại nếu trúng to :v
//	var t1 map[int]int64
//	t1, _, _ = CalcWonMoneys(
//		match.SlotResult, match.PayLineIndexs, match.MoneyPerLine)
//	if CalcSumPay(t1) >= 15000000 {
//		match.SlotResult = RandomSpin()
//	}
//	//
//	match.PlayerResult.SlotResult = match.SlotResult
//	match.PlayerResult.MapPaylineIndexToWonMoney,
//		match.PlayerResult.MapPaylineIndexToIsWin,
//		match.PlayerResult.MatchWonType = CalcWonMoneys(
//		match.SlotResult, match.PayLineIndexs, match.MoneyPerLine)
//	sumMoneyAfterSpin := CalcSumPay(match.PlayerResult.MapPaylineIndexToWonMoney)
//	_ = sumMoneyAfterSpin
//	match.Mutex.Unlock()
//	match.UpdateMatchStatus()
//	time.Sleep(DURATION_PHASE_1_SPIN)
//	//	 add % to jackpot
//	var jackpotObj *jackpot.Jackpot
//	var jacpotHitRate float64
//	if match.MoneyPerLine == 0 {
//		//
//	} else if match.MoneyPerLine <= 100 {
//		jackpotObj = match.Game.JackpotSmall
//		jacpotHitRate = float64(match.MoneyPerLine) / 100
//	} else if match.MoneyPerLine <= 1000 {
//		jackpotObj = match.Game.JackpotMedium
//		jacpotHitRate = float64(match.MoneyPerLine) / 1000
//	} else if match.MoneyPerLine <= 10000 {
//		jackpotObj = match.Game.JackpotBig
//		jacpotHitRate = float64(match.MoneyPerLine) / 10000
//	} else {
//	}
//
//	if jackpotObj != nil {
//		temp := match.MoneyPerLine * int64(len(match.PayLineIndexs))
//		temp = int64(0.01 * float64(temp))
//		jackpotObj.AddMoney(temp)
//
//		if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_JACKPOT {
//			amount := int64(float64(jackpotObj.Value()) * jacpotHitRate)
//			match.PlayerResult.SumWonMoney = amount
//			jackpotObj.AddMoney(-amount)
//			jackpotObj.NotifySomeoneHitJackpot(
//				match.GetGameCode(),
//				amount,
//				match.Player.Id(),
//				match.Player.Name(),
//			)
//		} else if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_AG {
//			eventName := fmt.Sprintf("EVENT_HOURLY_SLOTACP_%v", match.MoneyPerLine)
//			event_player.GlobalMutex.Lock()
//			e := event_player.MapEvents[eventName]
//			event_player.GlobalMutex.Unlock()
//			if e != nil {
//				e.GiveAPiece1(match.Player.Id())
//				match.Game.SendDataToPlayerId(
//					"SlotacpGiveAPiece",
//					map[string]interface{}{},
//					match.Player.Id(),
//				)
//			}
//			match.PlayerResult.SumWonMoney = sumMoneyAfterSpin
//		} else {
//			match.PlayerResult.SumWonMoney = sumMoneyAfterSpin
//		}
//	}
//	// _________________________________________________________________________
//	// end the match
//	// _________________________________________________________________________
//	action := gamemini.Action{
//		ActionName:   consts.ACTION_FINISH_SESSION,
//		ChanResponse: make(chan *gamemini.ActionResponse),
//	}
//	match.ChanActionReceiver <- &action
//	<-action.ChanResponse
//
//	match.Phase = consts.PHASE_4_RESULT
//	match.UpdateMatchStatus()
//
//	if match.PlayerResult.SumWonMoney > 0 {
//		match.Player.ChangeMoneyAndLog(
//			match.PlayerResult.SumWonMoney, match.GetCurrencyType(), false, "",
//			consts.ACTION_FINISH_SESSION, match.Game.GetGameCode(), match.MatchId)
//	}
//	if match.PlayerResult.SumWonMoney >= zmisc.GLOBAL_TEXT_LOWER_BOUND {
//		zmisc.InsertNewGlobalText(map[string]interface{}{
//			"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
//			"username": match.Player.DisplayName(),
//			"wonMoney": match.PlayerResult.SumWonMoney,
//			"gamecode": match.GetGameCode(),
//		})
//	}
//	// cập nhật lịch sửa 10 ván chơi gần nhất
//	match.Game.Mutex.Lock()
//	if _, isIn := match.Game.MapPlayerIdToHistory[match.Player.Id()]; !isIn {
//		temp := cardgame.NewSizedList(10)
//		match.Game.MapPlayerIdToHistory[match.Player.Id()] = &temp
//	}
//	match.Game.MapPlayerIdToHistory[match.Player.Id()].Append(
//		match.PlayerResult.String())
//	match.Game.Mutex.Unlock()
//	// cập nhật danh sách thắng lớn
//	if match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_BIG ||
//		match.PlayerResult.MatchWonType == consts.MATCH_WON_TYPE_JACKPOT {
//		match.Game.Mutex.Lock()
//		match.Game.BigWinList.Append(match.PlayerResult.String())
//		match.Game.Mutex.Unlock()
//	}
//
//	// LogMatchRecord2
//	var humanWon, humanLost, botWon, botLost int64
//	humanWon = match.PlayerResult.SumWonMoney
//	humanLost = -match.PlayerResult.SumLostMoney
//
//	playerIpAdds := map[int64]string{}
//	playerObj := match.Player
//	playerIpAdds[playerObj.Id()] = playerObj.IpAddress()
//
//	playerResults := make([]map[string]interface{}, 0)
//	r1p := match.PlayerResult
//	playerResults = append(playerResults, r1p.ToMap())
//
//	record.LogMatchRecord2(
//		match.Game.GetGameCode(), match.Game.GetCurrencyType(), match.MoneyPerLine, 0,
//		humanWon, humanLost, botWon, botLost,
//		match.MatchId, playerIpAdds,
//		playerResults)
//}
