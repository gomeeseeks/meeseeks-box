package meeseeks

import (
	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
	log "github.com/sirupsen/logrus"
)

func (m *Meeseeks) replyWithError(msg message.Message, err error) {
	content, err := m.formatter.Templates().RenderFailure(msg.GetUserLink(), err.Error(), "")
	if err != nil {
		log.Fatalf("could not render failure template: %s", err)
	}

	if err = m.client.Reply(content, m.formatter.ErrorColor(), msg.GetChannelID()); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithUnknownCommand(req request.Request) {
	log.Debugf("Could not find command '%s' in the command registry", req.Command)

	msg, err := m.formatter.Templates().RenderUnknownCommand(req.UserLink, req.Command)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	if err = m.client.Reply(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithHandshake(req request.Request, cmd command.Command) {
	if !cmd.HasHandshake() {
		return
	}
	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderHandshake(req.UserLink)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	if err = m.client.Reply(msg, m.formatter.InfoColor(), req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithUnauthorizedCommand(req request.Request, cmd command.Command) {
	log.Debugf("User %s is not allowed to run command '%s' on channel '%s'", req.Username,
		req.Command, req.Channel)

	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderUnauthorizedCommand(req.UserLink, req.Command)
	if err != nil {
		log.Fatalf("could not render unathorized command template %s", err)
	}

	if err = m.client.Reply(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithCommandFailed(req request.Request, cmd command.Command, err error, out string) {
	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderFailure(req.UserLink, err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	if err = m.client.Reply(msg, m.formatter.ErrorColor(), req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithSuccess(req request.Request, cmd command.Command, out string) {
	msg, err := m.formatter.WithTemplates(cmd.Templates()).RenderSuccess(req.UserLink, out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	if err = m.client.Reply(msg, m.formatter.SuccessColor(), req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}
