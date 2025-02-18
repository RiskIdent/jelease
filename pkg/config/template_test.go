// SPDX-FileCopyrightText: 2025 Risk.Ident GmbH <contact@riskident.com>
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

import "testing"

func TestTemplateRender(t *testing.T) {
	tmplCtx := map[string]any{
		"SomeNumber": 1234,
	}
	tests := []struct {
		name string
		tmpl *Template
		want string
	}{
		{
			name: "nil",
			tmpl: nil,
			want: "",
		},
		{
			name: "raw string",
			tmpl: MustTemplate("foo bar\nmoo doo"),
			want: "foo bar\nmoo doo",
		},
		{
			name: "use context",
			tmpl: MustTemplate("the number {{.SomeNumber}} is my favorite"),
			want: "the number 1234 is my favorite",
		},
		{
			name: "use func",
			tmpl: MustTemplate("i dont like the number {{add .SomeNumber 500}}."),
			want: "i dont like the number 1734.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.tmpl.Render(tmplCtx)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("wrong result\nwant: %q\ngot:  %q", test.want, got)
			}
		})
	}
}
