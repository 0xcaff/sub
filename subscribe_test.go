package sub

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"testing"
)

func TestSubscribe(t *testing.T) {
	callback := "https://my-server-name.com/callback_for_pubsub"
	topic := "https://this-is-the-best-network.com/content/feed.xml"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		byteBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		gotContentType := r.Header.Get("Content-Type")
		if gotContentType != "application/x-www-form-urlencoded" {
			t.Error("Incorrect content type", gotContentType)
		}

		body := string(byteBody)
		values, err := url.ParseQuery(body)
		if err != nil {
			t.Error(err)
		}

		gotMode := values.Get("hub.mode")
		if gotMode != "subscribe" {
			t.Error("Incorrect mode")
		}

		gotLease := values.Get("hub.lease_seconds")
		if gotLease != "" {
			t.Error("Incorrect lease time")
		}

		secret := values.Get("hub.secret")
		if len(secret) <= 0 || len(secret) >= 200 {
			t.Error("Incorrect secret length")
		}

		gotCb := values.Get("hub.callback")
		if gotCb != callback {
			t.Error("Incorrect callback.")
		}

		gotTopic := values.Get("hub.topic")
		if gotTopic != topic {
			t.Error("Incorrect topic")
		}

		w.WriteHeader(http.StatusAccepted)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	subscription := &Sub{
		Hub:      MustParseUrl(ts.URL),
		Topic:    MustParseUrl(topic),
		Callback: MustParseUrl(callback),
		Client:   http.DefaultClient,
	}

	err := subscription.Subscribe()
	if err != nil {
		t.Error(err)
	}
}

func TestUnsubscribe(t *testing.T) {
	topic := "http://example.com/feed.xml"
	callback := "https://my-server.com/subscriber"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		byteBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		gotContentType := r.Header.Get("Content-Type")
		if gotContentType != "application/x-www-form-urlencoded" {
			t.Error("Incorrect content type", gotContentType)
		}

		body := string(byteBody)
		values, err := url.ParseQuery(body)
		if err != nil {
			t.Error(err)
		}

		gotMode := values.Get("hub.mode")
		if gotMode != "unsubscribe" {
			t.Error("Incorrect mode", gotMode)
		}

		gotCb := values.Get("hub.callback")
		if gotCb != callback {
			t.Error("Incorrect callback.", gotCb)
		}

		gotTopic := values.Get("hub.topic")
		if gotTopic != topic {
			t.Error("Incorrect topic", gotTopic)
		}

		w.WriteHeader(http.StatusAccepted)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	sub := &Sub{
		Hub:      MustParseUrl(ts.URL),
		Callback: MustParseUrl(callback),
		Topic:    MustParseUrl(topic),
		Client:   http.DefaultClient,
	}

	err := sub.Unsubscribe()
	if err != nil {
		t.Error(err)
	}
}

func TestRenewal(t *testing.T) {
	// initialize
	waitChan := make(chan struct{})
	sub := &Sub{
		cancelRenew: make(chan struct{}),
		LeaseExpiry: time.Now(), // some time in the future/present so renewal will run
	}
	sub.OnRenewLease = func(s *Sub) {
		if s != sub {
			t.Error("The sub is invalid")
		}

		waitChan <- struct{}{}
	}

	// ensure that OnRenewLease is called
	sub.scheduleRenewal()
	<-waitChan
}

func TestCancelRenewal(t *testing.T) {
	// initialize
	sub := &Sub{
		cancelRenew: make(chan struct{}),
		OnRenewLease: func(s *Sub) {
			t.Error("The lease shouldn't have been renewed.")
		},

		// enough time to to schedule the renewal and cancel but not too
		// long to verify that OnRenewLease isn't called.
		LeaseExpiry: time.Now().Add(500 * time.Millisecond),
	}

	// ensure that when we cancel, it's sucessful and OnRenewLease isn't fired.
	sub.scheduleRenewal()
	success := sub.CancelRenewal()
	if !success {
		t.Error("Failed to cancel lease")
	}

	time.Sleep(500 * time.Millisecond)
}
