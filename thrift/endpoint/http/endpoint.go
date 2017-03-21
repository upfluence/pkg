package http

import (
	"fmt"
	"net/http"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/base_service"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift-http-go/http_thrift"
	stdThrift "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"
	"github.com/upfluence/goutils/httputil"
	"github.com/upfluence/goutils/thrift"
	"github.com/upfluence/goutils/thrift/handler"
)

type Endpoint struct {
	servers []*thrift.Server
	port    int
	mux     *http.ServeMux
}

func NewEndpoint(baseHandler *handler.Base, port int) *Endpoint {
	mux := http.NewServeMux()
	trans, _ := http_thrift.NewTHTTPServerFromMux(mux, "/base")

	return &Endpoint{
		servers: []*thrift.Server{
			thrift.NewServer(
				base_service.NewBaseServiceProcessor(baseHandler),
				trans,
			),
		},
		port: port,
		mux:  mux,
	}
}

func (e *Endpoint) Mount(processor stdThrift.TProcessor, path string) {
	trans, _ := http_thrift.NewTHTTPServerFromMux(e.mux, path)

	e.servers = append(e.servers, thrift.NewServer(processor, trans))
}

func (e *Endpoint) Serve() error {
	e.mux.HandleFunc("/healthcheck", httputil.HealthcheckHandler)

	for _, s := range e.servers {
		go func(server *thrift.Server) { server.Start() }(s)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", e.port), e.mux)
}
