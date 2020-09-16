package kms

import (
	"context"
	"errors"
	"fmt"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/athenianco/cloud-common/pkeyprocessor"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type KMS struct {
	processor pkeyprocessor.PKeyProcessor
	keyName   string
	cli       *kms.KeyManagementClient
}

func NewFromEnv(ctx context.Context, processor pkeyprocessor.PKeyProcessor) (pkeyprocessor.PKeyProcessor, error) {
	keyName := os.Getenv("GOOGLE_KMS_KEY")
	if keyName == "" {
		return nil, errors.New("GOOGLE_KMS_KEY was not set")
	}
	return New(ctx, keyName, processor)
}

func New(ctx context.Context, keyName string, processor pkeyprocessor.PKeyProcessor) (pkeyprocessor.PKeyProcessor, error) {
	kmsCli, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	return &KMS{
		processor: processor,
		keyName:   keyName,
		cli:       kmsCli,
	}, nil
}

func (k *KMS) ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) ([]byte, error) {
	if k.processor != nil {
		processedData, err := k.processor.ProcessPrivateKeyData(ctx, accID, data)
		if err != nil {
			return nil, err
		}
		data = processedData
	}
	return k.encryptPrivateKeyData(ctx, data)
}

func (k *KMS) encryptPrivateKeyData(ctx context.Context, data []byte) ([]byte, error) {
	req := &kmspb.EncryptRequest{
		Name:      k.keyName,
		Plaintext: data,
	}

	result, err := k.cli.Encrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}
	return result.GetCiphertext(), nil
}

func (k *KMS) Close() error {
	if k.processor != nil {
		if err := k.processor.Close(); err != nil {
			return err
		}
	}
	return k.cli.Close()
}
