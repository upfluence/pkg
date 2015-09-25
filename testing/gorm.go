package testing

import (
	"io/ioutil"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func BuildDabase(path string) (*gorm.DB, error) {
	blob, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	db, err := gorm.Open("sqlite3", ":memory:")

	if err != nil {
		return nil, err
	}

	if _, err := db.DB().Exec(string(blob)); err != nil {
		return nil, err
	}

	return &db, err
}
