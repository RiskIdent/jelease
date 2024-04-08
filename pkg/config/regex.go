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
	"regexp"

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

type RegexPattern regexp.Regexp

// Ensure the type implements the interfaces
var _ pflag.Value = &RegexPattern{}
var _ encoding.TextUnmarshaler = &RegexPattern{}
var _ jsonSchemaInterface = RegexPattern{}

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
		Examples: []any{
			"^appVersion: .*",
			"^version: .*",
		},
	}
}
