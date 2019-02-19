package utils

import (
	"sync"
)

type Int64SliceMapInterfaceMap struct {
	coreMap map[int64][]map[string]interface{}
	mutex   sync.RWMutex
}

func NewInt64SliceMapInterfaceMap() *Int64SliceMapInterfaceMap {
	return &Int64SliceMapInterfaceMap{
		coreMap: make(map[int64][]map[string]interface{}),
	}
}

func (mapObject *Int64SliceMapInterfaceMap) Set(key int64, value []map[string]interface{}) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64SliceMapInterfaceMap) Get(key int64) []map[string]interface{} {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64SliceMapInterfaceMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64SliceMapInterfaceMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64SliceMapInterfaceMap) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64SliceMapInterfaceMap) Copy() map[int64][]map[string]interface{} {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64][]map[string]interface{})
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

// =========================================

type Int64Int64Map struct {
	coreMap map[int64]int64
	mutex   sync.RWMutex
}

func NewInt64Int64Map() *Int64Int64Map {
	return &Int64Int64Map{
		coreMap: make(map[int64]int64),
	}
}

func (mapObject *Int64Int64Map) Set(key int64, value int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64Int64Map) Get(key int64) int64 {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64Int64Map) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64Int64Map) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64Int64Map) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64Int64Map) Copy() map[int64]int64 {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64]int64)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *Int64Int64Map) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *Int64Int64Map) ContainValueForKey(key int64) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
type Int64TimeOutMap struct {
	coreMap map[int64]*TimeOut
	mutex   sync.RWMutex
}

func NewInt64TimeOutMap() *Int64TimeOutMap {
	return &Int64TimeOutMap{
		coreMap: make(map[int64]*TimeOut),
	}
}

func (mapObject *Int64TimeOutMap) Set(key int64, value *TimeOut) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *Int64TimeOutMap) Get(key int64) *TimeOut {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *Int64TimeOutMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *Int64TimeOutMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *Int64TimeOutMap) Delete(key int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *Int64TimeOutMap) Copy() map[int64]*TimeOut {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[int64]*TimeOut)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *Int64TimeOutMap) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *Int64TimeOutMap) ContainValueForKey(key int64) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================

type StringInt64Map struct {
	coreMap map[string]int64
	mutex   sync.RWMutex
}

func NewStringInt64Map() *StringInt64Map {
	return &StringInt64Map{
		coreMap: make(map[string]int64),
	}
}

func (mapObject *StringInt64Map) Set(key string, value int64) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *StringInt64Map) Get(key string) int64 {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *StringInt64Map) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *StringInt64Map) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *StringInt64Map) Delete(key string) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *StringInt64Map) Copy() map[string]int64 {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[string]int64)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *StringInt64Map) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *StringInt64Map) ContainValueForKey(key string) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
