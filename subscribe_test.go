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

		body := string(byteBody)
		values, err := url.ParseQuery(body)
		if err != nil {
			t.Error(err)
		}

		if values.Get("hub.mode") != "subscribe" {
			t.Error("Incorrect mode")
		}

		if values.Get("hub.lease_seconds") != "" {
			t.Error("Incorrect lease time")
		}

		secret := values.Get("hub.secret")
		if len(secret) <= 0 || len(secret) >= 200 {
			t.Error("Incorrect secret length")
		}

		if values.Get("hub.callback") != callback {
			t.Error("Incorrect callback.")
		}

		if values.Get("hub.topic") != topic {
			t.Error("Incorrect topic")
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Error("Incorrect content type")
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

	subscription.Subscribe()
}
