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

type Provider struct {
	provider pkey.Provider
	keyName  string
	cli      *kms.KeyManagementClient
}

func NewProviderFromEnv(ctx context.Context, pkey pkey.Provider) (pkey.Provider, error) {
	keyName := os.Getenv("GOOGLE_KMS_KEY")
	if keyName == "" {
		return nil, errors.New("GOOGLE_KMS_KEY was not set")
	}
	return NewProvider(ctx, keyName, pkey)
}

func NewProvider(ctx context.Context, keyName string, pkey pkey.Provider) (pkey.Provider, error) {
	kmsCli, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Provider{
		provider: pkey,
		keyName:  keyName,
		cli:      kmsCli,
	}, nil
}

func (p *Provider) GetPrivateKeyData(ctx context.Context, appId int64) ([]byte, error) {
	encData, err := p.provider.GetPrivateKeyData(ctx, appId)
	if err != nil {
		return nil, err
	}
	return p.decryptPrivateKeyData(ctx, encData)
}

func (p *Provider) decryptPrivateKeyData(ctx context.Context, data []byte) ([]byte, error) {
	req := &kmspb.DecryptRequest{
		Name:       p.keyName,
		Ciphertext: data,
	}

	result, err := p.cli.Decrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}
	return result.GetPlaintext(), nil
}

func (p *Provider) Close() error { return p.cli.Close() }

type Processor struct {
	processor pkey.Processor
	keyName   string
	cli       *kms.KeyManagementClient
}

func NewProcessorFromEnv(ctx context.Context, processor pkey.Processor) (pkey.Processor, error) {
	keyName := os.Getenv("GOOGLE_KMS_KEY")
	if keyName == "" {
		return nil, errors.New("GOOGLE_KMS_KEY was not set")
	}
	return NewProcessor(ctx, keyName, processor)
}

func NewProcessor(ctx context.Context, keyName string, processor pkey.Processor) (pkey.Processor, error) {
	if processor == nil {
		return nil, errors.New("nested processor is nil")
	}
	kmsCli, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Processor{
		processor: processor,
		keyName:   keyName,
		cli:       kmsCli,
	}, nil
}

func (p *Processor) ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) error {
	data, err := p.encryptPrivateKeyData(ctx, data)
	if err != nil {
		return err
	}
	return p.processor.ProcessPrivateKeyData(ctx, accID, data)
}

func (p *Processor) encryptPrivateKeyData(ctx context.Context, data []byte) ([]byte, error) {
	req := &kmspb.EncryptRequest{
		Name:      p.keyName,
		Plaintext: data,
	}

	result, err := p.cli.Encrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}
	return result.GetCiphertext(), nil
}

func (p *Processor) DeletePrivateKeyData(ctx context.Context, accID int64) error {
	return p.processor.DeletePrivateKeyData(ctx, accID)
}

func (p *Processor) Close() error {
	if err := p.processor.Close(); err != nil {
		return err
	}
	return p.cli.Close()
}
