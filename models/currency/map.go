package currency

import (
	"sync"
)

type StringCurrencyMap struct {
	coreMap map[string]*Currency
	mutex   sync.RWMutex
}

func NewStringCurrencyMap() *StringCurrencyMap {
	return &StringCurrencyMap{
		coreMap: make(map[string]*Currency),
	}
}

func (mapObject *StringCurrencyMap) set(key string, value *Currency) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value

}

func (mapObject *StringCurrencyMap) get(key string) *Currency {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *StringCurrencyMap) rLock() {
	mapObject.mutex.RLock()
}

func (mapObject *StringCurrencyMap) rUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *StringCurrencyMap) delete(key string) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *StringCurrencyMap) copy() map[string]*Currency {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[string]*Currency)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

func (mapObject *StringCurrencyMap) len() int {
	return len(mapObject.coreMap)
}

func (mapObject *StringCurrencyMap) containValueForKey(key string) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

// =========================================
