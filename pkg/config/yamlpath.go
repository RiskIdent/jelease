// SPDX-FileCopyrightText: 2024 Risk.Ident GmbH <contact@riskident.com>
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

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
)

type YAMLPathPattern struct {
	YAMLPath *yamlpath.Path
	Source   string
}

// Ensure the type implements the interfaces
var _ pflag.Value = &YAMLPathPattern{}
var _ encoding.TextUnmarshaler = &YAMLPathPattern{}
var _ jsonSchemaInterface = YAMLPathPattern{}

func (r *YAMLPathPattern) String() string {
	return r.Source
}

func (r *YAMLPathPattern) Set(value string) error {
	path, err := yamlpath.NewPath(value)
	if err != nil {
		return err
	}
	r.YAMLPath = path
	r.Source = value
	return nil
}

func (r *YAMLPathPattern) Type() string {
	return "yamlpath"
}

func (r *YAMLPathPattern) UnmarshalText(text []byte) error {
	return r.Set(string(text))
}

func (r *YAMLPathPattern) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (YAMLPathPattern) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "YAML-Path pattern",
		Examples: []any{
			".appVersion",
			".version",
			"$..spec.containers[*].image",
		},
	}
}
