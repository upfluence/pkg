package metricutil

import (
	"github.com/upfluence/errors"
	"github.com/upfluence/errors/stats"
)

func WrapStatus(err error, status string) error {
	return errors.WithStatus(err, status)
}

func ResultStatus(err error) string {
	return stats.GetStatus(err)
}
