package storage

import (
	"errors"
	"strconv"
	"time"
	//"unsafe"
	"runtime"
	//"log"
	"sync"
	"log"
)
type LOCK_TYPE uint8
const (
	R LOCK_TYPE=iota
	W
)
func usedRam() uint64 {
	var memstats runtime.MemStats
	//force garbage collection
	runtime.GC()
	runtime.ReadMemStats(&memstats)
	return memstats.Alloc
}

type record struct {
	data  string
	timer *time.Timer
}

type dictionary map[uint64]record
type RamStorage struct {
	mutex sync.RWMutex
	dictionary
	counter uint64
}

func NewRamStorage() *RamStorage {
	return &RamStorage{dictionary: make(dictionary),
		counter: 0}
}

func (storage *RamStorage) unlock(lt LOCK_TYPE) {
	var ltstr string
	switch lt {
	case R:
		storage.mutex.RUnlock()
		ltstr="Unlocked for Reading"
	case W:
		storage.mutex.Unlock()
		ltstr="Unlocked for Writing"
	}
	log.Println(ltstr)
}
func (storage *RamStorage) lock(lt LOCK_TYPE) {
	var ltstr string
	switch lt {
	case R:
		storage.mutex.RLock()
		ltstr="Locked for Reading"
	case W:
		storage.mutex.Lock()
		ltstr="Locked for Writing"
	}
	log.Println(ltstr)
}

//we assume that storage.mutex is already locked !!!
func (storage *RamStorage) nextId() uint64 {
	storage.counter++
	return storage.counter
}

//we assume that storage.mutex is already locked !!!
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
	log.Println("DELETE")
	storage.lock(W)
	defer storage.unlock(W)
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
	log.Println("SET")
	storage.lock(W)
	defer storage.unlock(W)
	id := storage.nextId()
	storage.set(id, data, ttl)
	return id, nil
}

func (storage *RamStorage) Update(id uint64, data string, ttl time.Duration) bool {
	log.Println("UPDATE")
	storage.lock(W)
	defer storage.unlock(W)
	if _, ok := storage.dictionary[id]; !ok {
		return false
	}
	storage.set(id, data, ttl)
	return true
}

func (storage *RamStorage) GetValue(id uint64) (string, error) {
	log.Println("GET")
	storage.lock(R)
	defer storage.unlock(R)
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
		return usedRam(),nil
		//return uint64(unsafe.Sizeof(storage.dictionary)), nil
	case RPS:
		panic("RPS metric not implemented")
	default:
		return 0, errors.New("Invalid metric requested")
	}
}
