//go:generate stringer -type=State
package sub

type State int

const (
	Unsubscribed State = iota
	Requested
	Subscribed
)

const (
	subscribeMode   = "subscribe"
	unsubscribeMode = "unsubscribe"
	deniedMode      = "denied"
)
