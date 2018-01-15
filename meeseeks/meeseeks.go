package meeseeks

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"

	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/command"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/message"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
)

// Client interface that provides a way of replying to messages on a channel
type Client interface {
	Reply(text, color, channel string) error
	ReplyIM(text, color, user string) error
}

// Meeseeks is the command execution engine
type Meeseeks struct {
	client    Client
	config    config.Config
	commands  commands.Commands
	templates *template.TemplatesBuilder
}

// New creates a new Meeseeks service
func New(client Client, conf config.Config) Meeseeks {
	cmds, _ := commands.New(conf) // TODO handle the error
	templatesBuilder := template.NewBuilder().WithMessages(conf.Messages)
	return Meeseeks{
		client:    client,
		config:    conf,
		commands:  cmds,
		templates: templatesBuilder,
	}
}

// Process processes a received message
func (m Meeseeks) Process(msg message.Message) {
	req, err := request.FromMessage(msg)
	if err != nil {
		log.Debugf("Failed to parse message '%s' as a command: %s", msg.GetText(), err)
		m.replyWithInvalidMessage(msg, err)
		return
	}

	cmd, err := m.commands.Find(req.Command)
	if err == commands.ErrCommandNotFound {
		m.replyWithUnknownCommand(req)
		return
	}
	if err = auth.Check(req.Command, cmd); err != nil {
		m.replyWithUnauthorizedCommand(req, cmd)
		return
	}

	log.Infof("Accepted command '%s' from user '%s' on channel '%s' with args: %s",
		req.Command, req.Username, req.Channel, req.Args)
	j, err := jobs.Create(req)
	if err != nil {
		log.Errorf("could not create job: %s", err)
	}

	m.replyWithHandshake(j.Request, cmd)

	out, err := cmd.Execute(j)
	if err != nil {
		log.Errorf("Command '%s' from user '%s' failed execution with error: %s",
			req.Command, req.Username, err)
		m.replyWithCommandFailed(j.Request, cmd, err, out)
		jobs.Finish(j.ID, jobs.FailedStatus)
		return
	}

	log.Infof("Command '%s' from user '%s' succeeded execution", req.Command,
		req.Username)
	m.replyWithSuccess(j.Request, cmd, out)
	jobs.Finish(j.ID, jobs.SuccessStatus)
}

func (m Meeseeks) replyWithInvalidMessage(msg message.Message, err error) {
	content, err := m.templates.Build().RenderFailure(msg.GetUsernameID(), err.Error(), "")
	if err != nil {
		log.Fatalf("could not render failure template: %s", err)
	}

	if err = m.client.Reply(content, m.config.Colors.Error, msg.GetChannelID()); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m Meeseeks) replyWithUnknownCommand(req request.Request) {
	log.Debugf("Could not find command '%s' in the command registry", req.Command)

	msg, err := m.templates.Build().RenderUnknownCommand(req.UsernameID, req.Command)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Error, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m Meeseeks) replyWithHandshake(req request.Request, cmd command.Command) {
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

func (m Meeseeks) replyWithUnauthorizedCommand(req request.Request, cmd command.Command) {
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

func (m Meeseeks) replyWithCommandFailed(req request.Request, cmd command.Command, err error, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderFailure(req.UsernameID, err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Error, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m Meeseeks) replyWithSuccess(req request.Request, cmd command.Command, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderSuccess(req.UsernameID, out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	if err = m.client.Reply(msg, m.config.Colors.Success, req.ChannelID); err != nil {
		log.Errorf("Failed to reply: %s", err)
	}
}

func (m Meeseeks) buildTemplatesFor(cmd command.Command) template.Templates {
	return m.templates.Clone().WithTemplates(cmd.Templates()).Build()
}
