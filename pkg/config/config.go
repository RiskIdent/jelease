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
	"reflect"

	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/invopop/jsonschema"
)

type Config struct {
	DryRun   bool
	Packages []Package
	GitHub   GitHub
	Jira     Jira
	HTTP     HTTP
	Log      Log
}

func (c Config) TryFindPackage(pkgName string) (Package, bool) {
	for _, pkg := range c.Packages {
		if pkg.Name == pkgName {
			return pkg, true
		}
	}
	return Package{}, false
}

type Package struct {
	Name  string
	Repos []PackageRepo
}

type PackageRepo struct {
	URL     string `jsonschema_extras:"format=uri-reference"`
	Patches []PackageRepoPatch
}

type PackageRepoPatch struct {
	File  string
	Regex *PatchRegex `yaml:",omitempty" json:",omitempty" jsonschema:"oneof_required=regex"`
	YQ    *PatchYQ    `yaml:",omitempty" json:",omitempty" jsonschema:"oneof_required=yq"`
}

type PatchRegex struct {
	Match   *RegexPattern `jsonschema:"required"`
	Replace *Template     `jsonschema:"required"`
}

type PatchYQ struct {
	Expression string `jsonschema:"required"`
}

type GitHub struct {
	URL     *string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	TempDir *string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	Auth    GitHubAuth
	PR      GitHubPR
}

type GitHubAuth struct {
	Type  GitHubAuthType
	Token string
}

type GitHubPR struct {
	Title        *Template
	Description  *Template
	Branch       *Template
	Commit       *Template
	CommitAuthor GitHubCommitAuthor
}

type GitHubCommitAuthor struct {
	Name  *string `jsonschema:"oneof_type=string;null"`
	Email *string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=idn-email"`
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
