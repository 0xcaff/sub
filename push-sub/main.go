package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/caffinatedmonkey/sub"
)

const topic = "https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCF8wOZvgrBZzj4netRo3m2A"

// const topic = "http://push-pub.appspot.com/feed"

func main() {
	// discover hub
	s, err := sub.Discover(topic)
	if err != nil {
		panic(err)
	}

	// ensure secure hub
	s.Hub.Scheme = "https"

	// register server at random endpoint
	randomEndpoint := string(sub.RandAsciiBytes(99))
	s.Callback = MustParseUrl("https://hub.ydns.eu/" + randomEndpoint)
	fmt.Println(s.Callback)

	http.Handle("/"+randomEndpoint, &LoggerHandler{s})
	s.OnMessage = sub.MessageCallback(func(r *http.Request, body []byte) {
		fmt.Printf("Request: %#v\n Body: %s\n", body)
		fmt.Println()
	})

	s.OnError = sub.ErrorCallback(func(e error) {
		panic(e)
	})

	// localhost:17889
	go http.ListenAndServe(":8080", nil) // &LoggerHandler{http.DefaultServeMux})

	err = s.Subscribe()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(s.Secret))

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
	fmt.Printf("Request: %#v\n", r)
	h.Handler.ServeHTTP(rw, r)
}
