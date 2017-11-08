package log

import (
	"errors"
	"fmt"

	"github.com/op/go-logging"
	"github.com/upfluence/pkg/error_logger"
)

type errorLoggerBackend struct {
	client error_logger.ErrorLogger
}

func (b *errorLoggerBackend) Log(_ logging.Level, d int, r *logging.Record) error {
	var (
		err  error
		opts = make(map[string]interface{})
	)

	for i, arg := range r.Args {
		switch err2 := arg.(type) {
		case error:
			if err == nil {
				err = err2
			}

			opts[fmt.Sprintf("arg-%d", i)] = arg
		default:
			opts[fmt.Sprintf("arg-%d", i)] = arg
		}
	}

	if err == nil {
		err = errors.New(r.Formatted(d + 1))
	}

	return b.client.Capture(err, opts)
}
