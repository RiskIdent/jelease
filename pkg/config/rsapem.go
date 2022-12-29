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
