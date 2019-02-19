package server

import (
	"sync"
)

type ConnMap struct {
	coreMap map[int64]*Connection
	mutex   sync.RWMutex
}

func NewConnMap() *ConnMap {
	return &ConnMap{
		coreMap: make(map[int64]*Connection),
	}
}

func (mapObject *ConnMap) set(key int64, value *Connection) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *ConnMap) get(key int64) *Connection {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *ConnMap) rLock() {
	mapObject.mutex.RLock()
}

func (mapObject *ConnMap) rUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *ConnMap) delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

type PendingAuthConnMap struct {
	coreMap map[*Connection]bool
	mutex   sync.RWMutex
}

func NewPendingAuthConnMap() *PendingAuthConnMap {
	return &PendingAuthConnMap{
		coreMap: make(map[*Connection]bool),
	}
}

func (mapObject *PendingAuthConnMap) set(key *Connection, value bool) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *PendingAuthConnMap) get(key *Connection) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *PendingAuthConnMap) rLock() {
	mapObject.mutex.RLock()
}

func (mapObject *PendingAuthConnMap) rUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *PendingAuthConnMap) delete(key *Connection) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}
