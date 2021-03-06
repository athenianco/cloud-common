package gcs

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	"cloud.google.com/go/storage"

	"github.com/athenianco/cloud-common/dbs"
	"github.com/athenianco/cloud-common/pkey"
)

type Provider struct {
	bucket string
	cli    *storage.Client
}

func NewProviderFromEnv(ctx context.Context) (pkey.Provider, error) {
	bucket := os.Getenv("GOOGLE_KMS_BUCKET")
	if bucket == "" {
		return nil, errors.New("GOOGLE_KMS_BUCKET was not set")
	}
	return NewProvider(ctx, bucket)
}

func NewProvider(ctx context.Context, bucket string) (pkey.Provider, error) {
	storageCli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Provider{
		bucket: bucket,
		cli:    storageCli,
	}, nil
}

func (p *Provider) GetPrivateKeyData(ctx context.Context, appId int64) ([]byte, error) {
	return p.GetPrivateKey(ctx, strconv.Itoa(int(appId)))
}

func (p *Provider) GetPrivateKey(ctx context.Context, id string) ([]byte, error) {
	b := p.cli.Bucket(p.bucket)
	obj := b.Object(id)
	r, err := obj.NewReader(ctx)
	if err == storage.ErrBucketNotExist || err == storage.ErrObjectNotExist {
		return nil, dbs.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}

func (p *Provider) Close() error { return p.cli.Close() }

type Processor struct {
	bucket string
	cli    *storage.Client
}

func NewProcessorFromEnv(ctx context.Context) (pkey.Processor, error) {
	bucket := os.Getenv("GOOGLE_KMS_BUCKET")
	if bucket == "" {
		return nil, errors.New("GOOGLE_KMS_BUCKET was not set")
	}
	return NewProcessor(ctx, bucket)
}

func NewProcessor(ctx context.Context, bucket string) (pkey.Processor, error) {
	storageCli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Processor{
		bucket: bucket,
		cli:    storageCli,
	}, nil
}

func (p *Processor) ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) error {
	return p.PutPrivateKey(ctx, strconv.Itoa(int(accID)), data)
}

func (p *Processor) PutPrivateKey(ctx context.Context, id string, data []byte) error {
	b := p.cli.Bucket(p.bucket)
	obj := b.Object(id)
	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err := w.Write(data)
	return err
}

func (p *Processor) DeletePrivateKeyData(ctx context.Context, accID int64) error {
	return p.DelPrivateKey(ctx, strconv.Itoa(int(accID)))
}

func (p *Processor) DelPrivateKey(ctx context.Context, id string) error {
	b := p.cli.Bucket(p.bucket)
	obj := b.Object(id)

	return obj.Delete(ctx)
}

func (p *Processor) Close() error {
	return p.cli.Close()
}
