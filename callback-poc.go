package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var callbackPool map[int]chan string

func reqHandler(res http.ResponseWriter, req *http.Request) {
	// not thread safe
	var id int
	for {
		id = rand.Int()

		if _, ok := callbackPool[id]; !ok {
			callbackPool[id] = make(chan string, 1)
			break
		}
	}

	fmt.Printf("http://127.0.0.1:8080/callback?id=%d\n", id)

	select {
	case msg := <-callbackPool[id]:
		res.WriteHeader(200)
		fmt.Fprint(res, msg)

	case <-time.After(120 * time.Second):
		res.WriteHeader(500)
		fmt.Fprint(res, "Callback didn't happen.")
	}
	delete(callbackPool, id)
}

func callbackHandler(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Printf("Error parsing form: %s", err)
		res.WriteHeader(400)
		fmt.Fprint(res, "can't parse data")
		return
	}

	if idArray, ok := req.Form["id"]; ok {
		id, err := strconv.Atoi(idArray[0])
		if err != nil {
			res.WriteHeader(400)
			fmt.Fprint(res, "can't parse query string id to int")
			return
		}
		if _, idExist := callbackPool[id]; !idExist {
			res.WriteHeader(400)
			fmt.Fprint(res, "id not found")
			return
		}

		var msg string
		if msgArray, ok := req.Form["msg"]; ok {
			msg = msgArray[0]
		} else {
			msg = "--> not set <--"
		}

		callbackPool[id] <- msg

		res.WriteHeader(200)
		fmt.Fprintf(res, "Received id: %d, msg: %s", id, msg)
		return
	}

	res.WriteHeader(400)
	fmt.Fprint(res, "query string id not found")
}

func main() {
	callbackPool = make(map[int]chan string)

	http.HandleFunc("/req", reqHandler)
	http.HandleFunc("/callback", callbackHandler)

	http.ListenAndServe(":8080", nil)
}
