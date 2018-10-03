package local

import (
	"io/ioutil"
	"net/http"

	thttp "github.com/upfluence/proxy/http"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Requester struct {
	cl *http.Client
}

func NewRequester(cl *http.Client) *Requester {
	return &Requester{cl: cl}
}

func (r *Requester) Perform(ctx thrift.Context, treq *thttp.Request) (*thttp.Response, error) {
	// TODO: Share the parsing / serializing of the request with the proxy-server
	req, err := http.NewRequest("GET", treq.Url, nil)

	for k, v := range treq.Headers {
		req.Header[k] = []string{v}
	}

	if err != nil {
		return nil, err
	}

	res, err := r.cl.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return &thttp.Response{
		Status: int16(res.StatusCode),
		Body:   buf,
	}, nil
}
