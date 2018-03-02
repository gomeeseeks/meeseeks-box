package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/logs"
	"github.com/sirupsen/logrus"
)

// CommandOpts are the options used to build a new shell command
type CommandOpts struct {
	Cmd             string
	Args            []string
	AllowedGroups   []string
	AuthStrategy    string
	AllowedChannels []string
	ChannelStrategy string
	HasHandshake    bool
	Timeout         time.Duration
	Templates       map[string]string
	Help            meeseeks.Help
}

// New return a new ShellCommand based on the passed in opts
func New(opts CommandOpts) meeseeks.Command {
	return shellCommand{
		opts: opts,
	}
}

type shellCommand struct {
	opts CommandOpts
}

// Execute implements Command.Execute for the ShellCommand
func (c shellCommand) Execute(ctx context.Context, job meeseeks.Job) (string, error) {
	cmdArgs := append(c.Args(), job.Request.Args...)
	logrus.Debugf("Calling command %s with args %#v", c.Cmd(), cmdArgs)

	ctx, cancelFunc := context.WithTimeout(ctx, c.Timeout())
	defer cancelFunc()

	// TODO: load from config
	logConfig := logs.LoggerConfig{
		LoggerType: "local",
	}
	logR := logs.GetJobLogReader(logConfig, job.ID)
	logW := logs.GetJobLogWriter(logConfig, job.ID)

	AppendLogs := func(line string) {
		if e := logW.Append(line); e != nil {
			logrus.Errorf("Could not append '%s' to job %d logs: %s", line, job.ID, e)
		}
	}
	SetError := func(err error) error {
		if e := logW.SetError(err); e != nil {
			logrus.Errorf("Could set error to job %d: %s", job.ID, e)
		}
		return err
	}

	cmd := exec.CommandContext(ctx, c.Cmd(), cmdArgs...)
	op, err := cmd.StdoutPipe()
	if err != nil {
		return "", SetError(fmt.Errorf("could not create stdout pipe: %s", err))
	}
	ep, err := cmd.StderrPipe()
	if err != nil {
		return "", SetError(fmt.Errorf("could not create stderr pipe: %s", err))
	}

	tr := io.MultiReader(op, ep)

	done := make(chan struct{})

	go func() {
		s := bufio.NewScanner(tr)
		for s.Scan() {
			AppendLogs(s.Text())
		}
		done <- struct{}{}
	}()

	err = cmd.Start()
	if err != nil {
		logrus.Errorf("command failed to start: %s", err)
		return "", SetError(err)
	}

	// Wait for the command to be done or the context to be cancelled
	select {
	case <-ctx.Done():
		// We are finishing because the context was called
		err = ctx.Err()
	case <-done:
		// We are finishing because we are actually done
		err = cmd.Wait()
	}

	if err != nil {
		logrus.Errorf("command failed: %s", err)
		return "", SetError(err)
	}

	jobLog, err := logR.Get()
	if err != nil {
		logrus.Errorf("failed to read back output for job %d: %s", job.ID, err)
	}

	return jobLog.Output, jobLog.GetError()
}

func (c shellCommand) HasHandshake() bool {
	return c.opts.HasHandshake
}

func (c shellCommand) Templates() map[string]string {
	if c.opts.Templates == nil {
		return map[string]string{}
	}
	return c.opts.Templates
}

func (c shellCommand) AuthStrategy() string {
	if c.opts.AuthStrategy == "" {
		return "none"
	}
	return c.opts.AuthStrategy
}

func (c shellCommand) AllowedGroups() []string {
	if c.opts.AllowedGroups == nil {
		return []string{}
	}
	return c.opts.AllowedGroups
}

func (c shellCommand) ChannelStrategy() string {
	if c.opts.ChannelStrategy == "" {
		return "any"
	}
	return c.opts.ChannelStrategy
}

func (c shellCommand) AllowedChannels() []string {
	if c.opts.AllowedChannels == nil {
		return []string{}
	}
	return c.opts.AllowedChannels
}

func (c shellCommand) Args() []string {
	logrus.Debug("Returning shell command args ", c.opts.Args)
	if c.opts.Args == nil {
		return []string{}
	}
	return c.opts.Args
}

func (c shellCommand) Timeout() time.Duration {
	if c.opts.Timeout == 0 {
		return meeseeks.DefaultCommandTimeout
	}
	return c.opts.Timeout
}

func (c shellCommand) Cmd() string {
	return c.opts.Cmd
}

func (c shellCommand) Help() meeseeks.Help {
	return c.opts.Help
}

func (c shellCommand) Record() bool {
	return true
}

type shellHelp struct {
	summary string
	args    []string
}

func (h shellHelp) GetSummary() string {
	return h.summary
}

func (h shellHelp) GetArgs() []string {
	return h.args
}

// NewHelp returns a new command help implementation for the shell command
func NewHelp(summary string, args ...string) meeseeks.Help {
	return shellHelp{
		summary,
		append([]string{}, args...),
	}
}
