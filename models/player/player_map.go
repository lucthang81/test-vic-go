package player

import (
	"sync"
)

type Int64PlayerMap struct {
	coreMap map[int64]*Player
	mutex   sync.RWMutex
}

func NewInt64PlayerMap() *Int64PlayerMap {
	return &Int64PlayerMap{
		coreMap: make(map[int64]*Player),
	}
}

func (mapObject *Int64PlayerMap) Set(key int64, value *Player) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64PlayerMap) Get(key int64) *Player {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64PlayerMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64PlayerMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64PlayerMap) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64PlayerMap) Copy() map[int64]*Player {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64]*Player)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *Int64PlayerMap) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *Int64PlayerMap) ContainValueForKey(key int64) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
