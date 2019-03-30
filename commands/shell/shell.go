package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/sirupsen/logrus"
)

// New return a new ShellCommand based on the passed in opts
func New(opts meeseeks.CommandOpts) meeseeks.Command {
	return shellCommand{
		opts,
	}
}

type shellCommand struct {
	meeseeks.CommandOpts
}

// Execute implements Command.Execute for the ShellCommand
func (c shellCommand) Execute(ctx context.Context, job meeseeks.Job) (string, error) {
	cmdArgs := append(c.GetArgs(), job.Request.Args...)
	logrus.Debugf("Calling command %s with args %#v", c.GetCmd(), cmdArgs)

	ctx, cancelFunc := context.WithTimeout(ctx, c.GetTimeout())
	defer cancelFunc()

	outputBuffer := bytes.NewBufferString("")

	logW := persistence.LogWriter()

	AppendLogs := func(line string) {

		outputBuffer.WriteString(line)
		outputBuffer.WriteString("\n")

		if e := logW.Append(job.ID, line); e != nil {
			logrus.Errorf("Could not append '%s' to job %d logs: %s", line, job.ID, e)
		}

	}
	SetError := func(err error) error {
		if e := logW.SetError(job.ID, err); e != nil {
			logrus.Errorf("Could set error to job %d: %s", job.ID, e)
		}
		return err
	}

	cmd := exec.CommandContext(ctx, c.GetCmd(), cmdArgs...)
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

	return outputBuffer.String(), err
}
