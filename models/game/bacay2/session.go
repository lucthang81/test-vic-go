package bacay2

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/currency"
	sc "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

//
const (
	TAX_TO_JACKPOT_RATIO = float64(1 / 3.0)

	PHASE_1_START         = "PHASE_1_START"
	PHASE_2_MANDATORY_BET = "PHASE_2_MANDATORY_BET"
	// gồm cả góp gà và đánh biên
	PHASE_3_GROUP_BET    = "PHASE_3_GROUP_BET"
	PHASE_4_DEAL_CARDS   = "PHASE_4_DEAL_CARDS"
	PHASE_5_RESULT       = "PHASE_5_RESULT"
	PHASE_6_CHANGE_OWNER = "PHASE_6_CHANGE_OWNER"

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION" // not for players

	ACTION_MANDATORY_BET  = "ACTION_MANDATORY_BET"
	ACTION_JOIN_GROUP_BET = "ACTION_JOIN_GROUP_BET"
	ACTION_JOIN_PAIR_BET  = "ACTION_JOIN_PAIR_BET"
	ACTION_BECOME_OWNER   = "ACTION_BEST_HAND_BECOME_OWNER"

	OWNER_WON_ALL  = "OWNER_WON_ALL"
	OWNER_LOST_ALL = "OWNER_LOST_ALL"
)

// declare special jackpot hands
var jackPotHands map[string]bool
var mapMoneyUnitToJackpotRatio map[int64]float64

func init() {
	jackPotHands = map[string]bool{
		"d 8|d a|h a": true,
	}
	mapMoneyUnitToJackpotRatio = map[int64]float64{
		100: 0.005, 200: 0.005, 500: 0.005,
		1000: 0.01, 2000: 0.01, 5000: 0.01,
		10000: 0.025, 20000: 0.025, 50000: 0.025,
		100000: 0.05, 200000: 0.05, 500000: 0.05,
	}
	_ = zmisc.GLOBAL_TEXT_LOWER_BOUND
}

// []string
type Hand []string

func (hand Hand) ToString() string {
	sort.Strings(hand)
	return strings.Join(hand, "|")
}

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
	IsOwner  bool

	BacayScore          int
	EndGameWinningMoney int64
	Hand                string

	Changed int64
}

func (r *ResultOnePlayer) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["Id"] = r.Id
	result["Username"] = r.Username
	if r.IsOwner {
		result["IsOwner"] = r.IsOwner
	}

	result["BacayScore"] = r.BacayScore
	result["EndGameWinningMoney"] = r.EndGameWinningMoney
	result["Hand"] = r.Hand

	result["change"] = r.Changed
	return result
}

type Record struct {
	startedTime  time.Time
	players      map[int64]int
	startedCards map[int64][]string
	actions      []string
}

type PairBet struct {
	smallPlayerId int64
	bigPlayerId   int64
	moneyValue    int64
}

func (pairBet PairBet) ToString() string {
	return fmt.Sprintf("%v|%v|%v", pairBet.smallPlayerId, pairBet.bigPlayerId, pairBet.moneyValue)
}

func NewPairBet(pid1 int64, pid2 int64, moneyValue int64) PairBet {
	var smallPlayerId int64
	var bigPlayerId int64
	if pid1 < pid2 {
		smallPlayerId, bigPlayerId = pid1, pid2
	} else {
		smallPlayerId, bigPlayerId = pid2, pid1
	}
	return PairBet{
		smallPlayerId: smallPlayerId,
		bigPlayerId:   bigPlayerId,
		moneyValue:    moneyValue,
	}
}

type BaCaySession struct {
	game *BaCayGame
	room *game.Room

	startedTime time.Time
	matchId     string
	players     map[game.GamePlayer]int // include owner, map playerId to room sitting position

	phase       string
	owner       game.GamePlayer
	deck        []z.Card
	cards       map[int64][]string         // map player id to his cards
	betsVsOwner map[int64]int64            // map player id to money value
	betGroup    []int64                    // id những người chơi góp gà (nhất ăn hết, tiền cược là betEntry.min)
	betsPair    map[PairBet]map[int64]bool // biên 1v1: map[int64]bool ở đây gồm 2 người chơi, true là đồng ý

	ownerWinningType string

	canBeNewOwnerIds          []int64
	isCanBeNewOwnerIdsChanged bool
	// this var use for force become onwer
	tenPointPlayerId int64

	playerResults []*ResultOnePlayer
	record        *Record
	tax           int64

	ActionChan chan *Action // receive player action

	mutex sync.RWMutex // for newFollowers, betsVsOwner, betsGroup
}

func NewBaCaySession(gameInstance *BaCayGame, room *game.Room) *BaCaySession {
	session := &BaCaySession{
		game: gameInstance,
		room: room,

		startedTime: time.Now(),
		matchId:     fmt.Sprintf("#%v", time.Now().Unix()),
		players:     make(map[game.GamePlayer]int),

		phase:                     "",
		owner:                     room.Owner(),
		deck:                      z.NewBacayDeck(),
		cards:                     make(map[int64][]string),
		betsVsOwner:               make(map[int64]int64),
		betGroup:                  make([]int64, 0),
		betsPair:                  make(map[PairBet]map[int64]bool),
		canBeNewOwnerIds:          make([]int64, 0),
		tenPointPlayerId:          -1,
		isCanBeNewOwnerIdsChanged: false,

		playerResults: make([]*ResultOnePlayer, 0),
		record:        &Record{},

		ActionChan: make(chan *Action),
	}
	z.Shuffle(session.deck)

	for roomPosition, player := range session.room.Players().Copy() {
		if player != nil {
			session.players[player] = roomPosition
		}
	}
	for player, _ := range session.players {
		session.cards[player.Id()] = []string{"0 0", "0 0", "0 0"}
	}
	for player, _ := range session.players {
		r := &ResultOnePlayer{
			Id:       player.Id(),
			Username: player.Name(),
		}
		if session.owner != nil {
			r.IsOwner = (r.Id == session.owner.Id())
		}
		session.playerResults = append(session.playerResults, r)
	}

	go Start(session)
	go InSessionGameplayActionsReceiver(session)

	return session
}

// main match flow
func Start(session *BaCaySession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	currencyType := session.game.CurrencyType()
	reasonString := session.room.GetRoomIdentifierString()
	// _________________________________________________________________________
	durationPhase1 := session.game.duration_phase_1_start
	durationPhase2 := session.game.duration_phase_2_mandatory_bet
	durationPhase3 := session.game.duration_phase_3_group_bet
	durationPhase4 := session.game.duration_phase_4_deal_cards
	durationPhase5 := session.game.duration_phase_5_result
	durationPhase6 := session.game.duration_phase_6_change_owner
	session.mutex.RLock()
	if len(session.players) == 2 {
		durationPhase3 = 1 * time.Second
	}
	session.mutex.RUnlock()
	var alarm <-chan time.Time
	// _________________________________________________________________________
	session.room.DidStartGame(session)
	// _________________________________________________________________________
	alarm = time.After(durationPhase1)

	session.phase = PHASE_1_START
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	alarm = time.After(durationPhase2)

	session.phase = PHASE_2_MANDATORY_BET
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	alarm = time.After(durationPhase3)

	session.phase = PHASE_3_GROUP_BET
	// MANDATORY_BET
	for player, _ := range session.players {
		if player.Id() != session.owner.Id() {
			session.mutex.RLock()
			_, isIn := session.betsVsOwner[player.Id()]
			session.mutex.RUnlock()
			if isIn == false {
				InSessionMandatoryBet(session, player.Id(), betEntry.Min())
			}
		}
	}
	// đánh biên
	session.mutex.Lock()
	for _, moneyValue := range []int64{betEntry.Min(), 2 * betEntry.Min()} {
		for player1, _ := range session.players {
			if player1.Id() != session.owner.Id() {
				for player2, _ := range session.players {
					if (player2.Id() != player1.Id()) && (player2.Id() != session.owner.Id()) {
						session.betsPair[NewPairBet(player1.Id(), player2.Id(), moneyValue)] =
							map[int64]bool{player1.Id(): false, player2.Id(): false}
					}
				}
			}
		}
	}
	session.mutex.Unlock()
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	alarm = time.After(durationPhase4)

	session.phase = PHASE_4_DEAL_CARDS
	session.mutex.Lock()
	for player, _ := range session.players {
		if player.PlayerType() == "bot" {
			nTry := 0
			for {
				nTry += 1
				clone := z.Subtracted(session.deck, []z.Card{})
				z.Shuffle(clone)
				minahCards, _ := z.DealCards(&clone, 3)
				oldCards := components.ConvertMinahCardsToOldStrings(minahCards)
				s, _ := CalcScore(oldCards)
				if s <= 2 && nTry <= 10 {
					continue
				} else {
					session.deck = z.Subtracted(session.deck, minahCards)
					session.cards[player.Id()] = oldCards
					break
				}
			}
		} else {
			minahCards, _ := z.DealCards(&session.deck, 3)
			oldCards := components.ConvertMinahCardsToOldStrings(minahCards)
			session.cards[player.Id()] = oldCards
		}
		// for match_record
		session.GetPlayerResultObj(player.Id()).Hand =
			strings.Join(
				z.ToSliceString(
					components.ConvertOldStringsToMinahCards(
						session.cards[player.Id()])),
				" ",
			)
	}
	session.mutex.Unlock()
	session.room.DidChangeGameState(session)

	<-alarm
	// chia bài, nặn
	// _________________________________________________________________________
	alarm = time.After(durationPhase5)

	session.phase = PHASE_5_RESULT
	//
	// tiền đặt của người không phải chương, tiền gà, tiền biên đã trừ
	// ở đây chủ yếu cộng tiền

	// so bài với chương
	isOwnerWonAll := true
	isOwnerLostAll := true
	session.mutex.Lock()
	for playerId, moneyValue := range session.betsVsOwner {
		comparePlayerVsOwner, _ := CompareTwoBacayHand(
			session.cards[playerId], session.cards[session.owner.Id()],
		)
		bet := moneyValue
		playerScore, _ := CalcScore(session.cards[playerId])
		ownerScore, _ := CalcScore(session.cards[session.owner.Id()])
		if (playerScore == 10) || (ownerScore == 10) {
			bet = int64(2) * bet
		}
		temp := game.MoneyAfterTax(bet, betEntry)
		session.tax += bet - temp
		if comparePlayerVsOwner == COMPARE_RESULT_GREATER {
			isOwnerWonAll = false
			session.GetPlayerResultObj(playerId).EndGameWinningMoney += temp
			session.GetPlayerResultObj(session.owner.Id()).EndGameWinningMoney -= bet
		} else {
			isOwnerLostAll = false
			session.GetPlayerResultObj(playerId).EndGameWinningMoney -= bet
			session.GetPlayerResultObj(session.owner.Id()).EndGameWinningMoney += temp
		}
	}
	session.mutex.Unlock()
	if isOwnerWonAll && len(session.players) > 2 {
		session.ownerWinningType = OWNER_WON_ALL
	}
	if isOwnerLostAll && len(session.players) > 2 {
		session.ownerWinningType = OWNER_LOST_ALL
	}
	// kiểm tra có thể đổi chương
	bestHandPlayerId := session.owner.Id()
	bestHand := session.cards[bestHandPlayerId]
	bestHandScore, _ := CalcScore(bestHand)
	for player, _ := range session.players {
		compareVsBestHand, _ := CompareTwoBacayHand(
			session.cards[player.Id()],
			bestHand,
		)
		if compareVsBestHand == COMPARE_RESULT_GREATER {
			bestHandPlayerId = player.Id()
			bestHand = session.cards[bestHandPlayerId]
			bestHandScore, _ = CalcScore(bestHand)
		}
	}
	if (bestHandPlayerId != session.owner.Id()) &&
		(bestHandScore == 10) &&
		(session.GetPlayer(bestHandPlayerId).GetMoney(currencyType) >= int64(session.game.ownerRequirementMultiplier)*betEntry.Min()) {
		session.canBeNewOwnerIds = append(session.canBeNewOwnerIds, bestHandPlayerId)
		session.tenPointPlayerId = bestHandPlayerId
	}
	// so gà
	if len(session.betGroup) == 1 {
		player := session.GetPlayer(session.betGroup[0])
		session.GetPlayerResultObj(player.Id()).EndGameWinningMoney += betEntry.Min()
	} else if len(session.betGroup) >= 2 {
		bestHandGroupPlayerId := session.betGroup[0]
		bestHandGroup := session.cards[bestHandGroupPlayerId]
		for _, pid := range session.betGroup {
			compareVsBestHandGroup, _ := CompareTwoBacayHand(
				session.cards[pid],
				bestHandGroup,
			)
			if compareVsBestHandGroup == COMPARE_RESULT_GREATER {
				bestHandGroupPlayerId = pid
				bestHandGroup = session.cards[bestHandGroupPlayerId]
			}
		}
		winnerGroup := session.GetPlayer(bestHandGroupPlayerId)
		temp := game.MoneyAfterTax(int64(len(session.betGroup))*betEntry.Min(), betEntry)
		session.tax += int64(len(session.betGroup))*betEntry.Min() - temp
		session.GetPlayerResultObj(winnerGroup.Id()).EndGameWinningMoney += temp
	}
	//so từng cặp biên
	for pairBet, isAccepted := range session.betsPair {
		if (isAccepted[pairBet.bigPlayerId] == true) && (isAccepted[pairBet.smallPlayerId] == true) {
			compare, _ := CompareTwoBacayHand(
				session.cards[pairBet.smallPlayerId],
				session.cards[pairBet.bigPlayerId],
			)
			var winnerId int64
			if compare == COMPARE_RESULT_GREATER {
				winnerId = pairBet.smallPlayerId
			} else {
				winnerId = pairBet.bigPlayerId
			}
			temp := game.MoneyAfterTax(2*pairBet.moneyValue, betEntry)
			session.tax += 2*pairBet.moneyValue - temp
			session.GetPlayerResultObj(winnerId).EndGameWinningMoney += temp
		}
	}
	// add tax to jackpot
	jackpotCode := "all"
	jackpotInstance := jackpot.GetJackpot(jackpotCode, currencyType)
	if jackpotInstance != nil {
		if session.room.GetNumberOfHumans() != 0 {
			moneyToJackpot := int64(float64(session.tax) * TAX_TO_JACKPOT_RATIO)
			jackpotInstance.AddMoney(moneyToJackpot)
		}
		// check win jackpot
		var bh Hand
		bh = bestHand
		if _, isIn := jackPotHands[bh.ToString()]; isIn {
			if ratio, isIn := mapMoneyUnitToJackpotRatio[session.room.Requirement()]; isIn {
				temp := int64(float64(jackpotInstance.Value()) * ratio)
				session.GetPlayerResultObj(bestHandPlayerId).EndGameWinningMoney += temp
				jackpotInstance.AddMoney(-temp)
				jackpotInstance.NotifySomeoneHitJackpot(
					session.game.GameCode(),
					temp,
					session.GetPlayer(bestHandPlayerId).Id(),
					session.GetPlayer(bestHandPlayerId).Name(),
				)
			}
		}
	}
	//
	for _, result1p := range session.playerResults {
		s, _ := CalcScore(session.cards[result1p.Id])
		result1p.BacayScore = s
	}
	// cộng tiền
	for _, result1p := range session.playerResults {
		pObj := session.GetPlayer(result1p.Id)
		result1p.Changed += result1p.EndGameWinningMoney
		if result1p.EndGameWinningMoney < 0 {
			pObj.ChangeMoneyAndLog(
				result1p.EndGameWinningMoney, currencyType, true, reasonString,
				ACTION_FINISH_SESSION, session.game.GameCode(), session.matchId)
		} else {
			pObj.ChangeMoneyAndLog(
				result1p.EndGameWinningMoney, currencyType, false, "",
				ACTION_FINISH_SESSION, session.game.GameCode(), session.matchId)
		}

		if result1p.EndGameWinningMoney >= 3*zmisc.GLOBAL_TEXT_LOWER_BOUND &&
			session.game.currencyType == "money" &&
			pObj.PlayerType() == "normal" {
			zmisc.InsertNewGlobalText(map[string]interface{}{
				"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
				"username": result1p.Username,
				"wonMoney": result1p.EndGameWinningMoney,
				"gamecode": session.game.GameCode(),
			})
		}
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
	playerIpAdds := map[int64]string{}
	for playerObj, _ := range session.players {
		playerIpAdds[playerObj.Id()] = playerObj.IpAddress()
	}
	playerResults := make([]map[string]interface{}, 0)
	for _, r1p := range session.playerResults {
		playerResults = append(playerResults, r1p.ToMap())
	}
	// for debug betsPair ->
	temp := make([]string, 0)
	for pairKey, acceptedMap := range session.betsPair {
		if acceptedMap[pairKey.smallPlayerId] && acceptedMap[pairKey.bigPlayerId] {
			temp = append(temp, fmt.Sprintf("%v| %7d%7d",
				pairKey.moneyValue, pairKey.smallPlayerId, pairKey.bigPlayerId))
		}
	}
	// for bebug betsPair <-
	// event sc:
	sc.GlobalMutex.Lock()
	event := sc.MapEventSCs[sc.EVENTSC_TIENLEN_BAIDEP]
	sc.GlobalMutex.Unlock()
	winner := session.GetPlayer(bestHandPlayerId)
	if event != nil {
		event.Mutex.Lock()
		isLimited := false
		if event.MapPlayerIdToValue[winner.Id()] >= event.LimitNOBonus {
			isLimited = true
		}
		event.Mutex.Unlock()
		if bestHandScore == 10 && !isLimited &&
			session.game.currencyType != currency.CustomMoney {
			event.ChangeValue(winner.Id(), 1)
			winner.ChangeMoneyAndLog(
				2*betEntry.Min(), session.game.CurrencyType(), false, "",
				sc.ACTION_BONUS_EVENT, session.game.GameCode(), "")
			if winner.PlayerType() == "normal" {
				humanWon += 2 * betEntry.Min()
			}
		}
	}
	//
	if session.room.GetNumberOfHumans() != 0 {
		record.LogMatchRecord3(
			session.game.GameCode(), session.game.CurrencyType(), session.room.Requirement(), session.tax,
			humanWon, humanLost, botWon, botLost,
			session.matchId, playerIpAdds,
			playerResults, map[string]interface{}{
				"betsPair":    temp,
				"betGroup":    session.betGroup,
				"betsVsOwner": session.betsVsOwner,
			})
	}
	//
	event_player.GlobalMutex.Lock()
	e := event_player.MapEvents[event_player.EVENT_COLLECTING_PIECES]
	event_player.GlobalMutex.Unlock()
	if e != nil {
		for _, r1p := range session.playerResults {
			e.GiveAPiece(r1p.Id, false,
				currencyType == currency.TestMoney, r1p.Changed)
		}
	}
	//
	session.room.DidChangeGameState(session)

	<-alarm
	// _________________________________________________________________________
	if len(session.canBeNewOwnerIds) > 0 {
		alarm = time.After(durationPhase6)

		session.phase = PHASE_6_CHANGE_OWNER
		session.room.DidChangeGameState(session)

		<-alarm
		session.mutex.Lock()
		if (session.tenPointPlayerId != -1) &&
			(session.room.Owner().Id() == session.owner.Id()) {
			// co nguoi choi duoc 10
			// nguoi duoc 10 tu choi nhan chuong, nguoi khac cung tu choi
			// ep nguoi duoc 10 nhan chuong
			p := session.GetPlayer(session.tenPointPlayerId)
			if p.GetAvailableMoney(session.game.currencyType) >= int64(session.game.ownerRequirementMultiplier)*session.room.Requirement() {
				session.room.AssignOwner(p)
				time.Sleep(1 * time.Second)
			}
		}
		session.canBeNewOwnerIds = make([]int64, 0)
		session.mutex.Unlock()
	}
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
func (session *BaCaySession) CleanUp() {
}

func (session *BaCaySession) HandlePlayerRemovedFromGame(player game.GamePlayer) {

}
func (session *BaCaySession) HandlePlayerAddedToGame(player game.GamePlayer) {

}

func (session *BaCaySession) HandlePlayerOffline(player game.GamePlayer) {

}
func (session *BaCaySession) HandlePlayerOnline(player game.GamePlayer) {

}

func (session *BaCaySession) IsPlaying() bool {
	return true
}

func (session *BaCaySession) IsDelayingForNewGame() bool {
	return false
}

func (session *BaCaySession) SerializedData() map[string]interface{} {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	data := make(map[string]interface{})
	if session.owner != nil {
		data["owner_id"] = session.owner.Id()
	} else {
		data["owner_id"] = 0
	}
	data["game_code"] = session.game.GameCode()
	data["requirement"] = session.room.Requirement()
	data["matchId"] = session.matchId
	data["player_ids"] = session.GetPlayerIds()
	data["phase"] = session.phase
	temp1 := make([]map[string]interface{}, 0)
	if session.phase == PHASE_5_RESULT {
		data["ownerWinningType"] = session.ownerWinningType
		for _, result1p := range session.playerResults {
			temp1 = append(temp1, result1p.ToMap())
		}
	}
	data["results"] = temp1
	if session.phase == PHASE_6_CHANGE_OWNER {
		data["canBeNewOwnerIds"] = session.canBeNewOwnerIds
	}
	data["betGroup"] = session.betGroup
	temp := make([]map[string]interface{}, 0)
	for pairKey, acceptedMap := range session.betsPair {
		if acceptedMap[pairKey.smallPlayerId] && acceptedMap[pairKey.bigPlayerId] {
			temp = append(temp, map[string]interface{}{
				"NguoiMoi":  pairKey.smallPlayerId,
				"NguoiNhan": pairKey.bigPlayerId,
				"SoTien":    pairKey.moneyValue,
				"DaDongY":   true,
			})
		} else if acceptedMap[pairKey.smallPlayerId] && !acceptedMap[pairKey.bigPlayerId] {
			temp = append(temp, map[string]interface{}{
				"NguoiMoi":  pairKey.smallPlayerId,
				"NguoiNhan": pairKey.bigPlayerId,
				"SoTien":    pairKey.moneyValue,
				"DaDongY":   false,
			})
		} else if !acceptedMap[pairKey.smallPlayerId] && acceptedMap[pairKey.bigPlayerId] {
			temp = append(temp, map[string]interface{}{
				"NguoiMoi":  pairKey.bigPlayerId,
				"NguoiNhan": pairKey.smallPlayerId,
				"SoTien":    pairKey.moneyValue,
				"DaDongY":   false,
			})
		}
	}
	data["betsPair"] = temp
	temp2 := map[int64]int64{}
	for k, v := range session.betsVsOwner {
		temp2[k] = v
	}
	data["betsVsOwner"] = temp2
	return data
}

func (session *BaCaySession) ResultSerializedData() map[string]interface{} {
	return map[string]interface{}{}
}
func (session *BaCaySession) SerializedDataForPlayer(currentPlayer game.GamePlayer) map[string]interface{} {
	data := session.SerializedData()
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	playersData := make([]map[string]interface{}, 0)
	for player, _ := range session.players {
		playerData := make(map[string]interface{})
		playerData["id"] = player.Id()
		if (currentPlayer.Id() == player.Id()) ||
			(session.phase == PHASE_5_RESULT) ||
			(session.phase == PHASE_6_CHANGE_OWNER) {
			playerData["cards"] = session.cards[player.Id()]
		} else {
			playerData["cards"] = []string{"0 0", "0 0", "0 0"}
		}
		playerData["money"] = player.GetMoney(session.game.currencyType)
		playersData = append(playersData, playerData)
	}
	data["players_data"] = playersData
	return data
}

func (session *BaCaySession) GetPlayer(playerId int64) (player game.GamePlayer) {
	for player, _ := range session.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (session *BaCaySession) GetPlayerIds() []int64 {
	result := make([]int64, 0)
	for player, _ := range session.players {
		result = append(result, player.Id())
	}
	return result
}

func (session *BaCaySession) GetPlayerResultObj(pid int64) *ResultOnePlayer {
	for _, result1p := range session.playerResults {
		if result1p.Id == pid {
			return result1p
		}
	}
	return nil
}

// receive gameplay funcs from outside
func InSessionGameplayActionsReceiver(session *BaCaySession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	currencyType := session.game.CurrencyType()
	for {
		action := <-session.ActionChan
		actionName := action.actionName
		if actionName == ACTION_FINISH_SESSION {
			action.responseChan <- &ActionResponse{err: nil}
			break
		} else if actionName == ACTION_MANDATORY_BET {
			if session.phase != PHASE_2_MANDATORY_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else if action.playerId == session.owner.Id() {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0018))}
			} else {
				session.mutex.RLock()
				_, isIn := session.betsVsOwner[action.playerId]
				session.mutex.RUnlock()
				if isIn == true {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0019))}
				} else {
					moneyValue := utils.GetInt64AtPath(action.data, "moneyValue")
					if !((betEntry.Min() <= moneyValue) && (moneyValue <= betEntry.Max())) {
						action.responseChan <- &ActionResponse{err: errors.New("err:wrong_bet_amount")}
					} else {
						InSessionMandatoryBet(session, action.playerId, moneyValue)
						action.responseChan <- &ActionResponse{err: nil}
						session.room.DidChangeGameState(session)
					}
				}
			}
		} else if actionName == ACTION_JOIN_GROUP_BET {
			if session.phase != PHASE_3_GROUP_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				if session.GetPlayer(action.playerId).GetAvailableMoney(currencyType) < betEntry.Min() {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0016))}
				} else if Find3(session.betGroup, action.playerId) != -1 {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0019))}
				} else {
					InSessionJoinGroupBet(session, action.playerId)
					action.responseChan <- &ActionResponse{err: nil}
					session.room.DidChangeGameState(session)
				}
			}
		} else if actionName == ACTION_JOIN_PAIR_BET {
			if session.phase != PHASE_3_GROUP_BET {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				enemyId := utils.GetInt64AtPath(action.data, "enemyId")
				moneyValue := utils.GetInt64AtPath(action.data, "moneyValue")
				pairKey := NewPairBet(action.playerId, enemyId, moneyValue)
				session.mutex.RLock()
				pairBet, isIn := session.betsPair[pairKey]
				session.mutex.RUnlock()
				if isIn {
					if pairBet[action.playerId] != true {
						session.mutex.Lock()
						pairBet[action.playerId] = true
						session.mutex.Unlock()
						if (pairBet[pairKey.bigPlayerId] == true) && (pairBet[pairKey.smallPlayerId] == true) {
							if (session.GetPlayer(pairKey.bigPlayerId).GetAvailableMoney(currencyType) < betEntry.Min()) ||
								(session.GetPlayer(pairKey.smallPlayerId).GetAvailableMoney(currencyType) < betEntry.Min()) {
								session.mutex.Lock()
								pairBet[pairKey.bigPlayerId] = false
								pairBet[pairKey.smallPlayerId] = false
								session.mutex.Unlock()
								action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0016))}
								session.room.DidChangeGameState(session)
							} else {
								session.GetPlayer(pairKey.bigPlayerId).ChangeMoneyAndLog(
									-pairKey.moneyValue, currencyType, false, "",
									ACTION_JOIN_PAIR_BET, session.game.GameCode(), session.matchId)
								session.GetPlayerResultObj(pairKey.bigPlayerId).Changed += -pairKey.moneyValue

								session.GetPlayer(pairKey.smallPlayerId).ChangeMoneyAndLog(
									-pairKey.moneyValue, currencyType, false, "",
									ACTION_JOIN_PAIR_BET, session.game.GameCode(), session.matchId)
								session.GetPlayerResultObj(pairKey.smallPlayerId).Changed += -pairKey.moneyValue

								//
								action.responseChan <- &ActionResponse{err: nil}
								session.room.DidChangeGameState(session)
							}
						} else {
							action.responseChan <- &ActionResponse{err: nil}
							session.room.DidChangeGameState(session)
						}
					} else {
						//if pairBet[action.playerId] == true
						//lệnh vô dụng vì đã gạ kèo rồi
						// action.responseChan <- &ActionResponse{err: errors.New("err:already_accepted_this_pair_bet")}
						action.responseChan <- &ActionResponse{err: nil}
					}
				} else {
					action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0020))}
				}
			}
		} else if actionName == ACTION_BECOME_OWNER {
			if session.phase != PHASE_6_CHANGE_OWNER {
				action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0017))}
			} else {
				// l1
				session.mutex.Lock()
				if Find3(session.canBeNewOwnerIds, action.playerId) == -1 {
					// l1 ul
					session.mutex.Unlock()
					action.responseChan <- &ActionResponse{err: errors.New("err:u_cant_be_owner")}
				} else {
					choice := utils.GetBoolAtPath(action.data, "choice")
					if choice == true {
						player := session.GetPlayer(action.playerId)
						if player.GetMoney(currencyType) < int64(session.game.ownerRequirementMultiplier)*session.room.Requirement() {
							// l1 ul
							session.mutex.Unlock()
							action.responseChan <- &ActionResponse{err: errors.New("err:Không đủ tiền")}
						} else {
							err := session.room.AssignOwner(player)
							time.Sleep(1 * time.Second)
							// l1 ul
							session.mutex.Unlock()
							if err == nil {
								session.canBeNewOwnerIds = make([]int64, 0)
								action.responseChan <- &ActionResponse{err: nil}
							} else {
								action.responseChan <- &ActionResponse{err: err}
							}
						}
						session.room.DidChangeGameState(session)
					} else {
						if session.isCanBeNewOwnerIdsChanged == false { // thằng bestHand từ chối
							session.canBeNewOwnerIds = make([]int64, 0)
							for player, _ := range session.players {
								if (player.Id() != session.owner.Id()) && (player.Id() != action.playerId) {
									if player.GetMoney(currencyType) >= int64(session.game.ownerRequirementMultiplier)*session.room.Requirement() {
										session.canBeNewOwnerIds = append(session.canBeNewOwnerIds, player.Id())
									}
								}
							}
							// l1 ul
							session.mutex.Unlock()
							action.responseChan <- &ActionResponse{err: nil}
							session.room.DidChangeGameState(session)
							session.isCanBeNewOwnerIdsChanged = true
						} else { // thằng khác từ chối
							// l1 ul
							session.mutex.Unlock()
							action.responseChan <- &ActionResponse{err: nil}
						}
					}
				}
			}
		} else {
			fmt.Println("")
			action.responseChan <- &ActionResponse{err: errors.New(l.Get(l.M0021))}
		}
	}

}

// included session.mutex.Lock()
func InSessionMandatoryBet(session *BaCaySession, playerId int64, moneyValue int64) {
	session.mutex.Lock()
	session.betsVsOwner[playerId] = moneyValue
	session.mutex.Unlock()
}

// this func need check conditions before call
func InSessionJoinGroupBet(session *BaCaySession, playerId int64) {
	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	currencyType := session.game.CurrencyType()
	actionPlayer := session.GetPlayer(playerId)
	actionPlayer.ChangeMoneyAndLog(
		-betEntry.Min(), currencyType, false, "",
		ACTION_JOIN_GROUP_BET, session.game.GameCode(), session.matchId)
	session.GetPlayerResultObj(playerId).Changed += -betEntry.Min()
	session.mutex.Lock()
	session.betGroup = append(session.betGroup, playerId)
	session.mutex.Unlock()
}
