package main

import (
	. "cacheserver/storage"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
	"log"
)

var rps chan uint64=make(chan uint64,1)

func refresher()  {
	log.Println("time")
	rps<-0
	for range time.Tick(time.Second) {
		log.Println(<-rps)
		rps <- 0
	}
}

func incRps() {
	prevRps :=<-rps
	rps<- prevRps +1
}

type DataRecord struct {
	Data string `json:"data"`
	Ttl  int32  `json:"ttl"`
}

func getID(w http.ResponseWriter, params httprouter.Params) (uint64, bool) {
	id, err := strconv.ParseUint(params.ByName("id"), 10, 64)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Cannot fetch id: " + err.Error()))
		return 0, false
	}
	return id, true
}

func getValue(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	defer incRps()
	id, ok := getID(w, params)
	if !ok {
		return
	}
	record, err := storage.GetValue(id)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(record))
}

func setValue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer incRps()
	var rec DataRecord
	err := json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Cannot decode json: " + err.Error()))
		return
	}
	id, err := storage.Set(rec.Data, time.Duration(rec.Ttl)*time.Millisecond)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Cannot save value: " + err.Error()))
		return
	}
	w.Write([]byte(strconv.FormatUint(id, 10)))

}

func updateValue(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer incRps()
	id, ok := getID(w, params)
	if !ok {
		return
	}
	var rec DataRecord
	err := json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(400)
	}
	ok = storage.Update(id, rec.Data, time.Duration(rec.Ttl)*time.Millisecond)
	if !ok {
		w.WriteHeader(404)
	}

}

func deleteValue(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	defer incRps()
	id, ok := getID(w, params)
	if !ok {
		return
	}
	storage.Delete(id)

}

func getMetric(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	defer incRps()
	metric := params.ByName("metric")
	var value uint64
	var err error
	switch metric {
	case "ram":
		value, err = storage.GetMetric(RAM)
	case "cpu":
		value, err = storage.GetMetric(CPU)
	case "rps":
		//get current rps from channel
		value, err=<-rps, nil
		//...and push it back )))
		rps<-value
	default:
		w.WriteHeader(400)
		w.Write([]byte("Requested metric not implemented"))
	}
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Cannot get metric:" + err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(strconv.FormatUint(value, 10)))
}