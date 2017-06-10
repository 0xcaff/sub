// This packages implements a subscriber against version 0.4 of the PubSubHubbub
// specification (
// http://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html).

// NOTE:
// Some of these messages aren't secured, especially the protocol ones from the
// hub to us. To stay safe, use high entropy, HTTPS only callback urls.

package sub

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

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

// A helper function create and fill a slice of length n with characters from
// a-zA-Z0-9_-. It panics if there are any problems getting random bytes.
func RandAsciiBytes(n int) []byte {
	// The number of input bytes needed to have more than enough base64 output
	inputBytesNeeded := int64(n/4*3 + 1)

	// base64 encoded output
	output := bytes.NewBuffer(make([]byte, 0, n))

	// url safe, emits to output, input is writeRandTo
	enc := base64.RawURLEncoding
	writeRandTo := base64.NewEncoder(enc, output)
	_, err := io.CopyN(writeRandTo, rand.Reader, inputBytesNeeded)
	if err != nil {
		panic(err)
	}

	err = writeRandTo.Close()
	if err != nil {
		panic(err)
	}

	return output.Bytes()[:n]
}
