package sqlite3

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/source"
	_ "github.com/mattn/go-sqlite3"
)

func BuildDatabase(t testing.TB, driver source.Driver) *gorm.DB {
	f, _ := ioutil.TempFile("/tmp", "sqlite")
	db, err := gorm.Open("sqlite3", f.Name())

	if err != nil {
		t.Errorf("cant open migrate: %v", err)
	}

	if driver != nil {
		m, err := migrate.NewWithSourceInstance(
			"testing_source",
			driver,
			fmt.Sprintf("sqlite3://%s", f.Name()),
		)

		if err != nil {
			t.Errorf("cant open migrate: %v", err)
		}

		if err := m.Up(); err != nil {
			t.Errorf("cant run migration: %v", err)
		}
	}

	return db
}
