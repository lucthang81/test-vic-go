package event_player

import (
	"encoding/json"
	"fmt"
	//	"sort"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	//	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/record"
	//	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/slotacp/slotacpconfig"
	"github.com/vic/vic_go/models/player"
)

const (
	// weekly
	EVENT_COLLECTING_PIECES         = "EVENT_COLLECTING_PIECES"
	EVENT_COLLECTING_PIECES_MONTHLY = "EVENT_COLLECTING_PIECES_MONTHLY"

	EVENT_HOURLY_SLOTACP_1     = "EVENT_HOURLY_SLOTACP_1"
	EVENT_HOURLY_SLOTACP_50    = "EVENT_HOURLY_SLOTACP_50"
	EVENT_HOURLY_SLOTACP_100   = "EVENT_HOURLY_SLOTACP_100"
	EVENT_HOURLY_SLOTACP_250   = "EVENT_HOURLY_SLOTACP_250"
	EVENT_HOURLY_SLOTACP_500   = "EVENT_HOURLY_SLOTACP_500"
	EVENT_HOURLY_SLOTACP_1000  = "EVENT_HOURLY_SLOTACP_1000"
	EVENT_HOURLY_SLOTACP_2500  = "EVENT_HOURLY_SLOTACP_2500"
	EVENT_HOURLY_SLOTACP_5000  = "EVENT_HOURLY_SLOTACP_5000"
	EVENT_HOURLY_SLOTACP_10000 = "EVENT_HOURLY_SLOTACP_10000"
)

// map event name to eventTop object
var MapEvents map[string]*EventCollectingPieces

// for global var MapEvents
var GlobalMutex sync.Mutex

// load events from redis
func LoadSavedEvents() {
	MapEvents = make(map[string]*EventCollectingPieces)
	for _, eventName := range []string{
		EVENT_COLLECTING_PIECES,
		EVENT_COLLECTING_PIECES_MONTHLY,
		EVENT_HOURLY_SLOTACP_100,
		EVENT_HOURLY_SLOTACP_1000,
		EVENT_HOURLY_SLOTACP_10000,
		EVENT_HOURLY_SLOTACP_1,
		EVENT_HOURLY_SLOTACP_250,
		EVENT_HOURLY_SLOTACP_2500,
		EVENT_HOURLY_SLOTACP_50,
		EVENT_HOURLY_SLOTACP_500,
		EVENT_HOURLY_SLOTACP_5000,
	} {
		event := LoadEvent(eventName)
		if event != nil {
			GlobalMutex.Lock()
			MapEvents[eventName] = event
			GlobalMutex.Unlock()
			event.start()
		}
		//		fmt.Println("LoadEventCP", event)
	}
}

func init() {
	LoadSavedEvents()
}

type EventCollectingPieces struct {
	EventName             string
	StartingTime          time.Time
	FinishingTime         time.Time
	nPiecesToComplete     int // >= 2
	NLimitPrizes          int
	TotalPrize            int64
	ChanceToDropRarePiece float64

	// number of users's complete collection,
	// dont give rarePiece when nPrize >= nLimitPrize,
	// rarePiece index = nPiecesToComplete - 1
	nRarePieces int
	// MapPieces = Map piece to numberOfPieces
	MapPidToMapPieces map[int64]map[int]int

	// change to true when create a new event with the same name,
	// or call func event.finish
	IsFinished bool

	Mutex sync.Mutex
}

func (event *EventCollectingPieces) start() {
	go func() {
		time.Sleep(10 * time.Second) // w8 player module load db
		timer := time.After(event.FinishingTime.Sub(time.Now()))
		<-timer
		if event.IsFinished == false {
			event.finish()
		}
	}()
}

// call at the end of func start,
// free redis, free key in global MapEvents
func (event *EventCollectingPieces) finish() {
	// clear
	event.IsFinished = true
	ClearEvent(event.EventName)
	GlobalMutex.Lock()
	delete(MapEvents, event.EventName)
	GlobalMutex.Unlock()
	//
	mapPidToMapPiecesJson, _ := json.Marshal(event.MapPidToMapPieces)
	record.LogEventCollectingPiecesResult(
		event.EventName, event.StartingTime, event.FinishingTime,
		event.nPiecesToComplete, event.NLimitPrizes, event.nRarePieces,
		string(mapPidToMapPiecesJson), false)
}

// always drop a piece
func (event *EventCollectingPieces) GiveAPiece1(pid int64) {
	pObj, _ := player.GetPlayer(pid)
	if pObj == nil {
		fmt.Println("eventcp giveAPiece pObj == nil")
		return
	}
	nCompleteBefore := event.calcNCompleteForPlayer(pid)
	//
	event.Mutex.Lock()
	canDropRarePiece := true
	if event.nRarePieces >= event.NLimitPrizes {
		canDropRarePiece = false
	}
	var choosenPiece int
	if canDropRarePiece {
		if rand.Float64() < event.ChanceToDropRarePiece {
			choosenPiece = event.nPiecesToComplete - 1
		} else {
			choosenPiece = rand.Intn(event.nPiecesToComplete - 1)
		}
	} else {
		choosenPiece = rand.Intn(event.nPiecesToComplete - 1)
	}
	if choosenPiece == event.nPiecesToComplete-1 {
		event.nRarePieces += 1
	}
	_, isIn := event.MapPidToMapPieces[pid]
	if !isIn {
		temp := map[int]int{}
		for i := 0; i < event.nPiecesToComplete; i++ {
			temp[i] = 0
		}
		event.MapPidToMapPieces[pid] = temp
	}
	event.MapPidToMapPieces[pid][choosenPiece] += 1
	nChoosenPiece := event.MapPidToMapPieces[pid][choosenPiece]
	event.Mutex.Unlock()
	//
	Persist(event.EventName, pid, choosenPiece, nChoosenPiece)
	//	fmt.Println("persist cp ", event.EventName, pid, choosenPiece, pObj.PlayerType(), nChoosenPiece)
	if !isIn {
		for i := 0; i < event.nPiecesToComplete; i++ {
			if i != choosenPiece {
				Persist(event.EventName, pid, i, 0)
			}
		}
	}
	if choosenPiece == event.nPiecesToComplete-1 {
		Persist2(event.EventName, event.nRarePieces)
	}
	//
	s := fmt.Sprintf("Bạn đã nhận được mảnh ghép số  %v", choosenPiece+1)
	if !strings.Contains(event.EventName, "SLOTACP") {
		pObj.CreateRawMessage("Bạn trúng mảnh ghép", s)
		pObj.CreatePopUp(s)
	}
	//
	nCompleteAfter := event.calcNCompleteForPlayer(pid)
	if nCompleteAfter > nCompleteBefore {
		if strings.Contains(event.EventName, "SLOTACP") {
			bi := strings.LastIndex(event.EventName, "_") + 1
			moneyPerLineS := event.EventName[bi:len(event.EventName)]
			moneyPerLine, _ := strconv.ParseInt(moneyPerLineS, 10, 64)
			moneyValue := slotacpconfig.MapPicturePrize[moneyPerLine]
			pObj.ChangeMoneyAndLog(
				moneyValue, currency.Money, false, "",
				record.ACTION_SLOTACP_COMPLETE_PICTURE, "", "")
			record.LogMatchRecord2(
				slotacpconfig.SLOTACP_GAME_CODE, currency.Money, 0, 0,
				moneyValue, 0, 0, 0,
				"", map[int64]string{}, []map[string]interface{}{})
		} else {
			pObj.CreatePopUp(
				fmt.Sprintf("Bạn đã ghép thành công %v bức tranh", nCompleteAfter))
		}
	}
}

// have a chance to drop a piece,
// base on charging money amount,
// or money change in roomGame,
func (event *EventCollectingPieces) GiveAPiece(
	pid int64, isChargingMoney bool, isTestMoney bool, amount int64) {
	//
	pObj, _ := player.GetPlayer(pid)
	if pObj == nil {
		return
	} else if pObj.PlayerType() == "bot" {
		return
	}
	//
	if amount < 0 {
		// drop change is decrease when lost money
		amount = -amount / 2
	}
	if !isChargingMoney {
		amount = amount / 20
	}
	if isTestMoney {
		amount = amount / 100
	}
	// amount which will certainly give a piece
	amountCertainlyGiveAPiece := int64(float64(event.TotalPrize)*
		event.ChanceToDropRarePiece) * 5 // TotalPrize = 1/5 input money
	if amountCertainlyGiveAPiece == 0 {
		amountCertainlyGiveAPiece = 50000 // for error
	}
	b := amountCertainlyGiveAPiece
	q := amount / b
	r := amount - q*b
	for i := int64(0); i < q; i++ {
		event.GiveAPiece1(pid)
	}
	if rand.Int63n(b) < r {
		event.GiveAPiece1(pid)
	}
}

// include Lock,
// return number of completed picture for 1 player
func (event *EventCollectingPieces) calcNCompleteForPlayer(pid int64) int {
	result := 2147483647
	event.Mutex.Lock()
	mapPieces, isIn := event.MapPidToMapPieces[pid]
	if !isIn {
		result = 0
	} else {
		for piece := range mapPieces {
			if mapPieces[piece] < result {
				result = mapPieces[piece]
			}
		}
	}
	event.Mutex.Unlock()
	return result
}

// return json string
func (event *EventCollectingPieces) GetPiecesForPlayer(pid int64) string {
	event.Mutex.Lock()
	r, _ := json.Marshal(event.MapPidToMapPieces[pid])
	event.Mutex.Unlock()
	return string(r)
}

//
func (event *EventCollectingPieces) ChangeNLimitPrizes(newValue int) {

}

func (event *EventCollectingPieces) ToMap() map[string]interface{} {
	var EventDisplayName, Description string
	EventDisplayName = event.EventName
	if event.EventName == EVENT_COLLECTING_PIECES {
		Description = `Nhiệm vụ của các người chơi là sẽ thu thập các mảnh từ phần nạp tiền và chơi game để ghép thành một bức tranh hoàn chỉnh.

Các nguồn nhận mảnh ghép tranh:
- Với mỗi thẻ là bội số của 10.000 bạn có cơ hội nhận được một mảnh tranh (Nạp thẻ 10.000 sẽ có cơ hội nhận được một mảnh tranh nhưng nạp một thẻ 20.000 bạn cũng chỉ có cơ hội nhận được một mảnh tranh).
- Với mỗi thẻ là bội số của 50.000 bạn chắc chắn nhận được số mảnh tranh tương ứng (Nạp một thẻ 50.000 bạn chắc chắn có một mảnh tranh và nạp một thẻ 100.000 bạn chắc chắn có hai mảnh tranh).
- Ngoài ra khi bạn chơi game bạn cũng có cơ hội nhận được các mảnh tranh.

Giải thưởng: Ba người chơi ghép được hoàn chỉnh bức tranh đầu tiên trong tuần sẽ nhận được giải thưởng trị giá 1 triệu kim.`
		EventDisplayName = "Sự kiện ghép tranh tuần"
	} else if event.EventName == EVENT_COLLECTING_PIECES_MONTHLY {
		Description = "Reset sau mỗi tháng"
		EventDisplayName = "Sự kiện ghép tranh tháng"
	}
	return map[string]interface{}{
		"EventName":         event.EventName,
		"EventDisplayName":  EventDisplayName,
		"Description":       Description,
		"StartingTime":      event.StartingTime,
		"FinishingTime":     event.FinishingTime,
		"nPiecesToComplete": event.nPiecesToComplete,
		"NLimitPrizes":      event.NLimitPrizes,
		"nRarePieces":       event.nRarePieces,
	}
}

//
func NewEventCollectingPieces(
	eventName string, startingTime time.Time, finishingTime time.Time,
	nPiecesToComplete int, nLimitPrize int,
	TotalPrize int64, ChanceToDropRarePiece float64,
) *EventCollectingPieces {
	//
	GlobalMutex.Lock()
	oldEvent := MapEvents[eventName]
	GlobalMutex.Unlock()
	if oldEvent != nil {
		oldEvent.IsFinished = true
	}
	ClearEvent(eventName)
	//
	event := &EventCollectingPieces{
		EventName:             eventName,
		StartingTime:          startingTime,
		FinishingTime:         finishingTime,
		nPiecesToComplete:     nPiecesToComplete,
		NLimitPrizes:          nLimitPrize,
		MapPidToMapPieces:     map[int64]map[int]int{},
		TotalPrize:            TotalPrize,
		ChanceToDropRarePiece: ChanceToDropRarePiece,
	}
	GlobalMutex.Lock()
	MapEvents[eventName] = event
	GlobalMutex.Unlock()
	SaveEvent(event)
	event.start()
	fmt.Println(time.Now(), "MapEventsCollectingPieces", MapEvents)
	return event
}
