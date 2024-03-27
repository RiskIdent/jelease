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
	"net/url"
	"reflect"
	"strings"

	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/invopop/jsonschema"
)

var (
	redacted    = "--redacted--"
	redactedPtr = &redacted
)

type Config struct {
	DryRun      bool `yaml:"dryRun"`
	Packages    []Package
	GitHub      GitHub
	Jira        Jira
	NewReleases NewReleases
	HTTP        HTTP
	Log         Log
}

func (c Config) TryFindPackage(pkgName string) (Package, bool) {
	normalized := NormalizePackageName(pkgName)
	for _, pkg := range c.Packages {
		if pkg.NormalizedName() == normalized {
			return pkg, true
		}
	}
	return Package{}, false
}

func (c Config) Censored() Config {
	c.GitHub = c.GitHub.Censored()
	c.Jira = c.Jira.Censored()
	c.HTTP = c.HTTP.Censored()
	return c
}

type Package struct {
	Name  string
	Repos []PackageRepo
}

func (p Package) NormalizedName() string {
	return NormalizePackageName(p.Name)
}

func NormalizePackageName(pkgName string) string {
	return strings.ReplaceAll(pkgName, "/", "-")
}

type PackageRepo struct {
	URL     string `jsonschema_extras:"format=uri-reference"`
	Patches []PackageRepoPatch
}

type PackageRepoPatch struct {
	Regex *PatchRegex `yaml:",omitempty" json:",omitempty" jsonschema:"oneof_required=regex"`
	YAML  *PatchYAML  `yaml:",omitempty" json:",omitempty" jsonschema:"oneof_required=yaml"`
	Helm  *PatchHelm  `yaml:",omitempty" json:",omitempty" jsonschema:"oneof_required=helm"`
}

type PatchRegex struct {
	File    string        `jsonschema:"required"`
	Match   *RegexPattern `jsonschema:"required"`
	Replace *Template     `jsonschema:"required"`
}

type PatchYAML struct {
	File       string           `jsonschema:"required"`
	YAMLPath   *YAMLPathPattern `yaml:"yamlPath" jsonschema:"required"`
	Replace    *Template        `jsonschema:"required"`
	MaxMatches int              `yaml:"maxMatches,omitempty" jsonschema:"minimum=0"`
	Indent     int              `yaml:",omitempty" jsonschema:"minimum=0"`
}

type PatchHelm struct {
	Chart            *Template `jsonschema:"required,default=.,example=charts/jelease"`
	DependencyUpdate bool      `yaml:"dependencyUpdate,omitempty"`
}

type GitHub struct {
	URL     *string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	TempDir *string `yaml:"tempDir" jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	Auth    GitHubAuth
	PR      GitHubPR
}

func (gh GitHub) Censored() GitHub {
	gh.Auth = gh.Auth.Censored()
	return gh
}

type GitHubAuth struct {
	Type  GitHubAuthType
	Token *string `yaml:",omitempty" jsonschema:"oneof_type=string;null"`
	App   GitHubAuthApp
}

func (a GitHubAuth) Censored() GitHubAuth {
	if a.Token != nil {
		a.Token = redactedPtr
	}
	a.App = a.App.Censored()
	return a
}

type GitHubAuthApp struct {
	ID             int64
	PrivateKeyPath *string           `yaml:"privateKeyPath,omitempty" jsonschema:"oneof_type=string;null,oneof_required=privateKeyPath"`
	PrivateKeyPEM  *RSAPrivateKeyPEM `yaml:"privateKeyPem,omitempty" jsonschema:"oneof_required=privateKeyPem"`
}

func (a GitHubAuthApp) Censored() GitHubAuthApp {
	if a.PrivateKeyPath != nil {
		a.PrivateKeyPath = redactedPtr
	}
	if a.PrivateKeyPEM != nil {
		a.PrivateKeyPEM = &RSAPrivateKeyPEM{
			pem: []byte(redacted),
		}
	}
	return a
}

type GitHubPR struct {
	Title       *Template
	Description *Template
	Branch      *Template
	Commit      *Template
	Committer   GitHubCommitter
}

type GitHubCommitter struct {
	Name  *string `jsonschema:"oneof_type=string;null"`
	Email *string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=idn-email"`
}

type Jira struct {
	URL            string `jsonschema_extras:"format=uri"`
	SkipCertVerify bool   `yaml:"skipCertVerify"`
	Auth           JiraAuth
	Issue          JiraIssue
}

func (j Jira) Censored() Jira {
	j.Auth = j.Auth.Censored()
	return j
}

type JiraAuth struct {
	Type  JiraAuthType
	Token string
	User  string
}

func (a JiraAuth) Censored() JiraAuth {
	if a.Token != "" {
		a.Token = redacted
	}
	if a.User != "" {
		a.User = redacted
	}
	return a
}

// Jira Ticket type
type JiraIssue struct {
	Labels                 []string
	Status                 string
	Description            string
	Type                   string
	Project                string
	ProjectNameCustomField uint `yaml:"projectNameCustomField"`

	// PRDeferredCreation means Jelease will send a link to where user can
	// manually trigger the PR creation, instead of creating it automatically.
	PRDeferredCreation bool `yaml:"prDeferredCreation"`

	Comments JiraIssueComments
}

type JiraIssueComments struct {
	UpdatedIssue       *Template `yaml:"updatedIssue"`
	NoConfig           *Template `yaml:"noConfig"`
	NoPatches          *Template `yaml:"noPatches"`
	PRCreated          *Template `yaml:"prCreated"`
	PRFailed           *Template `yaml:"prFailed"`
	PRDeferredCreation *Template `yaml:"prDeferredCreation"`
}

type HTTP struct {
	Port      uint16
	PublicURL *URL `yaml:"publicUrl"`
}

func (h HTTP) Censored() HTTP {
	if h.PublicURL != nil {
		if h.PublicURL.User != nil {
			u := *h.PublicURL
			u.User = url.User(redacted)
			h.PublicURL = &u
		}
	}
	return h
}

type Log struct {
	Format LogFormat
	Level  LogLevel
}

type jsonSchemaInterface interface {
	JSONSchema() *jsonschema.Schema
}

func Schema() *jsonschema.Schema {
	r := new(jsonschema.Reflector)
	r.KeyNamer = util.ToCamelCase
	r.Namer = func(t reflect.Type) string {
		return util.ToCamelCase(t.Name())
	}
	r.RequiredFromJSONSchemaTags = true
	s := r.Reflect(&Config{})
	s.ID = "https://github.com/RiskIdent/jelease/raw/main/jelease.schema.json"
	return s
}
