package event

// cant call package models/player, player.money use this package
// pay prize in periodic_tasks

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/zconfig"
)

const (
	EVENT_OANTUTI_WINNING_STREAK = "chuỗi thắng Oẳn tù tì"
	EVENT_OANTUTI_LOSING_STREAK  = "chuỗi thua Oẳn tù tì"
	EVENT_TAIXIU_WINNING_STREAK  = "chuỗi thắng Tài xỉu"
	EVENT_SLOTBACAY_TEN_POINT    = "điểm 10 Mini ba cây"

	EVENT_CHARGING_MONEY     = "EVENT_CHARGING_MONEY"
	EVENT_EARNING_MONEY      = "EVENT_EARNING_MONEY"
	EVENT_EARNING_TEST_MONEY = "EVENT_EARNING_TEST_MONEY"

	NORMAL_TRACK_TAIXIU_EARNING_MONEY   = "NORMAL_TRACK_TAIXIU_EARNING_MONEY"
	NORMAL_TRACK_OANTUTI_WINNING_STREAK = "NORMAL_TRACK_OANTUTI_WINNING_STREAK"
	NORMAL_TRACK_OANTUTI_LOSING_STREAK  = "NORMAL_TRACK_OANTUTI_LOSING_STREAK"

	ACTION_FINISH_EVENT = "ACTION_FINISH_EVENT"
)

// map event name to eventTop object
var MapEvents map[string]*EventTop

// for global var MapEvents
var GlobalMutex sync.Mutex

// load events from redis
func LoadSavedEvents() {
	MapEvents = make(map[string]*EventTop)
	for _, eventName := range []string{
		NORMAL_TRACK_OANTUTI_WINNING_STREAK,
		NORMAL_TRACK_OANTUTI_LOSING_STREAK,
		NORMAL_TRACK_TAIXIU_EARNING_MONEY,
		EVENT_OANTUTI_WINNING_STREAK,
		EVENT_OANTUTI_LOSING_STREAK,
		EVENT_TAIXIU_WINNING_STREAK,
		EVENT_SLOTBACAY_TEN_POINT,

		EVENT_EARNING_TEST_MONEY,
		EVENT_EARNING_MONEY,
		EVENT_CHARGING_MONEY,
	} {
		if zconfig.ServerVersion == zconfig.SV_02 &&
			eventName == EVENT_EARNING_TEST_MONEY {
			continue
		}

		event := LoadEvent(eventName)
		if strings.Contains(eventName, "NORMAL_TRACK_") &&
			(event == nil) {
			event = NewEventTop(
				eventName,
				time.Now(),
				time.Now().Add(240000*time.Hour),
				map[int]int64{},
			)
		}
		if event != nil {
			GlobalMutex.Lock()
			MapEvents[eventName] = event
			GlobalMutex.Unlock()
			event.start()
		}
	}
}

func init() {
	//
	fmt.Print("")
	//
	LoadSavedEvents()
}

// help funcs for sort
type TopRow struct {
	PlayerId    int64
	CreatedTime time.Time
	Value       int64
}

// order first by value, second by createdTime
type OrderByValueDesc []TopRow

func (a OrderByValueDesc) Len() int {
	return len(a)
}
func (a OrderByValueDesc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a OrderByValueDesc) Less(i, j int) bool {
	if a[i].Value > a[j].Value {
		return true
	} else if a[i].Value == a[j].Value {
		return a[i].CreatedTime.Sub(a[j].CreatedTime) < 0
	} else {
		return false
	}
}

//
type EventTop struct {
	EventName          string
	StartingTime       time.Time
	FinishingTime      time.Time
	MapPositionToPrize map[int]int64

	// bigger value is better
	MapPlayerIdToValue map[int64]int64
	// sooner is better
	MapPlayerIdToCreatedTime map[int64]time.Time
	// map MapPlayerIdToValue is the best value,
	// this map is current value, pos for player base on best value
	MapPlayerIdToCurrentValue map[int64]int64

	// update each 1 minute,
	// json TopRows
	LeaderBoard string
	// for get position
	SortedValues []int64
	//
	FullOrder []TopRow

	// change to true when create a new event with the same name,
	// or call func event.finish
	IsFinished bool

	Mutex sync.Mutex
}

//
func (event *EventTop) start() {
	go loopUpdateLeaderBoard(event.EventName)
	go func() {
		time.Sleep(10 * time.Second) // w8 player module load db
		timer := time.After(event.FinishingTime.Sub(time.Now()))
		<-timer
		if event.IsFinished == false {
			event.finish()
		}
	}()
}

//
func (event *EventTop) updateLeaderBoard() {
	a := make([]TopRow, 0)
	event.Mutex.Lock()
	for pid, value := range event.MapPlayerIdToValue {
		a = append(a,
			TopRow{
				PlayerId:    pid,
				Value:       value,
				CreatedTime: event.MapPlayerIdToCreatedTime[pid],
			},
		)
	}
	event.Mutex.Unlock()
	sort.Sort(OrderByValueDesc(a))
	//
	sortedValues := []int64{}
	for _, v := range a {
		sortedValues = append(sortedValues, v.Value)
	}
	event.SortedValues = sortedValues
	//
	event.FullOrder = a
	//
	var b []TopRow
	if len(a) <= 30 {
		b = a[:len(a)]
	} else {
		b = a[:30]
	}
	bytes, _ := json.Marshal(b)
	event.LeaderBoard = string(bytes)
}

//
func loopUpdateLeaderBoard(eventName string) {
	time.Sleep(10 * time.Second) // w8 player module load db
	for {
		GlobalMutex.Lock()
		event := MapEvents[eventName]
		GlobalMutex.Unlock()
		if event != nil {
			event.updateLeaderBoard()
			time.Sleep(60 * time.Second)
		} else {
			break
		}
	}
}

// return json top30
func (event *EventTop) GetLeaderBoard() string {
	return event.LeaderBoard
}

// return position, value for player
func (event *EventTop) GetPosAndValue(pid int64) (int, int64) {
	if event.SortedValues == nil {
		return 0, 0
	} else {
		event.Mutex.Lock()
		pValue := event.MapPlayerIdToValue[pid]
		event.Mutex.Unlock()
		pos, err := z.BisectRight(event.SortedValues, pValue)
		if err != nil {
			fmt.Println("ERROR EventTop GetPosition", err)
		}
		return pos, pValue
	}
}

// include mutex lock and persist,
// change the currentValue,
// can lead to change the bestValue
func (event *EventTop) SetNewValue(pid int64, newValue int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToCurrentValue[pid] = newValue
	if event.MapPlayerIdToCurrentValue[pid] > event.MapPlayerIdToValue[pid] {
		event.MapPlayerIdToValue[pid] = event.MapPlayerIdToCurrentValue[pid]
		event.MapPlayerIdToCreatedTime[pid] = time.Now()
	}
	temp1 := event.MapPlayerIdToValue[pid]
	temp2 := event.MapPlayerIdToCreatedTime[pid]
	temp3 := event.MapPlayerIdToCurrentValue[pid]
	event.Mutex.Unlock()
	Persist(event.EventName, pid, temp1, temp2, temp3)
	return nil
}

// include mutex lock and persist,
// add amount to the currentValue,
// can lead to change the bestValue
func (event *EventTop) ChangeValue(pid int64, amount int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToCurrentValue[pid] += amount
	if event.MapPlayerIdToCurrentValue[pid] > event.MapPlayerIdToValue[pid] {
		event.MapPlayerIdToValue[pid] = event.MapPlayerIdToCurrentValue[pid]
		event.MapPlayerIdToCreatedTime[pid] = time.Now()
	}
	temp1 := event.MapPlayerIdToValue[pid]
	temp2 := event.MapPlayerIdToCreatedTime[pid]
	temp3 := event.MapPlayerIdToCurrentValue[pid]
	event.Mutex.Unlock()
	Persist(event.EventName, pid, temp1, temp2, temp3)
	return nil
}

// include mutex lock and persist,
// change the bestValue,
// dont care CurrentValue
func (event *EventTop) SetNewBestValue(pid int64, newValue int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToValue[pid] = newValue
	event.MapPlayerIdToCreatedTime[pid] = time.Now()
	temp1 := event.MapPlayerIdToValue[pid]
	temp2 := event.MapPlayerIdToCreatedTime[pid]
	temp3 := int64(0)
	event.Mutex.Unlock()
	Persist(event.EventName, pid, temp1, temp2, temp3)
	return nil
}

// include mutex lock and persist,
// add amount to the BestValue,
// dont care CurrentValue
func (event *EventTop) ChangeBestValue(pid int64, amount int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToValue[pid] += amount
	event.MapPlayerIdToCreatedTime[pid] = time.Now()
	temp1 := event.MapPlayerIdToValue[pid]
	temp2 := event.MapPlayerIdToCreatedTime[pid]
	temp3 := int64(0)
	event.Mutex.Unlock()
	Persist(event.EventName, pid, temp1, temp2, temp3)
	return nil
}

// call at the end of func start,
// free redis, free key in global MapEvents to stop loopUpdateLeaderBoard,
// save FullOrder to table event_top_result
func (event *EventTop) finish() {
	// clear
	event.IsFinished = true
	ClearEvent(event.EventName)
	GlobalMutex.Lock()
	delete(MapEvents, event.EventName)
	GlobalMutex.Unlock()
	event.updateLeaderBoard()
	//
	mapPositionToPrizeJson, _ := json.Marshal(event.MapPositionToPrize)
	fullOrderJson, _ := json.Marshal(event.FullOrder)
	record.LogEventTopResult(
		event.EventName, event.StartingTime, event.FinishingTime,
		string(mapPositionToPrizeJson), string(fullOrderJson), false)
}

func (event *EventTop) ToMap() map[string]interface{} {
	var EventDisplayName, Description, GameCode string
	EventDisplayName = event.EventName
	if event.EventName == EVENT_EARNING_MONEY {
		EventDisplayName = "Top thắng kim"
		Description = "Event's description"
	} else if event.EventName == EVENT_EARNING_TEST_MONEY {
		EventDisplayName = "Top thắng xu"
		Description = `Đua top xu

Người chơi tích lũy xu từ các việc chơi game và nạp tiền để đua top
Phần thưởng:
Hạng 1: 500.000 Kim
Hạng 2: 200.000 Kim
Hạng 3: 100.000 Kim
Ngoài ra khi bạn xếp hạng từ 1 đến 10 sẽ nhận ngẫu nhiên một con số từ 00 đến 99. Người nào trùng với hai số cuối của giải đặc biệt xổ số miền bắc phiên ngày thứ 2 của tuần tiếp theo sẽ trúng thưởng 1.000.000 Kim`
	} else if event.EventName == EVENT_CHARGING_MONEY {
		EventDisplayName = "Top nạp tiền"
		Description = "Event's description"
	} else if event.EventName == EVENT_OANTUTI_WINNING_STREAK {
		Description = "Event's description"
		GameCode = "oantuti"
	} else if event.EventName == EVENT_OANTUTI_LOSING_STREAK {
		Description = "Event's description"
		GameCode = "oantuti"
	} else if event.EventName == EVENT_SLOTBACAY_TEN_POINT {
		Description = "Event's description"
		GameCode = "slotbacay"
	} else if event.EventName == EVENT_TAIXIU_WINNING_STREAK {
		Description = "Event's description"
		GameCode = "taixiu"
	}
	return map[string]interface{}{
		"EventName":        event.EventName,
		"EventDisplayName": EventDisplayName,
		"Description":      Description,
		"GameCode":         GameCode,
		"StartingTime":     event.StartingTime,
		"FinishingTime":    event.FinishingTime,
	}
}

//
func NewEventTop(
	eventName string,
	startingTime time.Time, finishingTime time.Time,
	mapPositionToPrize map[int]int64,
) *EventTop {
	//
	GlobalMutex.Lock()
	oldEvent := MapEvents[eventName]
	GlobalMutex.Unlock()
	if oldEvent != nil {
		oldEvent.IsFinished = true
	}
	ClearEvent(eventName)
	//
	event := &EventTop{
		EventName:                 eventName,
		StartingTime:              startingTime,
		FinishingTime:             finishingTime,
		MapPositionToPrize:        mapPositionToPrize,
		MapPlayerIdToValue:        make(map[int64]int64),
		MapPlayerIdToCreatedTime:  make(map[int64]time.Time),
		MapPlayerIdToCurrentValue: make(map[int64]int64),
	}
	GlobalMutex.Lock()
	MapEvents[eventName] = event
	GlobalMutex.Unlock()
	SaveEvent(event)
	event.start()
	return event
}
