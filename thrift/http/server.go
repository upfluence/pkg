package http

import (
	"github.com/upfluence/thrift-http-go/http_thrift"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Server struct {
	processor  *thrift.TProcessor
	listenAddr string
}

func NewServer(processor *thrift.TProcessor, listenAddr string) *Server {
	return &Server{
		processor:  processor,
		listenAddr: listenAddr,
	}
}

func (s *Server) Start() error {
	httpServer, err := http_thrift.NewTHTTPServer(s.listenAddr)

	if err != nil {
		return err
	}

	server := thrift.NewTSimpleServer4(
		*s.processor,
		httpServer,
		thrift.NewTTransportFactory(),
		thrift.NewTBinaryProtocolFactoryDefault(),
	)

	return server.Serve()
}
