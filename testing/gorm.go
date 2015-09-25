package testing

import (
	"io/ioutil"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func BuildDatabase(path string) (*gorm.DB, error) {
	blob, err := ioutil.ReadFile(path)

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

	return &db, err
}
