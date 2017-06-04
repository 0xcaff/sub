package sub

import (
	"net/url"
)

// TODO: Test for Handler

// A helper function used for testing which parses urls and panics on malformed
// urls.
func MustParseUrl(raw string) *url.URL {
	url, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return url
}
