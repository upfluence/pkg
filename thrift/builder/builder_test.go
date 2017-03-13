package builder

import (
	"os"
	"reflect"
	"testing"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift/transport"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift/transport/http"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift/transport/rabbitmq"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/service/thrift_service"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"
)

func TestBuild(t *testing.T) {
	os.Setenv("RABBITMQ_URL", "amqp://localhost/%2f")

	httpClient, _ := thrift.NewTHttpPostClient("foo.com")
	amqpClient, _ := amqp_thrift.NewTAMQPClient(
		"amqp://localhost/%2f",
		"foo",
		"bar",
		defaultOpenTimeout,
	)

	for _, tCase := range []struct {
		in    *transport.Transport
		trans thrift.TTransport
		err   error
	}{
		// No transport
		{nil, nil, errNoTransport},
		{&transport.Transport{}, nil, errNoTransport},

		// HTTP transport
		{
			&transport.Transport{
				HttpTransport: &http.Transport{
					Method: http.Method_POST,
					Url:    "foo.com",
				},
			},
			httpClient,
			nil,
		},

		// Empty HTTP transport,
		{
			&transport.Transport{
				HttpTransport: &http.Transport{},
			},
			nil,
			errHTTPTransportNoURL,
		},

		// RabbitMQ transport
		{
			&transport.Transport{
				RabbitmqTransport: &rabbitmq.Transport{
					ExchangeName: "foo",
					RoutingKey:   "bar",
				},
			},
			amqpClient,
			nil,
		},

		// Both transport set
		{
			&transport.Transport{
				HttpTransport: &http.Transport{
					Method: http.Method_POST,
					Url:    "foo.com",
				},
				RabbitmqTransport: &rabbitmq.Transport{
					ExchangeName: "foo",
					RoutingKey:   "bar",
				},
			},
			httpClient,
			nil,
		},
	} {
		trans, prot, err := Build(&thrift_service.Service{Transport: tCase.in})

		if prot != protocolFactory {
			t.Errorf("Wrong protocol factory: %v", prot)
		}

		if err != tCase.err {
			t.Errorf("Wrong error: %v", err)
		}

		if r, ok := tCase.trans.(*amqp_thrift.TAMQPClient); ok {
			aTrans := trans.(*amqp_thrift.TAMQPClient)

			if aTrans.ExchangeName != r.ExchangeName ||
				aTrans.RoutingKey != r.RoutingKey {
				t.Errorf("Wrong transport: %+v", trans)
			}
		} else {
			if !reflect.DeepEqual(trans, tCase.trans) {
				t.Errorf("Wrong transport: %+v", trans)
			}
		}
	}
}
