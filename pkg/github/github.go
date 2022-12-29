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
	"errors"
	"fmt"
	"net/http"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Client interface {
	CreatePullRequest(pr NewPullRequest) (PullRequest, error)
}

func New(ghCfg *config.GitHub) (Client, error) {
	raw, err := newClient(ghCfg)
	if err != nil {
		return nil, err
	}
	return &client{
		raw: raw,
	}, nil
}

type client struct {
	raw *github.Client
}

func newClient(ghCfg *config.GitHub) (*github.Client, error) {
	httpClient, err := newHTTPClient(ghCfg)
	if err != nil {
		return nil, err
	}
	return newClientEnterpriceOrPublic(ghCfg.URL, httpClient)
}

func newHTTPClient(ghCfg *config.GitHub) (*http.Client, error) {
	switch ghCfg.Auth.Type {
	case config.GitHubAuthTypePAT:
		return newOAuthHTTPClient(ghCfg.Auth.Token), nil
	case config.GitHubAuthTypeApp:
		return newAppsHTTPClient(ghCfg)
	default:
		return nil, fmt.Errorf("unsupported GitHub auth type: %q", ghCfg.Auth.Type)
	}
}

func newOAuthHTTPClient(token string) *http.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(context.TODO(), tokenSource)
}

func newAppsHTTPClient(ghCfg *config.GitHub) (*http.Client, error) {
	appTransport, err := newAppsTransport(&ghCfg.Auth.App)
	if err != nil {
		return nil, err
	}
	if ghCfg.URL != nil {
		appTransport.BaseURL = *ghCfg.URL
	}
	return &http.Client{Transport: appTransport}, nil
}

func newAppsTransport(appCfg *config.GitHubAuthApp) (*ghinstallation.AppsTransport, error) {
	transport := http.DefaultTransport
	switch {
	case appCfg.PrivateKeyPEM != nil:
		appsTransport, err := ghinstallation.NewAppsTransport(transport, appCfg.ID, appCfg.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("read config privateKeyPem: %w", err)
		}
		return appsTransport, nil
	case appCfg.PrivateKeyPath != nil:
		appsTransport, err := ghinstallation.NewAppsTransportKeyFromFile(transport, appCfg.ID, *appCfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read config privateKeyPath: %w", err)
		}
		return appsTransport, nil
	default:
		return nil, errors.New("must set GitHub auth config privateKeyPem or privateKeyPath")
	}
}

func newClientEnterpriceOrPublic(ghURL *string, httpClient *http.Client) (*github.Client, error) {
	if ghURL != nil {
		return github.NewEnterpriseClient(*ghURL, "", httpClient)
	}
	return github.NewClient(httpClient), nil
}

func (c *client) CreatePullRequest(pr NewPullRequest) (PullRequest, error) {
	created, _, err := c.raw.PullRequests.Create(context.TODO(), pr.Owner, pr.Repo, &github.NewPullRequest{
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
		ID:          util.Deref(created.ID, -1),
		Number:      util.Deref(created.Number, -1),
		URL:         util.Deref(created.HTMLURL, ""),
		Title:       util.Deref(created.Title, ""),
		Description: util.Deref(created.Body, ""),
		Head:        util.Deref(created.Head.Label, ""),
		Base:        util.Deref(created.Base.Label, ""),
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
}
