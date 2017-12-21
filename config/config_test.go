package config_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

func Test_ConfigurationReading(t *testing.T) {
	tt := []struct {
		Name     string
		Content  string
		Expected config.Config
	}{
		{
			"Basic",
			"",
			config.Config{},
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
			},
		},
		{
			"With messages",
			dedent.Dedent(`
				commands:
				  something:
				    command: "ssh"
				`),
			config.Config{
				Commands: map[string]config.Command{
					"something": config.Command{
						Cmd: "ssh",
					},
				},
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
				t.Fatalf("configuration is not as expected; got %+v instead of %+v", c, tc.Expected)
			}

		})
	}

}
