package meeseeks

import (
	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
	"github.com/pcarranza/meeseeks-box/template"
	log "github.com/sirupsen/logrus"
)

func (m *Meeseeks) replyWithError(msg message.Message, err error) {
	content, err := m.templates.Build().RenderFailure(msg.GetUsernameID(), err.Error(), "")
	if err != nil {
		log.Fatalf("could not render failure template: %s", err)
	}

	if err = m.client.Reply(content, m.config.Colors.Error, msg.GetChannelID()); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithUnknownCommand(req request.Request) {
	log.Debugf("Could not find command '%s' in the command registry", req.Command)

	msg, err := m.templates.Build().RenderUnknownCommand(req.UsernameID, req.Command)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Error, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithHandshake(req request.Request, cmd command.Command) {
	if !cmd.HasHandshake() {
		return
	}
	msg, err := m.buildTemplatesFor(cmd).RenderHandshake(req.UsernameID)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Info, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithUnauthorizedCommand(req request.Request, cmd command.Command) {
	log.Debugf("User %s is not allowed to run command '%s' on channel '%s'", req.Username,
		req.Command, req.Channel)

	msg, err := m.buildTemplatesFor(cmd).RenderUnauthorizedCommand(req.UsernameID, req.Command)
	if err != nil {
		log.Fatalf("could not render unathorized command template %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Error, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithCommandFailed(req request.Request, cmd command.Command, err error, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderFailure(req.UsernameID, err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Error, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) replyWithSuccess(req request.Request, cmd command.Command, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderSuccess(req.UsernameID, out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Success, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m *Meeseeks) buildTemplatesFor(cmd command.Command) template.Templates {
	return m.templates.Clone().WithTemplates(cmd.Templates()).Build()
}
