package rabbitmq

import (
	"github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Server struct {
	Processor    thrift.TProcessor
	AmqpURL      string
	RoutingKey   string
	ExchangeName string
	QueueName    string
}

func (s *Server) Serve() error {
	amqpServer, err := amqp_thrift.NewTServerAMQP(
		s.AmqpURL,
		s.ExchangeName,
		s.RoutingKey,
		s.QueueName,
	)

	if err != nil {
		return err
	}

	server := thrift.NewTSimpleServer4(
		s.Processor,
		amqpServer,
		thrift.NewTTransportFactory(),
		thrift.NewTBinaryProtocolFactoryDefault(),
	)

	return server.Serve()
}
