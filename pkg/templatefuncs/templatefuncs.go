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
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/version"
	"gopkg.in/typ.v4/slices"
	"gopkg.in/yaml.v3"
)

var (
	pathCharRegex        = regexp.MustCompile(`[^a-zA-Z0-9/,._-]`)
	pathSegmentCharRegex = regexp.MustCompile(`[^a-zA-Z0-9,._-]`)
)

var FuncsMap = template.FuncMap{
	// Version functions
	"versionBump": func(add string, to string) string {
		// "1.2.3" | versionBump "0.1.0"
		// add = 0.1.0
		// to = 1.2.3
		addVer, err := version.Parse(add)
		if err != nil {
			panic(fmt.Sprintf("parse version %q: %s", add, err))
		}
		toVer, err := version.Parse(to)
		if err != nil {
			panic(fmt.Sprintf("parse version %q: %s", to, err))
		}
		return toVer.Bump(addVer).String()
	},

	// Type conversion functions
	"int": func(value any) int {
		switch value := value.(type) {
		case string:
			number, err := strconv.ParseInt(value, 0, 0)
			if err != nil {
				panic(err.Error())
			}
			return int(number)
		case int:
			return value
		case int8:
			return int(value)
		case int16:
			return int(value)
		case int32:
			return int(value)
		case int64:
			return int(value)
		case uint:
			return int(value)
		case uint8:
			return int(value)
		case uint16:
			return int(value)
		case uint32:
			return int(value)
		case uint64:
			return int(value)
		case float32:
			return int(value)
		case float64:
			return int(value)
		default:
			panic(fmt.Sprintf("unsupported type: %T", value))
		}
	},
	"float64": func(value any) float64 {
		switch value := value.(type) {
		case string:
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				panic(err.Error())
			}
			return number
		case int:
			return float64(value)
		case int8:
			return float64(value)
		case int16:
			return float64(value)
		case int32:
			return float64(value)
		case int64:
			return float64(value)
		case uint:
			return float64(value)
		case uint8:
			return float64(value)
		case uint16:
			return float64(value)
		case uint32:
			return float64(value)
		case uint64:
			return float64(value)
		case float32:
			return float64(value)
		case float64:
			return value
		default:
			panic(fmt.Sprintf("unsupported type: %T", value))
		}
	},

	// List functions
	"list": func(values ...any) []any {
		return values
	},
	"first": func(list []any) any {
		return list[0]
	},
	"rest": func(list []any) []any {
		return list[1:]
	},
	"last": func(list []any) any {
		return list[len(list)-1]
	},
	"initial": func(list []any) any {
		return list[:len(list)-1]
	},
	"append": func(list []any, values ...any) []any {
		return append(list, values...)
	},
	"prepend": func(list []any, values ...any) []any {
		return append(values, list...)
	},
	"concat": func(a, b []any) []any {
		return append(a, b...)
	},
	"reverse": func(list []any) []any {
		reversed := make([]any, len(list))
		copy(reversed, list)
		slices.Reverse(reversed)
		return reversed
	},

	// Dictionary functions
	"dict": func(keyAndValues ...any) map[string]any {
		if len(keyAndValues)%2 != 0 {
			panic("must be an even number of arguments, for pairs of key + value")
		}
		dict := make(map[string]any)
		for i := 0; i < len(keyAndValues); i += 2 {
			key, keyIsString := keyAndValues[i].(string)
			value := keyAndValues[i+1]

			if !keyIsString {
				panic(fmt.Sprintf("expected string for map key, got %T", keyAndValues[i]))
			}
			dict[key] = value
		}
		return dict
	},
	"get": func(dict map[string]any, key string) any {
		return dict[key]
	},
	"set": func(dict map[string]any, key string, value any) map[string]any {
		dict[key] = value
		return dict
	},
	"unset": func(dict map[string]any, key string) map[string]any {
		delete(dict, key)
		return dict
	},
	"hasKey": func(dict map[string]any, key string) bool {
		_, ok := dict[key]
		return ok
	},

	// Encoding functions
	"fromYaml": func(value string) (any, error) {
		var result any
		err := yaml.Unmarshal([]byte(value), &result)
		return result, err
	},
	"toYaml": func(value any) (string, error) {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(value); err != nil {
			return "", err
		}
		return buf.String(), nil
	},
	"toJson": func(value any) (string, error) {
		jsonValue, err := json.Marshal(value)
		return string(jsonValue), err
	},
	"toPrettyJson": func(value any) (string, error) {
		jsonValue, err := json.MarshalIndent(value, "", "  ")
		return string(jsonValue), err
	},
	"fromJson": func(value string) (any, error) {
		var result any
		err := json.Unmarshal([]byte(value), &result)
		return result, err
	},

	// String functions
	"trimPrefix": func(prefix, s string) string {
		return strings.TrimPrefix(s, prefix)
	},
	"trimSuffix": func(suffix, s string) string {
		return strings.TrimSuffix(s, suffix)
	},
	"trimSpace": func(s string) string {
		return strings.TrimSpace(s)
	},
	"trim": func(cutset, s string) string {
		return strings.Trim(s, cutset)
	},
	"trimLeft": func(cutset, s string) string {
		return strings.TrimLeft(s, cutset)
	},
	"trimRight": func(cutset, s string) string {
		return strings.TrimRight(s, cutset)
	},
	"hasPrefix": func(prefix, s string) bool {
		return strings.HasPrefix(s, prefix)
	},
	"hasSuffix": func(suffix, s string) bool {
		return strings.HasSuffix(s, suffix)
	},

	// Path functions
	"sanitizePath": func(path string) string {
		path = strings.ToLower(path)
		path = filepath.ToSlash(path)
		return pathCharRegex.ReplaceAllLiteralString(path, "-")
	},
	"sanitizePathSegment": func(path string) string {
		path = strings.ToLower(path)
		return pathSegmentCharRegex.ReplaceAllLiteralString(path, "-")
	},
	"basename": func(path string) string {
		return filepath.Base(path)
	},
	"dirname": func(path string) string {
		return filepath.Dir(path)
	},
	"fileext": func(path string) string {
		return filepath.Ext(path)
	},

	// Regex functions
	"regexReplaceAll": func(regex, replace, text string) string {
		re := regexp.MustCompile(regex)
		return re.ReplaceAllString(text, replace)
	},
	"regexMatch": func(regex, text string) bool {
		re := regexp.MustCompile(regex)
		return re.MatchString(text)
	},

	// Math functions
	"add": func(a, b any) any {
		aType := reflect.TypeOf(a)
		bType := reflect.TypeOf(b)
		if aType != bType {
			panic(fmt.Sprintf("args must be of same type, got %T and %T", a, b))
		}
		switch a := a.(type) {
		case string:
			return a + b.(string)
		case int:
			return a + b.(int)
		case int8:
			return a + b.(int8)
		case int16:
			return a + b.(int16)
		case int32:
			return a + b.(int32)
		case int64:
			return a + b.(int64)
		case uint:
			return a + b.(uint)
		case uint8:
			return a + b.(uint8)
		case uint16:
			return a + b.(uint16)
		case uint32:
			return a + b.(uint32)
		case uint64:
			return a + b.(uint64)
		case float32:
			return a + b.(float32)
		case float64:
			return a + b.(float64)
		case complex64:
			return a + b.(complex64)
		case complex128:
			return a + b.(complex128)
		default:
			panic(fmt.Sprintf("unsupported type: %T", a))
		}
	},
	"sub": func(a, b any) any {
		aType := reflect.TypeOf(a)
		bType := reflect.TypeOf(b)
		if aType != bType {
			panic(fmt.Sprintf("args must be of same type, got %T and %T", a, b))
		}
		switch a := a.(type) {
		case int:
			return a - b.(int)
		case int8:
			return a - b.(int8)
		case int16:
			return a - b.(int16)
		case int32:
			return a - b.(int32)
		case int64:
			return a - b.(int64)
		case uint:
			return a - b.(uint)
		case uint8:
			return a - b.(uint8)
		case uint16:
			return a - b.(uint16)
		case uint32:
			return a - b.(uint32)
		case uint64:
			return a - b.(uint64)
		case float32:
			return a - b.(float32)
		case float64:
			return a - b.(float64)
		case complex64:
			return a - b.(complex64)
		case complex128:
			return a - b.(complex128)
		default:
			panic(fmt.Sprintf("unsupported type: %T", a))
		}
	},
	"div": func(a, b any) any {
		aType := reflect.TypeOf(a)
		bType := reflect.TypeOf(b)
		if aType != bType {
			panic(fmt.Sprintf("args must be of same type, got %T and %T", a, b))
		}
		switch a := a.(type) {
		case int:
			return a / b.(int)
		case int8:
			return a / b.(int8)
		case int16:
			return a / b.(int16)
		case int32:
			return a / b.(int32)
		case int64:
			return a / b.(int64)
		case uint:
			return a / b.(uint)
		case uint8:
			return a / b.(uint8)
		case uint16:
			return a / b.(uint16)
		case uint32:
			return a / b.(uint32)
		case uint64:
			return a / b.(uint64)
		case float32:
			return a / b.(float32)
		case float64:
			return a / b.(float64)
		case complex64:
			return a / b.(complex64)
		case complex128:
			return a / b.(complex128)
		default:
			panic(fmt.Sprintf("unsupported type: %T", a))
		}
	},
	"mod": func(a, b any) any {
		aType := reflect.TypeOf(a)
		bType := reflect.TypeOf(b)
		if aType != bType {
			panic(fmt.Sprintf("args must be of same type, got %T and %T", a, b))
		}
		switch a := a.(type) {
		case int:
			return a % b.(int)
		case int8:
			return a % b.(int8)
		case int16:
			return a % b.(int16)
		case int32:
			return a % b.(int32)
		case int64:
			return a % b.(int64)
		case uint:
			return a % b.(uint)
		case uint8:
			return a % b.(uint8)
		case uint16:
			return a % b.(uint16)
		case uint32:
			return a % b.(uint32)
		case uint64:
			return a % b.(uint64)
		case float32:
			return float32(math.Mod(float64(a), float64(b.(float32))))
		case float64:
			return math.Mod(a, b.(float64))
		default:
			panic(fmt.Sprintf("unsupported type: %T", a))
		}
	},
	"mul": func(a, b any) any {
		aType := reflect.TypeOf(a)
		bType := reflect.TypeOf(b)
		if aType != bType {
			panic(fmt.Sprintf("args must be of same type, got %T and %T", a, b))
		}
		switch a := a.(type) {
		case int:
			return a * b.(int)
		case int8:
			return a * b.(int8)
		case int16:
			return a * b.(int16)
		case int32:
			return a * b.(int32)
		case int64:
			return a * b.(int64)
		case uint:
			return a * b.(uint)
		case uint8:
			return a * b.(uint8)
		case uint16:
			return a * b.(uint16)
		case uint32:
			return a * b.(uint32)
		case uint64:
			return a * b.(uint64)
		case float32:
			return a * b.(float32)
		case float64:
			return a * b.(float64)
		case complex64:
			return a * b.(complex64)
		case complex128:
			return a * b.(complex128)
		default:
			panic(fmt.Sprintf("unsupported type: %T", a))
		}
	},
	"floor": func(a any) any {
		switch value := a.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			return value
		case float32:
			return float32(math.Floor(float64(value)))
		case float64:
			return math.Floor(value)
		default:
			panic(fmt.Sprintf("unsupported type: %T", value))
		}
	},
	"ceil": func(a any) any {
		switch value := a.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			return value
		case float32:
			return float32(math.Ceil(float64(value)))
		case float64:
			return math.Ceil(value)
		default:
			panic(fmt.Sprintf("unsupported type: %T", value))
		}
	},
	"round": func(a any) any {
		switch value := a.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			return value
		case float32:
			return float32(math.Round(float64(value)))
		case float64:
			return math.Round(value)
		default:
			panic(fmt.Sprintf("unsupported type: %T", value))
		}
	},
}
