package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/caffinatedmonkey/sub"
)

// const topic = "https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCiGMIk8oeayv91jjTgm-CIw"

const topic = "http://push-pub.appspot.com/feed"

func main() {
	// discover hub
	s, err := sub.Discover(topic)
	if err != nil {
		panic(err)
	}

	// ensure secure hub
	s.Hub.Scheme = "https"
	fmt.Printf("Hub: %s\n", s.Hub)

	// register server at random endpoint
	randomEndpoint := RandStringBytes(99)
	s.Callback = MustParseUrl("https://hub.ydns.eu/" + randomEndpoint)
	fmt.Printf("Callback: %s\n", s.Callback)

	http.Handle("/"+randomEndpoint, &LoggerHandler{s})
	s.OnMessage = sub.MessageCallback(func(r *http.Request, body []byte) {
		fmt.Printf("%s", body)
		fmt.Println(r)
	})

	s.OnError = sub.ErrorCallback(func(e error) {
		panic(e)
	})

	// localhost:17889
	go http.ListenAndServe(":8080", nil) // &LoggerHandler{http.DefaultServeMux})

	err = s.Subscribe()
	if err != nil {
		if re, ok := err.(*sub.ResponseError); ok {
			message, err := ioutil.ReadAll(re.Response.Body)
			if err != nil {
				panic(err)
			}

			fmt.Println(string(message))
		}

		panic(err)
	}

	// wait forever
	time.Sleep(10000 * time.Hour)
}

func MustParseUrl(raw string) *url.URL {
	r, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return r
}

type LoggerHandler struct {
	http.Handler
}

func (h *LoggerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)

	h.Handler.ServeHTTP(rw, r)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
