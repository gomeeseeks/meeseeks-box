package parser_test

import (
	"reflect"
	"testing"

	parser "gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/command/parser"
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
			name:     "escaping escape chars",
			command:  "echo 'this is a \\ message'",
			expected: []string{"echo", "this is a \\ message"},
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
			args, err := parser.Parse(tc.command)
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
	_, err := parser.Parse("echo 'invalid arg because unclosed quote")
	if err != parser.ErrUnclosedQuoteInCommand {
		t.Fatalf("Got an invalid error, expected %s; got %s", parser.ErrUnclosedQuoteInCommand, err)
	}
}
