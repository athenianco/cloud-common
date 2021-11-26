package service

import (
	"context"
	"os"
)

func Register(ctx context.Context) (bool, error) {
	name := os.Getenv("SERVICE_NAME")
	dbURI := os.Getenv("SERVICE_DATABASE_URI")
	if name == "" || dbURI == "" {
		// service is not regulated
		return true, nil
	}
	db, err := OpenDatabaseFromEnv()
	if err != nil {
		return false, err
	}
	defer db.Close()
	return db.RegisterService(ctx, name)
}
