// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

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
	assertTemplateData(t, `{{toPrettyJson . }}`, "{\n  \"Name\": \"Berlin\"\n}", map[string]string{"Name": "Berlin"})
	assertTemplateData(t, `{{toJson .}}`, `{"Name":true}`, map[string]any{"Name": true})
	assertTemplateData(t, `{{fromJson . }}`, `map[Name:Bangladesh]`, `{"Name": "Bangladesh"}`)
	assertTemplateData(t, `{{fromYaml . }}`, `map[Name:Bangladesh]`, `{"Name": "Bangladesh"}`)
	assertTemplateData(t, `{{toYaml . }}`, "Name: Bangladesh\n", map[string]string{"Name": "Bangladesh"})
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
