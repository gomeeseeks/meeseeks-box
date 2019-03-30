package template_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
	"gitlab.com/yakshaving.art/meeseeks-box/text/template"
)

func Test_Templating(t *testing.T) {
	tt := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple static",
			template: "hello!",
			data:     map[string]interface{}{},
			expected: "hello!",
		},
		{
			name:     "simple static",
			template: "{{ \"hello!\" }}",
			data:     map[string]interface{}{},
			expected: "hello!",
		},
		{
			name:     "with one single value",
			template: "value: {{ .Value }}",
			data: map[string]interface{}{
				"Value": "a value",
			},
			expected: "value: a value",
		},
		{
			name:     "with a map of values",
			template: "list: {{ range $key, $element := .Values }}{{ $key }}={{ $element }} {{ end }}",
			data: map[string]interface{}{
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
		payload  map[string]interface{}
		expected string
	}{
		{
			name:     "no value",
			template: "{{ AnyValue \"Value\" . }}",
			payload:  map[string]interface{}{},
			expected: "failed to execute template no value: template: no value:1:3: executing \"no value\" at <AnyValue \"Value\" .>: error calling AnyValue: Value is not loaded in the payload",
		},
		{
			name:     "wrong value type",
			template: "{{ AnyValue \"Value\" . }}",
			payload:  map[string]interface{}{"Value": "something"},
			expected: "failed to execute template wrong value type: template: wrong value type:1:3: executing \"wrong value type\" at <AnyValue \"Value\" .>: error calling AnyValue: Value is not a string slice",
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
		template.Handshake: "hello!",
	}).Build()
	out, err := templates.Render(template.Handshake,
		map[string]interface{}{
			"user": "myuser",
		},
	)
	mocks.Must(t, "can't render changed handshake template", err)
	mocks.AssertEquals(t, "hello!", out)
}

func Test_ChangingMessages(t *testing.T) {
	templates := template.NewBuilder().WithMessages(map[string][]string{
		template.Handshake: {"yo!"},
	}).Build()
	out, err := templates.Render(template.Handshake, map[string]interface{}{})
	mocks.Must(t, "can't render changed handshake template", err)
	mocks.AssertEquals(t, "yo!", out)
}

func Test_DefaultTemplates(t *testing.T) {
	templates := template.NewBuilder().Build()

	handshakeMatcher, err := regexp.Compile(fmt.Sprintf("^(%s)$", strings.Join(template.DefaultHandshakeMessages, "|")))
	mocks.Must(t, "can't compile default hanshake matcher", err)

	successMatcher, err := regexp.Compile(fmt.Sprintf("^<@myself> (%s)$", strings.Join(template.DefaultSuccessMessages, "|")))
	mocks.Must(t, "can't compile default success matcher", err)

	successWithOutputMatcher, err := regexp.Compile(fmt.Sprintf("(?m)<@myself> (%s)[\\n `]*something happened", strings.Join(template.DefaultSuccessMessages, "|")))
	mocks.Must(t, "can't compile default success with output matcher", err)

	failureMatcher, err := regexp.Compile(fmt.Sprintf("^<@myself> (%s) :disappointed: it failed$", strings.Join(template.DefaultFailedMessages, "|")))
	mocks.Must(t, "can't compile default failure matcher", err)

	failureWithOutputMatcher, err := regexp.Compile(fmt.Sprintf("(?m)<@myself> (%s) :disappointed: it failed[\\n `]*some output", strings.Join(template.DefaultFailedMessages, "|")))
	mocks.Must(t, "can't compile default failure with output matcher", err)

	unknownCommandMatcher, err := regexp.Compile(fmt.Sprintf("<@myself> (%s) mycommand", strings.Join(template.DefaultUnknownCommandMessages, "|")))
	mocks.Must(t, "can't compile default unknown command matcher", err)

	unauthorizedCommandMatcher, err := regexp.Compile(fmt.Sprintf("<@myself> (%s) mycommand: just because", strings.Join(template.DefaultUnauthorizedMessages, "|")))
	mocks.Must(t, "can't compile default unauthorized command matcher", err)

	tt := []struct {
		name     string
		renderer func() (string, error)
		matcher  *regexp.Regexp
	}{
		{
			name: "Handshake",
			renderer: func() (string, error) {
				return templates.Render(template.Handshake, map[string]interface{}{})
			},
			matcher: handshakeMatcher,
		},
		{
			name: "Simple success",
			renderer: func() (string, error) {
				return templates.Render(template.Success, map[string]interface{}{
					"userlink": "<@myself>",
				})
			},
			matcher: successMatcher,
		},
		{
			name: "Success with output",
			renderer: func() (string, error) {
				return templates.Render(template.Success, map[string]interface{}{
					"userlink": "<@myself>",
					"output":   "something happened",
				})
			},
			matcher: successWithOutputMatcher,
		},
		{
			name: "Simple Failure",
			renderer: func() (string, error) {
				return templates.Render(template.Failure, map[string]interface{}{
					"userlink": "<@myself>",
					"error":    "it failed",
				})
				// return templates.RenderFailure("<@myself>", "it failed", "")
			},
			matcher: failureMatcher,
		},
		{
			name: "Failure with output",
			renderer: func() (string, error) {
				return templates.Render(template.Failure, map[string]interface{}{
					"userlink": "<@myself>",
					"error":    "it failed",
					"output":   "some output",
				})
			},
			matcher: failureWithOutputMatcher,
		},
		{
			name: "Unknown command",
			renderer: func() (string, error) {
				return templates.Render(template.UnknownCommand, map[string]interface{}{
					"userlink": "<@myself>",
					"command":  "mycommand",
				})
			},
			matcher: unknownCommandMatcher,
		},
		{
			name: "Unauthorized command",
			renderer: func() (string, error) {
				return templates.Render(template.Unauthorized, map[string]interface{}{
					"userlink": "<@myself>",
					"command":  "mycommand",
					"error":    "just because",
				})
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
