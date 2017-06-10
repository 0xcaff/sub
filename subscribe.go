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
// renewal to cancel or a renewal is currently in progress.

// Cancels the automatic renewal of the subscription. If there was no renewal to
// cancel, returns false.
func (s *Sub) CancelRenewal() bool {
	if s.cancelRenew == nil {
		// there is no automated renewal happening
		return false
	}

	s.cancelRenew <- struct{}{}
	return true
}

// Schedule a callback to run when the lease is close to expiration.
func (s *Sub) scheduleRenewal() {
	numTimeToExpiry := int64(float64(s.LeaseExpiry.Sub(time.Now())) * 0.75)
	if numTimeToExpiry < 0 {
		numTimeToExpiry = 0
	}

	// add channel to cancel renewals and signal that there is something to
	// cancel
	s.cancelRenew = make(chan struct{})

	timer := time.NewTimer(time.Duration(numTimeToExpiry))
	go func() {
		select {
		case <-timer.C:
			// handle update, renew lease
			s.OnRenewLease(s)

		case <-s.cancelRenew:
			// remove the channel to signal that there is no automatic renewal
			s.cancelRenew = nil

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

// Sends a subscription request to the hub suggesting a lease time of
// leaseSeconds. The hub gets the final decision on the lease time.
func (s *Sub) SubscribeWithLease(leaseSeconds int) error {
	s.State = Requested

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

// Sends a subscription request to the hub. If there is no error, the
// request completed sucessfully. State is only changed after the callback
// server verifies.
func (s *Sub) Subscribe() error {
	return s.SubscribeWithLease(0)
}

// Sends an un-subscription request to the hub. If no error is returned, the
// request completed sucessfully. State is only changed after the callback
// server verifies.
func (s *Sub) Unsubscribe() error {
	s.State = Requested

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
