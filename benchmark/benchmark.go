package main

import (
	"net/http"
	"container/list"
	"math/rand"
	"strings"
	"log"
	"io"
)
type Operation uint8
const (
	GET  = iota
	SET
	UPDATE
	DELETE
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

func main() {
	value:=`{"data":"first value","ttl":0}`
	client:=http.DefaultClient
	ids:=list.New()
	resp,err:=client.Post(URL,"text/json",strings.NewReader(value))
	if err==nil {
		ids.PushFront(readBody(resp))
	}
	for {
		switch rand.Int() % 4{
		case GET:
			resp,err=client.Get(URL+"/"+ids.Front().Value.(string))
			log.Println("GET")
			if err==nil {
				resp.Body.Close()
			}
		case SET:
			resp,err=client.Post(URL,"text/json",strings.NewReader(value))
			if err==nil {
				ids.PushFront(readBody(resp))
			}
			log.Println("SET")
		case UPDATE:
			req,_:=http.NewRequest(http.MethodPut,URL+"/"+ids.Front().Value.(string),strings.NewReader(value))
			resp,err=client.Do(req)
			if err==nil {
				resp.Body.Close()
			}
			log.Println("UPDATE")
		case DELETE:
			req,_:=http.NewRequest(http.MethodDelete,URL+"/"+ids.Front().Value.(string),nil)
			resp,err=client.Do(req)
			if err==nil {
				resp.Body.Close()
			}
			log.Println("DELETE")
		}
		resp,err=client.Get("http://localhost:8080/api/metrics/rps")
		if err==nil {
			log.Println(readBody(resp))
		} else {
			log.Println(err)
		}
	}
}
