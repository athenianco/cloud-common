package pkeyprocessor

import "context"

// TODO: remove key?

// PKeyProcessor is an interface that allows to encode, encrypt or store private key data
type PKeyProcessor interface {
	ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) ([]byte, error)
	Close() error
}
