package player

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"sort"
)

/*
one time claim event
event valid in specific time range
event

*/

const EventTypeOneTime string = "one_time"
const EventTypeTimeRange string = "time_range"

type Event struct {
	id          int64
	eventType   string
	priority    int
	title       string
	description string

	tipTitle       string
	tipDescription string

	iconUrl string
	data    map[string]interface{}
}

const EventCacheKey string = "event"
const EventDatabaseTableName string = "event"
const EventClassName string = "Event"

func (event *Event) CacheKey() string {
	return EventCacheKey
}

func (event *Event) DatabaseTableName() string {
	return EventDatabaseTableName
}

func (event *Event) ClassName() string {
	return EventClassName
}

func (event *Event) Id() int64 {
	return event.id
}

func (event *Event) SetId(id int64) {
	event.id = id
}

func (event *Event) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = event.Id()
	data["event_type"] = event.eventType
	data["data"] = event.data
	data["priority"] = event.priority
	data["title"] = event.title
	data["description"] = event.description
	data["icon_url"] = event.iconUrl
	data["tip_title"] = event.tipTitle
	data["tip_description"] = event.tipDescription
	return data
}

type EventManager struct {
	events []*Event
}

func NewEventManager() (manager *EventManager) {
	manager = &EventManager{
		events: make([]*Event, 0),
	}
	err := manager.fetchData()
	if err != nil {
		log.LogSerious("Fetch event error %v", err)
	}
	return manager
}

func (manager *EventManager) fetchData() (err error) {
	// get event
	queryString := fmt.Sprintf("SELECT id, event_type, data, priority, title, description,tip_title, tip_description, icon_url FROM %s", EventDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		return err
	}

	manager.events = make([]*Event, 0)
	for rows.Next() {
		var id int64
		var eventType string
		var dataString []byte
		var title []byte
		var description []byte
		var tipTitle []byte
		var tipDescription []byte
		var iconUrl []byte
		var priority int

		err = rows.Scan(&id, &eventType, &dataString, &priority, &title, &description, &tipTitle, &tipDescription, &iconUrl)
		if err != nil {
			rows.Close()
			return err
		}
		var data map[string]interface{}
		err := json.Unmarshal(dataString, &data)
		if err != nil {
			rows.Close()
			return err
		}
		event := &Event{}
		event.id = id
		event.eventType = eventType
		event.data = data
		event.priority = priority
		event.title = string(title)
		event.description = string(description)
		event.tipTitle = string(tipTitle)
		event.tipDescription = string(tipDescription)
		event.iconUrl = string(iconUrl)

		manager.events = append(manager.events, event)
	}
	rows.Close()

	return nil
}

type ByPriority []map[string]interface{}

func (a ByPriority) Len() int      { return len(a) }
func (a ByPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool {
	data1 := a[i]
	data2 := a[j]
	code1 := utils.GetIntAtPath(data1, "priority")
	code2 := utils.GetIntAtPath(data2, "priority")
	return code1 < code2
}

func GetEventManager() *EventManager {
	return eventManager
}

func getEventsData() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, event := range eventManager.events {
		results = append(results, event.SerializedData())
	}
	sort.Sort(ByPriority(results))
	return results
}

func createEvent(priority int,
	eventType string,
	title string,
	description string,
	tipTitle string,
	tipDescription string,
	iconUrl string,
	data map[string]interface{}) (err error) {
	dataString, _ := json.Marshal(data)
	queryString := fmt.Sprintf("INSERT INTO %s (priority, event_type, title, description,tip_title,tip_description, icon_url, data) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)", EventDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, priority, eventType, title, description, tipTitle, tipDescription, iconUrl, dataString)
	if err == nil {
		eventManager.fetchData()
	}
	return err
}

func getEventData(id int64) map[string]interface{} {
	for _, event := range eventManager.events {
		if event.Id() == id {
			return event.SerializedData()
		}
	}
	return nil
}

func editEvent(id int64,
	priority int,
	eventType string,
	title string,
	description string,
	tipTitle string,
	tipDescription string,
	iconUrl string,
	data map[string]interface{}) (err error) {
	dataString, _ := json.Marshal(data)
	queryString := fmt.Sprintf("UPDATE %s SET priority = $1, event_type = $2, title = $3, description = $4,tip_title = $5, tip_description = $6, icon_url = $7, data = $8 WHERE id = $9", EventDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, priority, eventType, title, description, tipTitle, tipDescription, iconUrl, dataString, id)
	if err == nil {
		eventManager.fetchData()
	}
	return err
}

func deleteEvent(id int64) (err error) {
	queryPrizeString := fmt.Sprintf("DELETE FROM %s WHERE id = $1", EventDatabaseTableName)
	fmt.Println(queryPrizeString, id)
	_, err = dataCenter.Db().Exec(queryPrizeString, id)
	if err == nil {
		eventManager.fetchData()
	}
	return err
}
