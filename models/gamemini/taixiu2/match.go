package baucua

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	//"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
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

	DURATION_PHASE_1_BET    = time.Duration(40 * time.Second)
	DURATION_PHASE_3_RESULT = time.Duration(10 * time.Second)

	TAIXIU_REFUND = "TAIXIU_REFUND"
)

func init() {
	fmt.Print("")
	_ = rand.Int()
}

type ResultOnePlayer struct {
	Id       int64
	Username string
	// tiền cộng cuối trận
	EndMatchWinningMoney int64
	FinishedMoney        int64
	// tổng tiền thay đổi trong trận đầu này (bao gồm cả tiền mất)
	Changed int64
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
	// giai đoạn của ván chơi PHASE_
	phase string
	// kết quả lắc xúc xắc
	shakingResult []int

	// fake info
	moneySteps []int64
	fakeMoneys []int64
	// need money is 500s
	fakeMoneys500 []int64
	fakeNops      []int

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
	match.moneySteps = make([]int64, 6)
	match.fakeMoneys = make([]int64, 6)
	match.fakeMoneys500 = make([]int64, 6)
	match.fakeNops = make([]int, 6)
	durationBetPhaseInSeconds := int64(DURATION_PHASE_1_BET/time.Second) - 2
	for i := 0; i < 6; i++ {
		endingMoney := 500000 + rand.Int63n(1600000)
		moneyStep := endingMoney / durationBetPhaseInSeconds
		match.moneySteps[i] = moneyStep
	}
	go func() {
		for s := int64(0); s <= durationBetPhaseInSeconds; s++ {
			time.Sleep(1 * time.Second)
			match.mutex.Lock()
			for i := 0; i < 6; i++ {
				match.fakeMoneys[i] += match.moneySteps[i] +
					rand.Int63n(20000) - 10000
				match.fakeMoneys500[i] = match.fakeMoneys[i] / 500 * 500
				match.fakeNops[i] += rand.Intn(2)
			}
			match.mutex.Unlock()
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
	match.updateMatchStatusForAll()

	<-alarm

	//
	time.Sleep(1 * time.Second)
	match.mutex.Lock()
	match.phase = PHASE_2_REFUND
	r0 := Balance(match.mapPlayerIdToBets)
	match.mapBetInfo = r0
	match.mutex.Unlock()
	match.game.lastBetInfo = r0

	// tính, trả tiền thắng cược
	match.phase = PHASE_3_RESULT
	alarm = time.After(DURATION_PHASE_3_RESULT)
	nTry := 0
	for {
		nTry += 1
		match.mutex.Lock()
		match.shakingResult = RandomShake()
		isGoodBalance := true
		// map[pid]changedMoney
		mapTempChanged := make(map[int64]int64)
		mapTempEmwm := CalcEndMatchWinningMoney(match.mapBetInfo, match.shakingResult)
		match.mutex.Unlock()
		for pid, emwm := range mapTempEmwm {
			mapTempChanged[pid] = match.GetROPObj(pid).Changed + emwm
		}
		tempSumChange := int64(0)
		for _, moneyChange := range mapTempChanged {
			tempSumChange += moneyChange
		}
		newBalance := match.game.balance - tempSumChange
		newSumUserBets := match.game.sumUserBets
		if newBalance < int64(match.game.stealingRate*float64(newSumUserBets)) {
			isGoodBalance = false
		}
		//		fmt.Println("shakingResult, mapTempChanged, isGoodBalance, nTry  ", match.shakingResult, mapTempChanged, isGoodBalance, nTry)
		if isGoodBalance || nTry > 100 || tempSumChange == 0 {
			break
		} else {
			if rand.Intn(100) < 20 {
				break
			}
		}
	}
	//	fmt.Println("baucua nTry", nTry)
	//
	match.mutex.Lock()
	matchDetail := map[string]interface{}{
		"shakingResult": match.shakingResult,
		"mapBetInfo":    match.mapBetInfo,
	}
	matchDetailS, errJsonDump := json.Marshal(matchDetail)
	match.mutex.Unlock()
	match.game.mutex.Lock()
	if len(match.shakingResult) >= 3 { // sure true
		match.game.taixiuHistory.Append(fmt.Sprintf("%v|%v|%v",
			match.shakingResult[0],
			match.shakingResult[1],
			match.shakingResult[2]))
		if errJsonDump == nil {
			match.game.history2.Append(string(matchDetailS))
		}
	}
	match.game.mutex.Unlock()
	// pay money
	match.game.mutex.Lock()
	mapEmwm := CalcEndMatchWinningMoney(match.mapBetInfo, match.shakingResult)
	match.game.mutex.Unlock()
	for pid, emwm := range mapEmwm {
		//
		tax := CalcTax(emwm, match.game.tax/2)
		wonMoney := emwm - tax
		match.tax += tax
		match.GetPlayer(pid).ChangeMoneyAndLog(
			wonMoney, match.CurrencyType(), false, "",
			ACTION_FINISH_SESSION, match.GameCode(), match.matchId)
		match.GetROPObj(pid).Changed += wonMoney
		match.GetROPObj(pid).EndMatchWinningMoney = wonMoney
		match.GetROPObj(pid).FinishedMoney = match.GetPlayer(pid).GetMoney(match.CurrencyType())

		event_player.GlobalMutex.Lock()
		e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
		event_player.GlobalMutex.Unlock()
		if e != nil {
			e.GiveAPiece(pid, false, false, match.GetROPObj(pid).Changed)
		}
	}
	match.updateMatchStatusForAll()
	// LogMatchRecord2
	var humanWon, humanLost, botWon, botLost int64
	for _, r1p := range match.playerResults {
		if match.GetPlayer(r1p.Id).PlayerType() == "bot" {
			if r1p.Changed >= 0 {
				botWon += r1p.Changed
			} else {
				botLost += -r1p.Changed // botLost is a positive number
			}
		} else {
			if r1p.Changed >= 0 {
				humanWon += r1p.Changed
			} else {
				humanLost += -r1p.Changed // humanLost is a positive number
			}
		}
	}
	match.game.balance += humanLost - humanWon
	match.game.sumUserBets += humanWon
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
				} else if action.actionName == ACTION_GET_MATCH_INFO {
					action.chanResponse <- &ActionResponse{err: nil}
					match.updateMatchStatusForPlayer(action.playerId)
				} else if action.actionName == ACTION_CHAT {
					senderName := utils.GetStringAtPath(action.data, "senderName")
					message := utils.GetStringAtPath(action.data, "message")
					InMatchChat(match, senderName, message)
					action.chanResponse <- &ActionResponse{err: nil}
				} else {
					action.chanResponse <- &ActionResponse{err: errors.New("wrongActionName")}
				}
			}(match, action)
		}
	}
}

//
func InMatchChat(match *TaixiuMatch, senderName string, message string) error {
	match.mutex.Lock()
	pids := match.GetAllPlayerIds()
	match.mutex.Unlock()
	for _, pid := range pids {
		match.game.SendDataToPlayerId(
			"BaucuaChat",
			map[string]interface{}{"senderName": senderName, "message": message},
			pid,
		)
	}
	return nil
}

//
func InMatchAddBet(match *TaixiuMatch, playerId int64, bet *Bet) error {
	currencyType := match.CurrencyType()
	if match.phase != PHASE_1_BET {
		return errors.New(l.Get(l.M0040))
	}
	playerObj := match.GetPlayer(playerId)
	if playerObj == nil {
		return errors.New("This error cant happen")
	}
	if _, isIn := ALL_SELECTIONS[bet.selection]; !isIn {
		return errors.New("Wrong selection")
	}
	if bet.moneyValue > playerObj.GetAvailableMoney(currencyType) {
		return errors.New(l.Get(l.M0016))
	}
	//
	playerObj.ChangeMoneyAndLog(
		-bet.moneyValue, currencyType, false, "",
		ACTION_ADD_BET, match.game.GameCode(), match.matchId)
	match.GetROPObj(playerObj.Id()).Changed += -bet.moneyValue

	match.mutex.Lock()
	_, isIn := match.mapPlayerIdToBets[playerId]
	if !isIn {
		match.mapPlayerIdToBets[playerId] = []*Bet{}
	}
	match.mapPlayerIdToBets[playerId] = append(match.mapPlayerIdToBets[playerId], bet)
	match.mutex.Unlock()
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
	//
	match.mutex.RLock()
	//
	result["gameCode"] = match.GameCode()
	result["currencyType"] = match.CurrencyType()
	result["startedTime"] = match.startedTime.Format(time.RFC3339)
	result["matchId"] = match.matchId
	//
	result["NOPlayersOn1"] = CalcNOPOnSelection(
		SELECTION_1, match.mapPlayerIdToBets) + match.fakeNops[0]
	result["NOPlayersOn2"] = CalcNOPOnSelection(
		SELECTION_2, match.mapPlayerIdToBets) + match.fakeNops[1]
	result["NOPlayersOn3"] = CalcNOPOnSelection(
		SELECTION_3, match.mapPlayerIdToBets) + match.fakeNops[2]
	result["NOPlayersOn4"] = CalcNOPOnSelection(
		SELECTION_4, match.mapPlayerIdToBets) + match.fakeNops[3]
	result["NOPlayersOn5"] = CalcNOPOnSelection(
		SELECTION_5, match.mapPlayerIdToBets) + match.fakeNops[4]
	result["NOPlayersOn6"] = CalcNOPOnSelection(
		SELECTION_6, match.mapPlayerIdToBets) + match.fakeNops[5]
	if match.mapBetInfo == nil {
		result["moneyOn1"] = CalcSumMoneyOnSelection(
			SELECTION_1, match.mapPlayerIdToBets) + match.fakeMoneys500[0]
		result["moneyOn2"] = CalcSumMoneyOnSelection(
			SELECTION_2, match.mapPlayerIdToBets) + match.fakeMoneys500[1]
		result["moneyOn3"] = CalcSumMoneyOnSelection(
			SELECTION_3, match.mapPlayerIdToBets) + match.fakeMoneys500[2]
		result["moneyOn4"] = CalcSumMoneyOnSelection(
			SELECTION_4, match.mapPlayerIdToBets) + match.fakeMoneys500[3]
		result["moneyOn5"] = CalcSumMoneyOnSelection(
			SELECTION_5, match.mapPlayerIdToBets) + match.fakeMoneys500[4]
		result["moneyOn6"] = CalcSumMoneyOnSelection(
			SELECTION_6, match.mapPlayerIdToBets) + match.fakeMoneys500[5]
	} else {
		result["moneyOn1"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_1, match.mapBetInfo) + match.fakeMoneys500[0]
		result["moneyOn2"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_2, match.mapBetInfo) + match.fakeMoneys500[1]
		result["moneyOn3"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_3, match.mapBetInfo) + match.fakeMoneys500[2]
		result["moneyOn4"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_4, match.mapBetInfo) + match.fakeMoneys500[3]
		result["moneyOn5"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_5, match.mapBetInfo) + match.fakeMoneys500[4]
		result["moneyOn6"] = CalcBalancedSumMoneyOnSelection(
			SELECTION_6, match.mapBetInfo) + match.fakeMoneys500[5]
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
			SELECTION_1: CalcSumMoneyOnSelectionForPlayer(SELECTION_1, match.mapPlayerIdToBets, playerId),
			SELECTION_2: CalcSumMoneyOnSelectionForPlayer(SELECTION_2, match.mapPlayerIdToBets, playerId),
			SELECTION_3: CalcSumMoneyOnSelectionForPlayer(SELECTION_3, match.mapPlayerIdToBets, playerId),
			SELECTION_4: CalcSumMoneyOnSelectionForPlayer(SELECTION_4, match.mapPlayerIdToBets, playerId),
			SELECTION_5: CalcSumMoneyOnSelectionForPlayer(SELECTION_5, match.mapPlayerIdToBets, playerId),
			SELECTION_6: CalcSumMoneyOnSelectionForPlayer(SELECTION_6, match.mapPlayerIdToBets, playerId),
		}
	} else {
		result["myBetInfo"] = match.mapBetInfo[playerId]
	}
	if match.phase == PHASE_3_RESULT {
		for _, resultOnePlayer := range match.playerResults {
			if resultOnePlayer != nil {
				if resultOnePlayer.Id == playerId {
					result["myEndMatchWinningMoney"] = resultOnePlayer.EndMatchWinningMoney
					result["changed"] = resultOnePlayer.Changed
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
		"BaucuaUpdateMatchStatusForPlayer",
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

// include match.mutex, return player or nil
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

// include match.mutex, get resultOnePlayer obj
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
