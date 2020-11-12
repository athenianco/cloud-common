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
	GetPrivateKeyData(ctx context.Context, appID int64) ([]byte, error)
	Close() error
}

// PKeyProcessor is an interface that allows to encode, encrypt or store private key data
type Processor interface {
	ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) error
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
