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

package util

import (
	"slices"
	"testing"
)

func TestConcat(t *testing.T) {
	tests := []struct {
		a, b []int
		want []int
	}{
		{
			a:    nil,
			b:    nil,
			want: nil,
		},
		{
			a:    []int{1},
			b:    nil,
			want: []int{1},
		},
		{
			a:    nil,
			b:    []int{1},
			want: []int{1},
		},
		{
			a:    []int{1, 2},
			b:    []int{3, 4},
			want: []int{1, 2, 3, 4},
		},
	}

	for _, tc := range tests {
		got := Concat(tc.a, tc.b)
		if !slices.Equal(tc.want, got) {
			t.Errorf("%v + %v: want %v, got %v", tc.a, tc.b, tc.want, got)
		}
	}
}
