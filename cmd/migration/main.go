package main

import (
	"errors"
	"log"

	"github.com/blkmlk/file-storage/env"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	uri, err := env.Get(env.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("starting migrations")

	m, err := migrate.New(
		"file://migrations",
		uri)
	if err != nil {
		log.Fatalf("error creating migrations migrations: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("error running migrations migrations: %v", err)
	}

	log.Println("db migrated successfully")
}
