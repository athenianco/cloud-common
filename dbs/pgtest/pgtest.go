package pgtest

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
)

const (
	pgVers = "11"
)

type DBFunc func(t testing.TB) (string, func())

func NewDatabasePool(t testing.TB, schemas ...string) (DBFunc, func()) {
	for _, schema := range schemas {
		_, err := os.Stat(schema)
		require.NoError(t, err)
	}
	return NewDatabasePoolWith(t, func(addr string) error {
		return importSchemas(addr, schemas...)
	})
}

func NewDatabasePoolMigrate(t testing.TB, dir string) (DBFunc, func()) {
	return NewDatabasePoolWith(t, func(addr string) error {
		m, err := migrate.New("file://"+dir, addr)
		if err != nil {
			return err
		}
		defer m.Close()
		return m.Up()
	})
}

func NewDatabasePoolWith(t testing.TB, schema func(addr string) error) (DBFunc, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	cont, err := pool.Run("postgres", pgVers, []string{
		"POSTGRES_PASSWORD=postgres",
	})
	if err != nil {
		t.Fatal(err)
	}

	const port = "5432/tcp"
	addrFmt := `postgres://postgres:postgres@` + cont.GetHostPort(port) + `/%s?sslmode=disable`
	addr := fmt.Sprintf(addrFmt, "")

	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", addr)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	if err != nil {
		cont.Close()
		t.Fatal(err)
	}

	var id int64

	return func(t testing.TB) (string, func()) {
			db, err := sql.Open("postgres", fmt.Sprintf(addrFmt, ""))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			dbname := fmt.Sprintf("testdb_%d", atomic.AddInt64(&id, 1))

			ctx := context.Background()

			_, err = db.ExecContext(ctx, `CREATE DATABASE `+dbname)
			if err != nil {
				t.Fatal(err)
			}

			addr := fmt.Sprintf(addrFmt, dbname)
			if schema != nil {
				err = schema(addr)
				if err != nil {
					_, _ = db.Exec(`DROP DATABASE ` + dbname)
					t.Fatal(err)
				}
			}

			return addr, func() {
				db, err := sql.Open("postgres", fmt.Sprintf(addrFmt, ""))
				if err != nil {
					return
				}
				defer db.Close()

				_, _ = db.Exec(`DROP DATABASE ` + dbname)
			}
		}, func() {
			cont.Close()
		}
}

func importSchemas(addr string, schemaPaths ...string) error {
	for _, s := range schemaPaths {
		if err := importSchema(addr, s); err != nil {
			return err
		}
	}
	return nil
}

func nextStmt(s []byte) (stmt, rest []byte) {
	i := bytes.IndexByte(s, ';')
	if i < 0 {
		return s, nil
	}
	j := bytes.Index(s, []byte("$$"))
	if i < j || j < 0 {
		return s[:i], s[i+1:]
	}
	k := bytes.Index(s[j+2:], []byte("$$"))
	k += j + 2
	i = bytes.IndexByte(s[k:], ';')
	i += k
	return s[:i], s[i+1:]
}

func sqlSplitStmts(data []byte) []string {
	// remove comments
	lines := bytes.Split(data, []byte("\n"))
	for i := range lines {
		if bytes.HasPrefix(lines[i], []byte("--")) {
			lines[i] = nil
		}
	}
	data = bytes.Join(lines, []byte("\n"))

	var stmts []string
	for len(data) > 0 {
		s, rest := nextStmt(data)
		data = rest
		s = bytes.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		stmts = append(stmts, string(s))
	}
	return stmts
}

func importSchema(addr, schemaPath string) error {
	data, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	stmts := sqlSplitStmts(data)

	db, err := sql.Open("postgres", addr)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	for _, stmt := range stmts {
		_, err = db.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("error importing schema: %v\nSQL: %s", err, stmts)
		}
	}
	return nil
}
