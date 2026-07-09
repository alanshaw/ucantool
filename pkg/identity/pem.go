package identity

import (
	"bytes"
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/multikey/ed25519"
)

// EncodeSignerToPEM encodes a signer to a PKCS#8 PEM format. The signer's key
// should be of a type supported by ["crypto/x509".MarshalPKCS8PrivateKey].
func EncodeSignerToPEM(signer multikey.Signer) ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(signer.PrivateKey())
	if err != nil {
		return nil, fmt.Errorf("marshaling private key of signer %s: %w", signer, err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	buffer := new(bytes.Buffer)
	if err := pem.Encode(buffer, privateKeyBlock); err != nil {
		return nil, fmt.Errorf("encoding private key of signer %s: %w", signer, err)
	}

	return buffer.Bytes(), nil
}

// DecodeSignerFromPEM loads a private key from a PKCS#8 PEM as a signer.
// Currently, only Ed25519 keys are supported.
func DecodeSignerFromPEM(pemData []byte) (multikey.Signer, error) {
	var privateKey *crypto_ed25519.PrivateKey
	rest := pemData
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining

		if block.Type == "PRIVATE KEY" {
			parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parsing PKCS#8 private key: %w", err)
			}

			key, ok := parsedKey.(crypto_ed25519.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("key is not an Ed25519 private key")
			}
			privateKey = &key
			break
		}
	}

	if privateKey == nil {
		return nil, fmt.Errorf("no PRIVATE KEY block found in PEM file")
	}

	return ed25519.FromRaw(privateKey.Seed())
}
