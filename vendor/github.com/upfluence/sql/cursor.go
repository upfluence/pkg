package sql

type ScanFunc func(Scanner) error

type Cursor interface {
	Scanner

	Close() error
	Err() error
	Next() bool
}

func ScrollCursor(c Cursor, fn ScanFunc) error {
	defer c.Close()

	for c.Next() {
		if err := fn(c); err != nil {
			return err
		}
	}

	return c.Err()
}
