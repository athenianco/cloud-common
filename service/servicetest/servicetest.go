package servicetest

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athenianco/cloud-common/dbs"
	"github.com/athenianco/cloud-common/service"
)

type DBServerFunc func(t testing.TB) (DBFunc, func())

type DBFunc func(t testing.TB) (service.TestDatabase, func())

func RunDatabaseTest(t *testing.T, pool DBServerFunc) {
	tests := []struct {
		name string
		run  func(testing.TB, service.TestDatabase)
	}{
		{"RegisterService", testRegisterGetService},
		{"EnableDisableService", testEnableDisableService},
		{"ListServices", testListServices},
	}

	fnc, closer := pool(t)
	defer closer()

	db, dbCloser := fnc(t)
	defer dbCloser()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.run(t, db)
			require.NoError(t, db.Cleanup(context.Background()))
		})
	}
}

func testRegisterGetService(t testing.TB, db service.TestDatabase) {
	ctx := context.Background()

	reg := func(name string, exists bool) {
		existsAct, err := db.RegisterService(ctx, name)
		require.NoError(t, err)
		require.Equal(t, exists, existsAct)
	}
	get := func(name string, exists bool) {
		svc, err := db.GetService(ctx, name)
		require.NoError(t, err)
		require.Equal(t, &service.Service{Name: name, Enabled: exists}, svc)
	}

	_, err := db.GetService(ctx, "aaaaa")
	require.Equal(t, dbs.ErrNotFound, err)

	_, err = db.RegisterService(ctx, "")
	require.Error(t, err)

	reg("svc", true)
	get("svc", true)
	reg("svc", true)
	get("svc", true)
	reg("svc1", true)
	get("svc", true)
	get("svc1", true)
}

func testEnableDisableService(t testing.TB, db service.TestDatabase) {
	ctx := context.Background()

	reg := func(name string, exists bool) {
		existsAct, err := db.RegisterService(ctx, name)
		require.NoError(t, err)
		require.Equal(t, exists, existsAct)
	}
	get := func(name string, exists bool) {
		svc, err := db.GetService(ctx, name)
		require.NoError(t, err)
		require.Equal(t, &service.Service{Name: name, Enabled: exists}, svc)
	}
	switchService := func(name string, enabled bool) {
		err := db.SwitchService(ctx, name, enabled)
		require.NoError(t, err)
	}
	enable := func(name string) { switchService(name, true) }
	disable := func(name string) { switchService(name, false) }

	require.Equal(t, dbs.ErrNotFound, db.SwitchService(ctx, "aaaa", false))
	require.Equal(t, dbs.ErrNotFound, db.SwitchService(ctx, "aaaa", true))

	reg("svc", true)
	enable("svc")
	get("svc", true)
	disable("svc")
	get("svc", false)
	disable("svc")
	get("svc", false)
	reg("svc", false)
	enable("svc")
	get("svc", true)

	checkPair := func(exists1, exists2 bool) {
		get("svc", exists1)
		get("svc1", exists2)
	}
	reg("svc1", true)
	checkPair(true, true)
	disable("svc")
	checkPair(false, true)
	enable("svc")
	checkPair(true, true)
	disable("svc")
	disable("svc1")
	checkPair(false, false)
	enable("svc")
	checkPair(true, false)
	enable("svc1")
	checkPair(true, true)
}

func testListServices(t testing.TB, db service.TestDatabase) {
	ctx := context.Background()

	reg := func(name string, exists bool) {
		existsAct, err := db.RegisterService(ctx, name)
		require.NoError(t, err)
		require.Equal(t, exists, existsAct)
	}

	servicesExp := make(map[string]service.Service)
	checkServices := func() {
		it, err := db.ListServices(ctx)
		require.NoError(t, err)
		defer it.Close()

		servicesAct := make(map[string]service.Service)
		for it.Next() {
			svc := *it.Value()
			servicesAct[svc.Name] = svc
		}
		require.NoError(t, it.Err())
		require.Equal(t, servicesExp, servicesAct)
	}
	checkServices()

	for i := 0; i < 4; i++ {
		key := "svc" + strconv.Itoa(i)
		servicesExp[key] = service.Service{Name: key, Enabled: true}
		reg(key, true)
	}
	checkServices()

	switchService := func(name string, enabled bool) {
		err := db.SwitchService(ctx, name, enabled)
		require.NoError(t, err)
		servicesExp[name] = service.Service{Name: name, Enabled: enabled}
	}
	enable := func(name string) { switchService(name, true) }
	disable := func(name string) { switchService(name, false) }
	enable("svc0")
	checkServices()
	disable("svc0")
	checkServices()
	enable("svc0")
	disable("svc1")
	checkServices()
	disable("svc3")
	checkServices()
}
