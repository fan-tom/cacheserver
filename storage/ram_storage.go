package storage

import (
	"errors"
	"strconv"
	"time"
	//"unsafe"
	"runtime"
	//"log"
)

type record struct {
	data  string
	timer *time.Timer
}

type dictionary map[uint64]record
type RamStorage struct {
	dictionary
	counter uint64
}

func NewRamStorage() *RamStorage {
	return &RamStorage{dictionary: make(dictionary),
		counter: 0}
}

func (storage *RamStorage) nextId() uint64 {
	storage.counter++
	return storage.counter
}

//just utility
func (storage *RamStorage) set(id uint64, data string, ttl time.Duration) {
	var timer *time.Timer=nil
	if ttl>0 {
		timer = time.AfterFunc(ttl, func() {
			//log.Println("Deleting id:",id)
			storage.Delete(id)
		})
	}
	storage.dictionary[id] = record{data: data, timer: timer}
}

func (storage *RamStorage) Delete(id uint64) {
	value,ok:=storage.dictionary[id]
	if ok {
		//stop timer to prevent future deletion that id
		if value.timer!=nil {
			value.timer.Stop()
		}
		delete(storage.dictionary, id)
	}
}

func (storage *RamStorage) Set(data string, ttl time.Duration) (uint64, error) {
	id := storage.nextId()
	storage.set(id, data, ttl)
	return id, nil
}

func (storage *RamStorage) Update(id uint64, data string, ttl time.Duration) bool {
	if _, ok := storage.dictionary[id]; !ok {
		return false
	}
	storage.set(id, data, ttl)
	return true
}

func (storage *RamStorage) GetValue(id uint64) (string, error) {
	v, ok := storage.dictionary[id]
	if !ok {
		return "", errors.New("No value for that key: " + strconv.FormatUint(id, 10))
	}
	return v.data, nil
}

func (storage *RamStorage) GetMetric(metric Metric) (uint64, error) {
	switch metric {
	case CPU:
		panic("CPU metric not implemented")
	case RAM:
		var memstats runtime.MemStats
		//force garbage collection
		runtime.GC()
		runtime.ReadMemStats(&memstats)
		return memstats.Alloc, nil
		//return uint64(unsafe.Sizeof(storage.dictionary)), nil
	case RPS:
		panic("RPS metric not implemented")
	default:
		return 0, errors.New("Invalid metric requested")
	}
}
