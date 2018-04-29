package request

// import (
// 	"errors"

// 	"github.com/gomeeseeks/meeseeks-box/meeseeks"
// 	"github.com/gomeeseeks/meeseeks-box/meeseeks/request/parser"
// 	"github.com/sirupsen/logrus"
// )

// Dummy function
func Dummy() {

}

// // ErrNoCommandToRun is returned when a request can't identify a command to run
// var ErrNoCommandToRun = errors.New("no command to run")

// // FromMessage gets a message and generates a valid request from it
// func FromMessage(msg meeseeks.Message) (meeseeks.Request, error) {
// 	args, err := parser.Parse(msg.GetText())
// 	logrus.Debugf("Command '%s' parsed as %#v", msg.GetText(), args)

// 	if err != nil {
// 		return meeseeks.Request{}, err
// 	}

// 	if len(args) == 0 {
// 		return meeseeks.Request{
// 			Username:    msg.GetUsername(),
// 			UserID:      msg.GetUserID(),
// 			UserLink:    msg.GetUserLink(),
// 			Channel:     msg.GetChannel(),
// 			ChannelID:   msg.GetChannelID(),
// 			ChannelLink: msg.GetChannelLink(),
// 			IsIM:        msg.IsIM(),
// 		}, ErrNoCommandToRun
// 	}

// 	return meeseeks.Request{
// 		Command:     args[0],
// 		Args:        args[1:],
// 		Username:    msg.GetUsername(),
// 		UserID:      msg.GetUserID(),
// 		UserLink:    msg.GetUserLink(),
// 		Channel:     msg.GetChannel(),
// 		ChannelID:   msg.GetChannelID(),
// 		ChannelLink: msg.GetChannelLink(),
// 		IsIM:        msg.IsIM(),
// 	}, nil
// }
