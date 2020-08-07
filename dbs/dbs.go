// Package dbs contains shared code for DB implementations.
package dbs

import "errors"

var (
	// ErrNotFound is returned when a DB record was not found.
	ErrNotFound = errors.New("db: not found")
)

// IteratorBase contains methods common to DB result set iterators.
type IteratorBase interface {
	Next() bool
	Err() error
	Close() error
}

// Scanner is a common interface for a row scanners.
type Scanner interface {
	Scan(args ...interface{}) error
}
