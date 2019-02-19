package gamemini

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/gamemini/consts"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

type Ag1CardX2 struct {
	// ag will change
	//		jackpotObj value and
	//		resultObj.SumWonMoney and
	//		resultObj.SumLostMoney
	playerObj         *player.Player
	jackpotObj        *jackpot.Jackpot
	resultObj         *ResultSlot
	sumMoneyAfterSpin int64
	currencyType      string
	//
	isFinished         bool
	ChanActionReceiver chan *Action
	//
	// from 0 to 6
	currentXxxLevel int
	// money to bet at current level,
	currentXxxMoney     int64
	requiredMoneyToGoOn int64
	winningMoneyIfStop  int64
	// whether player choose ACTION_SELECT_SMALL or ACTION_SELECT_BIG is right
	isRightSelection bool
	// a card, ex "6c", "8d", "Qh"..
	serverSelectionResult string
	// choose stop, small or big
	chanSelection chan string
}

//
func NewAg1CardX2(
	playerObj *player.Player,
	jackpotObj *jackpot.Jackpot,
	resultObj *ResultSlot,
	sumMoneyAfterSpin int64,
	currencyType string,
) AdditionalGameInterface {
	gameObj := Ag1CardX2{
		playerObj:         playerObj,
		jackpotObj:        jackpotObj,
		resultObj:         resultObj,
		sumMoneyAfterSpin: sumMoneyAfterSpin,
		currencyType:      currencyType,

		chanSelection:      make(chan string),
		ChanActionReceiver: make(chan *Action),
	}
	return &gameObj
}

func (ag *Ag1CardX2) ReceiveAction(action *Action) {
	ag.ChanActionReceiver <- action
	timeout := time.After(2 * time.Second)
	select {
	case <-action.ChanResponse:
		//
	case <-timeout:
		//
	}
}

//
func (ag *Ag1CardX2) serializedData() map[string]interface{} {
	return map[string]interface{}{
		"isFinished":            ag.isFinished,
		"currentXxxLevel":       ag.currentXxxLevel,
		"currentXxxMoney":       ag.currentXxxMoney,
		"serverSelectionResult": ag.serverSelectionResult,
		"requiredMoneyToGoOn":   ag.requiredMoneyToGoOn,
	}
}

//
func (ag *Ag1CardX2) updateAgStatus() {
	data := ag.serializedData()
	ServerObj.SendRequest(
		"Ag1CardX2UpdateAgStatus",
		data,
		ag.playerObj.Id(),
	)
}

func (ag *Ag1CardX2) agLoopReceiveActions() {
	for {
		action := <-ag.ChanActionReceiver
		if action.ActionName == consts.ACTION_FINISH_AG {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == consts.ACTION_STOP_PLAYING ||
					action.ActionName == consts.ACTION_SELECT_SMALL ||
					action.ActionName == consts.ACTION_SELECT_BIG {
					timer := time.After(3 * time.Second)
					select {
					case ag.chanSelection <- action.ActionName:
						action.ChanResponse <- &ActionResponse{Err: nil}
					case <-timer:
						action.ChanResponse <- &ActionResponse{Err: errors.New("timeout AgLoopReceiveActions Ag1CardX2")}
					}
				} else {
					action.ChanResponse <- &ActionResponse{
						Err: errors.New("Sent wrong action to AgLoopReceiveActions")}
				}
			}()
		}
	}
	//fmt.Println("finished AgLoopReceiveActions", ag)
}

func (ag *Ag1CardX2) Start() {
	ag.updateAgStatus()
	go ag.agLoopReceiveActions()
	//
	ag.winningMoneyIfStop = ag.sumMoneyAfterSpin
	ag.currentXxxMoney = ag.sumMoneyAfterSpin
	ag.requiredMoneyToGoOn = 0
	// loop x2 game, i is level counter
	i := 0
	for i < consts.MAX_XXX_LEVEL {
		if ag.playerObj.GetAvailableMoney(ag.currencyType) < ag.currentXxxMoney {
			break
		}
		ag.currentXxxLevel = i
		ag.updateAgStatus()
		timer := time.After(600 * time.Second)
		var phase3choice string
		select {
		case <-timer:
			phase3choice = consts.ACTION_STOP_PLAYING
		case phase3choice = <-ag.chanSelection:
			// received player selection
		}
		if phase3choice == consts.ACTION_STOP_PLAYING {
			break
		} else {
			ag.serverSelectionResult = random1Card(phase3choice, i)

			ag.isRightSelection = false
			var cardRank string
			if len(ag.serverSelectionResult) > 0 {
				cardRank = string(ag.serverSelectionResult[0])
			}
			if cardRank == "A" || cardRank == "2" || cardRank == "3" ||
				cardRank == "4" || cardRank == "5" || cardRank == "6" {
				if phase3choice == consts.ACTION_SELECT_SMALL {
					ag.isRightSelection = true
				}
			} else if cardRank == "7" {

			} else {
				if phase3choice == consts.ACTION_SELECT_BIG {
					ag.isRightSelection = true
				}
			}

			if ag.isRightSelection {
				ag.requiredMoneyToGoOn = 0
				ag.currentXxxMoney = 2 * ag.currentXxxMoney
				ag.winningMoneyIfStop = ag.currentXxxMoney
				i += 1
			} else {
				ag.winningMoneyIfStop = ag.currentXxxMoney
				ag.requiredMoneyToGoOn = ag.currentXxxMoney

				if ag.requiredMoneyToGoOn > 0 {
					ag.playerObj.ChangeMoneyAndLog(
						-ag.requiredMoneyToGoOn, ag.currencyType, false, "",
						phase3choice, consts.AGCODE_1CARDX2, "")
					//
					ag.resultObj.SumLostMoney -= ag.requiredMoneyToGoOn
					// add half money to jackpot
					ag.jackpotObj.AddMoney(ag.requiredMoneyToGoOn / 52)
				}
			}
		}
	} // end loop x2 game
	ag.currentXxxLevel = i
	if i == consts.MAX_XXX_LEVEL {
		amount := int64(float64(ag.jackpotObj.Value()) * 0.05)
		ag.winningMoneyIfStop += amount
		ag.jackpotObj.AddMoney(-amount)
		ag.jackpotObj.NotifySomeoneHitJackpot(
			ag.jackpotObj.GameCode,
			amount,
			ag.playerObj.Id(),
			ag.playerObj.Name(),
		)
	}
	ag.resultObj.SumWonMoney = ag.winningMoneyIfStop

	//
	ag.isFinished = true
	ag.updateAgStatus()
	action := Action{
		ActionName:   consts.ACTION_FINISH_AG,
		ChanResponse: make(chan *ActionResponse),
	}
	ag.ChanActionReceiver <- &action
	<-action.ChanResponse
}

// for agxxx,
// đã bao gồm hư cấu,
// i là level x2
func random1Card(userChoice string, i int) string {
	smallCards := []string{
		"2c", "2d", "2h", "2s", "3c", "3d", "3h", "3s",
		"4c", "4d", "4h", "4s", "5c", "5d", "5h", "5s",
		"6c", "6d", "6h", "6s", "Ac", "Ad", "Ah", "As"}
	bigCards := []string{
		"8c", "8d", "8h", "8s", "9c", "9d", "9h", "9s",
		"Jc", "Jd", "Jh", "Js", "Kc", "Kd", "Kh", "Ks",
		"Qc", "Qd", "Qh", "Qs", "Tc", "Td", "Th", "Ts"}
	var goodCards, badCards []string
	if userChoice == consts.ACTION_SELECT_SMALL {
		goodCards = smallCards
		badCards = bigCards
	} else {
		goodCards = bigCards
		badCards = smallCards
	}
	var result string
	r := rand.Intn(13)
	if r < 6 {
		result = goodCards[rand.Intn(len(goodCards))]
	} else {
		result = badCards[rand.Intn(len(badCards))]
	}

	if (i >= consts.MAX_XXX_LEVEL-1) && (rand.Intn(100) < 70) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-2) && (rand.Intn(100) < 60) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-3) && (rand.Intn(100) < 50) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-4) && (rand.Intn(100) < 40) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-5) && (rand.Intn(100) < 30) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-6) && (rand.Intn(100) < 20) {
		result = badCards[rand.Intn(len(badCards))]
	} else if (i >= consts.MAX_XXX_LEVEL-7) && (rand.Intn(100) < 10) {
		result = badCards[rand.Intn(len(badCards))]
	}
	return result
}

type AgRandomX1to5 struct {
	// ag will change
	//		jackpotObj value and
	//		resultObj.SumWonMoney and
	//		resultObj.SumLostMoney
	playerObj         *player.Player
	jackpotObj        *jackpot.Jackpot
	resultObj         *ResultSlot
	sumMoneyAfterSpin int64
	currencyType      string

	//
	isFinished         bool
	ChanActionReceiver chan *Action

	//
	chanSelection chan string
	// result in [1..5]
	result int64
}

//
func NewAgRandomX1to5(
	playerObj *player.Player,
	jackpotObj *jackpot.Jackpot,
	resultObj *ResultSlot,
	sumMoneyAfterSpin int64,
	currencyType string,
) AdditionalGameInterface {
	gameObj := AgRandomX1to5{
		playerObj:         playerObj,
		jackpotObj:        jackpotObj,
		resultObj:         resultObj,
		sumMoneyAfterSpin: sumMoneyAfterSpin,
		currencyType:      currencyType,

		chanSelection:      make(chan string),
		ChanActionReceiver: make(chan *Action),
	}
	return &gameObj
}

func (ag *AgRandomX1to5) ReceiveAction(action *Action) {
	ag.ChanActionReceiver <- action
	timeout := time.After(2 * time.Second)
	select {
	case <-action.ChanResponse:
		//
	case <-timeout:
		//
	}
}

//
func (ag *AgRandomX1to5) serializedData() map[string]interface{} {
	return map[string]interface{}{
		"isFinished": ag.isFinished,
		"result":     ag.result,
	}
}

//
func (ag *AgRandomX1to5) updateAgStatus() {
	data := ag.serializedData()
	ServerObj.SendRequest(
		"Ag1RandomX1to5UpdateAgStatus",
		data,
		ag.playerObj.Id(),
	)
}

func (ag *AgRandomX1to5) agLoopReceiveActions() {
	for {
		action := <-ag.ChanActionReceiver
		if action.ActionName == consts.ACTION_FINISH_AG {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == consts.ACTION_CHOOSE {
					timer := time.After(3 * time.Second)
					select {
					case ag.chanSelection <- action.ActionName:
						action.ChanResponse <- &ActionResponse{Err: nil}
					case <-timer:
						action.ChanResponse <- &ActionResponse{Err: errors.New("timeout AgLoopReceiveActions AgRandomX1to5 ")}
					}
				} else {
					action.ChanResponse <- &ActionResponse{
						Err: errors.New("Sent wrong action to AgLoopReceiveActions")}
				}
			}()
		}
	}
	//fmt.Println("finished AgLoopReceiveActions", ag)
}

func (ag *AgRandomX1to5) Start() {
	ag.updateAgStatus()
	go ag.agLoopReceiveActions()
	//
	timer := time.After(20 * time.Second)
	select {
	case <-timer:
	case <-ag.chanSelection:
	}
	r := rand.Float64()
	if 0 <= r && r < 0.4 {
		ag.result = 1
	} else if 0.4 <= r && r < 0.6 {
		ag.result = 2
	} else if 0.6 <= r && r < 0.85 {
		ag.result = 3
	} else if 0.85 <= r && r < 0.95 {
		ag.result = 4
	} else if 0.95 <= r && r < 1 {
		ag.result = 5
	}
	ag.resultObj.SumWonMoney = ag.result * ag.sumMoneyAfterSpin

	//
	ag.isFinished = true
	ag.updateAgStatus()
	action := Action{
		ActionName:   consts.ACTION_FINISH_AG,
		ChanResponse: make(chan *ActionResponse),
	}
	ag.ChanActionReceiver <- &action
	<-action.ChanResponse
}

type AgTaixiuX2 struct {
	// ag will change
	//		jackpotObj value and
	//		resultObj.SumWonMoney and
	//		resultObj.SumLostMoney
	playerObj         *player.Player
	jackpotObj        *jackpot.Jackpot
	resultObj         *ResultSlot
	sumMoneyAfterSpin int64
	currencyType      string
	//
	isFinished         bool
	ChanActionReceiver chan *Action
	//
	serverTaixiuResult []int
	isRightSelection   bool
	// a card, ex "6c", "8d", "Qh"..
	// choose stop, small or big
	chanSelection chan string
}

//
func NewAgTaixiuX2(
	playerObj *player.Player,
	jackpotObj *jackpot.Jackpot,
	resultObj *ResultSlot,
	sumMoneyAfterSpin int64,
	currencyType string,
) AdditionalGameInterface {
	gameObj := AgTaixiuX2{
		playerObj:         playerObj,
		jackpotObj:        jackpotObj,
		resultObj:         resultObj,
		sumMoneyAfterSpin: sumMoneyAfterSpin,
		currencyType:      currencyType,

		chanSelection:      make(chan string),
		ChanActionReceiver: make(chan *Action),
	}
	return &gameObj
}

func (ag *AgTaixiuX2) ReceiveAction(action *Action) {
	ag.ChanActionReceiver <- action
	timeout := time.After(2 * time.Second)
	select {
	case <-action.ChanResponse:
		//
	case <-timeout:
		//
	}
}

//
func (ag *AgTaixiuX2) serializedData() map[string]interface{} {
	return map[string]interface{}{
		"isFinished":         ag.isFinished,
		"isRightSelection":   ag.isRightSelection,
		"serverTaixiuResult": ag.serverTaixiuResult,
	}
}

//
func (ag *AgTaixiuX2) updateAgStatus() {
	data := ag.serializedData()
	ServerObj.SendRequest(
		"AgTaixiuX2UpdateAgStatus",
		data,
		ag.playerObj.Id(),
	)
}

func (ag *AgTaixiuX2) agLoopReceiveActions() {
	for {
		action := <-ag.ChanActionReceiver
		if action.ActionName == consts.ACTION_FINISH_AG {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == consts.ACTION_SELECT_SMALL ||
					action.ActionName == consts.ACTION_SELECT_BIG {
					timer := time.After(3 * time.Second)
					select {
					case ag.chanSelection <- action.ActionName:
						action.ChanResponse <- &ActionResponse{Err: nil}
					case <-timer:
						action.ChanResponse <- &ActionResponse{Err: errors.New("timeout AgLoopReceiveActions Ag1CardX2")}
					}
				} else {
					action.ChanResponse <- &ActionResponse{
						Err: errors.New("Sent wrong action to AgLoopReceiveActions")}
				}
			}()
		}
	}
	//fmt.Println("finished AgLoopReceiveActions", ag)
}

func (ag *AgTaixiuX2) Start() {
	ag.updateAgStatus()
	go ag.agLoopReceiveActions()

	timer := time.After(20 * time.Second)
	var userSelection string
	select {
	case <-timer:
		userSelection = consts.ACTION_SELECT_BIG
	case userSelection = <-ag.chanSelection:
		// received player selection
	}

	ag.serverTaixiuResult = randomTaixiu(userSelection)
	var temp string
	if ag.serverTaixiuResult[0]+ag.serverTaixiuResult[1]+
		ag.serverTaixiuResult[2] <= 10 {
		temp = consts.ACTION_SELECT_SMALL
	} else {
		temp = consts.ACTION_SELECT_BIG
	}
	if temp == userSelection {
		ag.isRightSelection = true
	} else {
		ag.isRightSelection = false
	}
	if ag.isRightSelection {
		ag.resultObj.SumWonMoney = 2 * ag.sumMoneyAfterSpin
	} else {
		ag.resultObj.SumWonMoney = 1 * ag.sumMoneyAfterSpin
	}

	//
	ag.isFinished = true
	ag.updateAgStatus()
	action := Action{
		ActionName:   consts.ACTION_FINISH_AG,
		ChanResponse: make(chan *ActionResponse),
	}
	ag.ChanActionReceiver <- &action
	<-action.ChanResponse
}

//
func randomTaixiu(userChoice string) []int {
	r := rand.Intn(100)
	userWinrate := 45
	result := []int{1, 1, 1}
	if (userChoice == consts.ACTION_SELECT_BIG && r < userWinrate) ||
		(userChoice == consts.ACTION_SELECT_SMALL && r >= userWinrate) {
		result[0] = 4 + rand.Intn(3)
		result[1] = 4 + rand.Intn(3)
		result[2] = 4 + rand.Intn(3)
	} else {
		result[0] = 1 + rand.Intn(3)
		result[1] = 1 + rand.Intn(3)
		result[2] = 1 + rand.Intn(3)
	}
	return result
}

type AgGoldMiner struct {
	// ag will change
	//		jackpotObj value and
	//		resultObj.SumWonMoney and
	//		resultObj.SumLostMoney
	playerObj         *player.Player
	jackpotObj        *jackpot.Jackpot
	resultObj         *ResultSlot
	sumMoneyAfterSpin int64
	currencyType      string

	//
	isFinished         bool
	ChanActionReceiver chan *Action

	// all pots info
	mapPotIndexToPrize map[int]int64
	// map the number of opened pots to moneyToOpenNextPot
	moneyToOpenPots map[int]int64
	// opened pots info, map potIndex to prize
	userChosenPots      map[int]int64
	userLastChosenIndex int
	chanSelection       chan int

	mutex sync.Mutex
}

//
func NewAgGoldMiner(
	playerObj *player.Player,
	jackpotObj *jackpot.Jackpot,
	resultObj *ResultSlot,
	sumMoneyAfterSpin int64,
	currencyType string,
) AdditionalGameInterface {
	gameObj := AgGoldMiner{
		playerObj:         playerObj,
		jackpotObj:        jackpotObj,
		resultObj:         resultObj,
		sumMoneyAfterSpin: sumMoneyAfterSpin,
		currencyType:      currencyType,

		ChanActionReceiver: make(chan *Action),

		chanSelection:      make(chan int),
		mapPotIndexToPrize: make(map[int]int64),
		moneyToOpenPots:    make(map[int]int64),
		userChosenPots:     make(map[int]int64),
	}
	return &gameObj
}

func (ag *AgGoldMiner) ReceiveAction(action *Action) {
	ag.ChanActionReceiver <- action
	timeout := time.After(2 * time.Second)
	select {
	case <-action.ChanResponse:
		//
	case <-timeout:
		//
	}
}

//
func (ag *AgGoldMiner) serializedData() map[string]interface{} {
	ag.mutex.Lock()
	defer ag.mutex.Unlock()
	clonedMapIndexPrizeRate := map[int]int64{}
	for k, v := range ag.mapPotIndexToPrize {
		clonedMapIndexPrizeRate[k] = v
	}
	clonedUserChosenIndexs := map[int]int64{}
	for k, v := range ag.userChosenPots {
		clonedUserChosenIndexs[k] = v
	}
	clonedMoneyToOpenPots := map[int]int64{}
	for k, v := range ag.moneyToOpenPots {
		clonedMoneyToOpenPots[k] = v
	}
	r := map[string]interface{}{
		"isFinished":          ag.isFinished,
		"mapPotIndexToPrize":  clonedMapIndexPrizeRate,
		"userChosenPots":      clonedUserChosenIndexs,
		"userLastChosenIndex": ag.userLastChosenIndex,
		"moneyToOpenPots":     clonedMoneyToOpenPots,
	}
	return r
}

//
func (ag *AgGoldMiner) updateAgStatus() {
	data := ag.serializedData()
	ServerObj.SendRequest(
		"Ag1RandomX1to5UpdateAgStatus",
		data,
		ag.playerObj.Id(),
	)
}

func (ag *AgGoldMiner) agLoopReceiveActions() {
	for {
		action := <-ag.ChanActionReceiver
		if action.ActionName == consts.ACTION_FINISH_AG {
			action.ChanResponse <- &ActionResponse{Err: nil}
			break
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				if action.ActionName == consts.ACTION_CHOOSE_GOLD_POT_INDEX ||
					action.ActionName == consts.ACTION_STOP_PLAYING {
					potIndex := utils.GetIntAtPath(action.Data, "potIndex")
					if action.ActionName == consts.ACTION_STOP_PLAYING {
						potIndex = -1
					}
					ag.mutex.Lock()
					_, isIn := ag.userChosenPots[potIndex]
					ag.mutex.Unlock()
					if isIn {
						action.ChanResponse <- &ActionResponse{Err: errors.New("Duplicate pot choice")}
					}
					timer := time.After(3 * time.Second)
					select {
					case ag.chanSelection <- potIndex:
						action.ChanResponse <- &ActionResponse{Err: nil}
					case <-timer:
						action.ChanResponse <- &ActionResponse{Err: errors.New("timeout AgLoopReceiveActions AgRandomX1to5 ")}
					}
				} else {
					action.ChanResponse <- &ActionResponse{
						Err: errors.New("Sent wrong action to AgLoopReceiveActions")}
				}
			}()
		}
	}
	//fmt.Println("finished AgLoopReceiveActions", ag)
}

func (ag *AgGoldMiner) Start() {
	go ag.agLoopReceiveActions()
	ag.mapPotIndexToPrize, ag.moneyToOpenPots =
		createAgGoldMinerPots(ag.sumMoneyAfterSpin)
	ag.updateAgStatus()
	//
ForLoop:
	for {
		nOpenedPots := len(ag.userChosenPots)
		moneyToOpenNextPot := ag.moneyToOpenPots[nOpenedPots]
		if ag.playerObj.GetAvailableMoney(ag.currencyType) < moneyToOpenNextPot {
			break ForLoop
		}
		timer := time.After(20 * time.Second)
		var userChosenIndex int
		select {
		case <-timer:
			break ForLoop
		case userChosenIndex = <-ag.chanSelection:
		}
		if userChosenIndex == -1 {
			break ForLoop
		} else {
			ag.mutex.Lock()
			ag.userLastChosenIndex = userChosenIndex
			ag.userChosenPots[userChosenIndex] = ag.mapPotIndexToPrize[userChosenIndex]
			ag.mutex.Unlock()

			ag.playerObj.ChangeMoneyAndLog(
				-moneyToOpenNextPot, ag.currencyType, false, "",
				consts.ACTION_CHOOSE_GOLD_POT_INDEX, consts.AGCODE_GOLDMINER, "")
			time.Sleep(200 * time.Millisecond)
			ag.playerObj.ChangeMoneyAndLog(
				ag.mapPotIndexToPrize[userChosenIndex], ag.currencyType, false, "",
				consts.ACTION_CHOOSE_GOLD_POT_INDEX, consts.AGCODE_GOLDMINER, "")
			ag.updateAgStatus()
		}
	}
	//
	ag.isFinished = true
	ag.updateAgStatus()
	action := Action{
		ActionName:   consts.ACTION_FINISH_AG,
		ChanResponse: make(chan *ActionResponse),
	}
	ag.ChanActionReceiver <- &action
	<-action.ChanResponse
}

func createAgGoldMinerPots(baseMoney int64) (
	mapPotIndexToPrize map[int]int64, moneyToOpenPots map[int]int64) {
	mapPotIndexToPrize = make(map[int]int64)
	moneyToOpenPots = make(map[int]int64)
	rates := []float64{1, 1, 1, 1, 1.5, 1.5, 2, 2, 5, 5, 10, 10}
	// ave = 3.4
	n := float64(len(rates))
	sum := float64(0)
	for _, rate := range rates {
		sum += rate
	}
	a := sum * 1.1 / (n * (n - 1) / 2)
	z.ShuffleFloat64s(rates)
	for i, rate := range rates {
		mapPotIndexToPrize[i] = int64(rate * float64(baseMoney))
	}
	for i := 0; i < int(n); i++ {
		moneyToOpenPots[i] = int64(float64(i) * a * float64(baseMoney))
		moneyToOpenPots[i] = int64(sum / (n - 1) * float64(baseMoney))
	}
	return
}
