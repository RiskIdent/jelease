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
	"encoding"
	"fmt"
	"regexp"
	"text/template"

	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

type Config struct {
	DryRun   bool
	Packages []Package
	Jira     Jira
	HTTP     HTTP
	Log      Log
}

/*
packages:
  - name: foobar
    patches:
      - repo: https://github.com/RiskIdent/jelease
        file: go.mod
        match: (github.com/joho/godotenv) v.*
        replace: "{{ .Groups[1] }} {{ .Version }}"
*/

type Package struct {
	Name    string
	Patches []PackagePatch
}

type PackagePatch struct {
	Repo    string `jsonschema_extras:"format=uri-reference"`
	File    string
	Match   *RegexPattern
	Replace *Template
}

type RegexPattern regexp.Regexp

// Ensure the type implements the interfaces
var _ pflag.Value = &RegexPattern{}
var _ encoding.TextUnmarshaler = &RegexPattern{}

func (r *RegexPattern) Regexp() *regexp.Regexp {
	return (*regexp.Regexp)(r)
}

func (r *RegexPattern) String() string {
	return r.Regexp().String()
}

func (r *RegexPattern) Set(value string) error {
	parsed, err := regexp.Compile(value)
	if err != nil {
		return err
	}
	*r = RegexPattern(*parsed)
	return nil
}

func (r *RegexPattern) Type() string {
	return "regex"
}

func (r *RegexPattern) UnmarshalText(text []byte) error {
	return r.Set(string(text))
}

func (r *RegexPattern) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (RegexPattern) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   "string",
		Title:  "Regular Expression pattern (regex)",
		Format: "regex",
	}
}

type Template template.Template

// Ensure the type implements the interfaces
var _ pflag.Value = &Template{}
var _ encoding.TextUnmarshaler = &Template{}

func (t *Template) Template() *template.Template {
	return (*template.Template)(t)
}

func (t *Template) String() string {
	return t.Template().Root.String()
}

func (t *Template) Set(value string) error {
	parsed, err := template.New("").Parse(value)
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
	}
}

type Jira struct {
	URL            string `jsonschema_extras:"format=uri"`
	SkipCertVerify bool
	Auth           JiraAuth
	Issue          JiraIssue
}

type JiraAuth struct {
	Type  JiraAuthType
	Token string
	User  string
}

type JiraAuthType string

const (
	JiraAuthTypePAT   JiraAuthType = "pat"
	JiraAuthTypeToken JiraAuthType = "token"
)

// Jira Ticket type
type JiraIssue struct {
	Labels                 []string
	Status                 string
	Description            string
	Type                   string
	Project                string
	ProjectNameCustomField uint
}

type HTTP struct {
	Port uint16
}

type Log struct {
	Format LogFormat
	Level  LogLevel
}

type LogLevel zerolog.Level

func _() {
	// Ensure the type implements the interfaces
	l := LogLevel(zerolog.DebugLevel)
	var _ pflag.Value = &l
	var _ encoding.TextUnmarshaler = &l
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	return l.Set(string(text))
}

func (l LogLevel) MarshalText() ([]byte, error) {
	return zerolog.Level(l).MarshalText()
}

func (l LogLevel) String() string {
	return zerolog.Level(l).String()
}

func (l *LogLevel) Set(value string) error {
	lvl, err := zerolog.ParseLevel(value)
	if err != nil {
		return err
	}
	*l = LogLevel(lvl)
	return nil
}

func (l *LogLevel) Type() string {
	return "level"
}

func (LogLevel) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging level",
		Enum: []any{
			LogLevel(zerolog.DebugLevel),
			LogLevel(zerolog.InfoLevel),
			LogLevel(zerolog.WarnLevel),
			LogLevel(zerolog.ErrorLevel),
			LogLevel(zerolog.FatalLevel),
			LogLevel(zerolog.PanicLevel),
			LogLevel(zerolog.Disabled),
			LogLevel(zerolog.TraceLevel),
		},
	}
}

type LogFormat string

const (
	LogFormatPretty LogFormat = "pretty"
	LogFormatJSON   LogFormat = "json"
)

func _() {
	// Ensure the type implements the interfaces
	f := LogFormatJSON
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
}

func (f LogFormat) String() string {
	return string(f)
}

func (f *LogFormat) Set(value string) error {
	switch LogFormat(value) {
	case LogFormatPretty:
		*f = LogFormatPretty
	case LogFormatJSON:
		*f = LogFormatJSON
	default:
		return fmt.Errorf("unknown log format: %q, must be one of: pretty, json", value)
	}
	return nil
}

func (f *LogFormat) Type() string {
	return "format"
}

func (f *LogFormat) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

func (LogFormat) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging format",
		Enum: []any{
			LogFormatPretty,
			LogFormatJSON,
		},
	}
}
