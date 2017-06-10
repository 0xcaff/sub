package sub

import (
	"io"
	"net/http"
	"net/http/httptest"

	"testing"
)

const Atom = `
<?xml version="1.0" encoding="utf-8"?>
<!--
  This is a static file. The corresponding feed (identified in the
  <link rel="self"/> tag) should be used as a topic on the
  http://pubsubhubbub.appspot.com hub to subscribe to video updates.
-->
<feed xmlns="http://www.w3.org/2005/Atom">
  <link rel="hub" href="http://pubsubhubbub.appspot.com"/>
  <link rel="self" href="https://www.youtube.com/xml/feeds/videos.xml" />
  <title>YouTube video feed</title>
</feed>
`

func TestDiscoverFromAtom(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, Atom)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	subscription := New()
	subscription.Topic = MustParseUrl(ts.URL)

	err := subscription.Discover()
	if err != nil {
		t.Error(err)
	}

	hubUrl := subscription.Hub.String()
	if hubUrl != "http://pubsubhubbub.appspot.com" {
		t.Error("Got incorrect hub url: " + hubUrl)
	}

	topicUrl := subscription.Topic.String()
	if topicUrl != ts.URL {
		t.Error("Got incorrect topic url: " + topicUrl)
	}
}

// func TestDiscoverFromHeaders(t *testing.T) {
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Add("Link", "<http://example.com/hub>; rel=hub")
// 		w.Header().Add("Link", "<http://example.com/feed>; rel=self")
// 	})
//
// 	ts := httptest.NewServer(handler)
// 	defer ts.Close()
//
// 	subscription, err := Discover(ts.URL)
// 	if err != nil {
// 		t.Error(err)
// 	}
//
// 	hubUrl := subscription.Hub.String()
// 	if hubUrl != "http://example.com/hub" {
// 		t.Error("Got incorrect hub url: " + hubUrl)
// 	}
//
// 	topicUrl := subscription.Topic.String()
// 	if topicUrl != "http://example.com/feed" {
// 		t.Error("Got incorrect topic url: " + topicUrl)
// 	}
// }
