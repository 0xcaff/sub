// This packages implements a subscriber against version 0.4 of the PubSubHubbub
// specification (
// http://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html).

// NOTE:
// Some of these messages aren't secured, especially the protocol ones from the
// hub to us. To stay safe, use high entropy, HTTPS only callback urls.

package sub

import (
	"net/http"
	"net/url"
	"time"
)

type MessageCallback func(*http.Request, []byte)
type ErrorCallback func(error)
type LeaseCallback func(*Sub) error

var DefaultLeaseCallback = func(s *Sub) error {
	return s.Subscribe()
}

// Represents a subscription to a topic on a PubSubHubbub hub.
type Sub struct {
	// The url of the hub.
	Hub *url.URL

	// The url of the topic which this subscription receives events for.
	Topic *url.URL

	// The url which is provided to the hub during subscription and renewal.
	// This url is allowed to contain query parameters.
	Callback *url.URL

	// The < 200 byte long secret used to validate that messages are coming from
	// the real server.
	Secret []byte

	// When a non-PubSubHubbub message from the hub arrives, this callback is
	// called.
	OnMessage MessageCallback

	// When a broken message is handled, the hub cancels our subscription or the
	// renewing the lease fails, errors are dispatched to this callback.
	OnError ErrorCallback

	// Called when it is time to renew the lease. This can be used to make
	// changes during lease renewals.
	OnRenewLease LeaseCallback

	// The client which is used to make requests to the hub.
	Client *http.Client

	// The current state of the client.
	State State

	// The time at which the lease will expire.
	LeaseExpiry time.Time

	// A channel to cancel any pending renewals.
	cancelRenew chan struct{}
}
