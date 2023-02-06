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
	"encoding"

	"github.com/golang-jwt/jwt/v4"
	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

type RSAPrivateKeyPEM struct {
	pem []byte
	key *rsa.PrivateKey
}

// Ensure the type implements the interfaces
var _ pflag.Value = &Template{}
var _ encoding.TextUnmarshaler = &Template{}
var _ jsonSchemaInterface = Template{}

func (k *RSAPrivateKeyPEM) PrivateKey() *rsa.PrivateKey {
	return k.key
}

func (k *RSAPrivateKeyPEM) String() string {
	return string(k.pem)
}

func (k *RSAPrivateKeyPEM) Set(value string) error {
	return k.UnmarshalText([]byte(value))
}

func (RSAPrivateKeyPEM) Type() string {
	return "RSA PEM"
}

func (k *RSAPrivateKeyPEM) UnmarshalText(text []byte) error {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(text)
	if err != nil {
		return err
	}
	*k = RSAPrivateKeyPEM{
		pem: text,
		key: key,
	}
	return nil
}

func (k *RSAPrivateKeyPEM) MarshalText() ([]byte, error) {
	return k.pem, nil
}

func (RSAPrivateKeyPEM) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:            "PEM-formatted RSA private key",
		ContentMediaType: "application/x-pem-file",
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			{
				Type: "null",
			},
		},
		Examples: []any{
			`-----BEGIN RSA PRIVATE KEY-----
bG9yZW0gaXBzdW0sIHNlY3JldCBrZXkgYmFzZTY0IGdvZXMgaGVyZS4u
-----END RSA PRIVATE KEY-----
`,
		},
	}
}
