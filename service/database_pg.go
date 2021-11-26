package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/athenianco/cloud-common/dbs"
)

const (
	pgConnMaxLifetime = time.Minute
	pgConnMaxIdleTime = 30 * time.Second
)

var _ Database = (*pgDatabase)(nil)

// pgDatabase is a postgres database where services info is stored.
type pgDatabase struct {
	db *pgxpool.Pool
}

// OpenDatabaseFromEnv opens default postgres database based on environment variable:
// SERVICE_DATABASE_URI
func OpenDatabaseFromEnv() (Database, error) {
	dbURI := os.Getenv("SERVICE_DATABASE_URI")
	if dbURI == "" {
		return nil, errors.New("SERVICE_DATABASE_URI is not set")
	}

	db, err := OpenDatabase(context.Background(), dbURI)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func OpenDatabase(ctx context.Context, dbURI string) (Database, error) {
	return openDatabase(ctx, dbURI)
}

func OpenTestDatabase(ctx context.Context, dbURI string) (TestDatabase, error) {
	return openDatabase(ctx, dbURI)
}

// openDatabase opens postgres connection
func openDatabase(ctx context.Context, dbURI string) (*pgDatabase, error) {
	config, err := pgxpool.ParseConfig(processAddress(dbURI))
	if err != nil {
		return nil, err
	}
	config.ConnConfig.PreferSimpleProtocol = true
	config.MaxConnLifetime = pgConnMaxLifetime
	config.MaxConnIdleTime = pgConnMaxIdleTime

	conn, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return &pgDatabase{db: conn}, nil
}

func (db *pgDatabase) RegisterService(ctx context.Context, name string) (bool, error) {
	tx, err := db.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	svc, err := db.getService(ctx, tx, name)
	if err != nil && err != dbs.ErrNotFound {
		return false, err
	} else if err == nil {
		return svc.Enabled, tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `INSERT INTO services(name) VALUES($1);`, newNullString(name))
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

func (db *pgDatabase) SwitchService(ctx context.Context, name string, enabled bool) error {
	tx, err := db.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	svc, err := db.getService(ctx, tx, name)
	if err != nil {
		return err
	}
	if enabled == svc.Enabled {
		return nil
	}
	if _, err := tx.Exec(ctx, `UPDATE services SET enabled = $2 WHERE name = $1;`, newNullString(name), enabled); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (db *pgDatabase) GetService(ctx context.Context, name string) (*Service, error) {
	tx, err := db.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	svc, err := db.getService(ctx, tx, name)
	if err != nil {
		return nil, err
	}
	return svc, tx.Commit(ctx)
}

func (db *pgDatabase) ListServices(ctx context.Context) (Iterator, error) {
	rows, err := db.db.Query(ctx, `SELECT name, enabled FROM services;`)
	if err != nil {
		return nil, err
	}
	return &serviceIter{
		rows: rows,
	}, nil
}

func (db *pgDatabase) Cleanup(ctx context.Context) error {
	_, err := db.db.Exec(ctx, `DELETE FROM services;`)
	return err
}

func (db *pgDatabase) getService(ctx context.Context, tx pgx.Tx, name string) (*Service, error) {
	row := tx.QueryRow(ctx, `SELECT name, enabled FROM services WHERE name = $1;`, name)
	svc, err := scanService(row)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (db *pgDatabase) Close() error {
	db.db.Close()
	return nil
}

func scanService(sc dbs.Scanner) (Service, error) {
	var (
		name    string
		enabled bool
	)
	err := sc.Scan(&name, &enabled)
	if err == pgx.ErrNoRows {
		err = dbs.ErrNotFound
	}

	return Service{
		Name:    name,
		Enabled: enabled,
	}, err
}

type serviceIter struct {
	rows    pgx.Rows
	current *Service

	err error
}

func (it *serviceIter) Next() bool {
	if it.err != nil {
		return false
	}
	if !it.rows.Next() {
		return false
	}
	iss, err := scanService(it.rows)
	if err != nil {
		it.err = err
		return false
	}
	it.current = &iss
	return true
}

func (it *serviceIter) Value() *Service {
	if it.err != nil {
		return nil
	}
	return it.current
}

func (it *serviceIter) Err() error {
	return it.err
}

func (it *serviceIter) Close() error {
	it.rows.Close()
	return it.err
}

func newNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func processAddress(addr string) string {
	return strings.Replace(addr, "&binary_parameters=yes", "", -1)
}
