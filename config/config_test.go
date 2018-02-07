package config_test

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/db"
	"github.com/renstrom/dedent"
)

func Test_ConfigurationReading(t *testing.T) {
	defaultColors := config.MessageColors{
		Info:    config.DefaultInfoColorMessage,
		Error:   config.DefaultErrColorMessage,
		Success: config.DefaultSuccessColorMessage,
	}
	defaultDatabase := db.DatabaseConfig{
		Path:    "meeseeks.db",
		Mode:    0600,
		Timeout: 2 * time.Second,
	}
	tt := []struct {
		Name     string
		Content  string
		Expected config.Config
	}{
		{
			"Default configuration",
			"",
			config.Config{
				Colors:   defaultColors,
				Database: defaultDatabase,
				Pool:     20,
			},
		},
		{
			"With messages",
			dedent.Dedent(`
				messages:
				  handshake: ["hallo"]
				`),
			config.Config{
				Messages: map[string][]string{
					"handshake": []string{"hallo"},
				},
				Colors:   defaultColors,
				Database: defaultDatabase,
				Pool:     20,
			},
		},
		{
			"With colors",
			dedent.Dedent(`
				colors:
				  info: "#FFFFFF"
				  success: "#CCCCCC"
				  error: "#000000"
				`),
			config.Config{
				Colors: config.MessageColors{
					Info:    "#FFFFFF",
					Success: "#CCCCCC",
					Error:   "#000000",
				},
				Database: defaultDatabase,
				Pool:     20,
			},
		},
		{
			"With commands",
			dedent.Dedent(`
				commands:
				  something:
				    command: "ssh"
				    authorized: ["someone"]
				    args: ["none"]
				`),
			config.Config{
				Commands: map[string]config.Command{
					"something": config.Command{
						Cmd:  "ssh",
						Args: []string{"none"},
					},
				},
				Colors:   defaultColors,
				Database: defaultDatabase,
				Pool:     20,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := config.New(strings.NewReader(tc.Content))
			if err != nil {
				t.Fatalf("failed to parse configuration: %s", err)
			}
			if !reflect.DeepEqual(tc.Expected, c) {
				t.Fatalf("configuration is not as expected; got %#v instead of %#v", c, tc.Expected)
			}

		})
	}
}

func Test_Errors(t *testing.T) {
	tc := []struct {
		name     string
		reader   io.Reader
		expected string
	}{
		{
			"invalid configuration",
			strings.NewReader("	invalid"),
			"could not parse configuration: yaml: found character that cannot start any token",
		},
		{
			"bad reader",
			badReader{},
			"could not read configuration: bad reader",
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			_, err := config.New(tc.reader)
			if err.Error() != tc.expected {
				t.Fatalf("wrong error, expected %s; got %s", tc.expected, err)
			}
		})
	}
}

type badReader struct {
}

func (badReader) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("bad reader")
}
