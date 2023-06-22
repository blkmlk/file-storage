package migrations

import (
	"fmt"
	"github.com/blkmlk/file-storage/env"
	"github.com/golang-migrate/migrate/v4"
	"path/filepath"
	"runtime"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewLocal() (*migrate.Migrate, error) {
	uri := env.GetOptional(env.DatabaseURL, "postgres://root:root@127.0.0.1:25432/file-storage-test?sslmode=disable")

	_, f, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(f)

	return migrate.New(fmt.Sprintf("file://%s", basePath), uri)
}
