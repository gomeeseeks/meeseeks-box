package message

// Message interface to interact with an abstract message
type Message interface {
	GetText() string
	GetChannel() string
	GetChannelID() string
	GetChannelLink() string
	GetUsernameID() string
	GetUsername() string
	IsIM() bool
}
