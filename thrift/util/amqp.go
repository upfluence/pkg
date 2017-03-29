package util

import (
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"
)

var defaultProtocolFactory = thrift.NewTBinaryProtocolFactoryDefault()

func BuildAMQPClient(amqpURL, exchangeName, routingKey string) (thrift.TTransport, thrift.TProtocolFactory, error) {
	trans, err := amqp_thrift.NewTAMQPClient(
		amqpURL,
		exchangeName,
		routingKey,
		0,
	)

	if err != nil {
		return nil, nil, err
	}

	if err := trans.Open(); err != nil {
		return nil, nil, err
	}

	return trans, defaultProtocolFactory, nil
}
