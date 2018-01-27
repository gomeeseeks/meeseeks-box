package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/jobs"
	"github.com/pcarranza/meeseeks-box/jobs/logs"
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
	Help          string
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
func (c shellCommand) Execute(job jobs.Job) (string, error) {
	cmdArgs := append(c.Args(), job.Request.Args...)

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.Timeout())
	defer cancelFunc()

	cmd := exec.CommandContext(ctx, c.Cmd(), cmdArgs...)

	var wg sync.WaitGroup
	wg.Add(2) // stdout and stderr

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

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
	outTeeReader := io.TeeReader(outReader, buffer)
	outScanner := bufio.NewScanner(outTeeReader)

	go func() {
		for outScanner.Scan() {
			line := fmt.Sprintln(outScanner.Text())
			AppendLogs(line)
		}
		wg.Done()
	}()

	errScanner := bufio.NewScanner(errReader)
	go func() {
		for errScanner.Scan() {
			line := fmt.Sprintln(errScanner.Text())
			AppendLogs(line)
		}
		wg.Done()
	}()

	err = cmd.Start()
	if err != nil {
		logrus.Errorf("Command failed to start: %s", err)
		return "", SetError(err)
	}

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		logrus.Errorf("Command failed to execute: %s", err)
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

func (c shellCommand) Help() string {
	return c.opts.Help
}

func (c shellCommand) Record() bool {
	return true
}
