package message

// Message interface to interact with an abstract message
type Message interface {
	// The text received without the matching portion
	GetText() string
	// The friendly name of the channel in which the message was issued
	GetChannel() string
	// The channel id used in replies so they are hyperlinks
	GetChannelID() string
	// The channel link? Doesn't seem to be in use anymore
	GetChannelLink() string
	// The friendly name of the user that has sent the message, used internally to match with groups and such
	GetUsername() string
	// The username id of the user that has sent the message, used in replies so they notify the user
	GetUsernameID() string
	// IsIM
	IsIM() bool
}
