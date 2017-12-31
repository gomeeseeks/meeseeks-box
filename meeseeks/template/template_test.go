package template_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
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

func Test_DefaultTemplates(t *testing.T) {
	templates := template.DefaultTemplates()
	tt := []struct {
		name     string
		payload  template.Payload
		renderer template.Renderer
		expected string
	}{
		{
			name: "Handshake",
			payload: template.Payload{
				"handshake": []string{"hello!"},
			},
			renderer: templates.Handshake,
			expected: "hello!",
		},
		{
			name: "Simple success",
			payload: template.Payload{
				"success": []string{"wooot!"},
				"user":    "<@myself>",
			},
			renderer: templates.Success,
			expected: "<@myself> wooot!",
		},
		{
			name: "Success with output",
			payload: template.Payload{
				"success": []string{"wooot!"},
				"user":    "<@myself>",
				"output":  "something happened\n",
			},
			renderer: templates.Success,
			expected: "<@myself> wooot!\n\nOutput:\n```\nsomething happened\n```",
		},
		{
			name: "Simple Failure",
			payload: template.Payload{
				"failed": []string{"bummer"},
				"user":   "<@myself>",
				"error":  "it failed",
			},
			renderer: templates.Failure,
			expected: "<@myself> bummer :disappointed:: it failed",
		},
		{
			name: "Failure with output",
			payload: template.Payload{
				"failed": []string{"bummer"},
				"user":   "<@myself>",
				"error":  "it failed",
				"output": "some output\n",
			},
			renderer: templates.Failure,
			expected: "<@myself> bummer :disappointed:: it failed\n\nOutput:\n```\nsome output\n```",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.renderer.Render(tc.payload)
			if err != nil {
				t.Fatalf("Render failed with error: %s", err)
			}
			if out != tc.expected {
				t.Fatalf("Wrong render. Expected %s; Got %s", tc.expected, out)
			}
		})
	}
}
