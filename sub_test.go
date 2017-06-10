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

func TestRandAsciiBytes(t *testing.T) {
	sizes := []int{100, 99}

	for _, size := range sizes {
		// get random bytes
		random := RandAsciiBytes(size)

		// check length
		if len(random) != size {
			t.Error("len(random)", len(random), "expected", size)
		}

		// check characters
	outerLoop:
		for idx, randomChar := range random {
			for _, char := range encodeURL {
				if char == rune(randomChar) {
					continue outerLoop
				}
			}

			// we've checked the entire string and the randomChar isn't a
			// encodeUrl char
			t.Error("Invalid Character", string(randomChar), "from", string(random), "at", idx)
		}
	}
}
