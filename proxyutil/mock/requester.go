package mock

import (
	"github.com/upfluence/proxy/http"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Requester struct {
	Response *http.Response
	Err      error

	Calls []*http.Request
}

func (r *Requester) Perform(_ thrift.Context, req *http.Request) (*http.Response, error) {
	r.Calls = append(r.Calls, req)

	return r.Response, r.Err
}
