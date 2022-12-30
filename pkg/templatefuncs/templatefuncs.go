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
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/version"
	"gopkg.in/yaml.v2"
)

var (
	pathCharRegex        = regexp.MustCompile(`[^a-zA-Z0-9/,._-]`)
	pathSegmentCharRegex = regexp.MustCompile(`[^a-zA-Z0-9,._-]`)
)

var FuncsMap = template.FuncMap{
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
	"regexReplaceAll": func(regex, replace, text string) string {
		re := regexp.MustCompile(regex)
		return re.ReplaceAllString(text, replace)
	},
	"regexMatch": func(regex, text string) bool {
		re := regexp.MustCompile(regex)
		return re.MatchString(text)
	},
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
	"fromYaml": func(value string) any {
		var result any
		if err := yaml.Unmarshal([]byte(value), &result); err != nil {
			panic(err.Error())
		}
		return result
	},
	"toYaml": func(value any) string {
		jsonValue, err := yaml.Marshal(value)
		if err != nil {
			panic(err.Error())
		}
		return string(jsonValue)
	},
	"toJson": func(value any) string {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			panic(err.Error())
		}
		return string(jsonValue)
	},
	"toPrettyJson": func(value any) string {
		jsonValue, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		return string(jsonValue)
	},
	"fromJson": func(value string) any {
		var result any
		if err := json.Unmarshal([]byte(value), &result); err != nil {
			panic(err.Error())
		}
		return result
	},
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
}
