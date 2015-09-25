package testing

import (
	"io/ioutil"
	test "testing"
)

func TestBuildDatabaseNotExist(t *test.T) {
	if _, err := BuildDatabase("/foo/bar"); err == nil {
		t.Errorf("Wrong file")
	}
}

func TestBuildDatabaseNotValid(t *test.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")

	f.WriteString("foo;\nbar;")

	if _, err := BuildDatabase(f.Name()); err == nil {
		t.Errorf("Execute wrong command")
	}
}

func TestBuildDatabaseValid(t *test.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")

	f.WriteString(
		`
		CREATE TABLE t(x INTEGER PRIMARY KEY ASC, y, z);
		CREATE TABLE y(x INTEGER PRIMARY KEY ASC, y, z);
		`,
	)

	db, err := BuildDatabase(f.Name())

	if err != nil {
		t.Errorf("Cannot execute sql command")
	}

	r := -1
	db.DB().QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = \"table\";",
	).Scan(&r)

	if r != 2 {
		t.Errorf("Wrong number of table: %v", r)
	}
}
