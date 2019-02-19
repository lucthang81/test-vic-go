package xocdia2

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

const (
	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION" // not for players

	ACTION_BECOME_HOST     = "ACTION_BECOME_HOST"
	ACTION_ADD_BET         = "ACTION_ADD_BET"
	ACTION_BET_EQUAL_LAST  = "ACTION_BET_AS_LAST"
	ACTION_BET_DOUBLE_LAST = "ACTION_BET_DOUBLE_LAST"
	ACTION_ACCEPT_BET      = "ACTION_ACCEPT_BET"

	PHASE_0_BECOME_HOST       = "PHASE_0_BECOME_HOST"
	PHASE_1_BET               = "PHASE_1_BET"
	PHASE_11_HOST_ACCEPT_BET  = "PHASE_11_HOST_ACCEPT_BET"
	PHASE_12_OTHER_ACCEPT_BET = "PHASE_12_OTHER_ACCEPT_BET"
	PHASE_2_SHAKE             = "PHASE_2_SHAKE"
	PHASE_3_RESULT            = "PHASE_3_RESULT"

	// reason for money change log
	XOCDIA_REFUND = "XOCDIA_REFUND"
)

type Action struct {
	actionName string
	playerId   int64

	data         map[string]interface{}
	responseChan chan *ActionResponse
}

func (action *Action) ToString() string {
	dataJson, err := json.Marshal(action.data)
	if err != nil {
		return ""
	}
	result, err := json.Marshal(map[string]interface{}{
		"actionTime": time.Now(),

		"actionName": action.actionName,
		"playerId":   action.playerId,
		"data":       dataJson,
	})
	if err != nil {
		return ""
	}
	return string(result)
}

type ActionResponse struct {
	err  error
	data map[string]interface{}
}

type ResultOnePlayer struct {
	Id       int64
	Username string

	FinishedMoney        int64
	EndMatchWinningMoney int64

	Changed int64 // tổng tiền thay đổi trong trận đầu này; gồm cả tiền bet, tiền trả về
}

func (r *ResultOnePlayer) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = r.Id
	result["username"] = r.Username
	result["FinishedMoney"] = r.FinishedMoney
	result["EndMatchWinningMoney"] = r.EndMatchWinningMoney
	result["change"] = r.Changed
	return result
}

type XocdiaSession struct {
	game *XocdiaGame
	room *game.Room

	startedTime time.Time
	matchId     string
	players     map[game.GamePlayer]int // include owner, map playerId to room sitting position

	phase         string
	playerResults []*ResultOnePlayer
	tax           int64 // sum tax

	hostPlayerId         int64  // 0 tức là 0 có host
	hostAcceptEvenOrOdd  string // SELECTION_ or ""
	hostMoneyOnEvenOrOdd int64
	betInfo              map[int64]map[string]int64 // map[playerId](map[selection]soTienCuoc), not include host
	acceptedMap          map[string]int64           // id của người cân 4 cửa ăn to, map[selection]playerId
	shakingResult        string                     // outcome

	ActionChan chan *Action // receive player action

	mutex sync.RWMutex
}

func NewXocdiaSession(gameInstance *XocdiaGame, room *game.Room) *XocdiaSession {
	session := &XocdiaSession{
		game: gameInstance,
		room: room,

		startedTime: time.Now(),
		matchId:     fmt.Sprintf("#%v", time.Now().Unix()),

		phase: "",
		tax:   0,

		ActionChan: make(chan *Action),
	}

	//
	session.players = make(map[game.GamePlayer]int)
	for roomPosition, player := range session.room.Players().Copy() {
		if player != nil {
			session.players[player] = roomPosition
		}
	}

	session.room.Mutex.Lock()
	hostId, _ := session.room.SharedData["hostId"].(int64)
	session.room.Mutex.Unlock()
	if session.GetPlayer(hostId) != nil {
		session.hostPlayerId = hostId
	} else {
		session.hostPlayerId = 0
		session.room.Mutex.Lock()
		session.room.SharedData["hostId"] = int64(0)
		session.room.Mutex.Unlock()
	}

	session.betInfo = make(map[int64]map[string]int64)
	for player, _ := range session.players {
		if player.Id() != session.hostPlayerId {
			session.betInfo[player.Id()] = make(map[string]int64)
			session.betInfo[player.Id()][SELECTION_EVEN] = 0
			session.betInfo[player.Id()][SELECTION_ODD] = 0
			session.betInfo[player.Id()][SELECTION_0_RED] = 0
			session.betInfo[player.Id()][SELECTION_1_RED] = 0
			session.betInfo[player.Id()][SELECTION_3_RED] = 0
			session.betInfo[player.Id()][SELECTION_4_RED] = 0
		}
	}

	session.acceptedMap = make(map[string]int64)
	for _, selection := range []string{SELECTION_0_RED, SELECTION_1_RED, SELECTION_3_RED, SELECTION_4_RED} {
		session.acceptedMap[selection] = 0
	}

	session.playerResults = make([]*ResultOnePlayer, 0)
	for player, _ := range session.players {
		session.playerResults = append(
			session.playerResults,
			&ResultOnePlayer{
				Id:       player.Id(),
				Username: player.Name(),
			})
	}

	//
	go Start(session)
	go InMatchLoopReceiveActions(session)
	//
	return session
}

// main match flow
func Start(session *XocdiaSession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	currencyType := session.game.CurrencyType()
	session.room.Mutex.RLock()
	// room.betEntry
	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	session.room.Mutex.RUnlock() // init actions when change phase
	// _________________________________________________________________________
	durationPhase0 := session.game.durationPhase0
	durationPhase1 := session.game.durationPhase1
	durationPhase11 := session.game.durationPhase11
	durationPhase12 := session.game.durationPhase12
	durationPhase2 := session.game.durationPhase2
	durationPhase3 := session.game.durationPhase3
	var alarm <-chan time.Time
	// _________________________________________________________________________
	session.room.DidStartGame(session)
	// _________________________________________________________________________
	maxMoney := int64(0)
	for player, _ := range session.players {
		if maxMoney < player.GetMoney(currencyType) {
			maxMoney = player.GetMoney(currencyType)
		}
	}
	session.room.Mutex.Lock()
	cond := (session.room.SharedData["hostId"].(int64) == 0) &&
		(maxMoney > session.room.SharedData["hostMinMoney"].(int64))
	session.room.Mutex.Unlock()
	if cond {
		alarm = time.After(durationPhase0)

		session.phase = PHASE_0_BECOME_HOST
		session.room.DidChangeGameState(session)

		<-alarm
	}
	// _________________________________________________________________________
	alarm = time.After(durationPhase1)

	session.phase = PHASE_1_BET
	session.room.DidChangeGameState(session)

	<-alarm
	session.mutex.RLock()
	temp := CopyAllBet(session.betInfo)
	session.mutex.RUnlock()
	session.room.Mutex.Lock()
	session.room.SharedData["lastBetInfo"] = temp
	session.room.Mutex.Unlock()
	// _________________________________________________________________________
	if session.hostPlayerId != 0 {
		alarm = time.After(durationPhase11)

		session.phase = PHASE_11_HOST_ACCEPT_BET
		session.room.DidChangeGameState(session)

		<-alarm
	}
	// _________________________________________________________________________
	if session.hostAcceptEvenOrOdd == "" {
		session.mutex.Lock()
		sumE := GetAllBetOn1Selection(session.betInfo, SELECTION_EVEN)
		sumA := GetAllBetOn1Selection(session.betInfo, SELECTION_ODD)
		session.mutex.Unlock()
		if sumA > sumE {
			refundAmount := sumA - sumE
			var refundRatio float64
			if sumA > 0 {
				refundRatio = float64(refundAmount) / float64(sumA)
			} else {
				refundRatio = 0
			}
			for pId, _ := range session.betInfo {
				session.mutex.Lock()
				temp := int64(refundRatio * float64(session.betInfo[pId][SELECTION_ODD]))
				session.betInfo[pId][SELECTION_ODD] -= temp
				session.mutex.Unlock()
				if temp > 0 {
					session.GetPlayer(pId).ChangeMoneyAndLog(
						temp, currencyType, false, "",
						XOCDIA_REFUND, session.game.GameCode(), session.matchId)
					session.GetROPObj(pId).Changed += temp
				}
			}
		} else if sumA < sumE {
			refundAmount := sumE - sumA
			var refundRatio float64
			if sumE > 0 {
				refundRatio = float64(refundAmount) / float64(sumE)
			} else {
				refundRatio = 0
			}
			for pId, _ := range session.betInfo {
				session.mutex.Lock()
				temp := int64(refundRatio * float64(session.betInfo[pId][SELECTION_EVEN]))
				session.betInfo[pId][SELECTION_EVEN] -= temp
				session.mutex.Unlock()
				if temp > 0 {
					session.GetPlayer(pId).ChangeMoneyAndLog(
						temp, currencyType, false, "",
						XOCDIA_REFUND, session.game.GameCode(), session.matchId)
					session.GetROPObj(pId).Changed += temp
				}
			}
		} else {

		}
	}
	// _________________________________________________________________________
	alarm = time.After(durationPhase12)

	session.phase = PHASE_12_OTHER_ACCEPT_BET
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	session.mutex.Lock()
	for selection, _ := range session.acceptedMap {
		if session.acceptedMap[selection] == 0 { // khong co ai can cac cua 4 16
			for pId, _ := range session.betInfo {
				temp := session.betInfo[pId][selection]
				session.GetPlayer(pId).ChangeMoneyAndLog(
					temp, currencyType, false, "",
					XOCDIA_REFUND, session.game.GameCode(), session.matchId)
				session.GetROPObj(pId).Changed += temp
				//
				session.betInfo[pId][selection] = 0
			}
		}
	}
	session.mutex.Unlock()
	// _________________________________________________________________________
	alarm = time.After(durationPhase2)

	session.phase = PHASE_2_SHAKE
	session.mutex.Lock()
	session.shakingResult = RandomShake()
	// cheat
	//	isGoodBalance := true
	//	tempWonMap := CalcResult(session.hostPlayerId, session.hostAcceptEvenOrOdd, session.hostMoneyOnEvenOrOdd, session.betInfo, session.acceptedMap, session.shakingResult)
	//	for pid, wonMoney := range tempWonMap {
	//		if session.GetPlayer(pid).PlayerType() == "normal" {
	//			tempWonMap[pid] = session.GetROPObj(pid).Changed + wonMoney
	//		} else {
	//			tempWonMap[pid] = 0
	//		}
	//	}
	//	tempSumChange := int64(0)
	//	for _, moneyChange := range tempWonMap {
	//		tempSumChange += moneyChange
	//	}
	//	newBalance := session.game.balance - tempSumChange
	//	newSumUserBets := session.game.sumUserBets
	//	if newBalance <= int64(session.game.stealingRate*float64(newSumUserBets)) {
	//		isGoodBalance = false
	//	}
	//	if isGoodBalance == false {
	//		var reverseTypeOutcome string
	//		if GetTypeOutcome(session.shakingResult) == SELECTION_EVEN {
	//			reverseTypeOutcome = SELECTION_ODD
	//		} else {
	//			reverseTypeOutcome = SELECTION_EVEN
	//		}
	//		session.shakingResult = CheatShake(reverseTypeOutcome)
	//	}
	session.mutex.Unlock()
	//
	session.room.Mutex.Lock()
	outcomeHistory := session.room.SharedData["outcomeHistory"].(*cardgame.SizedList)
	outcomeHistory.Append(session.shakingResult)
	session.room.SharedData["outcomeHistory"] = outcomeHistory
	session.room.Mutex.Unlock()
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	alarm = time.After(durationPhase3)

	session.phase = PHASE_3_RESULT
	session.mutex.Lock()
	mapPlayerIdToWinMoney := CalcResult(session.hostPlayerId, session.hostAcceptEvenOrOdd, session.hostMoneyOnEvenOrOdd, session.betInfo, session.acceptedMap, session.shakingResult)
	session.mutex.Unlock()
	//fmt.Println("mapPlayerIdToWinMoney", mapPlayerIdToWinMoney)
	for playerId, winMoney := range mapPlayerIdToWinMoney {
		temp := game.MoneyAfterTax(winMoney, betEntry)
		session.GetROPObj(playerId).EndMatchWinningMoney += temp
		session.GetROPObj(playerId).Changed += temp
		session.tax += winMoney - temp
	}
	for _, result1p := range session.playerResults {
		temp := result1p.EndMatchWinningMoney
		pId := result1p.Id
		session.GetPlayer(pId).ChangeMoneyAndLog(
			temp, currencyType, false, "",
			ACTION_FINISH_SESSION, session.game.GameCode(), session.matchId)
		result1p.FinishedMoney = session.GetPlayer(result1p.Id).GetMoney(currencyType)
	}
	//
	var humanWon, humanLost, botWon, botLost int64
	for _, r1p := range session.playerResults {
		if session.GetPlayer(r1p.Id).PlayerType() == "bot" {
			if r1p.Changed >= 0 {
				botWon += r1p.Changed
			} else {
				botLost += -r1p.Changed // botLose is a positive number
			}
		} else {
			if r1p.Changed >= 0 {
				humanWon += r1p.Changed
				rank.ChangeKey(rank.RANK_NUMBER_OF_WINS, r1p.Id, 1)
			} else {
				humanLost += -r1p.Changed // botLose is a positive number
			}
		}
	}
	session.game.balance += humanLost - humanWon
	session.game.sumUserBets += humanWon
	playerIpAdds := map[int64]string{}
	for playerObj, _ := range session.players {
		playerIpAdds[playerObj.Id()] = playerObj.IpAddress()
	}
	playerResults := make([]map[string]interface{}, 0)
	for _, r1p := range session.playerResults {
		playerResults = append(playerResults, r1p.ToMap())
	}
	if humanWon != 0 || humanLost != 0 {
		record.LogMatchRecord2(
			session.game.GameCode(), session.game.CurrencyType(), session.room.Requirement(), session.tax,
			humanWon, humanLost, botWon, botLost,
			session.matchId, playerIpAdds,
			playerResults)
	}
	//
	event_player.GlobalMutex.Lock()
	e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
	event_player.GlobalMutex.Unlock()
	if e != nil {
		for _, r1p := range session.playerResults {
			// fmt.Printf("xocdia eventcp %+v \n", r1p.ToMap())
			e.GiveAPiece(r1p.Id, false,
				currencyType == currency.TestMoney, r1p.Changed)
		}
	}
	// loại bỏ host thiếu tiền, host muốn thoát phòng
	session.room.Mutex.Lock()
	hostP := session.GetPlayer(session.hostPlayerId)
	if hostP != nil {
		if hostP.GetMoney(currencyType) < session.room.SharedData["hostMinMoney"].(int64) {
			session.room.SharedData["hostId"] = int64(0)
		}
		isHostWantToLeave := false
		for _, player := range session.room.GetListRegisterLeaveRoom() {
			if hostP.Id() == player.Id() {
				isHostWantToLeave = true
				break
			}
		}
		if isHostWantToLeave {
			session.room.SharedData["hostId"] = int64(0)
		}
	}
	session.room.Mutex.Unlock()
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	action := Action{
		actionName:   ACTION_FINISH_SESSION,
		responseChan: make(chan *ActionResponse),
	}
	session.ActionChan <- &action
	<-action.responseChan
	session.room.DidEndGame(map[string]interface{}{}, int(session.game.delayAfterEachGameInSeconds.Seconds())) // second to new match
}

// Interface
func (session *XocdiaSession) CleanUp() {
}

func (session *XocdiaSession) HandlePlayerRemovedFromGame(player game.GamePlayer) {

}
func (session *XocdiaSession) HandlePlayerAddedToGame(player game.GamePlayer) {

}

func (session *XocdiaSession) HandlePlayerOffline(player game.GamePlayer) {

}
func (session *XocdiaSession) HandlePlayerOnline(player game.GamePlayer) {

}

func (session *XocdiaSession) IsPlaying() bool {
	return true
}

func (session *XocdiaSession) IsDelayingForNewGame() bool {
	return true
}

func (session *XocdiaSession) SerializedData() map[string]interface{} {
	session.mutex.RLock()
	data := make(map[string]interface{})
	data["game_code"] = session.game.GameCode()
	data["requirement"] = session.room.Requirement()
	data["matchId"] = session.matchId
	data["player_ids"] = session.GetPlayerIds()
	data["phase"] = session.phase
	temp1 := make([]map[string]interface{}, 0)
	if session.phase == PHASE_3_RESULT {
		for _, result1p := range session.playerResults {
			temp1 = append(temp1, result1p.ToMap())
		}
	}

	data["betInfo"] = CopyAllBet(session.betInfo)
	data["hostPlayerId"] = session.hostPlayerId
	data["hostAcceptEvenOrOdd"] = session.hostAcceptEvenOrOdd
	data["hostMoneyOnEvenOrOdd"] = session.hostMoneyOnEvenOrOdd
	data["acceptedMap"] = session.acceptedMap

	data["results"] = temp1
	data["shakingResult"] = session.shakingResult
	session.mutex.RUnlock()

	session.room.Mutex.RLock()
	data["outcomeHistory"] = session.room.SharedData["outcomeHistory"]
	session.room.Mutex.RUnlock()
	return data
}

func (session *XocdiaSession) ResultSerializedData() map[string]interface{} {
	return map[string]interface{}{}
}

func (session *XocdiaSession) SerializedDataForPlayer(currentPlayer game.GamePlayer) map[string]interface{} {
	data := session.SerializedData()
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	playersData := make([]map[string]interface{}, 0)
	for player, _ := range session.players {
		playerData := make(map[string]interface{})
		playerData["id"] = player.Id()
		playerData["money"] = player.GetMoney(session.game.currencyType)
		playersData = append(playersData, playerData)
	}
	data["players_data"] = playersData
	return data
}

func (session *XocdiaSession) GetPlayer(playerId int64) (player game.GamePlayer) {
	for player, _ := range session.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (session *XocdiaSession) GetPlayerIds() []int64 {
	result := make([]int64, 0)
	for player, _ := range session.players {
		result = append(result, player.Id())
	}
	return result
}

// get resultOnePlayer obj
func (session *XocdiaSession) GetROPObj(playerId int64) *ResultOnePlayer {
	for _, ropo := range session.playerResults {
		if playerId == ropo.Id {
			return ropo
		}
	}
	return &ResultOnePlayer{} // cant happen with right logic
}

// receive gameplay funcs from outside
func InMatchLoopReceiveActions(session *XocdiaSession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()
	//
	currencyType := session.game.CurrencyType()
	session.room.Mutex.RLock()
	lastBetInfo := session.room.SharedData["lastBetInfo"].(map[int64]map[string]int64)
	session.room.Mutex.RUnlock() // init actions when change phase
	//
	for {
		action := <-session.ActionChan
		actionName := action.actionName
		actionPlayer := session.GetPlayer(action.playerId)
		if actionName == ACTION_FINISH_SESSION {
			action.responseChan <- &ActionResponse{err: nil}
			break
		} else if actionName == ACTION_BECOME_HOST {
			if session.phase != PHASE_0_BECOME_HOST {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				if session.hostPlayerId != 0 {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0032))}
				} else {
					if actionPlayer.GetMoney(currencyType) < session.room.SharedData["hostMinMoney"].(int64) {
						action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0016))}
					} else {
						session.mutex.Lock()
						session.hostPlayerId = action.playerId
						delete(session.betInfo, action.playerId)
						session.mutex.Unlock()
						session.room.Mutex.Lock()
						session.room.SharedData["hostId"] = action.playerId
						session.room.Mutex.Unlock()
						action.responseChan <- &ActionResponse{err: nil}
						session.room.DidChangeGameState(session)
					}
				}
			}
		} else if actionName == ACTION_ADD_BET {
			if session.phase != PHASE_1_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				betSelection := utils.GetStringAtPath(action.data, "betSelection")
				moneyValue := utils.GetInt64AtPath(action.data, "moneyValue")
				if action.playerId == session.hostPlayerId {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0033))}
				} else {
					if moneyValue > actionPlayer.GetMoney(currencyType) {
						action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0016))}
					} else {
						temp := moneyValue
						pId := action.playerId
						session.GetPlayer(pId).ChangeMoneyAndLog(
							-temp, currencyType, false, "",
							ACTION_ADD_BET, session.game.GameCode(), session.matchId)
						session.GetROPObj(pId).Changed += -temp
						//
						session.mutex.Lock()
						session.betInfo[action.playerId][betSelection] += moneyValue
						session.mutex.Unlock()
						action.responseChan <- &ActionResponse{err: nil}
						session.room.DidChangeGameState(session)
					}
				}
			}
		} else if actionName == ACTION_BET_EQUAL_LAST {
			if session.phase != PHASE_1_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				if _, isIn := lastBetInfo[action.playerId]; !isIn {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0034))}
				} else {
					session.mutex.Lock()
					temp := GetSumBet(session.betInfo[action.playerId])
					session.mutex.Unlock()
					if temp > 0 {
						action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0035))}
					} else {
						session.mutex.Lock()
						lastBet := CopyBet(action.playerId, lastBetInfo)
						session.mutex.Unlock()
						sumBet := GetSumBet(lastBet)
						if sumBet > actionPlayer.GetMoney(currencyType) {
							action.responseChan <- &ActionResponse{err: errors.New("not_enough_money")}
						} else {
							session.mutex.Lock()
							session.betInfo[action.playerId] = lastBet
							session.mutex.Unlock()
							//
							pId := action.playerId
							session.GetPlayer(pId).ChangeMoneyAndLog(
								-sumBet, currencyType, false, "",
								ACTION_BET_EQUAL_LAST, session.game.GameCode(), session.matchId)
							session.GetROPObj(pId).Changed += -sumBet
							//
							action.responseChan <- &ActionResponse{err: nil}
							session.room.DidChangeGameState(session)
						}
					}
				}
			}
		} else if actionName == ACTION_BET_DOUBLE_LAST {
			if session.phase != PHASE_1_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				if _, isIn := lastBetInfo[action.playerId]; !isIn {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0034))}
				} else {
					session.mutex.Lock()
					temp := GetSumBet(session.betInfo[action.playerId])
					session.mutex.Unlock()
					if temp > 0 {
						action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0035))}
					} else {
						session.mutex.Lock()
						lastBet := Copy2xBet(action.playerId, lastBetInfo)
						session.mutex.Unlock()
						sumBet := GetSumBet(lastBet)
						if sumBet > actionPlayer.GetMoney(currencyType) {
							action.responseChan <- &ActionResponse{err: errors.New("not_enough_money")}
						} else {
							session.mutex.Lock()
							session.betInfo[action.playerId] = lastBet
							session.mutex.Unlock()
							//
							pId := action.playerId
							session.GetPlayer(pId).ChangeMoneyAndLog(
								-sumBet, currencyType, false, "",
								ACTION_BET_DOUBLE_LAST, session.game.GameCode(), session.matchId)
							session.GetROPObj(pId).Changed += -sumBet
							//
							action.responseChan <- &ActionResponse{err: nil}
							session.room.DidChangeGameState(session)
						}
					}
				}
			}
		} else if actionName == ACTION_ACCEPT_BET {
			betSelection := utils.GetStringAtPath(action.data, "betSelection")
			ratio := utils.GetFloat64AtPath(action.data, "ratio")
			if (!CheckIsIn(betSelection, allSelections)) ||
				(ratio <= 0) ||
				(ratio > 1) {
				action.responseChan <- &ActionResponse{err: errors.New("wrong_selection_or_ratio")}
			} else {
				if ((action.playerId == session.hostPlayerId) &&
					(session.phase == PHASE_11_HOST_ACCEPT_BET)) ||
					((action.playerId != session.hostPlayerId) &&
						(session.phase == PHASE_12_OTHER_ACCEPT_BET)) {
					if (betSelection == SELECTION_EVEN) || (betSelection == SELECTION_ODD) {
						if action.playerId != session.hostPlayerId {
							action.responseChan <- &ActionResponse{err: errors.New("only_host_can_accept_bet_on_this_selection")}
						} else {
							// xử lí host cân cửa chẵn hoặc lẻ
							if session.hostAcceptEvenOrOdd != "" {
								action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0036))}
							} else {
								session.hostAcceptEvenOrOdd = betSelection
								// cửa mà nếu về thì host được tiền
								var hostWantS string
								if session.hostAcceptEvenOrOdd == SELECTION_EVEN {
									hostWantS = SELECTION_ODD
								} else {
									hostWantS = SELECTION_EVEN
								}
								session.mutex.Lock()
								sumE := GetAllBetOn1Selection(session.betInfo, session.hostAcceptEvenOrOdd)
								sumA := GetAllBetOn1Selection(session.betInfo, hostWantS)
								session.mutex.Unlock()
								deposit := int64(ratio * float64(sumE))
								if deposit > session.GetPlayer(session.hostPlayerId).GetMoney(currencyType) {
									deposit = session.GetPlayer(session.hostPlayerId).GetMoney(currencyType)
								}
								//
								pId := session.hostPlayerId
								session.GetPlayer(pId).ChangeMoneyAndLog(
									-deposit, currencyType, false, "",
									ACTION_ACCEPT_BET, session.game.GameCode(), session.matchId)
								session.GetROPObj(pId).Changed += -deposit
								//
								session.hostMoneyOnEvenOrOdd = deposit
								if sumA+deposit > sumE {
									refundAmount := sumA + deposit - sumE
									var refundRatio float64
									if float64(sumA) > 0.1 {
										refundRatio = float64(refundAmount) / float64(sumA)
									} else {
										refundRatio = 0
									}
									session.mutex.Lock()
									for pId, _ := range session.betInfo {
										temp := int64(refundRatio * float64(session.betInfo[pId][hostWantS]))
										//
										session.GetPlayer(pId).ChangeMoneyAndLog(
											temp, currencyType, false, "",
											XOCDIA_REFUND, session.game.GameCode(), session.matchId)
										session.GetROPObj(pId).Changed += temp
										//
										session.betInfo[pId][hostWantS] -= temp
									}
									session.mutex.Unlock()
								} else if sumA+deposit < sumE {
									refundAmount := sumE - (sumA + deposit)
									var refundRatio float64
									if float64(sumE) > 0.1 {
										refundRatio = float64(refundAmount) / float64(sumE)
									} else {
										refundRatio = 0
									}
									session.mutex.Lock()
									for pId, _ := range session.betInfo {
										temp := int64(refundRatio * float64(session.betInfo[pId][session.hostAcceptEvenOrOdd]))
										//
										session.GetPlayer(pId).ChangeMoneyAndLog(
											temp, currencyType, false, "",
											XOCDIA_REFUND, session.game.GameCode(), session.matchId)
										session.GetROPObj(pId).Changed += temp
										//
										session.betInfo[pId][session.hostAcceptEvenOrOdd] -= temp
									}
									session.mutex.Unlock()
								} else {
								}
								action.responseChan <- &ActionResponse{err: nil}
								session.room.DidChangeGameState(session)
							}
						}
					} else {
						// host hoặc người chơi cân các cửa 0 1 3 4
						if session.acceptedMap[betSelection] != 0 {
							action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0037))}
						} else {
							session.acceptedMap[betSelection] = action.playerId
							// trả tiền của người muốn cân cửa tại cửa đó về
							session.mutex.Lock()
							if session.betInfo[action.playerId][betSelection] > 0 {
								temp := session.betInfo[action.playerId][betSelection]
								pId := action.playerId
								session.GetPlayer(pId).ChangeMoneyAndLog(
									temp, currencyType, false, "",
									XOCDIA_REFUND, session.game.GameCode(), session.matchId)
								session.GetROPObj(pId).Changed += temp
								//
								session.betInfo[action.playerId][betSelection] = 0
							}
							session.mutex.Unlock()
							// đặt cọc
							session.mutex.Lock()
							sumBetOnSelection := GetAllBetOn1Selection(session.betInfo, betSelection)
							session.mutex.Unlock()
							var deposit int64
							if (betSelection == SELECTION_1_RED) || (betSelection == SELECTION_3_RED) {
								deposit = int64(3) * int64(ratio*float64(sumBetOnSelection))
							} else {
								deposit = int64(15) * int64(ratio*float64(sumBetOnSelection))
							}
							if deposit > actionPlayer.GetMoney(currencyType) {
								ratio = float64(actionPlayer.GetMoney(currencyType)) / float64(deposit)
								deposit = actionPlayer.GetMoney(currencyType)
							}
							//
							pId := action.playerId
							session.GetPlayer(pId).ChangeMoneyAndLog(
								-deposit, currencyType, false, "",
								ACTION_ACCEPT_BET, session.game.GameCode(), session.matchId)
							session.GetROPObj(pId).Changed += -deposit
							//
							refundRatio := 1 - ratio
							session.mutex.Lock()
							for _, selection := range []string{SELECTION_0_RED, SELECTION_1_RED, SELECTION_3_RED, SELECTION_4_RED} {
								for pId, _ := range session.betInfo {
									temp := int64(refundRatio * float64(session.betInfo[pId][selection]))
									//
									session.GetPlayer(pId).ChangeMoneyAndLog(
										temp, currencyType, false, "",
										XOCDIA_REFUND, session.game.GameCode(), session.matchId)
									session.GetROPObj(pId).Changed += temp
									//
									session.betInfo[pId][selection] -= temp
								}
							}
							session.mutex.Unlock()
							action.responseChan <- &ActionResponse{err: nil}
							session.room.DidStartGame(session)
						}
					}
				} else {
					action.responseChan <- &ActionResponse{err: errors.New("wrong_phase")}
				}
			}
		} else {
			fmt.Println("")
			action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0021))}
		}
	}
}
