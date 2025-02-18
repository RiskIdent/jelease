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

package config

import (
	"bytes"
	"encoding"
	"fmt"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/templatefuncs"
	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

// Template is a parsed Go [text/template] string, that has additional
// encoders implemented so it can be used in config files and CLI flags.
type Template template.Template

func NewTemplate(s string) (*Template, error) {
	var t Template
	if err := t.Set(s); err != nil {
		return nil, err
	}
	return &t, nil
}

func MustTemplate(s string) *Template {
	tmpl, err := NewTemplate(s)
	if err != nil {
		panic(fmt.Errorf("MustTemplate(%q): %w", s, err))
	}
	return tmpl
}

// TemplateContext is the common data passed into templates when executing them.
type TemplateContext struct {
	Package            string
	PackageDescription string
	Version            string
	JiraIssue          string
}

// Ensure the type implements the interfaces
var (
	_ pflag.Value              = &Template{}
	_ encoding.TextUnmarshaler = &Template{}
	_ jsonSchemaInterface      = Template{}
)

func (t *Template) Template() *template.Template {
	return (*template.Template)(t)
}

func (t *Template) String() string {
	return t.Template().Root.String()
}

func (t *Template) Set(value string) error {
	parsed, err := template.New("").Funcs(templatefuncs.FuncsMap).Parse(value)
	if err != nil {
		return err
	}
	*t = Template(*parsed)
	return nil
}

func (Template) Type() string {
	return "template"
}

func (t *Template) UnmarshalText(text []byte) error {
	return t.Set(string(text))
}

func (t *Template) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (Template) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Go template",
		Examples: []any{
			"{{ .Version }}",
			"version: {{ .Version | trimPrefix \"v\" }}",
			"version: {{ index .Groups 1 | versionBump \"0.0.1\" }}",
		},
	}
}

func (t *Template) Render(data any) (string, error) {
	if t == nil {
		return "", nil
	}
	var buf bytes.Buffer
	if err := t.Template().Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
