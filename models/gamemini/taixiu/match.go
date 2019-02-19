package taixiu

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	//"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
	//	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/language"
)

const (
	ACTION_STOP_GAME = "ACTION_STOP_GAME"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION"

	ACTION_GET_MATCH_INFO = "ACTION_GET_MATCH_INFO"
	ACTION_ADD_BET        = "ACTION_ADD_BET"
	ACTION_BET_AS_LAST    = "ACTION_BET_AS_LAST"
	ACTION_BET_X2_LAST    = "ACTION_BET_X2_LAST"
	ACTION_CHAT           = "ACTION_CHAT"

	PHASE_1_BET    = "PHASE_1_BET"
	PHASE_2_REFUND = "PHASE_2_REFUND"
	PHASE_3_RESULT = "PHASE_3_RESULT"

	DURATION_PHASE_1_BET    = time.Duration(40 * time.Second) // 40
	DURATION_PHASE_2_REFUND = time.Duration(3 * time.Second)  // 3
	DURATION_PHASE_3_RESULT = time.Duration(5 * time.Second)  // 5

	TAIXIU_REFUND = "TAIXIU_REFUND"

	botPid = int64(-1)
)

func init() {
	fmt.Print("")
}

type ResultOnePlayer struct {
	Id       int64
	Username string

	EndMatchWinningMoney int64 // tiền cộng cuối trận
	FinishedMoney        int64

	Changed int64 // tổng tiền thay đổi trong trận đầu này, gồm cả tiền ăn quân
}

func (r *ResultOnePlayer) ToMap() map[string]interface{} {
	// cần theo form đầu vào hàm record.LogMatchRecord2
	result := make(map[string]interface{})
	result["id"] = r.Id
	result["username"] = r.Username
	result["EndMatchWinningMoney"] = r.EndMatchWinningMoney
	result["FinishedMoney"] = r.FinishedMoney
	result["change"] = r.Changed
	return result
}

type TaixiuMatch struct {
	game *TaixiuGame
	// map player id to player obj
	players     map[int64]*player.Player
	startedTime time.Time
	matchId     string

	tax           int64
	playerResults []*ResultOnePlayer

	// list từng thao tác cược theo thời gian
	mapPlayerIdToBets map[int64][]*Bet
	// map[playerId](map[selection]soTienCuoc),
	// thông tin cược sau khi đã cân 2 cửa,
	// ai đặt muộn trả về
	mapBetInfo map[int64]map[string]int64
	// map pid to tiền thừa
	mapPlayerIdToRefund map[int64]int64
	// cửa cần trả tiền thừa
	refundSelection string
	// giai đoạn của ván chơi PHASE_
	phase string
	// kết quả lắc xúc xắc
	shakingResult []int

	// fake info
	moneySteps []int64
	nopSteps   []float64
	fakeMoneys []int64
	// need money is 500s
	fakeMoneys500 []int64
	fakeNops      []float64

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

// already embraced in taixiuG.mutex.Lock
func NewTaixiuMatch(taixiuG *TaixiuGame) *TaixiuMatch {
	match := &TaixiuMatch{
		game:               taixiuG,
		players:            map[int64]*player.Player{},
		startedTime:        time.Now(),
		matchId:            fmt.Sprintf("%v_%v_%v", taixiuG.GameCode(), taixiuG.matchCounter, time.Now().Unix()),
		playerResults:      []*ResultOnePlayer{},
		mapPlayerIdToBets:  make(map[int64][]*Bet),
		ChanActionReceiver: make(chan *Action),
		phase:              "PHASE_0_INITING",
		shakingResult:      []int{},
	}
	// init vars code in match here
	clonedHistory := make([]string, 0)
	for _, e := range match.game.taixiuHistory.Elements {
		// e = "SELECTION_XIU|[3 2 1]"
		pieces := strings.Split(e, "|")
		if len(pieces) > 0 {
			clonedHistory = append(clonedHistory, pieces[0])
		}
	}
	n := len(clonedHistory)
	dupI := -1
	dupRate := 1.0
	if n >= 3 {
		for i, s := range []string{SELECTION_TAI, SELECTION_XIU} {
			if s == clonedHistory[n-1] &&
				s == clonedHistory[n-2] &&
				s == clonedHistory[n-3] {
				dupI = i
				dupRate = 0.9
				if n >= 4 && s == clonedHistory[n-4] {
					dupRate = 0.7
				}
				if n >= 5 && s == clonedHistory[n-5] && s == clonedHistory[n-4] {
					dupRate = 0.5
				}
			}
		}
	}

	match.moneySteps = make([]int64, 6)
	match.nopSteps = make([]float64, 6)
	match.fakeMoneys = make([]int64, 6)
	match.fakeMoneys500 = make([]int64, 6)
	match.fakeNops = make([]float64, 6)

	durationBetPhaseInSeconds := int64(DURATION_PHASE_1_BET/time.Second) - 2
	for i := 0; i < 6; i++ {
		// approximate to 100M when system ccu = 250
		endingMoney := int64(float64(15000000+rand.Int63n(20000000)) *
			(float64(match.game.systemNHuman) / 250))
		moneyStep := endingMoney / durationBetPhaseInSeconds
		match.moneySteps[i] = moneyStep

		// approximate to system's ccu/4 on SV_01, ccu/2 on others
		var t float64
		if zconfig.ServerVersion == zconfig.SV_01 {
			t = 4
		} else {
			t = 2
		}
		endingNop := float64(match.game.systemNHuman) / t *
			(0.4 + 1.2*rand.Float64())
		nopStep := endingNop / float64(durationBetPhaseInSeconds)
		match.nopSteps[i] = nopStep
	}

	go func() {
		for s := int64(0); s <= durationBetPhaseInSeconds; s++ {
			//			fmt.Println("cp 0", s)
			time.Sleep(1 * time.Second)
			for i := 0; i < 6; i++ {
				rateI := float64(1)
				if i == dupI {
					rateI = dupRate
				}
				am := int64(float64(match.moneySteps[i])*
					(0.7+0.6*rand.Float64())*rateI) / 1000 * 1000
				match.mutex.Lock()
				match.fakeMoneys[i] += am
				match.fakeMoneys500[i] = match.fakeMoneys[i]
				match.fakeNops[i] += match.nopSteps[i] *
					(0.4 + 1.2*rand.Float64()) * rateI
				match.mutex.Unlock()
				var selection string
				if i == 0 {
					selection = SELECTION_TAI
				} else if i == 1 {
					selection = SELECTION_XIU
				}
				if selection != "" {
					InMatchAddBet(match, botPid, &Bet{
						playerId:   botPid,
						betTime:    time.Now(),
						selection:  selection,
						moneyValue: am,
					})
				}
			}

		}
	}()

	//
	go Start(match)
	go InMatchLoopReceiveActions(match)
	return match
}

// match main flow
func Start(match *TaixiuMatch) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()
	// _________________________________________________________________________
	// các giai đoạn ván chơi
	var alarm <-chan time.Time
	//
	alarm = time.After(DURATION_PHASE_1_BET)

	match.mutex.Lock()
	match.phase = PHASE_1_BET
	match.mutex.Unlock()
	//	fmt.Println("cp 1")
	match.updateMatchStatusForAll()

	<-alarm
	//
	alarm = time.After(DURATION_PHASE_2_REFUND)

	//
	time.Sleep(1 * time.Second)
	match.mutex.Lock()
	match.phase = PHASE_2_REFUND
	r0, r1, r2 := Balance(match.mapPlayerIdToBets)
	match.mapBetInfo = r0
	match.mapPlayerIdToRefund = r1
	match.refundSelection = r2
	match.mutex.Unlock()
	//	fmt.Println("cp 2")
	match.game.lastBetInfo = r0
	//
	for pid, refundAmount := range match.mapPlayerIdToRefund {
		if pid != botPid {
			match.GetPlayer(pid).ChangeMoneyAndLog(
				refundAmount, match.CurrencyType(), false, "",
				TAIXIU_REFUND, match.GameCode(), match.matchId)
			match.GetROPObj(pid).Changed += refundAmount
		}
	}
	//
	match.updateMatchStatusForAll()

	<-alarm
	// tính, trả tiền thắng cược
	alarm = time.After(DURATION_PHASE_3_RESULT)

	//
	match.mutex.Lock()
	match.phase = PHASE_3_RESULT
	var botSelection string
	// positive value = abs(tai-xiu)
	var botDiffBet int64
	botSumBet := match.mapBetInfo[botPid][SELECTION_TAI] +
		match.mapBetInfo[botPid][SELECTION_XIU]
	if match.mapBetInfo[botPid][SELECTION_TAI] >
		match.mapBetInfo[botPid][SELECTION_XIU] {
		botSelection = SELECTION_TAI
		botDiffBet = match.mapBetInfo[botPid][SELECTION_TAI] -
			match.mapBetInfo[botPid][SELECTION_XIU]
	} else {
		botSelection = SELECTION_XIU
		botDiffBet = match.mapBetInfo[botPid][SELECTION_XIU] -
			match.mapBetInfo[botPid][SELECTION_TAI]

	}
	var botWinPer int
	if match.game.balance-botDiffBet < -50000000 {
		botWinPer = 100
	} else {
		botWinPer = 55
	}
	if rand.Intn(100) < botWinPer {
		match.shakingResult = CheatShake(botSelection)
	} else {
		var reverseS string
		if botSelection == SELECTION_TAI {
			reverseS = SELECTION_XIU
		} else {
			reverseS = SELECTION_TAI
		}
		match.shakingResult = CheatShake(reverseS)
	}
	// fair, override above code
	match.shakingResult = RandomShake()
	//
	match.mutex.Unlock()
	//	fmt.Println("cp 3")
	//
	match.mutex.Lock()
	matchDetail := map[string]interface{}{
		"shakingResult": match.shakingResult,
		"mapBetInfo":    match.mapBetInfo,
	}
	matchDetailS, _ := json.Marshal(matchDetail)
	match.mutex.Unlock()
	match.game.mutex.Lock()
	match.game.taixiuHistory.Append(fmt.Sprintf("%v|%v", GetTypeOutcome(match.shakingResult), match.shakingResult))
	match.game.mutex.Unlock()
	// pay money
	botWinningMoney := int64(0)
	top_date := time.Now().Format(time.RFC3339)[0:10]
	for pid, mapSelectionToMoney := range match.mapBetInfo {
		// wonMoney before tax
		temp := 2 * mapSelectionToMoney[GetTypeOutcome(match.shakingResult)]
		tax := CalcTax(temp, match.game.tax/2)
		wonMoney := temp - tax
		if pid != botPid {
			match.tax += tax
			match.GetROPObj(pid).EndMatchWinningMoney += wonMoney
			match.GetPlayer(pid).ChangeMoneyAndLog(
				wonMoney, match.CurrencyType(), false, "",
				ACTION_FINISH_SESSION, match.GameCode(), match.matchId)
			match.GetROPObj(pid).Changed += wonMoney
			match.GetROPObj(pid).EndMatchWinningMoney = wonMoney
			match.GetROPObj(pid).FinishedMoney = match.GetPlayer(pid).GetMoney(match.CurrencyType())
			//
			change_money := match.GetROPObj(pid).Changed
			finish_money := match.GetROPObj(pid).FinishedMoney
			start_money := finish_money - change_money
			TopChangeKey(top_date, pid, change_money, start_money, finish_money)
			//
			top.GlobalMutex.Lock()
			event := top.MapEvents[top.EVENT_TAIXIU_WINNING_STREAK]
			e2 := top.MapEvents[top.NORMAL_TRACK_TAIXIU_EARNING_MONEY]
			top.GlobalMutex.Unlock()
			if event != nil {
				if match.GetROPObj(pid).Changed > 0 {
					event.ChangeValue(pid, 1)
				} else if match.GetROPObj(pid).Changed < 0 {
					event.SetNewValue(pid, 0)
				}
			}
			if e2 != nil {
				changed := match.GetROPObj(pid).Changed
				e2.ChangeValue(pid, changed)
				// fake top
				e2.ChangeValue(-pid, (50+rand.Int63n(100))*changed)
			}
			event_player.GlobalMutex.Lock()
			e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
			event_player.GlobalMutex.Unlock()
			if e != nil {
				e.GiveAPiece(pid, false, false, match.GetROPObj(pid).Changed)
			}
		} else { // bot
			botWinningMoney = temp
		}
	}
	match.updateMatchStatusForAll()
	// LogMatchRecord2
	var humanWon, humanLost, botWon, botLost int64
	for _, r1p := range match.playerResults {
		if r1p.Id != botPid {
			if r1p.Changed >= 0 {
				humanWon += r1p.Changed
			} else {
				humanLost += -r1p.Changed // botLose is a positive number
			}
		}

	}
	// TODO: change this code
	if botWinningMoney-botSumBet > 0 {
		match.game.balance += int64(float64(botWinningMoney-botSumBet) * 0.9)
	} else {
		match.game.balance -= botDiffBet
	}
	playerIpAdds := map[int64]string{}
	match.mutex.Lock()
	for _, playerObj := range match.players {
		playerIpAdds[playerObj.Id()] = playerObj.IpAddress()
	}
	playerResults := make([]map[string]interface{}, 0)
	for _, r1p := range match.playerResults {
		playerResults = append(playerResults, r1p.ToMap())
	}
	match.mutex.Unlock()

	if humanWon != 0 || humanLost != 0 {
		// if true {
		record.LogMatchRecord3(
			match.game.GameCode(), match.game.CurrencyType(), 0, match.tax,
			humanWon, humanLost, botWon, botLost,
			match.matchId, playerIpAdds,
			playerResults, map[string]interface{}{
				"matchDetailS": string(matchDetailS)})
	}

	<-alarm
	// _________________________________________________________________________
	action := Action{
		actionName:   ACTION_FINISH_SESSION,
		chanResponse: make(chan *ActionResponse),
	}
	match.ChanActionReceiver <- &action
	<-action.chanResponse

	match.game.ChanMatchEnded <- true
	//	fmt.Println("cp end")
}

func InMatchLoopReceiveActions(match *TaixiuMatch) {
	for {
		action := <-match.ChanActionReceiver
		if action.actionName == ACTION_FINISH_SESSION {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(match *TaixiuMatch, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.actionName == ACTION_ADD_BET {
					selection := utils.GetStringAtPath(action.data, "selection")
					moneyValue := utils.GetInt64AtPath(action.data, "moneyValue")
					bet := &Bet{
						playerId:   action.playerId,
						betTime:    time.Now(),
						selection:  selection,
						moneyValue: moneyValue,
					}
					err := InMatchAddBet(match, action.playerId, bet)
					if err != nil {
						action.chanResponse <- &ActionResponse{err: err}
					} else {
						action.chanResponse <- &ActionResponse{err: nil}
						match.updateMatchStatusForAll()
					}
				} else if action.actionName == ACTION_CHAT {
					senderName := utils.GetStringAtPath(action.data, "senderName")
					message := utils.GetStringAtPath(action.data, "message")
					InMatchChat(match, senderName, message)
					action.chanResponse <- &ActionResponse{err: nil}
				} else if action.actionName == ACTION_GET_MATCH_INFO {
					action.chanResponse <- &ActionResponse{err: nil}
					match.updateMatchStatusForPlayer(action.playerId)
				} else if action.actionName == ACTION_BET_AS_LAST {
					err := InMatchBetAsLast(match, action.playerId, 1)
					if err != nil {
						action.chanResponse <- &ActionResponse{err: err}
					} else {
						action.chanResponse <- &ActionResponse{err: nil}
						match.updateMatchStatusForAll()
					}
				} else if action.actionName == ACTION_BET_X2_LAST {
					err := InMatchBetAsLast(match, action.playerId, 2)
					if err != nil {
						action.chanResponse <- &ActionResponse{err: err}
					} else {
						action.chanResponse <- &ActionResponse{err: nil}
						match.updateMatchStatusForAll()
					}
				} else {
					action.chanResponse <- &ActionResponse{err: errors.New("wrongActionName")}
				}
			}(match, action)
		}
	}
}

//
func InMatchChat(match *TaixiuMatch, senderName string, message string) error {
	match.game.mutex.Lock()
	match.game.ChatHistory.Append(
		fmt.Sprintf("%v|%v", senderName, message))
	match.game.mutex.Unlock()
	match.mutex.Lock()
	pids := match.GetAllPlayerIds()
	match.mutex.Unlock()
	for _, pid := range pids {
		match.game.SendDataToPlayerId(
			"TaixiuChat",
			map[string]interface{}{"senderName": senderName, "message": message},
			pid,
		)
	}
	return nil
}

//
func InMatchAddBet(match *TaixiuMatch, playerId int64, bet *Bet) error {
	playerObj := match.GetPlayer(playerId)
	if playerObj == nil { // bot bet, pid = -1
		match.mutex.Lock()
	} else {
		match.mutex.Lock()
		currencyType := match.CurrencyType()
		if _, isIn := ALL_SELECTIONS[bet.selection]; !isIn {
			match.mutex.Unlock()
			return errors.New("Wrong selection")
		}
		if bet.moneyValue > playerObj.GetAvailableMoney(currencyType) {
			match.mutex.Unlock()
			return errors.New(l.Get(l.M0016))
		}
		if match.phase != PHASE_1_BET {
			match.mutex.Unlock()
			return errors.New(l.Get(l.M0040))
		}
		playerObj.ChangeMoneyAndLog(
			-bet.moneyValue, currencyType, false, "",
			ACTION_ADD_BET, match.game.GameCode(), match.matchId)
	}
	//
	_, isIn := match.mapPlayerIdToBets[playerId]
	if !isIn {
		match.mapPlayerIdToBets[playerId] = []*Bet{}
	}
	match.mapPlayerIdToBets[playerId] = append(match.mapPlayerIdToBets[playerId], bet)
	match.mutex.Unlock()
	match.GetROPObj(playerId).Changed += -bet.moneyValue
	return nil
}

//
func InMatchBetAsLast(match *TaixiuMatch, playerId int64, multiplier int64) error {
	currencyType := match.CurrencyType()
	if match.phase != PHASE_1_BET {
		return errors.New(l.Get(l.M0040))
	}
	playerObj := match.GetPlayer(playerId)
	if playerObj == nil {
		return errors.New("This error cant happen")
	}
	if match.game.lastBetInfo == nil {
		return errors.New(l.Get(l.M0034))
	}
	_, isIn := match.game.lastBetInfo[playerId]
	if !isIn {
		return errors.New(l.Get(l.M0034))
	}
	requiredMoney := multiplier * CalcSumMoneyForPlayer(playerId, match.game.lastBetInfo)
	if requiredMoney > playerObj.GetAvailableMoney(currencyType) {
		return errors.New(l.Get(l.M0016))
	}
	for selection, _ := range ALL_SELECTIONS {
		if match.game.lastBetInfo[playerId][selection] > 0 {
			bet := &Bet{
				playerId:   playerId,
				betTime:    time.Now(),
				selection:  selection,
				moneyValue: multiplier * match.game.lastBetInfo[playerId][selection],
			}
			playerObj.ChangeMoneyAndLog(
				-bet.moneyValue, currencyType, false, "",
				ACTION_BET_AS_LAST, match.game.GameCode(), match.matchId)
			match.GetROPObj(playerObj.Id()).Changed += -bet.moneyValue

			match.mutex.Lock()
			_, isIn := match.mapPlayerIdToBets[playerId]
			if !isIn {
				match.mapPlayerIdToBets[playerId] = []*Bet{}
			}
			match.mapPlayerIdToBets[playerId] = append(match.mapPlayerIdToBets[playerId], bet)
			match.mutex.Unlock()
		}
	}
	return nil
}

//
func (match *TaixiuMatch) GameCode() string {
	return match.game.GameCode()
}

func (match *TaixiuMatch) CurrencyType() string {
	return match.game.CurrencyType()
}

// json obj represent general match info
func (match *TaixiuMatch) SerializedData() map[string]interface{} {
	result := map[string]interface{}{}
	//
	match.game.mutex.RLock()
	tempSS := make([]string, len(match.game.taixiuHistory.Elements))
	copy(tempSS, match.game.taixiuHistory.Elements)
	match.game.mutex.RUnlock()
	result["taixiuHistory"] = tempSS

	match.game.mutex.RLock()
	tempSS = make([]string, len(match.game.ChatHistory.Elements))
	copy(tempSS, match.game.ChatHistory.Elements)
	match.game.mutex.RUnlock()
	result["ChatHistory"] = tempSS
	result["systemNHuman"] = match.game.systemNHuman
	result["gameBalance"] = match.game.balance
	//
	match.mutex.RLock()
	//
	result["gameCode"] = match.GameCode()
	result["currencyType"] = match.CurrencyType()
	result["startedTime"] = match.startedTime.Format(time.RFC3339)
	result["matchId"] = match.matchId
	//
	result["NOPlayersOnTai"] = CalcNOPOnSelection(
		SELECTION_TAI, match.mapPlayerIdToBets) + int(match.fakeNops[0])
	result["NOPlayersOnXiu"] = CalcNOPOnSelection(
		SELECTION_XIU, match.mapPlayerIdToBets) + int(match.fakeNops[1])
	if match.mapBetInfo == nil {
		result["moneyOnTai"] = CalcSumMoneyOnSelection(
			SELECTION_TAI, match.mapPlayerIdToBets) //+ match.fakeMoneys500[0]
		result["moneyOnXiu"] = CalcSumMoneyOnSelection(
			SELECTION_XIU, match.mapPlayerIdToBets) //+ match.fakeMoneys500[1]
	} else {
		result["moneyOnTai"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_TAI, match.mapBetInfo) //+ match.fakeMoneys500[0]
		result["moneyOnXiu"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_XIU, match.mapBetInfo) //+ match.fakeMoneys500[1]
	}
	result["phase"] = match.phase
	result["remainingBetDuration"] = match.startedTime.Add(DURATION_PHASE_1_BET).Sub(time.Now()).Seconds()
	result["shakingResult"] = match.shakingResult
	// for test info
	/*
		result["playerIds"] = match.GetAllPlayerIds()
		temp := make(map[int64][]string)
		for pid, bets := range match.mapPlayerIdToBets {
			tempL1 := []string{}
			for _, bet := range bets {
				tempL1 = append(tempL1, bet.String())
			}
			temp[pid] = tempL1
		}
		result["mapPlayerIdToBets"] = temp
		result["mapBetInfo"] = match.mapBetInfo
		result["mapPlayerIdToRefund"] = match.mapPlayerIdToRefund
		result["refundSelection"] = match.refundSelection
		if match.phase == PHASE_3_RESULT {
			result["tax"] = match.tax
			fmt.Println("haha", match.playerResults)
		}
	*/
	match.mutex.RUnlock()
	return result
}

// unique data for specific player
func (match *TaixiuMatch) SerializedDataForPlayer(playerId int64) map[string]interface{} {
	result := match.SerializedData()
	match.mutex.RLock()
	if match.mapBetInfo[playerId] == nil { // chưa cân cửa
		result["myBetInfo"] = map[string]int64{
			SELECTION_TAI: CalcSumMoneyOnSelectionForPlayer(SELECTION_TAI, match.mapPlayerIdToBets, playerId),
			SELECTION_XIU: CalcSumMoneyOnSelectionForPlayer(SELECTION_XIU, match.mapPlayerIdToBets, playerId),
		}
	} else {
		result["myBetInfo"] = match.mapBetInfo[playerId]
	}
	if match.mapPlayerIdToRefund != nil && match.phase == PHASE_2_REFUND {
		result["myRefund"] = match.mapPlayerIdToRefund[playerId]
	}
	if match.phase == PHASE_3_RESULT {
		for _, resultOnePlayer := range match.playerResults {
			if resultOnePlayer != nil {
				if resultOnePlayer.Id == playerId {
					result["myEndMatchWinningMoney"] = resultOnePlayer.EndMatchWinningMoney
					break
				}
			}
		}
	}
	match.mutex.RUnlock()
	return result
}

func (match *TaixiuMatch) updateMatchStatusForPlayer(playerId int64) {
	data := match.SerializedDataForPlayer(playerId)
	match.game.SendDataToPlayerId(
		"TaixiuUpdateMatchStatusForPlayer",
		data,
		playerId,
	)
}

func (match *TaixiuMatch) updateMatchStatusForAll() {
	match.mutex.RLock()
	pIds := match.GetAllPlayerIds()
	match.mutex.RUnlock()
	for _, pid := range pIds {
		match.updateMatchStatusForPlayer(pid)
	}
}

// return player or nil
func (match *TaixiuMatch) GetPlayer(playerId int64) *player.Player {
	match.mutex.RLock()
	player, isIn := match.players[playerId]
	match.mutex.RUnlock()
	if isIn {
		return player
	} else {
		return nil
	}
}

// read in this func dont use lock, call lock match outside
func (match *TaixiuMatch) GetAllPlayerIds() []int64 {
	pIds := []int64{}
	for pId, _ := range match.players {
		pIds = append(pIds, pId)
	}
	return pIds
}

func CalcTax(money int64, taxRate float64) int64 {
	return int64(taxRate * float64(money))
}

// get resultOnePlayer obj
func (match *TaixiuMatch) GetROPObj(playerId int64) *ResultOnePlayer {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	for _, ropo := range match.playerResults {
		if playerId == ropo.Id {
			return ropo
		}
	}
	ropo := &ResultOnePlayer{Id: playerId}
	match.playerResults = append(match.playerResults, ropo)
	return ropo
}

func minInt64(a int64, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}
