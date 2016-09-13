package main

import (
	"net/http"
	"strings"
	"log"
	"sync"
	"io"
)

const (URL="http://localhost:8080/api/records")

func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	buf:=make([]byte,resp.ContentLength)
	if _,err:=resp.Body.Read(buf); err==nil || err == io.EOF {
		return string(buf)
	} else {
		log.Println(buf)
		panic(err)
	}
}

var ids chan string=make(chan string,1000)
const value string=`{"data":"first value","ttl":0}`

func get()  {
	defer wg.Done()
	client:=http.Client{}
	for {
		log.Println("GET")
		id:=<-ids
		ids<-id
		resp,err:=client.Get(URL + "/" + id)
		if err==nil {
			resp.Body.Close()
		} else {
			log.Println(err)
		}
	}
}

func _set(client *http.Client){
	log.Println("SET")
	resp, err := client.Post(URL, "text/json", strings.NewReader(value))
	if err == nil {
		ids<-readBody(resp)
	} else {
		log.Println(err)
	}
}

func set()  {
	defer wg.Done()
	client:=http.Client{}
	for {
		_set(&client)
	}

}

func update()  {
	defer wg.Done()
	client:=http.Client{}
	for {
		log.Println("UPDATE")
		id:=<-ids
		ids<-id
		req, _ := http.NewRequest(http.MethodPut, URL + "/" + id, strings.NewReader(value))
		resp, err := client.Do(req)
		if err==nil {
			resp.Body.Close()
		} else {
			log.Println(err)
		}
	}
}

func del(){
	defer wg.Done()
	client:=http.Client{}
	for {
		log.Println("DELETE")
		req, _ := http.NewRequest(http.MethodDelete, URL + "/" + <-ids, nil)
		resp,err:=client.Do(req)
		if err==nil {
			resp.Body.Close()
		} else {
			log.Println(err)
		}
	}
}

var wg sync.WaitGroup
func main() {
	wg.Add(4)
	_set(http.DefaultClient)
	go get()
	go set()
	go update()
	go del()
	wg.Wait()
}
