package sub

import (
	"net/url"
	"testing"
)

// A helper function used for testing which parses urls and panics on malformed
// urls.
func MustParseUrl(raw string) *url.URL {
	url, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return url
}

func TestRandAsciiBytesEven(t *testing.T) {
	l := 10
	random := RandAsciiBytes(l)
	if len(random) != l {
		t.Error("Incorrect length", len(random), string(random))
	}
}

func TestRandAsciiBytesOdd(t *testing.T) {
	l := 101
	random := RandAsciiBytes(l)
	if len(random) != l {
		t.Error("Incorrect length", len(random), string(random))
	}
}
