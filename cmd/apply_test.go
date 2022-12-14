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

package cmd

import (
	"regexp"
	"testing"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/config"
)

func TestPatchSingleLineRegex(t *testing.T) {
	line := []byte("<<my-pkg v0.1.0>>")
	patch := config.PatchRegex{
		Match:   newRegex(t, `(my-pkg) v0.1.0`),
		Replace: newTemplate(t, `{{ index .Groups 1 }} {{ .Version }}`),
	}

	version := "v1.2.3"

	newLine, err := patchSingleLineRegex(patch, version, line)
	if err != nil {
		t.Fatal(err)
	}

	got := string(newLine)
	want := "<<my-pkg v1.2.3>>"

	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func newRegex(t *testing.T, text string) *config.RegexPattern {
	r, err := regexp.Compile(`(my-pkg) v0.1.0`)
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

func TestGetGitHubRepoRef(t *testing.T) {
	tests := []struct {
		name      string
		remote    string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "regular",
			remote:    "https://github.com/RiskIdent/jelease",
			wantOwner: "RiskIdent",
			wantRepo:  "jelease",
		},
		{
			name:      "with .git",
			remote:    "https://github.com/RiskIdent/jelease.git",
			wantOwner: "RiskIdent",
			wantRepo:  "jelease",
		},
		{
			name:      "enterprise",
			remote:    "https://some-github-enterprise.example.com/RiskIdent/jelease.git?ignore=this#please",
			wantOwner: "RiskIdent",
			wantRepo:  "jelease",
		},
		{
			name:      "ignores extra stuff",
			remote:    "https://some-github-enterprise.example.com/RiskIdent/jelease.git/woa?ignore=this#please",
			wantOwner: "RiskIdent",
			wantRepo:  "jelease",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getGitHubRepoRef(tc.remote)
			if err != nil {
				t.Fatal(err)
			}
			if got.Owner != tc.wantOwner {
				t.Errorf("want owner %q, got owner %q", tc.wantOwner, got.Owner)
			}
			if got.Repo != tc.wantRepo {
				t.Errorf("want repo %q, got repo %q", tc.wantRepo, got.Repo)
			}
		})
	}
}
