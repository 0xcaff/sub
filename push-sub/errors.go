package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Error returned by GetConfig when there are undecoded keys in the toml.
type UndecodedError struct {
	Keys []toml.Key
}

func (err *UndecodedError) Error() string {
	message := fmt.Sprintf("%d undecoded keys:\n", len(err.Keys))

	for keyIndex := range err.Keys {
		key := err.Keys[keyIndex]
		message += key.String() + "\n"
	}

	return message
}

// Error returned by GetConfigReader when a field is missing from a
// Subscription.
type FieldMissingError struct {
	Name  string
	Field string
	Sub   *Subscription
}

func (e *FieldMissingError) Error() string {
	return fmt.Sprintf("%s missing on subscription %s: %#v", e.Field, e.Name, e.Sub)
}
