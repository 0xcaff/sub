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
)

type MessageCallback func(*http.Request, []byte)
type ErrorCallback func(error)

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

	// When a broken message is handled, errors are dispatched to this callback.
	OnError ErrorCallback

	// The client which is used to make requests to the hub.
	Client *http.Client

	// The current state of the client.
	State State
}
