package service_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/athenianco/cloud-common/dbs/pgtest"
	"github.com/athenianco/cloud-common/service"
	"github.com/athenianco/cloud-common/service/servicetest"
)

func makeDatabasePool(t testing.TB) (servicetest.DBFunc, func()) {
	pool, closer := pgtest.NewDatabasePoolWith(t, func(addr string) error {
		ctx := context.Background()
		conn, err := pgxpool.Connect(ctx, addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		_, err = conn.Exec(ctx, `
CREATE TABLE services (
    name text NOT NULL,
    enabled bool NOT NULL DEFAULT TRUE,
    PRIMARY KEY(name)
);
`)
		return err
	})

	return func(t testing.TB) (service.TestDatabase, func()) {
		addr, closer := pool(t)

		gdb, err := service.OpenTestDatabase(context.Background(), addr)
		if err != nil {
			t.Fatal(err)
		}

		return gdb, func() {
			gdb.Close()
			closer()
		}
	}, closer
}

func TestPostgres(t *testing.T) {
	servicetest.RunDatabaseTest(t, makeDatabasePool)
}
