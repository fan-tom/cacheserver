package main

import (
	. "cacheserver/storage"
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"os"
	"strconv"
)

func config() (Storage, int) {
	//command-line flags
	var storage Storage
	st := flag.String("storage", "", "what storage to use. ram|redis|memcached")
	port := flag.Int("port", -1, "what port use to listen on")
	storage_server_addr := flag.String("server", "", "socket of storage server ip:port")
	password := *flag.String("password", "", "password for connection")

	flag.Parse()
	if *st == "" {
		//no command line flag->fetch value from env variable
		*st = os.Getenv("CACHESERVER_STORAGE")
	}
	if *port == -1 {
		prt, err := strconv.ParseInt(os.Getenv("CACHESERVER_PORT"), 10, 32)
		if err == nil {
			*port = int(prt)
		} else {
			panic("No port specified")
		}
	}

	switch *st {
	case "ram":
		tmp := NewRamStorage()
		storage = &tmp
	default:
		if *storage_server_addr == "" {
			*storage_server_addr = os.Getenv("CACHSERVER_STORAGE_SOCKET")
			if *storage_server_addr == "" {
				panic("No storage socket provided")
			}
		}

		switch *st {

		case "redis":
			tmp := NewRedisStorage(*storage_server_addr, password, 0)
			storage = &tmp
		case "memcached":
			tmp := NewMCStorage(*storage_server_addr)
			storage = &tmp
		default:
			panic("Wrong or empty storage specified: " + *st)
		}
	}
	fmt.Println(*st)
	fmt.Println(*port)
	return storage, *port
}

var storage Storage

func main() {
	var port int
	storage, port = config() //select storage
	router := httprouter.New()
	router.GET("/api/records/:id", getValue)
	router.GET("/api/metrics/:metric", getMetric)
	router.POST("/api/records", setValue)
	router.PUT("/api/records/:id", updateValue)
	router.DELETE("/api/records/:id", deleteValue)
	server := NewServer(int32(port), router)
	log.Fatal(server.ListenAndServe())
}
