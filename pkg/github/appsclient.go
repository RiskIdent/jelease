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
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog/log"
)

type appsClient struct {
	appsTransport   *ghinstallation.AppsTransport
	noInstallClient *github.Client
	clientsPerRepo  map[RepoRefSlim]installationClient
	baseURL         *string
}

type installationClient struct {
	*github.Client
	installationID int64
}

func NewAppClientFactory(ghCfg *config.GitHub) (ClientFactory, error) {
	appsTransport, err := newAppsTransport(ghCfg)
	if err != nil {
		return nil, err
	}
	nonInstallClient, err := newClientEnterpriceOrPublic(ghCfg.URL, &http.Client{Transport: appsTransport})
	if err != nil {
		return nil, err
	}
	return &appsClient{
		appsTransport:   appsTransport,
		noInstallClient: nonInstallClient,
		clientsPerRepo:  make(map[RepoRefSlim]installationClient),
		baseURL:         ghCfg.URL,
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

func (c *appsClient) NewClientForRepo(ctx context.Context, repo RepoRef) (*github.Client, error) {
	if client, ok := c.clientsPerRepo[repo.Slim()]; ok {
		log.Debug().
			Int64("installation", client.installationID).
			Stringer("repo", repo).
			Msg("Using cached client for repo.")
		return client.Client, nil
	}
	id, err := c.findInstallationIDForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	transport := ghinstallation.NewFromAppsTransport(c.appsTransport, id)
	client, err := newClientEnterpriceOrPublic(c.baseURL, &http.Client{Transport: transport})
	if err != nil {
		return nil, err
	}
	instClient := installationClient{
		Client:         client,
		installationID: id,
	}
	reposResp, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, repoData := range reposResp.Repositories {
		refSlim := RepoRefSlim{
			Owner: repoData.Owner.GetLogin(),
			Repo:  repoData.GetName(),
		}
		log.Debug().
			Int64("installation", id).
			Stringer("repo", refSlim).
			Msg("Caching client for repo.")
		c.clientsPerRepo[refSlim] = instClient
	}
	return client, nil
}

func (c *appsClient) findInstallationIDForRepo(ctx context.Context, repo RepoRef) (int64, error) {
	inst, _, err := c.noInstallClient.Apps.FindRepositoryInstallation(ctx, repo.Owner, repo.Repo)
	if err != nil {
		return 0, fmt.Errorf("find GitHub App installation for repo: %w", err)
	}
	return inst.GetID(), nil
}

func newAppsTransport(ghCfg *config.GitHub) (*ghinstallation.AppsTransport, error) {
	transport := http.DefaultTransport
	var appsTransport *ghinstallation.AppsTransport
	switch {
	case ghCfg.Auth.App.PrivateKeyPEM != nil:
		appsTransport = ghinstallation.NewAppsTransportFromPrivateKey(
			transport,
			ghCfg.Auth.App.ID,
			ghCfg.Auth.App.PrivateKeyPEM.PrivateKey())
	case ghCfg.Auth.App.PrivateKeyPath != nil:
		at, err := ghinstallation.NewAppsTransportKeyFromFile(
			transport,
			ghCfg.Auth.App.ID,
			*ghCfg.Auth.App.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read config privateKeyPath: %w", err)
		}
		appsTransport = at
	default:
		return nil, errors.New("must set GitHub auth config privateKeyPem or privateKeyPath")
	}
	if ghCfg.URL != nil {
		appsTransport.BaseURL = *ghCfg.URL
	}
	return appsTransport, nil
}
