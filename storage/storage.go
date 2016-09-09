package storage

import (
	"strconv"
	"time"
)

type Metric uint8

const (
	CPU Metric = iota
	RAM
	RPS
)

type Storage interface {
	Set(data string, ttl time.Duration) (uint64, error)
	GetValue(id uint64) (string, error)
	Delete(id uint64)
	Update(id uint64, data string, ttl time.Duration) bool
	GetMetric(metric Metric) (uint64, error)
}

//prepares id in suitable format for that storage
func strId(id uint64) string {
	return strconv.FormatUint(id, 10)
}
