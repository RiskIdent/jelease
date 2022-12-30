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
		// Version functions
		{
			name: "versionBump",
			tmpl: `{{"1.2.3" | versionBump "0.1.0"}}`,
			want: "1.3.0",
		},
		{
			name: "versionBump_2",
			tmpl: `{{"v1.2.3-rc.123" | versionBump "v0.1.0"}}`,
			want: "v1.3.0",
		},

		// Type conversion functions
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

		// List functions
		{
			name: "list",
			tmpl: `{{list 1 true 3.5 "four"}}`,
			want: `[1 true 3.5 four]`,
		},
		{
			name: "first",
			tmpl: `{{first .}}`,
			want: `a`,
			data: []any{"a", "b", "c", "d"},
		},
		{
			name: "rest",
			tmpl: `{{rest .}}`,
			want: `[b c d]`,
			data: []any{"a", "b", "c", "d"},
		},
		{
			name: "last",
			tmpl: `{{last .}}`,
			want: `d`,
			data: []any{"a", "b", "c", "d"},
		},
		{
			name: "initial",
			tmpl: `{{initial .}}`,
			want: `[a b c]`,
			data: []any{"a", "b", "c", "d"},
		},
		{
			name: "append",
			tmpl: `{{append . 1 2}}`,
			want: `[a b 1 2]`,
			data: []any{"a", "b"},
		},
		{
			name: "prepend",
			tmpl: `{{prepend . 1 2}}`,
			want: `[1 2 a b]`,
			data: []any{"a", "b"},
		},
		{
			name: "concat",
			tmpl: `{{concat .A .B}}`,
			want: `[a b c d]`,
			data: map[string]any{
				"A": []any{"a", "b"},
				"B": []any{"c", "d"},
			},
		},
		{
			name: "reverse",
			tmpl: `{{reverse .}}`,
			want: `[d c b a]`,
			data: []any{"a", "b", "c", "d"},
		},

		// Dictionary functions
		{
			name: "dict",
			tmpl: `{{dict "a" 1}}`,
			want: `map[a:1]`,
		},
		{
			name: "get",
			tmpl: `{{get . "a"}}`,
			want: `1`,
			data: map[string]any{"a": 1},
		},
		{
			name: "set",
			tmpl: `{{set . "a" 5}}`,
			want: `map[a:5]`,
			data: map[string]any{"a": 1},
		},
		{
			name: "unset",
			tmpl: `{{unset . "a"}}`,
			want: `map[]`,
			data: map[string]any{"a": 1},
		},
		{
			name: "hasKey",
			tmpl: `{{hasKey . "a"}}`,
			want: `true`,
			data: map[string]any{"a": 1},
		},

		// Encoding functions
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

		// String functions
		{
			name: "trimPrefix",
			tmpl: `{{"www.example.cc" | trimPrefix "w"}}`,
			want: "ww.example.cc",
		},
		{
			name: "trimSuffix",
			tmpl: `{{"www.example.cc" | trimSuffix "c"}}`,
			want: "www.example.c",
		},
		{
			name: "trimSpace",
			tmpl: `{{"  \t www.example.cc \n" | trimSpace}}`,
			want: "www.example.cc",
		},
		{
			name: "trim",
			tmpl: `{{"www.example.cc" | trim "wc."}}`,
			want: "example",
		},
		{
			name: "trimRight",
			tmpl: `{{"www.example.cc" | trimRight "wc."}}`,
			want: "www.example",
		},
		{
			name: "trimLeft",
			tmpl: `{{"www.example.cc" | trimLeft "wc."}}`,
			want: "example.cc",
		},
		{
			name: "hasPrefix",
			tmpl: `{{"www.example.cc" | hasPrefix "www."}}`,
			want: "true",
		},
		{
			name: "hasSuffix",
			tmpl: `{{"www.example.cc" | hasSuffix ".cc"}}`,
			want: "true",
		},

		// Path functions
		{
			name: "sanitizePath",
			tmpl: `{{sanitizePath "/some wiärd päth/to/some/place"}}`,
			want: "/some-wi-rd-p-th/to/some/place",
		},
		{
			name: "sanitizePathSegment",
			tmpl: `{{sanitizePathSegment "/some wiärd päth/to/some/place"}}`,
			want: "-some-wi-rd-p-th-to-some-place",
		},
		{
			name: "basename",
			tmpl: `{{basename "/some/long/path/to/a/file.txt"}}`,
			want: "file.txt",
		},
		{
			name: "dirname",
			tmpl: `{{dirname "/some/long/path/to/a/file.txt"}}`,
			want: "/some/long/path/to/a",
		},
		{
			name: "fileext",
			tmpl: `{{fileext "/some/long/path/to/a/file.txt"}}`,
			want: ".txt",
		},

		// Regex functions
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

		// Math functions
		{
			name: "add_int",
			tmpl: `{{add 1 2}}`,
			want: `3`,
		},
		{
			name: "add_float",
			tmpl: `{{add 1.5 2.1}}`,
			want: `3.6`,
		},
		{
			name: "add_string",
			tmpl: `{{add "ab" "ba"}}`,
			want: `abba`,
		},
		{
			name: "sub_int",
			tmpl: `{{sub 1 2}}`,
			want: `-1`,
		},
		{
			name: "sub_float",
			tmpl: `{{sub 1.5 2.1 | printf "%.1f"}}`,
			want: `-0.6`,
		},
		{
			name: "div_int",
			tmpl: `{{div 10 2}}`,
			want: `5`,
		},
		{
			name: "div_float",
			tmpl: `{{div 4.5 3.0 | printf "%.1f"}}`,
			want: `1.5`,
		},
		{
			name: "mod_int",
			tmpl: `{{mod 10 2}}`,
			want: `0`,
		},
		{
			name: "mod_float",
			tmpl: `{{mod 1.1 1.0 | printf "%.1f"}}`,
			want: `0.1`,
		},
		{
			name: "mul_int",
			tmpl: `{{mul 10 2}}`,
			want: `20`,
		},
		{
			name: "mul_float",
			tmpl: `{{mul 1.1 10.0 | printf "%.1f"}}`,
			want: `11.0`,
		},
		{
			name: "floor",
			tmpl: `{{floor 10.9}}`,
			want: `10`,
		},
		{
			name: "ceil",
			tmpl: `{{ceil 10.9}}`,
			want: `11`,
		},
		{
			name: "round_up",
			tmpl: `{{round 10.9}}`,
			want: `11`,
		},
		{
			name: "round_down",
			tmpl: `{{round 10.2}}`,
			want: `10`,
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
