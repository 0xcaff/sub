package sub

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestSubscriptionVerification(t *testing.T) {
	// initialize subscription
	s := &Sub{
		Topic: MustParseUrl("https://example.com/awesome-stuff/feed"),
		State: Requested,
	}

	// start listening on server
	ts := httptest.NewServer(s)
	defer ts.Close()

	// build verification request
	subscriber := MustParseUrl(ts.URL)
	challenge := "Thisisaprettyhardchallenge!!!"
	leaseTime := int64(10000)

	query := url.Values{}
	query.Add("hub.topic", s.Topic.String())
	query.Add("hub.mode", subscribeMode)
	query.Add("hub.challenge", challenge)
	query.Add("hub.lease_seconds", strconv.FormatInt(leaseTime, 10))
	subscriber.RawQuery = query.Encode()

	// send verification request
	resp, err := http.Get(subscriber.String())
	if err != nil {
		t.Error(err)
	}

	// read body
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	// check that the challenge was echoed
	if string(rawBody) != challenge {
		t.Error("The challenge wasn't echoed.")
	}

	// check that the subscription's state was updated
	if s.State != Subscribed {
		t.Error("Subscription in unexpected state. Was in state: " + s.State.String())
	}

	// check that the lease time was updated
	realLease := int64(s.LeaseExpiry.Sub(time.Now()).Seconds())
	if realLease < leaseTime-10 || realLease > leaseTime+10 {
		t.Error("Expected: ", leaseTime, "Real: ", realLease)
	}
}

// TODO: Test message sending.
