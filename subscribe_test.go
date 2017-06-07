package sub

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

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

// TODO: Test Renewal
