package main

import (
	"net/http"
	"strconv"
)

func NewServer(port int32, router http.Handler) http.Server {
	server := http.Server{
		Addr:    ":" + strconv.FormatInt(int64(port), 10),
		Handler: router,
	}
	return server
}
