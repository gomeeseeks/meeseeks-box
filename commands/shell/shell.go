package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/gomeeseeks/meeseeks-box/command"
	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/jobs/logs"
	"github.com/sirupsen/logrus"
)

// CommandOpts are the options used to build a new shell command
type CommandOpts struct {
	Cmd           string
	Args          []string
	AllowedGroups []string
	AuthStrategy  string
	Timeout       time.Duration
	Templates     map[string]string
	Help          command.CommandHelp
}

// New return a new ShellCommand based on the passed in opts
func New(opts CommandOpts) command.Command {
	return shellCommand{
		opts: opts,
	}
}

type shellCommand struct {
	opts CommandOpts
}

// Execute implements Command.Execute for the ShellCommand
func (c shellCommand) Execute(ctx context.Context, job jobs.Job) (string, error) {
	cmdArgs := append(c.Args(), job.Request.Args...)
	logrus.Debugf("Calling command %s with args %#v", c.Cmd(), cmdArgs)

	ctx, cancelFunc := context.WithTimeout(ctx, c.Timeout())
	defer cancelFunc()

	AppendLogs := func(line string) {
		if e := logs.Append(job.ID, line); e != nil {
			logrus.Errorf("Could not append '%s' to job %d logs: %s", line, job.ID, e)
		}
	}
	SetError := func(err error) error {
		if e := logs.SetError(job.ID, err); e != nil {
			logrus.Errorf("Could set error to job %d: %s", job.ID, e)
		}
		return err
	}

	buffer := bytes.NewBufferString("")

	cmd := exec.CommandContext(ctx, c.Cmd(), cmdArgs...)
	op, err := cmd.StdoutPipe()
	if err != nil {
		return "", SetError(fmt.Errorf("Could not create stdout pipe: %s", err))
	}
	ep, err := cmd.StderrPipe()
	if err != nil {
		return "", SetError(fmt.Errorf("Could not create stderr pipe: %s", err))
	}

	tr := io.TeeReader(io.MultiReader(op, ep), buffer)

	done := make(chan struct{})

	go func() {
		s := bufio.NewScanner(tr)
		for s.Scan() {
			AppendLogs(fmt.Sprintln(s.Text()))
		}
		done <- struct{}{}
	}()

	err = cmd.Start()
	if err != nil {
		logrus.Errorf("Command failed to start: %s", err)
		return "", SetError(err)
	}

	<-done

	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("Command failed: %s", err)
		return "", SetError(err)
	}

	return buffer.String(), err
}

func (c shellCommand) HasHandshake() bool {
	return true
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

func (c shellCommand) Args() []string {
	logrus.Debug("Returning shell command args ", c.opts.Args)
	if c.opts.Args == nil {
		return []string{}
	}
	return c.opts.Args
}

func (c shellCommand) Timeout() time.Duration {
	if c.opts.Timeout == 0 {
		return command.DefaultCommandTimeout
	}
	return c.opts.Timeout
}

func (c shellCommand) Cmd() string {
	return c.opts.Cmd
}

func (c shellCommand) Help() command.CommandHelp {
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
func NewHelp(summary string, args ...string) command.CommandHelp {
	return shellHelp{
		summary,
		append([]string{}, args...),
	}

}
