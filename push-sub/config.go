package main

import (
	"io"
	"net/url"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Subscriptions map[string]*Subscription

	// The TCP address to listen to connections on. This value is passed to
	// net.Listen.
	Address string

	// The publicly accessible path of this callback server. This is used by
	// hubs to send notifications.
	BasePath *URL
}

type Subscription struct {
	// The topic url of the feed we are subscribing to. This field is required.
	Topic *URL

	// The hub providing updates about the topic. This can be discovered by
	// looking up the topic url.
	Hub *URL

	// A command which is called when a verified message is received. The
	// message body is passed to the program through standard input.
	Bin  string
	Args []string

	// Allow connecting to a discovered hub over http. This is insecure because
	// there is no way to verify state change requests to the callback server
	// originate from the hub.
	AllowInsecure bool
}

// Parses and validates the toml configuration.
func GetConfigReader(r io.Reader) (*Config, error) {
	// decode configuration
	config := &Config{}
	metadata, err := toml.DecodeReader(r, config)
	if err != nil {
		return nil, err
	}

	// report extra fields
	extras := metadata.Undecoded()
	if len(extras) > 0 {
		return nil, &UndecodedError{Keys: extras}
	}

	// report missing information
	for name, sub := range config.Subscriptions {
		// we always need the topic
		if sub.Topic == nil {
			return nil, &FieldMissingError{name, "topic", sub}
		}

		if len(sub.Bin) == 0 {
			return nil, &FieldMissingError{name, "bin", sub}
		}
	}

	return config, nil
}

func GetConfig(path string) (*Config, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return GetConfigReader(file)
}

// A type which wraps the real URL but adds unmarshalling superpowers.
type URL struct {
	url.URL
}

func (u *URL) UnmarshalText(text []byte) error {
	return u.URL.UnmarshalBinary(text)
}

var ConfigBaseDir = getEnv("XDG_CONFIG_HOME", "$HOME/.config")

// Returns the value of the environment variable. If the variable is empty,
// returns the os.ExpandEnv value of def.
func getEnv(env, def string) string {
	val := os.Getenv(env)
	if len(val) == 0 {
		return os.ExpandEnv(def)
	}

	return val
}
