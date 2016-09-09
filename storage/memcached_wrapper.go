package storage

import (
	"errors"
	"github.com/Cristofori/kmud/telnet"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var telnet_client *telnet.Telnet

//var conn net.Conn
var used_memory_regexp_mc *regexp.Regexp

func init() {
	used_memory_regexp_mc, _ = regexp.Compile(`STAT bytes (\d+)`)
}

type MCStorage struct {
	client *memcache.Client
}

func NewMCStorage(sockets ...string) MCStorage {
	//var err error
	log.Println("Connect to" + sockets[0])
	conn, err := net.Dial("tcp", sockets[0])
	telnet_client = telnet.NewTelnet(conn)
	if err != nil {
		panic("Cannot connect to specified host")
	}
	client := memcache.New(sockets...)
	client.Add(&memcache.Item{Key: "counter",
		Value: []byte("0")})
	return MCStorage{client: client}
}
func (storage *MCStorage) nextId() (uint64, error) {
	return storage.client.Increment("counter", 1)
}

//prepare Item struct from arguments
func item(id uint64, data string, ttl time.Duration) *memcache.Item {
	return &memcache.Item{Key: strId(id),
		Value:      []byte(data),
		Expiration: int32(ttl / time.Second),
	}
}

func (storage *MCStorage) Delete(id uint64) {
	storage.client.Delete(strId(id))
}

func (storage *MCStorage) Set(data string, ttl time.Duration) (uint64, error) {
	id, err := storage.nextId()
	if err != nil {
		return 0, err
	}
	return id, storage.client.Add(item(id, data, ttl))
}

func (storage *MCStorage) Update(id uint64, data string, ttl time.Duration) bool {
	return storage.client.Replace(item(id, data, ttl)) == nil
}

func (storage *MCStorage) GetValue(id uint64) (string, error) {
	v, err := storage.client.Get(strId(id))
	if err != nil {
		return "", err
	}
	return string(v.Value), nil
}

func (storage *MCStorage) GetMetric(metric Metric) (uint64, error) {
	switch metric {
	case CPU:
		panic("CPU metric not implemented")
	case RAM:
		telnet_client.Write([]byte("stats\n"))
		//WARNING!!! buffer size may be not enough
		buffer := make([]byte, 2000)
		n, n_ := 0, 0
		for !strings.HasSuffix(string(buffer[:n]), "END\r\n") {
			n_, _ = telnet_client.Read(buffer[n:])
			n += n_
		}
		matches := used_memory_regexp_mc.FindStringSubmatch(string(buffer))
		if matches == nil {
			return 0, errors.New("Wrong storage response")
		}
		return strconv.ParseUint(matches[1], 10, 32)
	case RPS:
		panic("RPS metric not implemented")
	default:
		return 0, errors.New("Invalid metric requested")
	}
}
