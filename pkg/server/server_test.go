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
	"net/url"
	"testing"

	"github.com/RiskIdent/jelease/pkg/config"
)

func TestCreateDeferredCreationURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		pkgName string
		req     CreatePRRequest
		want    string
	}{
		{
			name:    "regular path",
			url:     "http://localhost:8080",
			pkgName: "my-package",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name:    "ignores fragment and query",
			url:     "http://localhost:8080?somequery=helloworld#foobar",
			pkgName: "my-package",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name:    "optional fields",
			url:     "http://localhost:8080",
			pkgName: "my-package",
			req:     CreatePRRequest{},
			want:    "http://localhost:8080/packages/my-package/create-pr",
		},
		{
			name:    "ok with trailing slash",
			url:     "http://localhost:8080/",
			pkgName: "my-package",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name:    "ok with base path",
			url:     "http://localhost:8080/base/",
			pkgName: "my-package",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "http://localhost:8080/base/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name:    "normalizes name",
			url:     "http://localhost:8080",
			pkgName: "my-org/my-pkg",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "http://localhost:8080/packages/my-org-my-pkg/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name:    "keeps scheme",
			url:     "https+loremipsum://localhost:8080",
			pkgName: "my-org/my-pkg",
			req: CreatePRRequest{
				Version:   "v1.2.3",
				JiraIssue: "OP-1234",
				PRCreate:  true,
			},
			want: "https+loremipsum://localhost:8080/packages/my-org-my-pkg/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			publicURL, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("parse public URL: %s", err)
			}
			got := createDeferredCreationURL(publicURL, tc.pkgName, tc.req)
			if got == nil {
				t.Fatal("expected URL returned, got nil")
			}
			gotStr := got.String()
			if gotStr != tc.want {
				t.Errorf("did not match!\nwant: %q\ngot:  %q", tc.want, gotStr)
			}
		})
	}
}

func TestSetTemplateContextPackageDescription(t *testing.T) {
	tests := []struct {
		name string
		ctx  config.TemplateContext
		desc *config.Template
		want string
	}{
		{
			name: "empty",
			ctx:  config.TemplateContext{},
			desc: nil,
			want: "",
		},
		{
			name: "only package name",
			ctx:  config.TemplateContext{Package: "my-pkg"},
			desc: nil,
			want: "",
		},
		{
			name: "only package desc",
			ctx:  config.TemplateContext{},
			desc: config.MustTemplate("Pkg `{{.Package}}` desc"),
			want: "Pkg `` desc",
		},
		{
			name: "package name and desc",
			ctx:  config.TemplateContext{Package: "my-pkg"},
			desc: config.MustTemplate("Pkg `{{.Package}}` desc"),
			want: "Pkg `my-pkg` desc",
		},
		{
			name: "package name, version, and desc",
			ctx:  config.TemplateContext{Package: "my-pkg", Version: "v1.2.3"},
			desc: config.MustTemplate("Pkg `{{.Package}}@{{.Version}}` desc"),
			want: "Pkg `my-pkg@v1.2.3` desc",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := setTemplateContextPackageDescription(test.ctx, test.desc)
			if err != nil {
				t.Fatal(err)
			}
			if got.PackageDescription != test.want {
				t.Errorf("wrong result\nwant: %#v\ngot:  %#v", test.want, got.PackageDescription)
			}
		})
	}
}
