package sub

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
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

	// check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Error("The status code is bad", resp.StatusCode)
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
		t.Error("Subscription in unexpected state. Was in state: " +
			s.State.String())
	}

	// check that the lease time was updated
	realLease := int64(s.LeaseExpiry.Sub(time.Now()).Seconds())
	if realLease < leaseTime-10 || realLease > leaseTime+10 {
		t.Error("Expected: ", leaseTime, "Real: ", realLease)
	}
}

func TestMessageReceive(t *testing.T) {
	expectedBody := "This is a secure message!"

	// initialize subscription
	s := &Sub{
		OnMessage: func(req *http.Request, rawBody []byte) {
			body := string(rawBody)

			// compare body
			if body != expectedBody {
				t.Error("Body incorrect. Expected: " + expectedBody + " Got: " + body)
			}
		},
		Secret: []byte("orGXgeXafOMfMzJONfuJyUQZoILaCWKYNGaTvjOwnldjtkAmYyebVhkXrnUgLvBAZhkUMhczxsJcaxBKOgfaYqYKezwdWuLKHeA"),
	}

	// start listening on server
	ts := httptest.NewServer(s)
	defer ts.Close()

	// build request
	mac := hmac.New(sha1.New, s.Secret)
	mac.Write([]byte(expectedBody))
	macSum := mac.Sum(nil)

	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(expectedBody))
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("X-Hub-Signature", "sha1="+string(toHex(macSum)))

	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	// check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Error("The status code is bad", resp.StatusCode)
	}
}

func toHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}
