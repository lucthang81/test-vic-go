package congrat_queue

import (
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
	"math/rand"
	"sync"
	"time"
)

type Congrat struct {
	gameCode string
	username string
	value    int64
}

func (congrat *Congrat) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["game_code"] = congrat.gameCode
	data["username"] = congrat.username
	data["value"] = congrat.value
	return data
}

var queue []*Congrat
var currentList []*Congrat
var mutex sync.Mutex

func LoadCongratQueue() {
	queue = make([]*Congrat, 0)
	currentList = make([]*Congrat, 0)
	go scheduledUpdate()
}

func addToQueue(congrat *Congrat) {
	mutex.Lock()
	defer mutex.Unlock()

	queue = append(queue, congrat)
}

func GetCurrentList() []*Congrat {
	mutex.Lock()
	defer mutex.Unlock()
	return currentList
}

func GetQueue() []*Congrat {
	return queue
}

func SerializedQueue(theQueue []*Congrat) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, object := range theQueue {
		results = append(results, object.SerializedData())
	}
	return results
}

func scheduledUpdate() {
	ticker := time.NewTicker(game_config.CongratUpdateTick())
	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			if len(queue) > game_config.CongratQueueMax() {
				offset := len(queue) - game_config.CongratQueueMax()
				queue = queue[offset:]
			}

			currentList = make([]*Congrat, 0)
			if len(queue) > 0 {
				for i := 0; i < utils.MinInt(len(queue), game_config.CongratFetchCount()); i++ {
					indexToGet := rand.Intn(len(queue))
					currentList = append(currentList, queue[indexToGet])
					queue = cutIndexInSlice(queue, indexToGet)
				}
			}

			mutex.Unlock()
		}
	}
}

func AddWinCongrat(username string, gameCode string, amount int64) {
	congrat := &Congrat{
		gameCode: gameCode,
		username: username,
		value:    amount,
	}
	addToQueue(congrat)
}

// helper

func cutIndexInSlice(a []*Congrat, index int) []*Congrat {
	i := index
	j := index + 1
	copy(a[i:], a[j:])
	for k, n := len(a)-j+i, len(a); k < n; k++ {
		a[k] = nil // or the zero value of T
	}
	a = a[:len(a)-j+i]
	return a
}
