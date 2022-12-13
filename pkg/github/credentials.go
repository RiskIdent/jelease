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
	"crypto/rsa"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Credentials interface {
	Apply(req *http.Request) error
}

func NewTokenCredentils(username, token string) Credentials {
	return tokenCredentials{username: username, token: token}
}

type tokenCredentials struct {
	username string
	token    string
}

func (c tokenCredentials) Apply(req *http.Request) error {
	req.URL.User = url.UserPassword(c.username, c.token)
	return nil
}

func NewAppCredentials(rsaKey *rsa.PrivateKey, appID string) Credentials {
	return appCredentials{rsaKey: rsaKey, appID: appID}
}

type appCredentials struct {
	rsaKey *rsa.PrivateKey
	appID  string
}

// [https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app]
func (c appCredentials) Apply(req *http.Request) error {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		Issuer:    c.appID,
	})
	tokenStr, err := token.SignedString(c.rsaKey)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	// TODO: Request the access token for the installation,
	// required for performing requests targeting repos
	return errors.New("not implemented")
}
