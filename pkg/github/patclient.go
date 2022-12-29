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
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type patClient struct {
	cred git.Credentials
	gh   *github.Client
}

func NewPATClientFactory(ghCfg *config.GitHub) (ClientFactory, error) {
	httpClient := newOAuthHTTPClient(ghCfg.Auth.Token)
	gh, err := newClientEnterpriceOrPublic(ghCfg.URL, httpClient)
	if err != nil {
		return nil, err
	}
	return &patClient{
		cred: git.Credentials{
			Username: "git",
			Password: ghCfg.Auth.Token,
		},
		gh: gh,
	}, nil
}

func (c *patClient) TestConnection(ctx context.Context) error {
	// get empty means get the currently authenticated user
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("get current GitHub user: %w", err)
	}
	log.Debug().
		Str("login", user.GetLogin()).
		Str("name", user.GetName()).
		Msg("Authenticated as GitHub user.")
	return nil
}

func (c *patClient) GitCredentialsForRepo(context.Context, RepoRef) (git.Credentials, error) {
	// use the same credentials for all repos
	return c.cred, nil
}

func (c *patClient) NewClientForRepo(context.Context, RepoRef) (*github.Client, error) {
	// use the same client for all repos
	return c.gh, nil
}

func newOAuthHTTPClient(token string) *http.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(context.TODO(), tokenSource)
}
