package stringslice

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type StringSlice struct {
	Strings []string
	Valid   bool
}

func (n StringSlice) Value() (driver.Value, error) {
	if len(n.Strings) == 0 {
		return nil, nil
	}

	return strings.Join(n.Strings, ","), nil
}

func (n *StringSlice) Scan(src interface{}) error {
	if src == nil {
		n.Valid = false
		return nil
	}

	if err := n.convertValue(src); err != nil {
		return err
	}

	n.Valid = true
	return nil
}

func (n *StringSlice) convertValue(src interface{}) error {
	switch vv := src.(type) {
	case string:
		n.Strings = strings.Split(vv, ",")
	case []byte:
		n.Strings = strings.Split(string(vv), ",")
	default:
		return fmt.Errorf("invalid type parsed: %T", src)
	}

	return nil
}
