package deps

import (
	"github.com/blkmlk/file-storage/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	dsn, err := env.Get(env.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return gorm.Open(postgres.Open(dsn))
}

func NewLocalDB() (*gorm.DB, error) {
	dsn := env.GetOptional(env.DatabaseURL, "postgres://root:root@127.0.0.1:25432/file-storage-test?sslmode=disable")

	return gorm.Open(postgres.Open(dsn))
}
