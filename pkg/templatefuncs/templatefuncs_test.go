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

package templatefuncs

import (
	"bytes"
	"testing"
	"text/template"
)

func TestTemplate(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		want string
		data any
	}{
		{
			name: "regexReplaceAll",
			tmpl: `{{"abbaba" | regexReplaceAll "b+" "."}}`,
			want: "a.a.a",
		},
		{
			name: "regexMatch",
			tmpl: `{{"abba" | regexMatch "b*"}}`,
			want: "true",
		},
		{
			name: "int_fromString",
			tmpl: `{{int "36"}}`,
			want: "36",
		},
		{
			name: "int_fromInt",
			tmpl: `{{int 36}}`,
			want: "36",
		},
		{
			name: "int_fromFloat",
			tmpl: `{{int 36.05}}`,
			want: "36",
		},
		{
			name: "float64_fromString",
			tmpl: `{{float64 "36.05"}}`,
			want: "36.05",
		},
		{
			name: "float64_fromInt",
			tmpl: `{{float64 36}}`,
			want: "36",
		},
		{
			name: "float64_fromFloat",
			tmpl: `{{float64 36.05}}`,
			want: "36.05",
		},
		{
			name: "toPrettyJson",
			tmpl: `{{toPrettyJson .}}`,
			want: `{
  "Name": "Berlin"
}`,
			data: map[string]string{"Name": "Berlin"},
		},
		{
			name: "toJson",
			tmpl: `{{toJson .}}`,
			want: `{"Name":true}`,
			data: map[string]any{"Name": true},
		},
		{
			name: "fromJson",
			tmpl: `{{fromJson .}}`,
			want: `map[Name:Bangladesh]`,
			data: `{"Name": "Bangladesh"}`,
		},
		{
			name: "fromYaml",
			tmpl: `{{fromYaml .}}`,
			want: `map[Name:Bangladesh]`,
			data: `Name: Bangladesh`,
		},
		{
			name: "toYaml",
			tmpl: `{{toYaml .}}`,
			want: `City:
  Name: Bangladesh
`,
			data: map[string]any{
				"City": map[string]any{
					"Name": "Bangladesh",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Template: %s", tc.tmpl)
			tmpl, err := template.New(tc.name).Funcs(FuncsMap).Parse(tc.tmpl)
			if err != nil {
				t.Errorf("error creating template: %s", err)
				return
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, tc.data); err != nil {
				t.Errorf("error: %s", err)
				return
			}
			got := buf.String()
			if got != tc.want {
				t.Errorf("want %q got %q", tc.want, got)
			}
		})
	}
}
