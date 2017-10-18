package postgres

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
	"github.com/mattes/migrate/source"

	"github.com/upfluence/pkg/cfg"
)

var postgresURL = cfg.FetchString(
	"POSTGRES_URL",
	"postgres://localhost:5432/test_database?sslmode=disable",
)

func parseURL(t testing.TB) (string, string) {
	var (
		u, err = url.Parse(postgresURL)
		db     = "test_database"
	)

	if err != nil {
		t.Errorf("Postgres URL not valid: %v", err)
	}

	if len(u.Path) > 1 {
		db = u.Path[1:]
	}

	return postgresURL, db
}

func BuildDatabase(t testing.TB, driver source.Driver) *gorm.DB {
	var (
		url, database = parseURL(t)
		plainDB, err  = sql.Open("postgres", postgresURL)
	)

	if err != nil {
		t.Errorf("cant open the DB: %v", err)
	}

	plainDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
	plainDB.Exec(fmt.Sprintf("CREATE DATABASE %s", database))

	plainDB.Close()

	db, err := gorm.Open("postgres", postgresURL)

	if err != nil {
		t.Errorf("cant open the DB: %v", err)
	}

	if driver != nil {
		m, err := migrate.NewWithSourceInstance(
			"testing_source",
			driver,
			url,
		)

		if err != nil {
			t.Errorf("cant open migrate: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			t.Errorf("cant run migration: %v", err)
		}
	}

	return db
}
