package service

import (
	"context"

	"github.com/athenianco/cloud-common/dbs"
)

type Service struct {
	Name    string
	Enabled bool
}

type Database interface {
	RegisterService(ctx context.Context, name string) (bool, error)
	SwitchService(ctx context.Context, name string, enabled bool) error
	GetService(ctx context.Context, name string) (*Service, error)
	ListServices(ctx context.Context) (Iterator, error)
	Close() error
}

type TestDatabase interface {
	Database
	Cleanup(ctx context.Context) error
}

type Iterator interface {
	dbs.IteratorBase
	Value() *Service
}
