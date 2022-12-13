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

package version

import "testing"

func TestParse_intactAfterStringer(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "1-seg",
			version: "0",
		},
		{
			name:    "v1-seg",
			version: "v1",
		},
		{
			name:    "3-seg",
			version: "0.1.2",
		},
		{
			name:    "v3-seg",
			version: "v0.1.2",
		},
		{
			name:    "3-seg-suffix",
			version: "0.1.2-rc.12+3331",
		},
		{
			name:    "v3-seg-suffix",
			version: "v0.1.2-rc.12+3331",
		},
		{
			name:    "v5-seg-suffix",
			version: "v0.1.2.3.4-rc.12+3331",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := Parse(tc.version)
			if err != nil {
				t.Fatalf("want %q, got error: %s", tc.version, err)
			}
			s := v.String()
			if s != tc.version {
				t.Errorf("want %q, got %q", tc.version, s)
			}
		})
	}
}

func TestParse_errors(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "prefix too long",
			version: "ver0.0.0",
		},
		{
			name:    "prefix uppercase",
			version: "V0.0.0",
		},
		{
			name:    "suffix invalid chars",
			version: "0.0.0-ß",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.version)
			if err == nil {
				t.Fatalf("want error for %q, got nil", tc.version)
			}
		})
	}
}
