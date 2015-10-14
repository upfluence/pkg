package thrift

import (
	"github.com/upfluence/goutils/error_logger/opbeat"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Server struct {
	processor thrift.TProcessor
	transport thrift.TServerTransport
}

func NewServer(processor thrift.TProcessor, transport thrift.TServerTransport) *Server {
	return &Server{
		processor: processor,
		transport: transport,
	}
}

func (s *Server) Start() error {
	server := thrift.NewTSimpleServer4(
		s.processor,
		s.transport,
		thrift.NewTTransportFactory(),
		thrift.NewTBinaryProtocolFactoryDefault(),
	)

	opbeatLogger := opbeat.NewErrorLogger()
	errLog := func(err error) {
		opbeatLogger.Capture(err, nil)
	}

	server.SetErrorLogger(errLog)

	err := server.Serve()
	opbeatLogger.Close()
	return err
}
