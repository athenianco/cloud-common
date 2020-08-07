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

type GCS struct {
	bucket string
	cli    *storage.Client
}

func NewFromEnv(ctx context.Context) (pkey.PKeyProvider, error) {
	bucket := os.Getenv("GOOGLE_KMS_BUCKET")
	if bucket == "" {
		return nil, errors.New("GOOGLE_KMS_BUCKET was not set")
	}
	return New(ctx, bucket)
}

func New(ctx context.Context, bucket string) (pkey.PKeyProvider, error) {
	storageCli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GCS{
		bucket: bucket,
		cli:    storageCli,
	}, nil
}

func (g *GCS) GetPrivateKeyData(ctx context.Context, appId int64) ([]byte, error) {
	b := g.cli.Bucket(g.bucket)
	obj := b.Object(strconv.Itoa(int(appId)))
	r, err := obj.NewReader(ctx)
	if err == storage.ErrBucketNotExist || err == storage.ErrObjectNotExist {
		return nil, dbs.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}

func (g *GCS) Close() error { return g.cli.Close() }
