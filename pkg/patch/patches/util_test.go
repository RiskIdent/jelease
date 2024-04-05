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

package patches

import (
	"regexp"
	"testing"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
)

// No tests in this file.
// This file only contains test utils.

func newRegex(t *testing.T, text string) *config.RegexPattern {
	r, err := regexp.Compile(text)
	if err != nil {
		t.Fatal(err)
	}
	return (*config.RegexPattern)(r)
}

func newTemplate(t *testing.T, text string) *config.Template {
	tmpl, err := template.New("").Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return (*config.Template)(tmpl)
}

func newYAMLPath(t *testing.T, text string) *config.YAMLPathPattern {
	p, err := yamlpath.NewPath(text)
	if err != nil {
		t.Fatal(err)
	}
	return &config.YAMLPathPattern{
		YAMLPath: p,
		Source:   text,
	}
}
