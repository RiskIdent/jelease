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

package github

import (
	"fmt"
	"net/url"
	"strings"
)

type RepoRef struct {
	URL   string
	Owner string
	Repo  string
}

func ParseRepoRef(remote string) (RepoRef, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return RepoRef{}, err
	}
	u.User = nil
	path := u.Path
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) < 2 {
		return RepoRef{}, fmt.Errorf("expected https://host/OWNER/REPO in URL, got: %s", u.String())
	}
	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")

	u.Path = fmt.Sprintf("%s/%s", owner, repo)
	u.Fragment = ""
	u.RawFragment = ""
	u.RawQuery = ""
	return RepoRef{
		URL:   u.String(),
		Owner: owner,
		Repo:  repo,
	}, nil
}
