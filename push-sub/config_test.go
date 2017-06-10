package main

import (
	"strings"
	"testing"
)

const config = `
address="127.0.0.1:17889"
basepath="https://sub.shipped.com/"

[subscriptions.blog_feed]
topic="https://example.com/feed.xml"
hub="https://example.com/hub"
bin="/usr/bin/env"
args=["tee", "/tmp/output.txt"]
`

func TestParseConfig(t *testing.T) {
	r := strings.NewReader(config)
	_, err := GetConfigReader(r)
	if err != nil {
		t.Error(err)
	}
}
