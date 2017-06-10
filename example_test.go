package sub_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/0xcaff/sub"
)

func Example() {
	listenAddr := ":80"

	// create and initialize subscription
	s := sub.New()
	s.Topic = MustParseUrl("https://example.com/feed.xml")

	// discover the hub
	s.Discover()

	// register handler, we are using a randomly generate endpoint because
	// anyone can make a GET request to change our internal state if they know
	// the subscription callback url.
	randomEndpoint := "/subs/" + string(sub.RandAlphanumBytes(99))
	s.Callback = MustParseUrl("https://this-server.com" + randomEndpoint)

	// add callbacks
	s.OnError = func(err error) {
		// these errors are non-critical, just log them
		fmt.Println("Error", err)
	}

	s.OnMessage = func(req *http.Request, body []byte) {
		fmt.Println("Message", body)
	}

	s.OnRenewLease = func(s *sub.Sub) {
		err := s.Subscribe()
		if err != nil {
			fmt.Println("Subscription Error", err)
		}

		os.Exit(1)
	}

	mux := &http.ServeMux{}
	mux.Handle(randomEndpoint, s)

	// start listening before we start subscribing so that validation by the hub
	// doesn't fail
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	// subscribe
	err = s.Subscribe()
	if err != nil {
		panic(err)
	}

	// start serving
	server := &http.Server{Addr: listenAddr, Handler: mux}
	err = server.Serve(listener)
	if err != nil {
		panic(err)
	}
}

func MustParseUrl(raw string) *url.URL {
	url, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return url
}
