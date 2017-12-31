package meeseeks_test

import (
	"reflect"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
)

func Test_ParsingCommandsCorrectly(t *testing.T) {
	tt := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "double quote echo",
			command:  `echo "this is a message"`,
			expected: []string{"echo", "this is a message"},
		},
		{
			name:     "single quote echo",
			command:  "echo 'this is a message'",
			expected: []string{"echo", "this is a message"},
		},
		{
			name:     "multiarg echo",
			command:  "echo this is a message",
			expected: []string{"echo", "this", "is", "a", "message"},
		},
		{
			name:     "with options and such",
			command:  "echo -n    'this is a message'",
			expected: []string{"echo", "-n", "this is a message"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			args, err := meeseeks.ParseCommand(tc.command)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(args, tc.expected) {
				t.Fatalf("Args are wrong, got: %+v; expecting: %+v", args, tc.expected)
			}
		})
	}
}

func Test_InvalidCommand(t *testing.T) {
	_, err := meeseeks.ParseCommand("echo 'invalid arg because unclosed quote")
	if err != meeseeks.ErrUnclosedQuoteInCommand {
		t.Fatalf("Got an invalid error, expected %s; got %s", meeseeks.ErrUnclosedQuoteInCommand, err)
	}
}
