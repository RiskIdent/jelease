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
	"context"
	"fmt"
	"net/http"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/google/go-github/v48/github"
)

type Client interface {
	CreatePullRequest(ctx context.Context, pr NewPullRequest) (PullRequest, error)
	TestConnection(ctx context.Context) error
	GitCredentialsForRepo(ctx context.Context, repo RepoRef) (git.Credentials, error)
}

func New(ghCfg *config.GitHub) (Client, error) {
	switch ghCfg.Auth.Type {
	case config.GitHubAuthTypePAT:
		return NewPATClient(ghCfg)
	case config.GitHubAuthTypeApp:
		return NewAppClient(ghCfg)
	default:
		return nil, fmt.Errorf("unsupported GitHub auth type: %q", ghCfg.Auth.Type)
	}
}

func newClientEnterpriceOrPublic(ghURL *string, httpClient *http.Client) (*github.Client, error) {
	if ghURL != nil {
		return github.NewEnterpriseClient(*ghURL, *ghURL, httpClient)
	}
	return github.NewClient(httpClient), nil
}

func CreatePullRequest(ctx context.Context, gh *github.Client, pr NewPullRequest) (PullRequest, error) {
	created, _, err := gh.PullRequests.Create(ctx, pr.Owner, pr.Repo, &github.NewPullRequest{
		Title:               &pr.Title,
		Body:                &pr.Description,
		Head:                &pr.Head,
		Base:                &pr.Base,
		MaintainerCanModify: util.Ref(true),
	})
	if err != nil {
		return PullRequest{}, err
	}
	return PullRequest{
		RepoRef:     pr.RepoRef,
		ID:          created.GetID(),
		Number:      created.GetNumber(),
		URL:         created.GetHTMLURL(),
		Title:       created.GetTitle(),
		Description: created.GetBody(),
		Head:        created.Head.GetLabel(),
		Base:        created.Base.GetLabel(),
	}, nil
}

type NewPullRequest struct {
	RepoRef
	Title       string
	Description string
	Head        string
	Base        string
}

type PullRequest struct {
	RepoRef
	ID          int64
	Number      int
	URL         string
	Title       string
	Description string
	Head        string
	Base        string
	Commit      git.Commit
}
