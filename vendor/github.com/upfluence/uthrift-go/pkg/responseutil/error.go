package responseutil

import (
	"github.com/pkg/errors"
	"github.com/upfluence/thrift/lib/go/thrift"
)

func FetchError(res thrift.TResponse, err error) error {
	if err != nil {
		return errors.Cause(err)
	}

	return res.GetError()
}
