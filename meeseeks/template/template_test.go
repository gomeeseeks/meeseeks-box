package template_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

func Test_Templating(t *testing.T) {
	tt := []struct {
		name     string
		template string
		data     template.Payload
		expected string
	}{
		{
			name:     "simple static",
			template: "hello!",
			data:     template.Payload{},
			expected: "hello!",
		},
		{
			name:     "simple static",
			template: "{{ \"hello!\" }}",
			data:     template.Payload{},
			expected: "hello!",
		},
		{
			name:     "with one single value",
			template: "value: {{ .Value }}",
			data: template.Payload{
				"Value": "a value",
			},
			expected: "value: a value",
		},
		{
			name:     "with a map of values",
			template: "list: {{ range $key, $element := .Values }}{{ $key }}={{ $element }} {{ end }}",
			data: template.Payload{
				"Values": map[string]string{
					"first":  "one",
					"second": "two",
				},
			},
			expected: "list: first=one second=two ",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			renderer, err := template.New(tc.name, tc.template)
			if err != nil {
				t.Fatalf("could not create template %s: %s", tc.name, err)
			}
			out, err := renderer.Render(tc.data)
			if err != nil {
				t.Fatalf("failed to render template %s: %s", tc.name, err)
			}
			if tc.expected != out {
				t.Fatalf("output is wrong, expected: %s; got: %s", tc.expected, out)
			}
		})
	}
}

func Test_InvalidTemplate(t *testing.T) {
	tt := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "invalid template",
			template: "{{ invalid }}",
			expected: "could not parse template invalid template: template: invalid template:1: function \"invalid\" not defined",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New(tc.name, tc.template)
			if err == nil {
				t.Fatalf("We should have failed with an error, but instead got and output: %+v", tmpl)
			}

			if err.Error() != tc.expected {
				t.Fatalf("wrong error, expected %s; got %s", tc.expected, err)
			}
		})
	}
}

func Test_InvalidData(t *testing.T) {
	tt := []struct {
		name     string
		template string
		payload  template.Payload
		expected string
	}{
		{
			name:     "no value",
			template: "{{ AnyValue \"Value\" . }}",
			payload:  template.Payload{},
			expected: "failed to execute template no value: template: no value:1:3: executing \"no value\" at <AnyValue \"Value\" .>: error calling AnyValue: ERROR: Value is not loaded in the payload",
		},
		{
			name:     "wrong value type",
			template: "{{ AnyValue \"Value\" . }}",
			payload:  template.Payload{"Value": "something"},
			expected: "failed to execute template wrong value type: template: wrong value type:1:3: executing \"wrong value type\" at <AnyValue \"Value\" .>: error calling AnyValue: ERROR: Value is not a string slice",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New(tc.name, tc.template)
			if err != nil {
				t.Fatalf("Failed to parse the template %s: %s", tc.name, err)
			}

			out, err := tmpl.Render(tc.payload)
			if err == nil {
				t.Fatalf("We should have failed with an error, but instead got and output: %s", out)
			}

			if err.Error() != tc.expected {
				t.Fatalf("wrong error, expected %s; got %s", tc.expected, err)
			}
		})
	}
}

func Test_ChangingTemplate(t *testing.T) {
	templates := template.NewBuilder().WithTemplates(map[string]string{
		template.HandshakeKey: "hello!",
	}).Build()
	out, err := templates.RenderHandshake("myuser")
	stubs.Must(t, "can't render changed handshake tempalte", err)
	stubs.AssertEquals(t, "hello!", out)
}

func Test_ChangingMessages(t *testing.T) {
	templates := template.NewBuilder().WithMessages(map[string][]string{
		template.HandshakeKey: []string{"yo!"},
	}).Build()
	out, err := templates.RenderHandshake("myuser")
	stubs.Must(t, "can't render changed handshake tempalte", err)
	stubs.AssertEquals(t, "yo!", out)
}
func Test_DefaultTemplates(t *testing.T) {
	templates := template.NewBuilder().Build()

	handshakeMatcher, err := regexp.Compile(fmt.Sprintf("^(%s)$", strings.Join(template.DefaultHandshakeMessages, "|")))
	stubs.Must(t, "can't compile default hanshake matcher", err)

	successMatcher, err := regexp.Compile(fmt.Sprintf("^<@myself> (%s)$", strings.Join(template.DefaultSuccessMessages, "|")))
	stubs.Must(t, "can't compile default success matcher", err)

	successWithOutputMatcher, err := regexp.Compile(fmt.Sprintf("(?m)<@myself> (%s)[\\n `]*something happened", strings.Join(template.DefaultSuccessMessages, "|")))
	stubs.Must(t, "can't compile default success with output matcher", err)

	failureMatcher, err := regexp.Compile(fmt.Sprintf("^<@myself> (%s) :disappointed: it failed$", strings.Join(template.DefaultFailedMessages, "|")))
	stubs.Must(t, "can't compile default failure matcher", err)

	failureWithOutputMatcher, err := regexp.Compile(fmt.Sprintf("(?m)<@myself> (%s) :disappointed: it failed[\\n `]*some output", strings.Join(template.DefaultFailedMessages, "|")))
	stubs.Must(t, "can't compile default failure with output matcher", err)

	unknownCommandMatcher, err := regexp.Compile(fmt.Sprintf("<@myself> (%s) mycommand", strings.Join(template.DefaultUnknownCommandMessages, "|")))
	stubs.Must(t, "can't compile default unknown command matcher", err)

	unauthorizedCommandMatcher, err := regexp.Compile(fmt.Sprintf("<@myself> (%s) mycommand", strings.Join(template.DefaultUnauthorizedMessages, "|")))
	stubs.Must(t, "can't compile default unauthorized command matcher", err)

	tt := []struct {
		name     string
		renderer func() (string, error)
		expected string
		matcher  *regexp.Regexp
	}{
		{
			name: "Handshake",
			renderer: func() (string, error) {
				return templates.RenderHandshake("my user")
			},
			matcher: handshakeMatcher,
		},
		{
			name: "Simple success",
			renderer: func() (string, error) {
				return templates.RenderSuccess("<@myself>", "")
			},
			matcher: successMatcher,
		},
		{
			name: "Success with output",
			renderer: func() (string, error) {
				return templates.RenderSuccess("<@myself>", "something happened")
			},
			matcher: successWithOutputMatcher,
		},
		{
			name: "Simple Failure",
			renderer: func() (string, error) {
				return templates.RenderFailure("<@myself>", "it failed", "")
			},
			matcher: failureMatcher,
		},
		{
			name: "Failure with output",
			renderer: func() (string, error) {
				return templates.RenderFailure("<@myself>", "it failed", "some output")
			},
			matcher: failureWithOutputMatcher,
		},
		{
			name: "Unknown command",
			renderer: func() (string, error) {
				return templates.RenderUnknownCommand("<@myself>", "mycommand")
			},
			matcher: unknownCommandMatcher,
		},
		{
			name: "Unauthorized command",
			renderer: func() (string, error) {
				return templates.RenderUnauthorizedCommand("<@myself>", "mycommand")
			},
			matcher: unauthorizedCommandMatcher,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.renderer()
			if err != nil {
				t.Fatalf("Render failed with error: %s", err)
			}
			if !tc.matcher.MatchString(out) {
				t.Fatalf("Bad message, expected %s; got %s", tc.matcher, out)
			}
		})
	}
}
