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
	"net/url"

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

type URL url.URL

// Ensure the type implements the interfaces
var _ pflag.Value = &URL{}
var _ encoding.TextUnmarshaler = &URL{}
var _ jsonSchemaInterface = URL{}

func (u *URL) URL() *url.URL {
	return (*url.URL)(u)
}

func (u *URL) String() string {
	return u.URL().String()
}

func (u *URL) Set(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	*u = URL(*parsed)
	return nil
}

func (URL) Type() string {
	return "url"
}

func (u *URL) UnmarshalText(text []byte) error {
	return u.Set(string(text))
}

func (u *URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (URL) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   "string",
		Title:  "URL",
		Format: "uri",
		Examples: []any{
			"http://localhost:8080",
			"https://example.com",
		},
	}
}
