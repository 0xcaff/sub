/*
This packages provides a subscriber for v0.4 of the PubSubHubbub protocol. It
provides features like discovery and automated renewal while aiming to remain
flexible and secure.
*/
package sub

// NOTE:
// Some of these messages aren't secured, especially the protocol ones from the
// hub to us. To stay safe, use high entropy, HTTPS only callback urls and only
// talk to the hub over HTTMS.

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Represents a subscription to a topic on a PubSubHubbub hub.
type Sub struct {
	// The url of the topic which this subscription receives events for.
	Topic *url.URL

	// The url of the hub. This can be discovered using Sub.Discover()
	Hub *url.URL

	// The url which is provided to the hub during subscription and renewal.
	// This url is allowed to contain query parameters.
	Callback *url.URL

	// When a non-PubSubHubbub message from the hub arrives, this callback is
	// called. The body on request is always closed.
	OnMessage func(request *http.Request, body []byte)

	// When a broken message is handled or the hub cancels our subscription,
	// this callback is called.
	OnError func(err error)

	// Called when it is time to renew the lease. This can be used to make
	// changes during lease renewals. Errors returned from here are sent to
	// OnError.
	OnRenewLease func(subscription *Sub)

	// The client which is used to make requests to the hub.
	Client *http.Client

	// The < 200 byte long secret used to validate that messages are coming from
	// the real server.
	Secret []byte

	// The current state of the client.
	State State

	// The time at which the lease will expire.
	LeaseExpiry time.Time

	// A channel to cancel any pending renewals.
	cancelRenew chan struct{}

	// A mutex which ensures that multiple callback requests don't leave the
	// response handler in an inconsistent state.
	requestLock sync.Mutex
}

func New() *Sub {
	return &Sub{
		Client: http.DefaultClient,
	}
}

// len(encodeURL) == 64. This allows (x <= 265) x % 64 to have an even
// distribution.
const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// TODO: Rename?

// A helper function create and fill a slice of length n with characters from
// a-zA-Z0-9_-. It panics if there are any problems getting random bytes.
func RandAlphanumBytes(n int) []byte {
	output := make([]byte, n)

	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)

	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])

		// random % 64
		randomPos := random % uint8(len(encodeURL))

		// put into output
		output[pos] = encodeURL[randomPos]
	}

	return output
}
