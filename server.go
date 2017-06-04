package sub

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"

	"fmt"
)

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
		if mode == deniedMode.String() {
			// a subscription can be denied without requesting a change
			s.State = Unsubscribed

			// forward error
			if s.OnError != nil {
				s.OnError(&DeniedError{
					Topic:  topic,
					Reason: values.Get("hub.reason"),
				})
			}

			// TODO: Cancel resubscription
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

		if mode == subMode.String() {
			s.State = Subscribed
		} else if mode == unsubMode.String() {
			s.State = Unsubscribed
		} else {
			// bad request
			if s.OnError != nil {
				s.OnError(&RequestError{
					Request: r,
					Message: "Unknown message mode",
				})
			}

			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		// echo the challenge
		challenge := values.Get("hub.challenge")

		fmt.Println(challenge)
		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, challenge)

		// TODO: Register a listener for the lease
		return
	}

	if r.Method == http.MethodPost {
		// handle as a message
		if s.OnMessage == nil {
			// nothing to do
			return
		}

		// check hmac
		messageMac := []byte(r.Header.Get("X-Hub-Signature"))
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
		isValid := checkMAC(message, messageMac, s.Secret)
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
	http.Error(rw, "", http.StatusBadRequest)
}

// A helper function which computes and securely compares the sha1 mac of a
// message.
func checkMAC(message, messageMac, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)

	calculatedMac := mac.Sum(nil)
	calcMacHeader := []byte("sha1=")
	calcMacHeader = append(calcMacHeader, toHex(calculatedMac)...)

	println(string(calcMacHeader))
	return hmac.Equal(messageMac, calcMacHeader)
}

func toHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}