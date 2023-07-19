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

package config

import (
	"net/url"
	"testing"
)

func TestCensored(t *testing.T) {
	token := "secret-token"
	tokenURL := "http://:" + token + "@localhost:8080/"
	cfg := &Config{
		GitHub: GitHub{
			Auth: GitHubAuth{
				Token: &token,
			},
		},
		Jira: Jira{
			Auth: JiraAuth{
				Token: token,
			},
		},
		HTTP: HTTP{
			PublicURL: (*URL)(mustParseURL(t, tokenURL)),
		},
	}
	censored := cfg.Censored()

	if censored.GitHub.Auth.Token == nil {
		t.Fatal("unexpected nil GitHub.Auth.Token")
	}
	if *censored.GitHub.Auth.Token == token {
		t.Fatal("did not censor GitHub.Auth.Token")
	}
	if *cfg.GitHub.Auth.Token != token {
		t.Fatalf("changed original config GitHub.Auth.Token to %q, but should not do that", *cfg.GitHub.Auth.Token)
	}

	if censored.Jira.Auth.Token == token {
		t.Fatal("did not censor Jira.Auth.Token")
	}
	if cfg.Jira.Auth.Token != token {
		t.Fatalf("changed original config Jira.Auth.Token to %q, but should not do that", cfg.Jira.Auth.Token)
	}
	if cfg.HTTP.PublicURL == nil {
		t.Fatalf("unexpected nil HTTP.PublicURL")
	}
	if censored.HTTP.PublicURL.String() == tokenURL {
		t.Fatalf("did not censor HTTP.PublicURL")
	}
	if cfg.HTTP.PublicURL.String() != tokenURL {
		t.Fatalf("changed original config HTTP.PublicURL to %q, but should not do that", cfg.HTTP.PublicURL)
	}
}

func mustParseURL(t *testing.T, value string) *url.URL {
	t.Helper()
	u, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
