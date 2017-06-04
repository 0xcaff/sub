package sub

import (
	"fmt"
	"net/http"
)

// This error is sent to Sub.OnError when unexpected requests are sent.
type RequestError struct {
	Request *http.Request
	Message string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("%s %#v", e.Message, e.Request)
}

// This error is sent to Sub.OnError when a subscription is denied.
type DeniedError struct {
	Topic  string
	Reason string
}

func (e *DeniedError) Error() string {
	return e.Topic + ": " + e.Reason
}

type ResponseError struct {
	Response *http.Response
	Message  string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%s %#v", e.Message, e.Response)
}
