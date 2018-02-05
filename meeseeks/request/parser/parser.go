package parser

import (
	"errors"
	"strings"
)

const (
	startState  = "start"
	quotesState = "quotes"
	argState    = "arg"
)

// ErrUnclosedQuoteInCommand means that the command is not correctly escaped
var ErrUnclosedQuoteInCommand = errors.New("Unclosed quote on command")

// Parse parses a command and returns a slice of strings and an error if the command is wrongly built
func Parse(command string) ([]string, error) {
	args := make([]string, 0)
	state := startState
	current := ""
	quote := "\""
	escapeNext := true

	command = strings.TrimSpace(command)

	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == quotesState {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = startState
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' || c == '`' {
			state = quotesState
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = startState
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = argState
			current += string(c)
		}
	}

	if state == quotesState {
		return []string{}, ErrUnclosedQuoteInCommand
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}
