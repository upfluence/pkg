package thrift

import (
	"github.com/upfluence/goutils/error_logger/opbeat"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	DefaultTransportFactory thrift.TTransportFactory = thrift.NewTTransportFactory()
	DefaultProtocolFactory  thrift.TProtocolFactory  = thrift.NewTBinaryProtocolFactoryDefault()
)

type Server struct {
	server thrift.TServer
}

func NewServerFromTServer(server thrift.TServer) *Server {
	return &Server{server}
}

func NewServer(processor thrift.TProcessor, transport thrift.TServerTransport) *Server {
	return &Server{
		thrift.NewTSimpleServer4(
			processor,
			transport,
			DefaultTransportFactory,
			DefaultProtocolFactory,
		),
	}
}

func (s *Server) Start() error {
	opbeatLogger := opbeat.NewErrorLogger()
	errLog := func(err error) {
		opbeatLogger.Capture(err, nil)
	}

	s.server.SetErrorLogger(errLog)

	err := s.server.Serve()
	opbeatLogger.Close()
	return err
}
