package cmd

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/config"
)

func TestTemplate(t *testing.T) {
	assertTemplate(t, `{{regexReplaceAll "Jelease regex replace all" "regex" "regexp"}}`, "Jelease regexp replace all")
	assertTemplate(t, `{{regexMatch "Jelease regex replace all" "regex"}}`, "true")
	assertTemplate(t, `{{int "36"}}`, "36")
	assertTemplate(t, `{{float "36.05"}}`, "36.05")
	assertTemplateData(t, `{{toPrettyJson . }}`, "{\n\t\"Name\": \"Berlin\"\n}", map[string]string{"Name": "Berlin"})
	assertTemplateData(t, `{{toJson .}}`, `{"Name":true}`, map[string]any{"Name": true})
	assertTemplateData(t, `{{fromJson . }}`, `map[Name:Bangladesh]`, `{"Name": "Bangladesh"}`)
	assertTemplateData(t, `{{toYaml . }}`, `map[Name:Bangladesh]`, `{"Name": "Bangladesh"}`)
	assertTemplateData(t, `{{fromYaml .}}`, "Name: Bangladesh\n", map[string]string{"Name": "Bangladesh"})
}

func assertTemplate(t *testing.T, templateString string, want string) {
	tmpl, err := template.New("").Funcs(config.FuncsMap).Parse(templateString)
	if err != nil {
		t.Errorf("Template %q: error %q", templateString, err)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Errorf("Template %q: buf error %q", templateString, err)
		return
	}
	got := buf.String()

	if got != want {
		t.Errorf("Template %q: want %q got %q", templateString, want, got)
	}
}

func assertTemplateData(t *testing.T, templateString string, want string, data any) {
	tmpl, err := template.New("").Funcs(config.FuncsMap).Parse(templateString)
	if err != nil {
		t.Errorf("Template %q: error %q", templateString, err)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Errorf("Template %q: buf error %q", templateString, err)
		return
	}
	got := buf.String()

	if got != want {
		t.Errorf("Template %q: want %q got %q", templateString, want, got)
	}
}
