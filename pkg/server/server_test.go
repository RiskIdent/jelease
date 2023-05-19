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
)

func TestCreateDeferredCreationURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		req  CreatePRRequest
		want string
	}{
		{
			name: "regular path",
			url:  "http://localhost:8080",
			req: CreatePRRequest{
				PackageName: "my-package",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name: "ignores fragment and query",
			url:  "http://localhost:8080?somequery=helloworld#foobar",
			req: CreatePRRequest{
				PackageName: "my-package",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name: "optional fields",
			url:  "http://localhost:8080",
			req: CreatePRRequest{
				PackageName: "my-package",
			},
			want: "http://localhost:8080/packages/my-package/create-pr",
		},
		{
			name: "ok with trailing slash",
			url:  "http://localhost:8080/",
			req: CreatePRRequest{
				PackageName: "my-package",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
			},
			want: "http://localhost:8080/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name: "ok with base path",
			url:  "http://localhost:8080/base/",
			req: CreatePRRequest{
				PackageName: "my-package",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
			},
			want: "http://localhost:8080/base/packages/my-package/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name: "normalizes name",
			url:  "http://localhost:8080",
			req: CreatePRRequest{
				PackageName: "my-org/my-pkg",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
			},
			want: "http://localhost:8080/packages/my-org-my-pkg/create-pr?jiraIssue=OP-1234&prCreate=true&version=v1.2.3",
		},
		{
			name: "keeps scheme",
			url:  "https+loremipsum://localhost:8080",
			req: CreatePRRequest{
				PackageName: "my-org/my-pkg",
				Version:     "v1.2.3",
				JiraIssue:   "OP-1234",
				PRCreate:    true,
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
			got := createDeferredCreationURL(publicURL, tc.req)
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
