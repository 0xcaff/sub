//go:generate stringer -type=State
package sub

type State int

const (
	Unsubscribed State = iota
	Requested
	Subscribed
)

type mode int

const (
	subMode mode = iota
	unsubMode
	deniedMode
)

func (m mode) String() string {
	switch m {
	case subMode:
		return "subscribe"

	case unsubMode:
		return "unsubscribe"

	case deniedMode:
		return "denied"

	default:
		return ""
	}
}
