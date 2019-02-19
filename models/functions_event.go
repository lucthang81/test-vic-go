package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	top "github.com/vic/vic_go/models/event"
	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

func getEventList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["results"] = player.GetEventsData()
	return responseData, nil
}

func EventTopGetLeaderBoard(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	eventName := utils.GetStringAtPath(data, "eventName")
	top.GlobalMutex.Lock()
	event := top.MapEvents[eventName]
	top.GlobalMutex.Unlock()
	if event != nil {
		rows := make([]map[string]interface{}, 0)
		for i, e := range event.FullOrder {
			pObj, _ := player.GetPlayer2(e.PlayerId)
			if i <= 30 {
				var row map[string]interface{}
				if pObj != nil {
					row = map[string]interface{}{
						"PlayerId":    e.PlayerId,
						"PlayerName":  pObj.DisplayName(),
						"Value":       e.Value,
						"CreatedTime": e.CreatedTime,
					}
				} else {
					fakeName := player.BotUsernames[e.PlayerId%int64(len(player.BotUsernames))]
					row = map[string]interface{}{
						"PlayerId":    e.PlayerId,
						"PlayerName":  fakeName,
						"Value":       e.Value,
						"CreatedTime": e.CreatedTime,
					}
				}
				rows = append(rows, row)
			}
		}
		responseData := map[string]interface{}{
			"eventName":   eventName,
			"leaderBoard": rows,
		}
		return responseData, nil
	} else {
		return nil, errors.New("event == nil ")
	}
}

func EventTopGetPosAndValue(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	eventName := utils.GetStringAtPath(data, "eventName")

	top.GlobalMutex.Lock()
	event := top.MapEvents[eventName]
	top.GlobalMutex.Unlock()
	if event != nil {
		p, v := event.GetPosAndValue(playerId)
		responseData := map[string]interface{}{
			"eventName": eventName,
			"position":  p,
			"value":     v,
		}
		return responseData, nil
	} else {
		return nil, errors.New("event == nil ")
	}
}

func EventGetList(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	eventTopNames := []map[string]interface{}{}
	for name, event := range top.MapEvents {
		if !strings.Contains(name, "NORMAL_TRACK_") &&
			name != top.EVENT_EARNING_MONEY &&
			name != top.EVENT_CHARGING_MONEY {
			eventTopNames = append(eventTopNames, event.ToMap())
		}
	}
	eventSCNames := []map[string]interface{}{}
	for _, event := range top.MapEventSCs {
		eventSCNames = append(eventSCNames, event.ToMap())
	}
	EventCPNames := []map[string]interface{}{}
	for name, event := range event_player.MapEvents {
		if !strings.Contains(name, "SLOTACP") {
			EventCPNames = append(EventCPNames, event.ToMap())
		}
	}
	return map[string]interface{}{
		"EventTops": eventTopNames,
		"EventSCs":  eventSCNames,
		"EventCPs":  EventCPNames,
	}, nil
}

func EventCollectingPiecesInfo(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	eventName := utils.GetStringAtPath(data, "eventName")
	if eventName == "" {
		eventName = event_player.EVENT_COLLECTING_PIECES
	}

	event_player.GlobalMutex.Lock()
	event := event_player.MapEvents[eventName]
	event_player.GlobalMutex.Unlock()
	if event != nil {
		iJson := event.GetPiecesForPlayer(playerId)
		var mapPieces map[string]interface{}
		err := json.Unmarshal([]byte(iJson), &mapPieces)
		if err != nil {
			return nil, err
		} else {
			return map[string]interface{}{
				"mapPieces": mapPieces,
				"event":     event.ToMap(),
				"remainingDurationInSeconds": event.FinishingTime.Sub(time.Now()).Seconds(),
			}, nil
		}
	} else {
		return nil, errors.New("event == nil ")
	}
}

func EventCollectingPiecesMonthlyInfo(
	models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	eventName := event_player.EVENT_COLLECTING_PIECES_MONTHLY

	event_player.GlobalMutex.Lock()
	event := event_player.MapEvents[eventName]
	event_player.GlobalMutex.Unlock()
	if event != nil {
		iJson := event.GetPiecesForPlayer(playerId)
		var responseData map[string]interface{}
		err := json.Unmarshal([]byte(iJson), &responseData)
		if err != nil {
			return nil, err
		} else {
			return responseData, nil
		}
	} else {
		return nil, errors.New("event == nil ")
	}
}
