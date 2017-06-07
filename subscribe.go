package sub

import (
	"math/rand" // TODO: Detereministic
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Tries to cancel automatic renewal of the subscription. If the renewal was
// sucessfully cancelled, returns true. The renewal can fail when there is no
// renewal to cancel, a renewal is currently in progress.
func (s *Sub) CancelRenewal() bool {
	select {
	case s.cancelRenew <- struct{}{}:
		return true

	default:
		return false
	}
}

// Schedule a callback to run when the lease is close to expiration.
func (s *Sub) scheduleRenewal() {
	numTimeToExpiry := int64(float64(s.LeaseExpiry.Sub(time.Now())) * 0.75)
	if numTimeToExpiry < 0 {
		numTimeToExpiry = 0
	}

	timer := time.NewTimer(time.Duration(numTimeToExpiry))
	go func() {
		select {
		case <-timer.C:
			// handle update
			// re-subscribe
			err := s.Subscribe()
			if err != nil {
				if s.OnError != nil {
					s.OnError(err)
				}

				return
			}

		case <-s.cancelRenew:
			// stop the timer so it doesn't leak
			timer.Stop()
			return
		}
	}()
}

// A helper function to send requests to the hub. It populates the hub.callback
// and hub.topic fields then builds and sends a request with the encoded values
// and appropriate content type. If the request returns a non-202 status code,
// an RequestError is returned.
func (s *Sub) sendHubReq(values url.Values) error {
	// identifying information
	values.Set("hub.callback", s.Callback.String())
	values.Set("hub.topic", s.Topic.String())

	// encode request
	bodyReader := strings.NewReader(values.Encode())
	req, err := http.NewRequest(http.MethodPost, s.Hub.String(), bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// send requst
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}

	// check for application errors
	if resp.StatusCode != http.StatusAccepted {
		return &ResponseError{
			Response: resp,
			Message:  "Received non 200 status code",
		}
	}

	return nil
}

const maxSecretLen = 200 - 1

// Subscribes suggesting a lease time of leaseSeconds. The hub gets the final
// decision on what the lease time actually is.
func (s *Sub) SubscribeWithLease(leaseSeconds int) error {
	values := url.Values{}

	// add mode
	values.Set("hub.mode", subscribeMode)

	// add lease time
	if leaseSeconds != 0 {
		values.Set("hub.lease_seconds", strconv.Itoa(leaseSeconds))
	}

	// generate secret
	if len(s.Secret) == 0 {
		s.Secret = RandAsciiBytes(maxSecretLen)
	}

	// add secret
	values.Set("hub.secret", string(s.Secret))

	// send request
	return s.sendHubReq(values)
}

func (s *Sub) Subscribe() error {
	return s.SubscribeWithLease(0)
}

func (s *Sub) Unsubscribe() error {
	values := url.Values{}
	values.Set("hub.mode", unsubscribeMode)
	return s.sendHubReq(values)
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
