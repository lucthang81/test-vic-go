package phom

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/language"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/rank"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

const (
	TAX_TO_JACKPOT_RATIO = float64(1 / 3.0)

	ACTION_FINISH_SESSION = "ACTION_FINISH_SESSION" // not for players

	ACTION_POP_CARD           = "ACTION_POP_CARD"           // đánh
	ACTION_EAT_CARD           = "ACTION_EAT_CARD"           // ăn
	ACTION_DRAW_CARD          = "ACTION_DRAW_CARD"          // rút
	ACTION_SHOW_COMBO_BY_USER = "ACTION_SHOW_COMBO_BY_USER" // hạ phỏm user gửi lên [][]Card
	ACTION_AUTO_SHOW_COMBOS   = "ACTION_AUTO_SHOW_COMBOS"   // tự động hạ ít điểm nhất, nếu đã có thao tác từ client thì hạ nốt
	ACTION_HANG_CARD          = "ACTION_HANG_CARD"          // gửi 1 lá bài
	ACTION_AUTO_HANG_CARDS    = "ACTION_AUTO_HANG_CARDS"    // tự động gửi bài
)

var mapMoneyUnitToJackpotRatio map[int64]float64

func init() {
	mapMoneyUnitToJackpotRatio = map[int64]float64{
		100: 0.05, 200: 0.05, 500: 0.05,
		1000: 0.1, 2000: 0.1, 5000: 0.1,
		10000: 0.25, 20000: 0.25, 50000: 0.25,
		100000: 0.5, 200000: 0.5, 500000: 0.5,
	}
	_, _ = json.Marshal([]int{})
}

type Action struct {
	actionName   string
	playerId     int64
	data         map[string]interface{}
	chanResponse chan *ActionResponse
}

func (action *Action) ToMap() map[string]interface{} {
	dataJson, _ := json.Marshal(action.data)
	result := map[string]interface{}{
		"actionTime": time.Now(),
		"actionName": action.actionName,
		"playerId":   action.playerId,
		"data":       string(dataJson),
	}
	return result
}

type ActionResponse struct {
	err  error
	data map[string]interface{}
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

// Custom timer, have Add func to modify timer
type MinahTimer struct {
	Timer    *time.Timer
	EndPoint time.Time
}

func NewMinahTimer(duration time.Duration) *MinahTimer {
	result := &MinahTimer{
		Timer:    time.NewTimer(duration),
		EndPoint: time.Now().Add(duration),
	}
	return result
}

// add more time to the timer
func (minahTimer *MinahTimer) Add(duration time.Duration) {
	temp := minahTimer.Timer.Stop()
	if temp == false {
		// timer was fired before this func call
		// dont do anything
	} else {
		minahTimer.EndPoint = minahTimer.EndPoint.Add(duration)
		remainingDur := minahTimer.EndPoint.Sub(time.Now())
		minahTimer.Timer.Reset(remainingDur)
	}
}

type PhomSession struct {
	game *PhomGame
	room *game.Room

	startedTime time.Time
	matchId     string
	players     map[game.GamePlayer]int
	// bot or normal
	mapPlayerIdToPtype map[int64]string

	playerResults []*ResultOnePlayer
	tax           int64

	// bộ bài giữa sân
	deck []z.Card

	// các biến reset khi có turn mới
	// người chơi đang có lượt
	turnPlayerId int64
	// thời điểm bắt đầu lượt đánh
	turnStartedTime time.Time
	isFirstTurn     bool
	// turnPlayer đã ăn hoặc bốc bài chưa
	isDrawnOrEaten bool
	// turnPlayer đã hạ phỏm chưa
	isShowedCombos bool
	// phục vụ khóa lệnh đánh bài khi đang chờ xử lí tự đánh
	isWaitingForAutoPopCard bool
	// lá bài bốc được
	drawnCard z.Card
	// các cách hạ phỏm
	waysToShowCombos [][][]z.Card
	// lưu phỏm client thao tác hạ
	clientShowedCombos [][]z.Card

	// biến thông báo có ăn bài, chỉ true trong 1 gói tin,
	// thay đổi trước sau gửi tin chứ không phải đổi trong các hàm xử lí lệnh
	notifyEaten bool

	// lá bài người turn trước hoặc turn này đánh
	poppedCard z.Card

	// mapPlayerIdToNextPlayerId
	mapPlayerIdToNextPlayerId map[int64]int64
	// bài trên tay
	mapPlayerIdToHand map[int64][]z.Card
	// những lá bài bị khoá không được đánh và tạo phỏm (ăn của người khác)
	// Locks = map[int64][]Card, map 1, 2, 3, 4 to []Card, tiền được ứng với lá
	// sau khi hạ bài sẽ giải phóng
	mapPlayerIdToLocks map[int64]map[int64][]z.Card
	// gần giống map lock, nhưng không thay đổi khi hạ bài, dùng để tính tiền
	mapPlayerIdToEatens map[int64]map[int64][]z.Card
	// những phỏm đã hạ, có phần tử cuối là các quân lẻ
	mapPlayerIdToShowedCombos map[int64][][]z.Card
	// những quân đã đánh
	mapPlayerIdToPoppedCards map[int64][]z.Card
	// id người hạ đầu
	firstShowPlayerId int64
	// lưu các lá đã gửi
	mapPlayerIdToHungCards map[int64][]z.Card
	// kq chỉ tiền ù nhất nhì ba bét, quân ăn đã tính trước
	mapPlayerIdToWinMoney map[int64]int64
	//
	mapPlayerIdToIsAfkLastTurn map[int64]bool

	// nhận lệnh từ người chơi
	ChanAction chan *Action

	// send from InSessionGameplayActionsReceiver when poped or had full combos
	// to end turn
	ChanIsPoppedCard chan bool
	TurnTimer        *MinahTimer
	// đã có người ù
	isSomeoneHas9FullCombos  bool
	isSomeoneHas10FullCombos bool
	isNoDrawsWin             bool
	instantWinHand           [][]z.Card
	isEndedTurnLoop          bool

	// data for func record.LogMatchRecord3 (log to database)
	moreMatchData map[string]interface{}
	playerActions []map[string]interface{}

	// debug vars
	checkpoint1 bool
	checkpoint2 bool
	checkpoint3 bool
	checkpoint4 bool

	mutex sync.RWMutex
}

func NewPhomSession(gameInstance *PhomGame, room *game.Room) *PhomSession {
	//
	isTesting := false
	//
	session := &PhomSession{
		game: gameInstance,
		room: room,

		startedTime: time.Now(),
		matchId:     fmt.Sprintf("#%v", time.Now().Unix()),

		tax: 0,

		ChanIsPoppedCard: make(chan bool),
		ChanAction:       make(chan *Action),

		moreMatchData:              map[string]interface{}{},
		playerActions:              []map[string]interface{}{},
		mapPlayerIdToIsAfkLastTurn: make(map[int64]bool),
	}
	//
	if isTesting {
		// session.game.turnTimeInSeconds = 3 * time.Second
	}
	//
	session.players = make(map[game.GamePlayer]int)
	session.mapPlayerIdToPtype = make(map[int64]string)
	for roomPosition, player := range session.room.Players().Copy() {
		if player != nil {
			session.players[player] = roomPosition
			session.mapPlayerIdToPtype[player.Id()] = player.PlayerType()
		}
	}
	//
	session.isFirstTurn = true
	//
	session.mapPlayerIdToNextPlayerId = make(map[int64]int64)
	// playerIds sort by position
	temp := make([]int64, 0)
	for _, player := range session.room.Players().Copy() {
		if player != nil {
			temp = append(temp, player.Id())
		}
	}
	for i, pId := range temp {
		if i < len(temp)-1 {
			session.mapPlayerIdToNextPlayerId[pId] = temp[i+1]
		} else {
			session.mapPlayerIdToNextPlayerId[pId] = temp[0]
		}
	}
	session.room.Mutex.RLock()
	lastWinnerId := session.room.SharedData["lastWinnerId"].(int64)
	session.room.Mutex.RUnlock()
	if session.GetPlayer(lastWinnerId) != nil {
		session.turnPlayerId = lastWinnerId
	} else {
		session.turnPlayerId = temp[rand.Intn(len(temp))]
	}
	if isTesting {
		minId := temp[0]
		for _, id := range temp {
			if id < minId {
				minId = id
			}
		}
		session.turnPlayerId = minId
	}
	//
	isFixCard := false
	session.mapPlayerIdToHand = make(map[int64][]z.Card)
	//	for player, _ := range session.players {
	//		if player.Id() == 26 {
	//			redis_fixbai := fmt.Sprintf("%v_fix", session.game.GameCode())
	//			data, err := dataCenter.GetCardsFix(redis_fixbai)
	//			if err != nil {
	//				break
	//			}
	//			var v interface{}
	//			json.Unmarshal(data, &v)
	//			//v := {"bai_loc":["S 2"],"fix":["s 3"],"other":[["s 4"],["s 4"]]}
	//			new_v := v.(map[string]interface{})
	//
	//			bai_loc := new_v["bai_loc"].([]interface{})
	//			session.deck = []z.Card{}
	//			// gan bai loc
	//			for _, vl := range bai_loc {
	//				session.deck = append(session.deck, z.ToCard(vl.(string)))
	//			}
	//
	//			get_fix := new_v["fix"].([]interface{})
	//			get_other_fix := new_v["other"].([]interface{})
	//
	//			if len(get_other_fix)+1 < len(temp) || len(get_fix) != 10 {
	//				break
	//			}
	//			isFixCard = true
	//			//iCount bien dem cua other car
	//			var iCount int
	//			for _, pId := range temp {
	//				session.mapPlayerIdToHand[pId] = []z.Card{}
	//				if pId == session.turnPlayerId {
	//					for i := 0; i < len(get_fix); i++ {
	//						session.mapPlayerIdToHand[pId] = append(session.mapPlayerIdToHand[pId], z.ToCard(get_fix[i].(string)))
	//					}
	//
	//				} else {
	//					get_fix = get_other_fix[iCount].([]interface{})
	//					if len(get_fix) != 9 {
	//						isFixCard = false
	//						break
	//					}
	//					iCount++
	//					for i := 0; i < len(get_fix); i++ {
	//						session.mapPlayerIdToHand[pId] = append(session.mapPlayerIdToHand[pId], z.ToCard(get_fix[i].(string)))
	//					}
	//				}
	//			}
	//
	//			break
	//		}
	//	}
	if isFixCard == false {
		session.deck = z.NewDeck()
		z.Shuffle(session.deck)
		_, _ = z.DealCards(&session.deck, 13*(4-len(session.players)))
		for _, pId := range temp {
			var dealtCards []z.Card
			if pId == session.turnPlayerId {
				dealtCards, _ = z.DealCards(&session.deck, 10)
			} else {
				dealtCards, _ = z.DealCards(&session.deck, 9)
			}
			session.mapPlayerIdToHand[pId] = dealtCards
		}
	}
	//
	session.mapPlayerIdToLocks = make(map[int64]map[int64][]z.Card)
	for player, _ := range session.players {
		session.mapPlayerIdToLocks[player.Id()] = map[int64][]z.Card{
			1: []z.Card{},
			2: []z.Card{},
			3: []z.Card{},
			4: []z.Card{}, // quân chốt
		}
	}
	session.mapPlayerIdToEatens = make(map[int64]map[int64][]z.Card)
	for player, _ := range session.players {
		session.mapPlayerIdToEatens[player.Id()] = map[int64][]z.Card{
			1: []z.Card{},
			2: []z.Card{},
			3: []z.Card{},
			4: []z.Card{}, // quân chốt
		}
	}
	//
	session.mapPlayerIdToShowedCombos = make(map[int64][][]z.Card)
	for player, _ := range session.players {
		session.mapPlayerIdToShowedCombos[player.Id()] = make([][]z.Card, 0)
	}
	//
	session.mapPlayerIdToPoppedCards = make(map[int64][]z.Card)
	for player, _ := range session.players {
		session.mapPlayerIdToPoppedCards[player.Id()] = make([]z.Card, 0)
	}
	//
	session.mapPlayerIdToHungCards = make(map[int64][]z.Card)
	for player, _ := range session.players {
		session.mapPlayerIdToHungCards[player.Id()] = make([]z.Card, 0)
	}
	//
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
	session.moreMatchData["startingMapPlayerIdToHand"] = session.mapPlayerIdToHand
	session.moreMatchData["startingDeck"] = session.deck
	session.moreMatchData["mapPlayerIdToNextPlayerId"] = session.mapPlayerIdToNextPlayerId
	//
	go Start(session)
	go InSessionGameplayActionsReceiver(session)
	//
	return session
}

// match main flow
func Start(session *PhomSession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	// _________________________________________________________________________
	session.room.DidStartGame(session)
	// _________________________________________________________________________

	// check ù khan khi bắt đầu ván
	session.mutex.Lock()
	checkNoDrawsPlayerId := session.turnPlayerId
	for _ = range z.Range(len(session.players)) {
		draws := GetAllDraws(session.mapPlayerIdToHand[checkNoDrawsPlayerId])
		if len(draws) == 0 {
			session.isNoDrawsWin = true
			session.turnPlayerId = checkNoDrawsPlayerId
			session.instantWinHand = [][]z.Card{session.mapPlayerIdToHand[checkNoDrawsPlayerId]}
			break
		} else {
			checkNoDrawsPlayerId = session.mapPlayerIdToNextPlayerId[checkNoDrawsPlayerId]
		}
	}
	session.mutex.Unlock()
	//
	if !session.isNoDrawsWin {
		// mỗi vòng lặp là 1 turn của 1 thằng
		// turn gồm có: rút, hạ phỏm (chỉ ở vòng hạ), gửi (chỉ ở vòng hạ), vứt
		// vòng lặp kết thúc:
		//    - sau turn mà deck hết bài
		//    - hoặc ù10 sau ăn / bốc / gửi
		//    - hoặc ù9 sau vứt bài
		for (len(session.deck) > 0) &&
			(session.isSomeoneHas10FullCombos == false) &&
			(session.isSomeoneHas9FullCombos == false) {
			session.mutex.Lock()
			if session.isFirstTurn {
				session.isFirstTurn = false
				session.turnPlayerId = session.turnPlayerId
				session.isDrawnOrEaten = true
			} else {
				session.turnPlayerId = session.mapPlayerIdToNextPlayerId[session.turnPlayerId]
				session.isDrawnOrEaten = false
			}
			session.turnStartedTime = time.Now()
			session.isShowedCombos = false
			session.drawnCard = z.Card{}
			session.waysToShowCombos = [][][]z.Card{}
			session.clientShowedCombos = [][]z.Card{}
			if session.mapPlayerIdToIsAfkLastTurn[session.turnPlayerId] == true {
				session.TurnTimer = NewMinahTimer(10 * time.Second)
			} else {
				session.TurnTimer = NewMinahTimer(session.game.turnTimeInSeconds)
			}
			session.isWaitingForAutoPopCard = false
			session.mutex.Unlock()
			session.room.DidChangeGameState(session)
			//
			select {
			case <-session.ChanIsPoppedCard:
				session.mutex.Lock()
				session.mapPlayerIdToIsAfkLastTurn[session.turnPlayerId] = false
				session.mutex.Unlock()
				// người chơi đã gửi lệnh đánh bài và được xử lí xong ở action loop
			case <-session.TurnTimer.Timer.C:
				session.mutex.Lock()
				session.mapPlayerIdToIsAfkLastTurn[session.turnPlayerId] = true
				session.mutex.Unlock()
				session.isWaitingForAutoPopCard = true
				// tự động rút, hạ, gửi, đánh
				// rút bài
				if session.isDrawnOrEaten == false {
					InSessionDrawCard(session)
					session.room.DidChangeGameState(session)
					time.Sleep(500 * time.Millisecond)
				}
				// nếu là vòng cuối
				if len(session.deck) <= len(session.players)-1 {
					// hạ bài
					if !session.isShowedCombos {
						if len(session.waysToShowCombos) >= 1 {
							InSessionAutoShowCombos(session) // tự động chọn cách hạ đầu tiên
							session.room.DidChangeGameState(session)
							time.Sleep(500 * time.Millisecond)
						} else {
							// cant happen
						}
					}
				}
				// tự đánh bài
				if !session.isSomeoneHas10FullCombos {
					session.mutex.RLock()
					ways := ShowCombosMinPoint(
						session.mapPlayerIdToHand[session.turnPlayerId],
						GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
					)
					session.mutex.RUnlock()
					if len(ways) > 0 {
						bestWay := ways[0]
						// bài lẻ theo cách hạ ít điểm nhất
						sh := bestWay[len(bestWay)-1]
						//	fmt.Println("debug phom bestWay, sh", bestWay, sh)
						for _, card := range sh {
							//	fmt.Println("debug phom card hihi", card)
							if CheckCanPopCard(
								session.mapPlayerIdToHand[session.turnPlayerId],
								GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
								card,
							) {
								//	fmt.Println("debug phom card hihi CheckCanPopCard true", card)
								InSessionPopCard(session, card)
								break
							} else {
								//fmt.Println("debug phom card hihi CheckCanPopCard false", card)
							}
						}
					}
				}
			}
		}
	}
	session.isEndedTurnLoop = true
	// đuổi người afk
	afkPids := make([]int64, 0)
	session.mutex.Lock()
	for pid, isAfk := range session.mapPlayerIdToIsAfkLastTurn {
		if isAfk {
			afkPids = append(afkPids, pid)
		}
	}
	session.mutex.Unlock()
	for _, pid := range afkPids {
		pObj := session.GetPlayer(pid)
		if pObj != nil {
			game.RegisterLeaveRoom(session.game, pObj)
			pObj2, _ := player.GetPlayer(pid)
			if pObj2 != nil {
				pObj2.CreatePopUp("Bạn đã không thao tác quá lâu")
			}
		}
	}
	// _________________________________________________________________________
	action := Action{
		actionName:   ACTION_FINISH_SESSION,
		chanResponse: make(chan *ActionResponse),
	}
	t1 := time.After(2 * time.Second)
	select {
	case session.ChanAction <- &action:
		t2 := time.After(2 * time.Second)
		select {
		case <-action.chanResponse:
		case <-t2:
		}
	case <-t1:
	}
	session.checkpoint1 = true
	// _________________________________________________________________________
	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	currencyType := session.game.CurrencyType()
	reasonString := session.room.GetRoomIdentifierString()
	session.mutex.Lock()
	session.checkpoint2 = true
	// chỉ tiền ù nhất nhì ba bét, quân ăn đã tính trước
	mapPlayerIdToWinMoney, winner, isHitJackpot := CalcResult(
		session.room.Requirement(),
		session.isNoDrawsWin,
		session.isSomeoneHas10FullCombos,
		session.isSomeoneHas9FullCombos,
		session.instantWinHand,
		session.turnPlayerId,
		session.firstShowPlayerId,
		len(session.deck),
		session.mapPlayerIdToNextPlayerId,
		session.mapPlayerIdToHand,
		session.mapPlayerIdToEatens,
		session.mapPlayerIdToShowedCombos,
	)
	session.mapPlayerIdToWinMoney = mapPlayerIdToWinMoney
	// add tax to jacpot
	jackpotCode := "all"
	jackpotInstance := jackpot.GetJackpot(jackpotCode, currencyType)
	if jackpotInstance != nil {
		if session.room.GetNumberOfHumans() != 0 {
			moneyToJackpot := int64(float64(session.tax) * TAX_TO_JACKPOT_RATIO)
			jackpotInstance.AddMoney(moneyToJackpot)
		}
		if isHitJackpot {
			if ratio, isIn := mapMoneyUnitToJackpotRatio[session.room.Requirement()]; isIn {
				temp := int64(float64(jackpotInstance.Value()) * ratio)
				jackpotInstance.AddMoney(-temp)
				mapPlayerIdToWinMoney[winner] += temp
				jackpotInstance.NotifySomeoneHitJackpot(
					session.game.GameCode(),
					temp,
					session.GetPlayer(winner).Id(),
					session.GetPlayer(winner).Name())
			}
		}
	}
	// fmt.Println("mapPlayerIdToWinMoney", mapPlayerIdToWinMoney)
	for playerId, winMoney := range mapPlayerIdToWinMoney {
		if winMoney >= 0 {
			temp := game.MoneyAfterTax(winMoney, betEntry)
			session.tax += winMoney - temp
			session.GetPlayer(playerId).ChangeMoneyAndLog(
				temp, currencyType, false, "",
				ACTION_FINISH_SESSION, session.game.GameCode(), session.matchId)
			session.GetROPObj(playerId).EndMatchWinningMoney += temp
			session.GetROPObj(playerId).Changed += temp
		} else { // winMoney < 0, trừ tiền đã đóng băng khi vào phòng
			session.GetPlayer(playerId).ChangeMoneyAndLog(
				winMoney, currencyType, true, reasonString,
				ACTION_FINISH_SESSION, session.game.GameCode(), session.matchId)
			session.GetROPObj(playerId).EndMatchWinningMoney += winMoney
			session.GetROPObj(playerId).Changed += winMoney
		}

		if winMoney >= zmisc.GLOBAL_TEXT_LOWER_BOUND &&
			session.game.currencyType == "money" &&
			session.GetPlayer(playerId).PlayerType() == "normal" {
			zmisc.InsertNewGlobalText(map[string]interface{}{
				"type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
				"username": session.GetPlayer(playerId).DisplayName(),
				"wonMoney": winMoney,
				"gamecode": session.game.GameCode(),
			})
		}
	}
	session.mutex.Unlock()
	session.checkpoint3 = true
	//
	for _, result1p := range session.playerResults {
		result1p.FinishedMoney = session.GetPlayer(result1p.Id).GetMoney(currencyType)
	}
	// LogMatchRecord2
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
	session.moreMatchData["finishingMapPlayerIdToHand"] = session.mapPlayerIdToHand
	session.moreMatchData["finishingMapPlayerIdToEatens"] = session.mapPlayerIdToEatens
	session.moreMatchData["finishingMapPlayerIdToShowedCombos"] = session.mapPlayerIdToShowedCombos
	session.moreMatchData["finishingMapPlayerIdToEatens"] = session.mapPlayerIdToEatens
	session.moreMatchData["playerActions"] = session.playerActions

	if session.room.GetNumberOfHumans() != 0 {
		record.LogMatchRecord3(
			session.game.GameCode(), session.game.CurrencyType(), session.room.Requirement(), session.tax,
			humanWon, humanLost, botWon, botLost,
			session.matchId, playerIpAdds,
			playerResults, session.moreMatchData)
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
	//
	session.room.Mutex.Lock()
	session.room.SharedData["lastWinnerId"] = winner
	session.room.Mutex.Unlock()
	//
	session.room.DidEndGame(map[string]interface{}{}, int(session.game.delayAfterEachGameInSeconds.Seconds())) // second to new match
}

// in session func, for split code

// receive gameplay funcs from outside
func InSessionGameplayActionsReceiver(session *PhomSession) {
	defer func() {
		if r := recover(); r != nil {
			bytes := debug.Stack()
			fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
		}
	}()

	for {
		action := <-session.ChanAction
		actionName := action.actionName
		session.playerActions = append(session.playerActions, action.ToMap())
		if actionName == ACTION_FINISH_SESSION {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else if actionName == ACTION_DRAW_CARD {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if session.isDrawnOrEaten {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0025))}
				} else {
					InSessionDrawCard(session)
					action.chanResponse <- &ActionResponse{err: nil}
					session.room.DidChangeGameState(session)
					if session.isSomeoneHas10FullCombos {
						session.ChanIsPoppedCard <- true
					} else {
						InSessionAfterDrawOrEat(session)
					}
				}
			}
		} else if actionName == ACTION_EAT_CARD {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if session.isDrawnOrEaten {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0025))}
				} else {
					if CheckCanEatCard(
						session.mapPlayerIdToHand[session.turnPlayerId],
						GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
						session.poppedCard,
					) {
						InSessionEatCard(session)
						action.chanResponse <- &ActionResponse{err: nil}
						session.notifyEaten = true
						session.room.DidChangeGameState(session)
						session.notifyEaten = false
						if session.isSomeoneHas10FullCombos {
							session.ChanIsPoppedCard <- true
						} else {
							InSessionAfterDrawOrEat(session)
						}
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0026))}
					}
				}
			}
		} else if actionName == ACTION_AUTO_SHOW_COMBOS {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if !session.isDrawnOrEaten {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0027))}
				} else {
					if len(session.deck) > len(session.players)-1 {
						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0028))}
					} else {
						InSessionAutoShowCombos(session)
						action.chanResponse <- &ActionResponse{err: nil}
						session.room.DidChangeGameState(session)
					}
				}
			}
		} else if actionName == ACTION_SHOW_COMBO_BY_USER {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if !session.isDrawnOrEaten {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0027))}
				} else {
					if len(session.deck) > len(session.players)-1 {
						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0028))}
					} else {
						clientCardsEI := action.data["cardsToShow"]
						clientCardsStrings, isOk := clientCardsEI.([]string)
						if !isOk {
							// cant happen
							action.chanResponse <- &ActionResponse{err: errors.New("cardsTypeMustBe[]string")}
						} else {
							clientCards, err := z.ToCardsFromStrings(clientCardsStrings)
							if err != nil {
								action.chanResponse <- &ActionResponse{err: errors.New("wrongCardText")}
							} else {
								session.mutex.RLock()
								cr, combos, err := CheckIsLegalShowCards(
									session.mapPlayerIdToHand[session.turnPlayerId],
									session.mapPlayerIdToLocks[session.turnPlayerId],
									clientCards,
									session.clientShowedCombos,
								)
								session.mutex.RUnlock()
								if !cr {
									action.chanResponse <- &ActionResponse{err: err}
								} else {
									session.mutex.Lock()
									session.clientShowedCombos = append(session.clientShowedCombos, combos...)
									session.mutex.Unlock()
									action.chanResponse <- &ActionResponse{err: nil}
									session.room.DidChangeGameState(session)
								}
							}
						}
					}
				}
			}
		} else if actionName == ACTION_HANG_CARD {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if len(session.deck) > len(session.players)-1 {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0028))}
				} else {
					if session.isShowedCombos == false {
						InSessionAutoShowCombos(session)
						session.room.DidChangeGameState(session)
						time.Sleep(500 * time.Millisecond)
					}
					cardString := utils.GetStringAtPath(action.data, "cardString")
					targetPlayerId := utils.GetInt64AtPath(action.data, "targetPlayerId")
					comboId := utils.GetIntAtPath(action.data, "comboId")
					card := z.FNewCardFS(cardString)
					err := InSessionHangCard(session, card, targetPlayerId, comboId)
					if err != nil {
						action.chanResponse <- &ActionResponse{err: err}
					} else {
						action.chanResponse <- &ActionResponse{err: nil}
						session.TurnTimer.Add(5 * time.Second)
						session.room.DidChangeGameState(session)
						if session.isSomeoneHas10FullCombos {
							session.ChanIsPoppedCard <- true
						}
					}
				}
			}
		} else if actionName == ACTION_AUTO_HANG_CARDS {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else {
				if len(session.deck) > len(session.players)-1 {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0028))}
				} else {
					if !session.isShowedCombos {
						InSessionAutoShowCombos(session)
						session.room.DidChangeGameState(session)
						time.Sleep(500 * time.Millisecond)
					}
					InSessionAutoHangCards(session)
					action.chanResponse <- &ActionResponse{err: nil}
					session.room.DidChangeGameState(session)
				}
			}
		} else if actionName == ACTION_POP_CARD {
			if session.turnPlayerId != action.playerId {
				action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0024))}
			} else if session.isWaitingForAutoPopCard {
				action.chanResponse <- &ActionResponse{err: errors.New("turn timeout, waiting for auto pop")}
			} else {
				if !session.isDrawnOrEaten {
					action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0027))}
				} else {
					if len(session.deck) <= len(session.players)-1 &&
						!session.isShowedCombos {
						InSessionAutoShowCombos(session)
						session.room.DidChangeGameState(session)
						time.Sleep(500 * time.Millisecond)
					}
					cardString := utils.GetStringAtPath(action.data, "cardString")
					card := z.FNewCardFS(cardString)
					if CheckCanPopCard(
						session.mapPlayerIdToHand[session.turnPlayerId],
						GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
						card,
					) {
						InSessionPopCard(session, card)
						session.ChanIsPoppedCard <- true
						action.chanResponse <- &ActionResponse{err: nil}
						// session.room.DidChangeGameState(session)
						// dont need notify game change state, notify on next turn
					} else {
						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0029))}
					}

				}
			}
		} else {
			fmt.Println("")
			action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0021))}
		}
	}
	// loop ended
	session.checkpoint4 = true
}

// player or system auto draw a card for the turnPlayer
// nếu trong vòng hạ thì tạo các cách hạ bài sau khi rút xong
func InSessionDrawCard(session *PhomSession) {
	session.mutex.Lock()
	session.isDrawnOrEaten = true
	drawnCards, err := z.DealCards(&session.deck, 1)
	if err == nil {
		drawnCard := drawnCards[0]
		session.drawnCard = drawnCard
		session.mapPlayerIdToHand[session.turnPlayerId] = append(session.mapPlayerIdToHand[session.turnPlayerId], drawnCard)
	} else {
		// cant happend, deck is empty
	}
	// nhớ người hạ đầu
	if len(session.deck) == len(session.mapPlayerIdToNextPlayerId)-1 {
		session.firstShowPlayerId = session.turnPlayerId
	}
	// check ù sau khi rút bài
	cr, way := CheckIsFullCombos(
		session.mapPlayerIdToHand[session.turnPlayerId],
		GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
	)
	if cr {
		session.isSomeoneHas10FullCombos = true
		session.instantWinHand = way
	}
	// tạo các cách hạ bài cho người chơi chọn sau khi rút bài ở vòng hạ
	if len(session.deck) <= len(session.players)-1 {
		session.waysToShowCombos = ShowCombosMinPoint(
			session.mapPlayerIdToHand[session.turnPlayerId],
			GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
		)
	}
	session.mutex.Unlock()
}

// check can eat before call this func
func InSessionEatCard(session *PhomSession) {
	//
	session.mutex.Lock()
	session.isDrawnOrEaten = true
	session.mapPlayerIdToHand[session.turnPlayerId] = append(
		session.mapPlayerIdToHand[session.turnPlayerId],
		session.poppedCard,
	)
	//
	prevPlayerId := GetPrevId(session.turnPlayerId, session.mapPlayerIdToNextPlayerId)
	//
	mapMoneyChange := map[int64]int64{
		session.turnPlayerId: 0,
		prevPlayerId:         0,
	}
	//
	if len(session.mapPlayerIdToPoppedCards[prevPlayerId]) >= 1 {
		temp := make([]z.Card, len(session.mapPlayerIdToPoppedCards[prevPlayerId])-1)
		copy(temp, session.mapPlayerIdToPoppedCards[prevPlayerId])
		session.mapPlayerIdToPoppedCards[prevPlayerId] = temp
	} else {
		// debug, khong xay ra neu logic dung
		fmt.Println("debug ERROR InSessionEatCard:824 session.turnPlayerId, prevPlayerId, session.poppedCard, session.mapPlayerIdToHand, session.mapPlayerIdToPoppedCards",
			session.turnPlayerId, prevPlayerId, session.poppedCard, session.mapPlayerIdToHand, session.mapPlayerIdToPoppedCards)
	}
	//
	if len(session.deck) <= len(session.players) { // ăn quân chốt
		session.mapPlayerIdToLocks[session.turnPlayerId][4] = append(
			session.mapPlayerIdToLocks[session.turnPlayerId][4],
			session.poppedCard)
		session.mapPlayerIdToEatens[session.turnPlayerId][4] = append(
			session.mapPlayerIdToEatens[session.turnPlayerId][4],
			session.poppedCard)
		mapMoneyChange[session.turnPlayerId] += 4 * session.room.Requirement()
		mapMoneyChange[prevPlayerId] -= 4 * session.room.Requirement()
	} else { // ăn quân không phải là chốt
		numberOfEatenCards := len(GetListFromLocks(session.mapPlayerIdToEatens[session.turnPlayerId]))
		session.mapPlayerIdToLocks[session.turnPlayerId][int64(numberOfEatenCards+1)] = append(
			session.mapPlayerIdToLocks[session.turnPlayerId][int64(numberOfEatenCards+1)],
			session.poppedCard)
		session.mapPlayerIdToEatens[session.turnPlayerId][int64(numberOfEatenCards+1)] = append(
			session.mapPlayerIdToEatens[session.turnPlayerId][int64(numberOfEatenCards+1)],
			session.poppedCard)
		mapMoneyChange[session.turnPlayerId] += int64(numberOfEatenCards+1) * session.room.Requirement()
		mapMoneyChange[prevPlayerId] -= int64(numberOfEatenCards+1) * session.room.Requirement()
	}
	//
	cr, way := CheckIsFullCombos(
		session.mapPlayerIdToHand[session.turnPlayerId],
		GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
	)
	if cr {
		session.isSomeoneHas10FullCombos = true
		session.instantWinHand = way
	}
	// tạo các cách hạ bài cho người chơi chọn sau khi ăn bài ở vòng hạ
	if len(session.deck) <= len(session.players)-1 {
		session.waysToShowCombos = ShowCombosMinPoint(
			session.mapPlayerIdToHand[session.turnPlayerId],
			GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
		)
	}
	session.mutex.Unlock()
	// cộng trừ tiền ăn bài
	currencyType := session.room.CurrencyType()
	reasonString := session.room.GetRoomIdentifierString()

	betEntry := session.room.Game().BetData().GetEntry(session.room.Requirement())
	temp := game.MoneyAfterTax(mapMoneyChange[session.turnPlayerId], betEntry)
	session.tax += mapMoneyChange[session.turnPlayerId] - temp
	session.GetPlayer(session.turnPlayerId).ChangeMoneyAndLog(
		temp, currencyType, false, "",
		ACTION_EAT_CARD, session.game.GameCode(), session.matchId)
	session.GetROPObj(session.turnPlayerId).Changed += mapMoneyChange[session.turnPlayerId]

	session.GetPlayer(prevPlayerId).ChangeMoneyAndLog(
		mapMoneyChange[prevPlayerId], currencyType, true, reasonString,
		ACTION_EAT_CARD, session.game.GameCode(), session.matchId)
	session.GetROPObj(prevPlayerId).Changed += mapMoneyChange[prevPlayerId]

	// notify money change for room
	session.room.SendAMap(
		"PhomNotifyMoneyEatCard",
		map[string]interface{}{
			fmt.Sprintf("%v", session.turnPlayerId): mapMoneyChange[session.turnPlayerId],
			fmt.Sprintf("%v", prevPlayerId):         mapMoneyChange[prevPlayerId],
		},
	)
}

// tự đánh bài và kết thúc khi có thể ù 9,
func InSessionAfterDrawOrEat(session *PhomSession) {
	session.mutex.RLock()
	ways := ShowCombosMinPoint(
		session.mapPlayerIdToHand[session.turnPlayerId],
		GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
	)
	session.mutex.RUnlock()
	if len(ways) > 0 {
		bestWay := ways[0]
		// bài lẻ theo cách hạ ít điểm nhất
		sh := bestWay[len(bestWay)-1]
		if len(sh) == 1 {
			time.Sleep(1500 * time.Millisecond)
			InSessionPopCard(session, sh[0])
			session.ChanIsPoppedCard <- true
		}
	}
}

// check can pop before call this func
func InSessionPopCard(session *PhomSession, card z.Card) {
	session.mutex.Lock()
	session.mapPlayerIdToHand[session.turnPlayerId] = z.Subtracted(
		session.mapPlayerIdToHand[session.turnPlayerId],
		[]z.Card{card})
	session.poppedCard = card
	session.mapPlayerIdToPoppedCards[session.turnPlayerId] = append(session.mapPlayerIdToPoppedCards[session.turnPlayerId], card)
	cr, way := CheckIsFullCombos(
		session.mapPlayerIdToHand[session.turnPlayerId],
		GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
	)
	//fmt.Println("CheckIsFullCombos InSessionPopCard", cr, way)
	if cr {
		session.isSomeoneHas9FullCombos = true
		if session.isShowedCombos {
			nTemp := len(session.mapPlayerIdToShowedCombos[session.turnPlayerId])
			if nTemp >= 1 {
				session.instantWinHand = session.mapPlayerIdToShowedCombos[session.turnPlayerId][:nTemp-1]
			} else { // cant happend
				session.instantWinHand = session.mapPlayerIdToShowedCombos[session.turnPlayerId]
			}
		} else {
			session.instantWinHand = way
		}
	}
	session.mutex.Unlock()
}

// check cond before call this func
func InSessionAutoShowCombos(session *PhomSession) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	//
	var choosenWayToShow [][]z.Card
	if len(session.clientShowedCombos) == 0 {
		ways := ShowCombosMinPoint(
			session.mapPlayerIdToHand[session.turnPlayerId],
			GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
		)
		choosenWayToShow = ways[0]
	} else {
		clientWays := [][][]z.Card{}
		allShowedWays := ShowCombos(
			session.mapPlayerIdToHand[session.turnPlayerId],
			GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
		)
		for _, showedWay := range allShowedWays {
			if CheckIsCombosInWay(session.clientShowedCombos, showedWay) {
				clientWays = append(clientWays, showedWay)
			}
		}
		// fmt.Println("clientWays", clientWays)
		minDiffNoc := 10
		minIndex := 0
		for i, clientWay := range clientWays {
			lenDif := len(clientWay) - len(session.clientShowedCombos)
			if lenDif < 0 {
				lenDif = -lenDif
			}
			if lenDif < minDiffNoc {
				minDiffNoc = lenDif
				minIndex = i
			}
		}
		if len(clientWays) > 0 {
			choosenWayToShow = clientWays[minIndex]
		} else {
			// cant happen
			theHand := make([]z.Card, len(session.mapPlayerIdToHand[session.turnPlayerId]))
			copy(theHand, session.mapPlayerIdToHand[session.turnPlayerId])
			choosenWayToShow = [][]z.Card{theHand}
		}
		//		fmt.Println("clientWays minIndex ", clientWays, minIndex)
	}

	//	fmt.Println("Hand[turnPlayerId]: ", session.mapPlayerIdToHand[session.turnPlayerId])
	//	fmt.Println("locks", GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]))
	//	fmt.Println("session.clientShowedCombos: ", session.clientShowedCombos)
	//	fmt.Println("choosenWayToShow: ", choosenWayToShow)

	// noc = số phỏm + 1
	noc := len(session.mapPlayerIdToShowedCombos[session.turnPlayerId])
	if noc == 0 { // lần đầu hạ phỏm
		session.mapPlayerIdToShowedCombos[session.turnPlayerId] = choosenWayToShow
	} else {
		if len(choosenWayToShow) > 1 { // có phỏm mới, chèn thêm vào
			temp := make([][]z.Card, noc-1)
			copy(temp, session.mapPlayerIdToShowedCombos[session.turnPlayerId]) // copy phỏm cũ
			temp = append(temp, choosenWayToShow[0])                            // thêm phỏm mới
			temp = append(temp, choosenWayToShow[1])                            // thêm quân lẻ
			session.mapPlayerIdToShowedCombos[session.turnPlayerId] = temp
		} else {
		}
	}
	cardsInCombos := []z.Card{}
	for i, combo := range session.mapPlayerIdToShowedCombos[session.turnPlayerId] {
		if i != len(session.mapPlayerIdToShowedCombos[session.turnPlayerId])-1 {
			cardsInCombos = append(cardsInCombos, combo...)
		}
	}
	session.mapPlayerIdToHand[session.turnPlayerId] = z.Subtracted(
		session.mapPlayerIdToHand[session.turnPlayerId],
		cardsInCombos,
	)
	session.mapPlayerIdToLocks[session.turnPlayerId] = map[int64][]z.Card{
		1: []z.Card{},
		2: []z.Card{},
		3: []z.Card{},
		4: []z.Card{},
	}
	session.isShowedCombos = true
}

// check isShowedCombos == true before call this func
// if targetPlayerId = -1, auto find
func InSessionHangCard(session *PhomSession, card z.Card, targetPlayerId int64, comboId int) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	n := len(session.mapPlayerIdToShowedCombos[session.turnPlayerId])
	if n >= 2 { // check người chơi có phỏm
		sCard := card
		if z.FindCardInSlice(
			sCard,
			session.mapPlayerIdToHand[session.turnPlayerId],
		) != -1 { // check quân cần hạ là ở trên tay
			pId := targetPlayerId
			showedCombos := session.mapPlayerIdToShowedCombos[pId]
			if session.GetPlayer(pId) != nil {
				i := comboId
				if i < len(showedCombos)-1 { // cards ở index -1 là quân lẻ
					showedCombo := showedCombos[i]
					cs := make([]z.Card, len(showedCombo))
					copy(cs, showedCombo)
					cs = append(cs, sCard)
					if r, _ := CheckIsCombo(cs); r == true {
						session.mapPlayerIdToShowedCombos[pId][i] = cs
						session.mapPlayerIdToHand[session.turnPlayerId] = z.Subtracted(
							session.mapPlayerIdToHand[session.turnPlayerId],
							[]z.Card{sCard},
						)
						session.mapPlayerIdToHungCards[session.turnPlayerId] = append(session.mapPlayerIdToHungCards[session.turnPlayerId], sCard)
						// check ù sau khi gửi bài
						cr, way := CheckIsFullCombos(
							session.mapPlayerIdToHand[session.turnPlayerId],
							GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
						)
						if cr {
							session.isSomeoneHas10FullCombos = true
							session.instantWinHand = way
						}
						return nil
					} else {
						return errors.New(l.Get(l.M0030))
					}
				} else {
					return errors.New("wrong_combo_id")
				}
			} else { // targetPlayerId = -1
				for pId, showedCombos := range session.mapPlayerIdToShowedCombos {
					if len(showedCombos) > 1 {
						for i := 0; i < len(showedCombos)-1; i++ { // cards ở index -1 là quân lẻ
							showedCombo := showedCombos[i]
							cs := make([]z.Card, len(showedCombo))
							copy(cs, showedCombo)
							cs = append(cs, sCard)
							if r, _ := CheckIsCombo(cs); r == true {
								session.mapPlayerIdToShowedCombos[pId][i] = cs
								session.mapPlayerIdToHand[session.turnPlayerId] = z.Subtracted(
									session.mapPlayerIdToHand[session.turnPlayerId],
									[]z.Card{sCard},
								)
								session.mapPlayerIdToHungCards[session.turnPlayerId] = append(session.mapPlayerIdToHungCards[session.turnPlayerId], sCard)
								// check ù sau khi gửi bài
								cr, way := CheckIsFullCombos(
									session.mapPlayerIdToHand[session.turnPlayerId],
									GetListFromLocks(session.mapPlayerIdToLocks[session.turnPlayerId]),
								)
								if cr {
									session.isSomeoneHas10FullCombos = true
									session.instantWinHand = way
								}
								return nil
							}
						}
					}
				}
				return errors.New("cant_find_combo_to_hang")
			}
		} else {
			return errors.New("cant_hang_card_not_in_hand")
		}
	} else {
		return errors.New(l.Get(l.M0031))
	}
}

// tự động gửi bài cho người chơi
// rồi check xem có ù tròn không
func InSessionAutoHangCards(session *PhomSession) {
	// TODO: địch có 3s4s5s, gửi 7s trước 6s sau thì xịt 7s.
	n := len(session.mapPlayerIdToShowedCombos[session.turnPlayerId])
	if n >= 1 { // check lại người chơi đã hạ phỏm
		singleCards := session.mapPlayerIdToShowedCombos[session.turnPlayerId][n-1]
		for _, sCard := range singleCards {
			for pId, showedCombos := range session.mapPlayerIdToShowedCombos {
				if len(showedCombos) > 1 {
					for i, _ := range showedCombos[:len(showedCombos)-1] {
						InSessionHangCard(session, sCard, pId, i)
					}
				}
			}
		}
	}
}

// Interface

func (session *PhomSession) CleanUp() {
}

func (session *PhomSession) HandlePlayerRemovedFromGame(player game.GamePlayer) {

}
func (session *PhomSession) HandlePlayerAddedToGame(player game.GamePlayer) {

}

func (session *PhomSession) HandlePlayerOffline(player game.GamePlayer) {

}
func (session *PhomSession) HandlePlayerOnline(player game.GamePlayer) {

}

func (session *PhomSession) IsPlaying() bool {
	return true
}

func (session *PhomSession) IsDelayingForNewGame() bool {
	return true
}

func (session *PhomSession) SerializedData() map[string]interface{} {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	data := make(map[string]interface{})
	// data["game_code"] = session.game.GameCode()
	// data["requirement"] = session.room.Requirement()
	data["matchId"] = session.matchId
	data["player_ids"] = session.GetPlayerIds()
	data["isNoDrawsWin"] = session.isNoDrawsWin
	data["isSomeoneHas9FullCombos"] = session.isSomeoneHas9FullCombos
	data["isSomeoneHas10FullCombos"] = session.isSomeoneHas10FullCombos
	data["instantWinHand"] = z.ToStringss(session.instantWinHand)
	data["deckLen"] = len(session.deck)
	//	data["deck"] = z.ToSliceString(session.deck)
	data["turnPlayerId"] = session.turnPlayerId
	if session.TurnTimer != nil {
		data["turnRemainingTime"] = session.TurnTimer.EndPoint.Sub(time.Now()).Seconds()
	}
	data["mapPlayerIdToNextPlayerId"] = session.mapPlayerIdToNextPlayerId

	temp := make(map[int64][]string)
	for pId, _ := range session.mapPlayerIdToLocks {
		temp[pId] = z.ToSliceString(GetListFromLocks(session.mapPlayerIdToLocks[pId]))
	}
	data["mapPlayerIdToLocks"] = temp

	data["isDrawnOrEaten"] = session.isDrawnOrEaten
	data["isShowedCombos"] = session.isShowedCombos
	data["poppedCard"] = session.poppedCard.String()
	data["clientShowedCombos"] = z.ToStringss(session.clientShowedCombos)

	temp2 := make(map[int64][][]string)
	for pId, combos := range session.mapPlayerIdToShowedCombos {
		if len(combos) >= 1 {
			temp2[pId] = z.ToStringss(combos[:len(combos)-1]) // không show các quân lẻ
		} else {
			temp2[pId] = [][]string{}
		}
	}
	data["mapPlayerIdToShowedCombos"] = temp2

	temp3 := make(map[int64][]string)
	for pId, poppedCards := range session.mapPlayerIdToPoppedCards {
		temp3[pId] = z.ToSliceString(poppedCards)
	}
	data["mapPlayerIdToPoppedCards"] = temp3

	temp4 := make(map[int64][]string)
	for pId, cards := range session.mapPlayerIdToHungCards {
		temp4[pId] = z.ToSliceString(cards)
	}
	data["mapPlayerIdToHungCards"] = temp4

	if session.isEndedTurnLoop {
		temp1 := make([]map[string]interface{}, 0)
		for _, result1p := range session.playerResults {
			temp1 = append(temp1, result1p.ToMap())
		}
		data["results"] = temp1
		data["resultsNhatNhiBaBet"] = session.mapPlayerIdToWinMoney
	}

	// biến thêm chỉ giai đoạn, tính từ các biến khác
	phase := ""
	if !session.isDrawnOrEaten {
		phase = "phase_draw_or_eat"
	} else {
		if len(session.deck) <= len(session.players)-1 {
			phase = "phase_show_hang_pop"
		} else {
			phase = "phase_pop"
		}
	}
	data["phase"] = phase

	// client sẽ cập nhật linh tinh cả nếu biến này là true
	data["isEaten"] = session.notifyEaten
	data["mapPlayerIdToPtype"] = session.mapPlayerIdToPtype // this map only for read, dont need to clone

	data["checkpoint1"] = session.checkpoint1
	data["checkpoint2"] = session.checkpoint2
	data["checkpoint3"] = session.checkpoint3
	data["checkpoint4"] = session.checkpoint4

	return data
}

func (session *PhomSession) SerializedDataForPlayer(currentPlayer game.GamePlayer) map[string]interface{} {
	data := session.SerializedData()

	if currentPlayer == nil {
		// vào phòng đúng thời điểm bắt đầu ván chơi
		return data
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()
	mapPlayerIdToHand := make(map[int64][]string)
	for player, _ := range session.players {
		if player.Id() == currentPlayer.Id() {
			mapPlayerIdToHand[player.Id()] = z.ToSliceString(session.mapPlayerIdToHand[player.Id()])
		} else {
			//			if session.isEndedTurnLoop {
			if true {
				mapPlayerIdToHand[player.Id()] = z.ToSliceString(session.mapPlayerIdToHand[player.Id()])
			}
		}
	}
	data["mapPlayerIdToHand"] = mapPlayerIdToHand

	if currentPlayer.Id() == session.turnPlayerId {
		data["drawnCard"] = session.drawnCard.String()
		// data["waysToShowCombos"] = z.ToStringsss(session.waysToShowCombos)
	}

	return data
}

func (session *PhomSession) ResultSerializedData() map[string]interface{} {
	return map[string]interface{}{}
}

func (session *PhomSession) GetPlayer(playerId int64) (player game.GamePlayer) {
	for player, _ := range session.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (session *PhomSession) GetPlayerIds() []int64 {
	result := make([]int64, 0)
	for player, _ := range session.players {
		result = append(result, player.Id())
	}
	return result
}

// get resultOnePlayer obj
func (session *PhomSession) GetROPObj(playerId int64) *ResultOnePlayer {
	for _, ropo := range session.playerResults {
		if playerId == ropo.Id {
			return ropo
		}
	}
	return &ResultOnePlayer{} // cant happen with right logic
}
