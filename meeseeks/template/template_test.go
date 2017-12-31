package template_test

import (
	"strings"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
)

func Test_Templating(t *testing.T) {
	tt := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "simple static",
			template: "hello!",
			data:     struct{}{},
			expected: "hello!",
		},
		{
			name:     "simple static",
			template: "{{ \"hello!\" }}",
			data:     struct{}{},
			expected: "hello!",
		},
		{
			name:     "with one single value",
			template: "value: {{ .Value }}",
			data: map[string]string{
				"Value": "a value",
			},
			expected: "value: a value",
		},
		{
			name:     "with a map of values",
			template: "list: {{ range $key, $element := .Values }}{{ $key }}={{ $element }} {{ end }}",
			data: map[string]map[string]string{
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
			tmpl, err := template.New(tc.name, tc.template)
			if err != nil {
				t.Fatalf("could not create template %s: %s", tc.name, err)
			}
			out, err := tmpl.Render(tc.data)
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
	_, err := template.New("invalid", "{{ invalid }}")

	expected := "could not parse template invalid: template: invalid:1: function \"invalid\" not defined"
	if err.Error() != expected {
		t.Fatalf("wrong error, expected %s; got %s", expected, err)
	}
}

func Test_InvalidData(t *testing.T) {
	tmpl, _ := template.New("invalid-data", "{{ .Value }}")

	_, err := tmpl.Render(struct{}{})
	expected := "template: invalid-data:1:3: executing \"invalid-data\" at <.Value>: can't evaluate field Value in type struct {}"
	if !strings.HasSuffix(err.Error(), expected) {
		t.Fatalf("wrong error, expected tail of %s; got %s", expected, err)
	}
}

func Test_DefaultTemplates_Handshake(t *testing.T) {
	templates := template.DefaultTemplates()
	data := template.TemplateData{
		"Handshake": []string{"hello!", "how you doing?!"},
	}
	out, err := templates.Handshake.Render(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Error(out)
}

func Test_DefaultTemplates_Success(t *testing.T) {
	templates := template.DefaultTemplates()
	data := template.TemplateData{
		"Success": []string{"Done!", "All done!"},
		"User":    "myself",
	}
	out, err := templates.Success.Render(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Error(out)
}
