package gcs

import (
	"context"
	"errors"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/athenianco/cloud-common/pkeyprocessor"
)

type GCS struct {
	processor pkeyprocessor.PKeyProcessor
	bucket    string
	cli       *storage.Client
}

func NewFromEnv(ctx context.Context, processor pkeyprocessor.PKeyProcessor) (pkeyprocessor.PKeyProcessor, error) {
	bucket := os.Getenv("GOOGLE_KMS_BUCKET")
	if bucket == "" {
		return nil, errors.New("GOOGLE_KMS_BUCKET was not set")
	}
	return New(ctx, bucket, processor)
}

func New(ctx context.Context, bucket string, processor pkeyprocessor.PKeyProcessor) (pkeyprocessor.PKeyProcessor, error) {
	storageCli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GCS{
		processor: processor,
		bucket:    bucket,
		cli:       storageCli,
	}, nil
}

func (g *GCS) ProcessPrivateKeyData(ctx context.Context, accID int64, data []byte) ([]byte, error) {
	if g.processor != nil {
		processedData, err := g.processor.ProcessPrivateKeyData(ctx, accID, data)
		if err != nil {
			return nil, err
		}
		data = processedData
	}
	b := g.cli.Bucket(g.bucket)
	obj := b.Object(strconv.Itoa(int(accID)))
	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err := w.Write(data)
	return data, err
}

func (g *GCS) Close() error {
	if g.processor != nil {
		if err := g.processor.Close(); err != nil {
			return err
		}
	}
	return g.cli.Close()
}
