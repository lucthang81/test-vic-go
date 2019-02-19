package game

import (
	"sync"
)

type IntGamePlayerMap struct {
	coreMap map[int]GamePlayer
	mutex   sync.RWMutex
}

func NewIntGamePlayerMap() *IntGamePlayerMap {
	return &IntGamePlayerMap{
		coreMap: make(map[int]GamePlayer),
	}
}

func (mapObject *IntGamePlayerMap) Set(key int, value GamePlayer) {
	mapObject.set(key, value)

}

func (mapObject *IntGamePlayerMap) Get(key int) GamePlayer {
	return mapObject.get(key)
}

func (mapObject *IntGamePlayerMap) RLock() {
	mapObject.rLock()
}

func (mapObject *IntGamePlayerMap) RUnlock() {
	mapObject.rUnlock()
}

func (mapObject *IntGamePlayerMap) Delete(key int) {
	mapObject.delete(key)
}

func (mapObject *IntGamePlayerMap) Copy() map[int]GamePlayer {
	return mapObject.copy()
}

func (mapObject *IntGamePlayerMap) set(key int, value GamePlayer) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *IntGamePlayerMap) get(key int) GamePlayer {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *IntGamePlayerMap) rLock() {
	mapObject.mutex.RLock()
}

func (mapObject *IntGamePlayerMap) rUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *IntGamePlayerMap) delete(key int) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *IntGamePlayerMap) copy() map[int]GamePlayer {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int]GamePlayer)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *IntGamePlayerMap) Len() int {
	return len(mapObject.coreMap)
}

// =========================================

type Int64GamePlayerMap struct {
	coreMap map[int64]GamePlayer
	mutex   sync.RWMutex
}

func NewInt64GamePlayerMap() *Int64GamePlayerMap {
	return &Int64GamePlayerMap{
		coreMap: make(map[int64]GamePlayer),
	}
}

func (mapObject *Int64GamePlayerMap) Set(key int64, value GamePlayer) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64GamePlayerMap) Get(key int64) GamePlayer {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64GamePlayerMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64GamePlayerMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64GamePlayerMap) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64GamePlayerMap) Copy() map[int64]GamePlayer {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64]GamePlayer)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *Int64GamePlayerMap) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *Int64GamePlayerMap) ContainValueForKey(key int64) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
type Int64RoomMap struct {
	coreMap map[int64]*Room
	mutex   sync.RWMutex
}

func (m *Int64RoomMap) CoreMap() map[int64]*Room {
	return m.coreMap
}

func NewInt64RoomMap() *Int64RoomMap {
	return &Int64RoomMap{
		coreMap: make(map[int64]*Room),
	}
}

func (mapObject *Int64RoomMap) Set(key int64, value *Room) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64RoomMap) Get(key int64) *Room {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64RoomMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64RoomMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64RoomMap) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64RoomMap) Copy() map[int64]*Room {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64]*Room)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *Int64RoomMap) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *Int64RoomMap) ContainValueForKey(key int64) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
