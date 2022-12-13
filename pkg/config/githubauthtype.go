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

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

type GitHubAuthType string

const (
	GitHubAuthTypePAT GitHubAuthType = "pat"
)

func _() {
	// Ensure the type implements the interfaces
	f := GitHubAuthTypePAT
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
	var _ jsonSchemaInterface = f
}

func (f GitHubAuthType) String() string {
	return string(f)
}

func (f *GitHubAuthType) Set(value string) error {
	switch GitHubAuthType(value) {
	case GitHubAuthTypePAT:
		*f = GitHubAuthTypePAT
	default:
		return fmt.Errorf("unknown auth type: %q, must be one of: pat, token", value)
	}
	return nil
}

func (f *GitHubAuthType) Type() string {
	return "auth"
}

func (f *GitHubAuthType) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

func (GitHubAuthType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "GitHub auth type",
		Enum: []any{
			GitHubAuthTypePAT,
		},
	}
}
