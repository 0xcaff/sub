package sub

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const sha1Header = "sha1="

// Handles the incoming request for the subscriber.
func (s *Sub) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// handle as a protocol message
		values := r.URL.Query()

		// check topic
		topic := values.Get("hub.topic")
		if topic != s.Topic.String() {
			// we received an invalid topic, this could be the doing of a
			// malicious actor

			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: "Topic not found",
				})
			}

			http.Error(rw, "Topic not found", http.StatusNotFound)
			return
		}

		mode := values.Get("hub.mode")
		if mode == deniedMode {
			// a subscription can be denied without requesting a change
			s.State = Unsubscribed

			// forward error
			if s.OnError != nil {
				s.OnError(&DeniedError{
					Topic:  topic,
					Reason: values.Get("hub.reason"),
				})
			}

			s.CancelRenewal()
			return
		}

		// reject all events if we aren't in a requesting state
		if s.State != Requested {
			// log
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: "Received mode: " + mode +
						" in state: " + s.State.String(),
				})
			}

			// invalid state, message sent at wrong time
			rw.WriteHeader(http.StatusOK)
			return
		}

		if mode == subscribeMode {
			s.State = Subscribed
		} else if mode == unsubscribeMode {
			s.State = Unsubscribed
		} else {
			// bad request
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: "Unknown message mode",
				})
			}

			http.Error(rw, "Unknown request type", http.StatusBadRequest)
			return
		}

		// get the lease time
		leaseStr := values.Get("hub.lease_seconds")
		leaseSecs, err := strconv.ParseInt(leaseStr, 10, 64)
		if err != nil {
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: err.Error(),
				})
			}

			http.Error(rw, "Invalid lease seconds format", http.StatusBadRequest)
			return
		}
		s.LeaseExpiry = time.Now().Add(time.Duration(leaseSecs) * time.Second)
		s.scheduleRenewal()

		// echo the challenge
		challenge := values.Get("hub.challenge")

		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, challenge)

		return
	}

	if r.Method == http.MethodPost {
		// handle as a message
		if s.OnMessage == nil {
			// nothing to do
			return
		}

		// check hmac
		messageMac := r.Header.Get("X-Hub-Signature")

		// messageMac starts with sha1=
		ourHeader, digest := messageMac[0:len(sha1Header)],
			messageMac[len(sha1Header):]

		if ourHeader != sha1Header {
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: "The header was invalid.",
				})
			}

			http.Error(rw, "X-Hub-Signature header is invalid", http.StatusBadRequest)
			return
		}

		// de-hex encode
		decoded, err := hex.DecodeString(digest)
		if err != nil {
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: err.Error(),
				})
			}

			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		message, err := ioutil.ReadAll(r.Body)
		if err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}

			// unable to read stream
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		// this only checks the authenticity of the body. The headers could still
		// be tampered with.
		isValid := checkMAC(message, decoded, s.Secret)
		if !isValid && s.OnError != nil {
			// invalid hmac signature
			s.OnError(&RequestError{
				Request: r,
				Message: "Invalid HMAC Signature",
			})
		}

		if isValid {
			// call callback
			s.OnMessage(r, message)
		}

		rw.WriteHeader(http.StatusOK)
		return
	}

	// unknown request
	if s.OnError != nil {
		s.OnError(&RequestError{
			Request: r,
			Message: "Unknown request",
		})
	}
	http.Error(rw, "Unknown request", http.StatusBadRequest)
}

// A helper function which computes and securely compares the sha1 mac of a
// message.
func checkMAC(message, messageMac, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	calculatedMac := mac.Sum(nil)

	return hmac.Equal(messageMac, calculatedMac)
}
