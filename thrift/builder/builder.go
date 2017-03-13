package builder

import (
	"errors"
	"os"
	"time"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift/transport/http"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift/transport/rabbitmq"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift_service"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"
)

const (
	defaultOpenTimeout = 30 * time.Second
	defaultRabbitMQURL = "amqp://localhost:5672/%2f"
)

var (
	errNoTransport        = errors.New("No transport")
	errHTTPTransportNoURL = errors.New("No URL provided")
	protocolFactory       = thrift.NewTBinaryProtocolFactoryDefault()
)

func Build(service *thrift_service.Service) (thrift.TTransport, thrift.TProtocolFactory, error) {
	var (
		transport thrift.TTransport
		err       error
	)

	if trans := service.Transport; trans != nil {
		if t := trans.HttpTransport; t != nil {
			transport, err = buildHTTPTransport(t)
		} else if t := trans.RabbitmqTransport; t != nil {
			transport, err = buildRabbitMQTransport(t)
		}
	}

	if transport == nil && err == nil {
		err = errNoTransport
	}

	return transport, protocolFactory, err
}

func buildHTTPTransport(transport *http.Transport) (thrift.TTransport, error) {
	if url := transport.GetUrl(); url == "" {
		return nil, errHTTPTransportNoURL
	}

	return thrift.NewTHttpPostClient(transport.GetUrl())
}

func buildRabbitMQTransport(transport *rabbitmq.Transport) (thrift.TTransport, error) {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	if rabbitMQURL == "" {
		rabbitMQURL = defaultRabbitMQURL
	}

	return amqp_thrift.NewTAMQPClient(
		rabbitMQURL,
		transport.GetExchangeName(),
		transport.GetRoutingKey(),
		defaultOpenTimeout,
	)
}
