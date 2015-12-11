package testing

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/jinzhu/gorm"
	_ "github.com/mattes/migrate/driver/sqlite3"
	"github.com/mattes/migrate/migrate"
	_ "github.com/mattn/go-sqlite3"
)

func BuildDatabase(schemaPath string) (*gorm.DB, error) {
	blob, err := ioutil.ReadFile(schemaPath)

	if err != nil {
		return nil, err
	}

	f, _ := ioutil.TempFile("/tmp", "sqlite")
	db, err := gorm.Open("sqlite3", f.Name())

	if err != nil {
		return nil, err
	}

	if _, err := db.DB().Exec(string(blob)); err != nil {
		return nil, err
	}

	_, file, _, _ := runtime.Caller(1)
	dbPath := fmt.Sprintf("sqlite3://%s", f.Name())
	migrationsPath := path.Join(path.Dir(file), "..", "..", "migrations")
	errs, ok := migrate.UpSync(dbPath, migrationsPath)
	if !ok {
		for migrationError := range errs {
			fmt.Println(migrationError)
		}
	}

	return &db, err
}
