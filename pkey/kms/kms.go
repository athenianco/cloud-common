package kms

import (
	"context"
	"errors"
	"fmt"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/athenianco/cloud-common/pkey"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type KMS struct {
	provider pkey.PKeyProvider
	keyName  string
	cli      *kms.KeyManagementClient
}

func NewFromEnv(ctx context.Context, pkey pkey.PKeyProvider) (pkey.PKeyProvider, error) {
	keyName := os.Getenv("GOOGLE_KMS_KEY")
	if keyName == "" {
		return nil, errors.New("GOOGLE_KMS_KEY was not set")
	}
	return New(ctx, keyName, pkey)
}

func New(ctx context.Context, keyName string, pkey pkey.PKeyProvider) (pkey.PKeyProvider, error) {
	kmsCli, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	return &KMS{
		provider: pkey,
		keyName:  keyName,
		cli:      kmsCli,
	}, nil
}

func (k *KMS) GetPrivateKeyData(ctx context.Context, appId int64) ([]byte, error) {
	encData, err := k.provider.GetPrivateKeyData(ctx, appId)
	if err != nil {
		return nil, err
	}
	return k.decryptPrivateKeyData(ctx, encData)
}

func (k *KMS) decryptPrivateKeyData(ctx context.Context, data []byte) ([]byte, error) {
	req := &kmspb.DecryptRequest{
		Name:       k.keyName,
		Ciphertext: data,
	}

	result, err := k.cli.Decrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}
	return result.GetPlaintext(), nil
}

func (k *KMS) Close() error { return k.cli.Close() }
