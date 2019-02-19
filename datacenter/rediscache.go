package datacenter

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/vic/vic_go/log"
)

type redisCache struct {
	pool *redis.Pool
}

func startRedis(address string) (redisCacheObject *redisCache) {
	pool := &redis.Pool{
		MaxIdle:   200,
		MaxActive: 400, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				log.LogSerious(err.Error())
				return nil, err
			}
			_, err = c.Do("SELECT", 2)
			if err != nil {
				log.LogSerious(err.Error())
				c.Close()
				return nil, err
			}
			return c, err
		},
	}
	redisCacheObject = &redisCache{
		pool: pool,
	}
	return redisCacheObject
}

func (redisCache *redisCache) close() {
	redisCache.pool.Close()
}

func (redisCache *redisCache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	poolConnection := redisCache.pool.Get()
	if poolConnection == nil {
		return nil, errors.New("err:redis_exhausted")
	}
	reply, err = poolConnection.Do(commandName, args...)
	poolConnection.Close()
	return reply, err
}

func generateKey(dataObject DataObject) string {
	return fmt.Sprintf("%s:%d", dataObject.CacheKey(), dataObject.Id())
}

func getStringFromReply(reply interface{}) string {
	if reply == nil {
		return ""
	}
	return string(reply.([]byte))
}

func (redisCache *redisCache) save(dataObject DataObject, key string, value string) error {
	cacheKey := generateKey(dataObject)
	_, err := redisCache.do("HSET", cacheKey, key, value)
	return err
}

func (redisCache *redisCache) saveMany(dataObject DataObject, keys []string, values []string) error {
	if len(keys) != len(values) {
		return errors.New("err:keys_values_len_mismatch")
	}
	cacheKey := generateKey(dataObject)
	conn := redisCache.pool.Get()
	conn.Send("MULTI")
	for index, key := range keys {
		conn.Send("HSET", cacheKey, key, values[index])
	}
	_, err := conn.Do("EXEC")
	conn.Close()
	return err
}

func (redisCache *redisCache) get(dataObject DataObject, key string) (value string, err error) {
	cacheKey := generateKey(dataObject)
	reply, err := redisCache.do("HGET", cacheKey, key)
	if err != nil {
		return "", err
	}
	return getStringFromReply(reply), nil
}

func (redisCache *redisCache) getMany(dataObject DataObject, keys []string) (values []string, err error) {
	cacheKey := generateKey(dataObject)
	values = make([]string, len(keys))
	for index, key := range keys {
		reply, err := redisCache.do("HGET", cacheKey, key)
		if err != nil {
			return []string{}, err
		}
		values[index] = getStringFromReply(reply)
	}
	return values, nil
}

func (redisCache *redisCache) saveGroupKeyValue(group string, key string, value string) (err error) {
	_, err = redisCache.do("HSET", group, key, value)
	return err
}

func (redisCache *redisCache) loadGroupKey(group string, key string) (value string, err error) {
	fmt.Println("load", group, key)
	reply, err := redisCache.do("HGET", group, key)
	if err != nil {
		return "", err
	}
	return getStringFromReply(reply), nil
}

func (redisCache *redisCache) removeGroupKey(group string, key string) (err error) {
	_, err = redisCache.do("HDEL", group, key)
	return err
}

func (redisCache *redisCache) loadAllKeysInGroup(group string) (keys []string, err error) {
	return redis.Strings(redisCache.do("HKEYS", group))
}
