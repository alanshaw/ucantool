package identity

import (
	"bytes"
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/fil-forge/ucantone/principal/ed25519"
)

// EncodeEd25519SignerToPEM encodes an Ed25519 signer to a PKCS#8 PEM format.
func EncodeEd25519SignerToPEM(signer ed25519.Signer) ([]byte, error) {
	privateKey := crypto_ed25519.PrivateKey(signer.Raw())
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling ed25519 private key: %w", err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	buffer := new(bytes.Buffer)
	if err := pem.Encode(buffer, privateKeyBlock); err != nil {
		return nil, fmt.Errorf("encoding ed25519 private key: %w", err)
	}

	return buffer.Bytes(), nil
}

// DecodeEd25519SignerFromPEM loads an Ed25519 private key from a PKCS#8 PEM.
func DecodeEd25519SignerFromPEM(pemData []byte) (ed25519.Signer, error) {
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
