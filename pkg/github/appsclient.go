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
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog/log"
)

type appsClient struct {
	appsTransport       *ghinstallation.AppsTransport
	noInstallClient     *github.Client
	installationPerRepo map[RepoRefSlim]installation
	baseURL             *string
}

type installation struct {
	client         *github.Client
	transport      *ghinstallation.Transport
	installationID int64
}

func NewAppClient(ghCfg *config.GitHub) (Client, error) {
	appsTransport, err := newAppsTransport(ghCfg)
	if err != nil {
		return nil, err
	}
	nonInstallClient, err := newClientEnterpriceOrPublic(ghCfg.URL, &http.Client{Transport: appsTransport})
	if err != nil {
		return nil, err
	}
	appsTransport.BaseURL = strings.TrimRight(nonInstallClient.BaseURL.String(), "/")
	return &appsClient{
		appsTransport:       appsTransport,
		noInstallClient:     nonInstallClient,
		installationPerRepo: make(map[RepoRefSlim]installation),
		baseURL:             ghCfg.URL,
	}, nil
}

func (c *appsClient) TestConnection(ctx context.Context) error {
	// get empty means get the currently authenticated app
	app, _, err := c.noInstallClient.Apps.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("get GitHub App metadata: %w", err)
	}
	log.Debug().
		Str("app", app.GetName()).
		Str("url", app.GetHTMLURL()).
		Msg("Authenticated as GitHub App.")
	return nil
}

func (c *appsClient) GitCredentialsForRepo(ctx context.Context, repo RepoRef) (git.Credentials, error) {
	inst, err := c.findInstallationForRepo(ctx, repo)
	if err != nil {
		return git.Credentials{}, err
	}
	token, err := inst.transport.Token(ctx)
	if err != nil {
		return git.Credentials{}, err
	}
	// Using the "HTTP-based Git access by an installation" auth here
	// https://docs.github.com/en/enterprise-cloud@latest/developers/apps/building-github-apps/authenticating-with-github-apps#http-based-git-access-by-an-installation
	// Simpler than setting up a deploy token, and keeps us to only use HTTPS
	// traffic, instead of also requiring SSH traffic. Fewer ports to open.
	return git.Credentials{
		Username: "x-access-token",
		Password: token,
	}, nil
}

func (c *appsClient) CreatePullRequest(ctx context.Context, pr NewPullRequest) (PullRequest, error) {
	inst, err := c.findInstallationForRepo(ctx, pr.RepoRef)
	if err != nil {
		return PullRequest{}, err
	}
	return CreatePullRequest(ctx, inst.client, pr)
}

func (c *appsClient) findInstallationForRepo(ctx context.Context, repo RepoRef) (installation, error) {
	if inst, ok := c.installationPerRepo[repo.Slim()]; ok {
		return inst, nil
	}
	id, err := c.findInstallationIDForRepo(ctx, repo)
	if err != nil {
		return installation{}, err
	}
	transport := ghinstallation.NewFromAppsTransport(c.appsTransport, id)
	client, err := newClientEnterpriceOrPublic(c.baseURL, &http.Client{Transport: transport})
	if err != nil {
		return installation{}, err
	}
	inst := installation{
		client:         client,
		transport:      transport,
		installationID: id,
	}
	if err := c.cacheInstallationClient(ctx, inst); err != nil {
		return installation{}, err
	}
	return inst, nil
}

func (c *appsClient) cacheInstallationClient(ctx context.Context, inst installation) error {
	// Only lists repos for this installation.
	// GitHub API endpoint requires installation-specific credentials for this.
	reposResp, _, err := inst.client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		return fmt.Errorf("list which repos to cache client for: %w", err)
	}
	for _, repoData := range reposResp.Repositories {
		refSlim := RepoRefSlim{
			Owner: repoData.Owner.GetLogin(),
			Repo:  repoData.GetName(),
		}
		if _, exists := c.installationPerRepo[refSlim]; exists {
			// no need to override existing ones
			continue
		}
		log.Debug().
			Int64("installation", inst.installationID).
			Stringer("repo", refSlim).
			Msg("Caching GitHub client for repo.")
		c.installationPerRepo[refSlim] = inst
	}
	return nil
}

func (c *appsClient) findInstallationIDForRepo(ctx context.Context, repo RepoRef) (int64, error) {
	inst, resp, err := c.noInstallClient.Apps.FindRepositoryInstallation(ctx, repo.Owner, repo.Repo)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return 0, fmt.Errorf("GitHub App is not installed for repo: %s", repo)
		}
		return 0, fmt.Errorf("find GitHub App installation for repo: %w", err)
	}
	return inst.GetID(), nil
}

func newAppsTransport(ghCfg *config.GitHub) (*ghinstallation.AppsTransport, error) {
	transport := http.DefaultTransport
	switch {
	case ghCfg.Auth.App.PrivateKeyPEM != nil:
		return ghinstallation.NewAppsTransportFromPrivateKey(
			transport,
			ghCfg.Auth.App.ID,
			ghCfg.Auth.App.PrivateKeyPEM.PrivateKey()), nil
	case ghCfg.Auth.App.PrivateKeyPath != nil:
		appsTransport, err := ghinstallation.NewAppsTransportKeyFromFile(
			transport,
			ghCfg.Auth.App.ID,
			*ghCfg.Auth.App.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read config privateKeyPath: %w", err)
		}
		return appsTransport, nil
	default:
		return nil, errors.New("must set GitHub auth config privateKeyPem or privateKeyPath")
	}
}
