package sub

import (
	"math/rand" // TODO: Detereministic
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// A helper function to send requests to the hub.
func (s *Sub) buildHubReq(values url.Values) (*http.Request, error) {
	// identifying information
	values.Set("hub.callback", s.Callback.String())
	values.Set("hub.topic", s.Topic.String())

	bodyReader := strings.NewReader(values.Encode())
	req, err := http.NewRequest(http.MethodPost, s.Hub.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

const maxSecretLen = 200 - 1

func (s *Sub) SubscribeWithLease(leaseSeconds int) error {
	values := url.Values{}

	// add mode
	values.Set("hub.mode", subMode.String())

	// add lease time
	if leaseSeconds != 0 {
		values.Set("hub.lease_seconds", strconv.Itoa(leaseSeconds))
	}

	// generate secret
	if len(s.Secret) == 0 {
		s.Secret = RandAsciiBytes(maxSecretLen)
	}

	// send secret
	values.Set("hub.secret", string(s.Secret))

	// send request
	req, err := s.buildHubReq(values)
	if err != nil {
		return err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}

	// check application errors
	if resp.StatusCode != http.StatusAccepted {
		return &ResponseError{
			Response: resp,
			Message:  "Received non 200 status code",
		}
	}

	return nil
}

func (s *Sub) Subscribe() error {
	return s.SubscribeWithLease(0)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// A helper function create and fill a slice of length n with characters from
// a-zA-Z0-9.
func RandAsciiBytes(n int) []byte {
	b := make([]byte, n)
	count := len(letterBytes)
	for i := range b {
		b[i] = letterBytes[rand.Intn(count)]
	}

	return b
}
