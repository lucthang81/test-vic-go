package event

import (
	//	"encoding/json"
	"fmt"
	//	"sort"
	"sync"
	"time"
)

const (
	EVENTSC_MAUBINH_BAIDEP = "săn bài đẹp mậu binh"
	EVENTSC_TIENLEN_BAIDEP = "săn bài đẹp tiến lên"
	EVENTSC_BACAY_BAIDEP   = "săn bài đẹp ba cây"

	ACTION_BONUS_EVENT = "ACTION_BONUS_EVENT"

	TIME_UNIT_24H = 24 * time.Hour
)

// map event name to eventSC object
var MapEventSCs map[string]*EventSC

// load events from redis
func LoadSavedEvents2() {
	MapEventSCs = make(map[string]*EventSC)
	for _, eventName := range []string{
		EVENTSC_MAUBINH_BAIDEP,
		EVENTSC_TIENLEN_BAIDEP,
		EVENTSC_BACAY_BAIDEP,
	} {
		event := LoadEventSC(eventName)
		if event != nil {
			GlobalMutex.Lock()
			MapEventSCs[eventName] = event
			GlobalMutex.Unlock()
			event.start()
		}
	}
}

func init() {
	//
	fmt.Print("")
	//
	LoadSavedEvents2()
}

// just for track nBonus, give prize in game
type EventSC struct {
	EventName     string
	StartingTime  time.Time
	FinishingTime time.Time
	//
	TimeUnit time.Duration
	// limit number of bonus 1 user can get per TimeUnit
	LimitNOBonus int64

	//
	MapPlayerIdToValue map[int64]int64

	// change to true when create a new event with the same name,
	// or call func event.finish
	IsFinished bool

	Mutex sync.Mutex
}

//
func (event *EventSC) start() {
	go loopResetNOBonus(event.EventName)
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
// pay prize,
// record prize info to database,
// free redis, free key in global MapEvents2,
func (event *EventSC) finish() {
	// clear
	event.IsFinished = true
	ClearEventSC(event.EventName)
	GlobalMutex.Lock()
	delete(MapEventSCs, event.EventName)
	GlobalMutex.Unlock()
}

// include mutex lock and persist,
// change the currentValue,
// lead to change the bestValue
func (event *EventSC) SetNewValue(pid int64, newValue int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToValue[pid] = newValue
	temp1 := event.MapPlayerIdToValue[pid]
	event.Mutex.Unlock()
	PersistSC(event.EventName, pid, temp1)
	return nil
}

//
func (event *EventSC) ChangeValue(pid int64, amount int64) error {
	event.Mutex.Lock()
	event.MapPlayerIdToValue[pid] += amount
	temp1 := event.MapPlayerIdToValue[pid]
	event.Mutex.Unlock()
	PersistSC(event.EventName, pid, temp1)
	return nil
}

func (event *EventSC) ToMap() map[string]interface{} {
	var EventDisplayName, Description, GameCode string
	EventDisplayName = event.EventName
	if event.EventName == EVENTSC_BACAY_BAIDEP {
		Description = "Event's description"
		GameCode = "bacay2"
	} else if event.EventName == EVENTSC_MAUBINH_BAIDEP {
		Description = "Event's description"
		GameCode = "maubinh"
	} else if event.EventName == EVENTSC_TIENLEN_BAIDEP {
		Description = "Event's description"
		GameCode = "tienlen"
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

func loopResetNOBonus(eventName string) {
	time.Sleep(10 * time.Second)
	//
	var remainingT time.Duration
	GlobalMutex.Lock()
	event := MapEventSCs[eventName]
	if event != nil {
		temp := time.Now().Sub(event.StartingTime) % event.TimeUnit
		remainingT = event.TimeUnit - temp
	}
	GlobalMutex.Unlock()
	time.Sleep(remainingT)
	//
	for {
		GlobalMutex.Lock()
		event := MapEventSCs[eventName]
		GlobalMutex.Unlock()
		if event != nil {
			event.Mutex.Lock()
			for pid, _ := range event.MapPlayerIdToValue {
				newValue := int64(0)
				event.MapPlayerIdToValue[pid] = newValue
				PersistSC(event.EventName, pid, newValue)
			}
			event.Mutex.Unlock()
			time.Sleep(event.TimeUnit)
		} else {
			break
		}
	}
}

//
func NewEventSC(
	eventName string,
	startingTime time.Time, finishingTime time.Time,
	limitNOBonus int64, timeUnit time.Duration,
) *EventSC {
	//
	GlobalMutex.Lock()
	oldEvent := MapEventSCs[eventName]
	GlobalMutex.Unlock()
	if oldEvent != nil {
		oldEvent.IsFinished = true
	}
	ClearEventSC(eventName)
	//
	event := &EventSC{
		EventName:          eventName,
		StartingTime:       startingTime,
		FinishingTime:      finishingTime,
		LimitNOBonus:       limitNOBonus,
		TimeUnit:           timeUnit,
		MapPlayerIdToValue: make(map[int64]int64),
	}
	GlobalMutex.Lock()
	MapEventSCs[eventName] = event
	GlobalMutex.Unlock()
	SaveEventSC(event)
	event.start()
	return event
}
