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
	"strings"
	"unicode"
)

func Concat[S ~[]E, E any](slices ...S) S {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	result := make(S, totalLen)
	var offset int
	for _, s := range slices {
		copy(result[offset:], s)
		offset += len(s)
	}
	return result
}

func Ref[T any](v T) *T {
	return &v
}

func Deref[T any](v *T, or T) T {
	if v == nil {
		return or
	}
	return *v
}

var camelCaseReplacer = strings.NewReplacer(
	"ID", "Id",
	"URL", "Url",
	"HTTP", "Http",
	"JSON", "Json",
	"JQ", "Jq",
	"YAML", "Yaml",
	"YQ", "Yq",
	"GitHub", "Github",
	"PR", "Pr",
)

// ToCamelCase is a very stupid implementation for converting
// PascalCase to camelCase.
//
// NOTE: If we need to convert user-provided strings to camelCase,
// then we should replace this with the community strcase package.
func ToCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	s = camelCaseReplacer.Replace(s)
	b := []byte(s)
	b[0] = byte(unicode.ToLower(rune(b[0])))
	return string(b)
}
