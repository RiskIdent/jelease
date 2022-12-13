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

type JiraAuthType string

const (
	JiraAuthTypePAT   JiraAuthType = "pat"
	JiraAuthTypeToken JiraAuthType = "token"
)

func _() {
	// Ensure the type implements the interfaces
	f := JiraAuthTypePAT
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
	var _ jsonSchemaInterface = f
}

func (f JiraAuthType) String() string {
	return string(f)
}

func (f *JiraAuthType) Set(value string) error {
	switch JiraAuthType(value) {
	case JiraAuthTypePAT:
		*f = JiraAuthTypePAT
	case JiraAuthTypeToken:
		*f = JiraAuthTypeToken
	default:
		return fmt.Errorf("unknown auth type: %q, must be one of: pat, token", value)
	}
	return nil
}

func (f *JiraAuthType) Type() string {
	return "auth"
}

func (f *JiraAuthType) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

func (JiraAuthType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Jira auth type",
		Enum: []any{
			JiraAuthTypePAT,
			JiraAuthTypeToken,
		},
	}
}
