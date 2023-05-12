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

package server

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"

	"gopkg.in/yaml.v3"
)

var FuncMap = template.FuncMap{
	// Encodings
	"toYaml": func(o any) (string, error) {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(o); err != nil {
			return "", err
		}
		return buf.String(), nil
	},

	// Math
	"add": func(values ...any) (float64, error) {
		floats, err := toFloats(values)
		if err != nil {
			return 0, err
		}
		var sum float64
		for _, f := range floats {
			sum += f
		}
		return sum, nil
	},
}

func toFloats(values []any) ([]float64, error) {
	floats := make([]float64, len(values))
	for i, val := range values {
		f, err := toFloat(val)
		if err != nil {
			return nil, fmt.Errorf("argument index %d: %w", i, err)
		}
		floats[i] = f
	}
	return floats, nil
}

func toFloat(value any) (float64, error) {
	switch value := value.(type) {
	case string:
		return strconv.ParseFloat(value, 64)
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	default:
		return 0, fmt.Errorf("cannot convert type to float64: %T", value)
	}
}
