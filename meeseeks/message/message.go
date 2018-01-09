package message

// Message interface to interact with an abstract message
type Message interface {
	GetText() string
	GetChannel() string
	GetReplyTo() string
	GetUsername() string
	IsIM() bool
}
