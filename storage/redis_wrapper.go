package storage

import (
	"errors"
	"gopkg.in/redis.v4"
	"regexp"
	"strconv"
	"time"
)

type RedisStorage struct {
	client *redis.Client
}

var used_memory_regexp_redis *regexp.Regexp
var used_cpu_user *regexp.Regexp

func init() {
	used_memory_regexp_redis, _ = regexp.Compile(`used_memory:(\d+)`)
	used_cpu_user, _ = regexp.Compile(`used_cpu_user:(\d+)`)
}

func (storage *RedisStorage) nextId() (uint64, error) {
	if err := storage.client.Incr("counter").Err(); err != nil {
		return 0, err
	}
	id, err := storage.client.Get("counter").Uint64()
	return id, err
}

func NewRedisStorage(addr string, password string, db int) *RedisStorage {
	return &RedisStorage{client: redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}),
	}
}

func (storage *RedisStorage) Delete(id uint64) {
	storage.client.Del(strId(id))
}

func (storage *RedisStorage) Set(data string, ttl time.Duration) (uint64, error) {
	id, err := storage.nextId()
	if err != nil {
		return 0, err
	}
	return id, storage.client.Set(strId(id), data, ttl).Err()
}

//Warning, change ttl
func (storage *RedisStorage) Update(id uint64, data string, ttl time.Duration) bool {
	return storage.client.Set(strId(id), data, ttl).Err() == nil
}

func (storage *RedisStorage) GetValue(id uint64) (string, error) {
	v, err := storage.client.Get(strId(id)).Result()
	return v, err
}

func (storage *RedisStorage) GetMetric(metric Metric) (uint64, error) {
	switch metric {
	case CPU:
		res, err := storage.client.Info("cpu").Result()
		if err == nil {
			return strconv.ParseUint(used_cpu_user.FindStringSubmatch(res)[1], 10, 64)
		}
		return 0, err
	case RAM:
		res, err := storage.client.Info("memory").Result()
		if err == nil {
			return strconv.ParseUint(used_memory_regexp_redis.FindStringSubmatch(res)[1], 10, 64)
		}
		return 0, err
	case RPS:
		panic("RPS metric not implemented")
	default:
		return 0, errors.New("Invalid metric requested")
	}
}
