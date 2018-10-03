package binary

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/balancer/random"
	"github.com/upfluence/pkg/discovery/resolver/static"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/uthrift/context/httputil"
	"github.com/upfluence/uthrift-go/uthrift/transport/base"
)

func NewSingleEndpointFactory(url string) *TransportFactory {
	return NewTransportFactory(
		random.NewBalancer(static.NewResolverFromStrings([]string{url})),
	)
}

func NewTransportFactory(b balancer.Balancer) *TransportFactory {
	return &TransportFactory{balancer: b}
}

type TransportFactory struct {
	balancer balancer.Balancer
}

func (f *TransportFactory) GetTransport(thrift.TTransport) thrift.TTransport {
	return NewTransport(f.balancer)
}

type Transport struct {
	*base.Transport

	balancer balancer.Balancer

	rBuf *bytes.Buffer

	cl *http.Client
}

func NewTransport(b balancer.Balancer) *Transport {
	return &Transport{
		Transport: base.NewTransport(),
		balancer:  b,
		cl:        http.DefaultClient,
		rBuf:      &bytes.Buffer{},
	}
}

func (t *Transport) Open() error {
	return t.balancer.Open(context.Background())
}

func (t *Transport) IsOpen() bool {
	return t.balancer.IsOpen()
}

func (t *Transport) Close() error {
	return t.balancer.Close()
}

func (t *Transport) Read(p []byte) (int, error) {
	n, err := t.rBuf.Read(p)

	return n, thrift.NewTTransportExceptionFromError(err)
}

func (t *Transport) Flush() error {
	defer t.ResetBuffers()
	t.rBuf.Reset()

	peer, err := t.balancer.Get(t.Ctx, balancer.BalancerGetOptions{})

	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	req, err := http.NewRequest(http.MethodPost, peer.Addr, t.Buf)

	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	req.Header.Set("Content-Type", "application/x-thrift")

	res, err := t.cl.Do(httputil.WithContext(req, t.Ctx))

	if err == nil {
		if res.StatusCode == 200 {
			defer res.Body.Close()

			io.Copy(t.rBuf, res.Body)
		} else {
			err = thrift.NewTTransportException(
				thrift.UNKNOWN_TRANSPORT_EXCEPTION,
				fmt.Sprintf("Status code: %d", res.StatusCode),
			)
		}
	}

	return thrift.NewTTransportExceptionFromError(err)
}
