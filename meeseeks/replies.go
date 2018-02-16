package meeseeks

// import (
// 	"github.com/gomeeseeks/meeseeks-box/command"
// 	"github.com/gomeeseeks/meeseeks-box/meeseeks/message"
// 	"github.com/gomeeseeks/meeseeks-box/meeseeks/request"
// 	log "github.com/sirupsen/logrus"
// )

// // Error
// // Request:
// //  - GetUserLink
// //  - GetChannelID
// // err: Error
// func (m *Meeseeks) replyWithError(msg message.Message, err error) {
// 	content, err := m.formatter.Templates().RenderFailure(msg.GetUserLink(), err.Error(), "")
// 	if err != nil {
// 		log.Fatalf("could not render failure template: %s", err)
// 	}

// 	if err = m.client.ReplyWithAttachment(content, m.formatter.ErrorColor(), msg.GetChannelID()); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }

// // Error
// // Request:
// //   - Command, UserLink, ChannelID
// func (m *Meeseeks) replyWithUnknownCommand(req request.Request) {
// 	log.Debugf("Could not find command '%s' in the command registry", req.Command)

// 	msg, err := m.formatter.Templates().RenderUnknownCommand(req.UserLink, req.Command)
// 	if err != nil {
// 		log.Fatalf("could not render unknown command template: %s", err)
// 	}

// 	if err = m.client.ReplyWithAttachment(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }

// // Error
// // Request:
// //   - username, command, channel -> for auditory logging
// //   - userlink, command -> to render
// //   - channelID -> to reply
// // Command: templates <- this is wrong
// func (m *Meeseeks) replyWithUnauthorizedCommand(req request.Request, cmd command.Command) {
// 	log.Debugf("User %s is not allowed to run command '%s' on channel '%s'", req.Username,
// 		req.Command, req.Channel)

// 	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderUnauthorizedCommand(req.UserLink, req.Command)
// 	if err != nil {
// 		log.Fatalf("could not render unathorized command template %s", err)
// 	}

// 	if err = m.client.ReplyWithAttachment(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }

// // Error
// // Request: user link, channelID
// // Command: templates
// // Err: the error
// // Out: the output
// func (m *Meeseeks) replyWithCommandFailed(req request.Request, cmd command.Command, err error, out string) {
// 	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderFailure(req.UserLink, err.Error(), out)
// 	if err != nil {
// 		log.Fatalf("could not render failure template %s", err)
// 	}

// 	if err = m.client.ReplyWithAttachment(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }

// // Info request and command -
// // Request: channelID, UserLink
// // Command: templates, and to check if there's a handshake (could be avoided?)
// func (m *Meeseeks) replyWithHandshake(req request.Request, cmd command.Command) {
// 	if !cmd.HasHandshake() {
// 		return
// 	}
// 	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderHandshake(req.UserLink)
// 	if err != nil {
// 		log.Fatalf("could not render unknown command template: %s", err)
// 	}

// 	if err = m.client.ReplyWithAttachment(msg, m.formatter.InfoColor(), req.ChannelID); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }

// // Success request, command, and text
// // Request: channelID, UserLink
// // Command: templates
// // out: text
// func (m *Meeseeks) replyWithSuccess(req request.Request, cmd command.Command, out string) {
// 	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderSuccess(req.UserLink, out)

// 	if err != nil {
// 		log.Fatalf("could not render success template %s", err)
// 	}

// 	if err = m.client.Reply(msg, req.ChannelID); err != nil {
// 		log.Errorf("Failed to reply: %s", err)
// 	}
// }
