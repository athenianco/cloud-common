// Package dbs contains shared code for DB implementations.

package dbs

import "errors"

var (
	// ErrNotFound is returned when a DB record was not found.
	ErrNotFound = errors.New("db: not found")
)
