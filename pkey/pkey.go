package pkey

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const pemPKeyBlock = "RSA PRIVATE KEY"

// TODO: rename because this interface will be used not only for the private key data

type Provider interface {
	// GetPrivateKey fetches the private key.
	GetPrivateKey(ctx context.Context, id string) ([]byte, error)

	// GetPrivateKeyData fetches the private key.
	//
	// Deprecated: accepts int ID, which is harder to keep unique. Use GetPrivateKey instead.
	GetPrivateKeyData(ctx context.Context, appID int64) ([]byte, error)

	Close() error
}

// Processor is an interface that allows to encode, encrypt or store private key data.
type Processor interface {
	// PutPrivateKey stores the private key.
	PutPrivateKey(ctx context.Context, id string, data []byte) error
	// DelPrivateKey removes the private key.
	DelPrivateKey(ctx context.Context, id string) error

	// ProcessPrivateKeyData stores the private key.
	//
	// Deprecated: accepts int ID, which is harder to keep unique. Use PutPrivateKey instead.
	ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) error
	// DeletePrivateKeyData removes the private key.
	//
	// Deprecated: accepts int ID, which is harder to keep unique. Use DelPrivateKey instead.
	DeletePrivateKeyData(ctx context.Context, accID int64) error

	Close() error
}

func ParsePKey(data []byte) (*rsa.PrivateKey, error) {
	for {
		p, rest := pem.Decode(data)
		if p == nil {
			return nil, errors.New("cannot decode PEM private key")
		}
		if p.Type != pemPKeyBlock {
			data = rest
			continue
		}
		data = p.Bytes
		break
	}
	pKey, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		pKey, err = x509.ParsePKCS1PrivateKey(data)
		if err != nil {
			return nil, err
		}
	}
	rpKey, ok := pKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is invalid")
	}
	return rpKey, nil
}
