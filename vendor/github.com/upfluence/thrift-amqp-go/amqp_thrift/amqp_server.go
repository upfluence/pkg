package amqp_thrift

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/streadway/amqp"
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/thrift/lib/go/thrift"
)

const (
	exchangeType        = "direct"
	DefaultAMQPURI      = "amqp://guest:guest@localhost:5672/%2F"
	DefaultQueueName    = "rpc-server-queue"
	DefaultRoutingKey   = "thrift"
	DefaultExchangeName = "rpc-server-exchange"
)

var (
	errTimeout       = errors.New("thrift/transport/amqp: Timeout reached")
	errEmptyMessage  = errors.New("thrift/transport/amqp: Message empty")
	errClosedChannel = errors.New("thrift/transport/amqp: Channel closed")
)

type TAMQPServer struct {
	quit chan struct{}

	processorFactory      thrift.TProcessorFactory
	inputProtocolFactory  thrift.TProtocolFactory
	outputProtocolFactory thrift.TProtocolFactory

	connection *amqp.Connection
	channel    *amqp.Channel
	deliveries <-chan amqp.Delivery
	options    ServerOptions

	errorLogger *func(error)
}

type ServerOptions struct {
	Prefetch     uint
	AmqpURI      string
	ExchangeName string
	RoutingKey   string
	QueueName    string
	ConsumerTag  string
	Timeout      time.Duration
}

func NewTAMQPServer(
	processor thrift.TProcessor,
	protocolFactory thrift.TProtocolFactory,
	opts ServerOptions,
) (*TAMQPServer, error) {
	return &TAMQPServer{
		processorFactory:      thrift.NewTProcessorFactory(processor),
		inputProtocolFactory:  protocolFactory,
		outputProtocolFactory: protocolFactory,
		options:               opts,
		quit:                  make(chan struct{}, 1),
	}, nil
}

func (s *TAMQPServer) Listen() error {
	var err error

	if s.connection == nil {
		uri := s.options.AmqpURI

		if uri == "" {
			uri = DefaultAMQPURI
		}

		if s.connection, err = amqp.Dial(uri); err != nil {
			return err
		}
	}

	if s.channel == nil {
		if s.channel, err = s.connection.Channel(); err != nil {
			return err
		}
	}

	if opts := s.options; opts.Prefetch != 0 {
		if err := s.channel.Qos(int(opts.Prefetch), 0, false); err != nil {
			return err
		}
	}

	exchangeName := s.options.ExchangeName
	if exchangeName == "" {
		exchangeName = DefaultExchangeName
	}

	queueName := s.options.QueueName
	if queueName == "" {
		queueName = DefaultQueueName
	}

	routingKey := s.options.RoutingKey
	if routingKey == "" {
		routingKey = DefaultRoutingKey
	}

	if err = s.channel.ExchangeDeclare(
		exchangeName, // name osf the exchange
		exchangeType, // type
		false,        // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return err
	}

	if _, err = s.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return err
	}

	if err = s.channel.QueueBind(
		queueName,    // name of the queue
		routingKey,   // bindingKey
		exchangeName, // sourceExchange
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return err
	}

	s.deliveries, err = s.channel.Consume(
		queueName,             // name
		s.options.ConsumerTag, // consumerTag,
		false, // noAck
		false, // exclusive
		false, //            noLocal
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *TAMQPServer) Stop() error {
	s.quit <- struct{}{}

	if s.connection == nil {
		return errors.New("The connection is not opened")
	}

	defer func() {
		s.channel = nil
		s.connection = nil
	}()

	s.connection.Close()

	return nil
}

func (p *TAMQPServer) ProcessorFactory() thrift.TProcessorFactory {
	return p.processorFactory
}

func (p *TAMQPServer) ServerTransport() thrift.TServerTransport {
	return nil
}

func (p *TAMQPServer) InputTransportFactory() thrift.TTransportFactory {
	return nil
}

func (p *TAMQPServer) OutputTransportFactory() thrift.TTransportFactory {
	return nil
}

func (p *TAMQPServer) InputProtocolFactory() thrift.TProtocolFactory {
	return p.inputProtocolFactory
}

func (p *TAMQPServer) OutputProtocolFactory() thrift.TProtocolFactory {
	return p.outputProtocolFactory
}

func (p *TAMQPServer) SetErrorLogger(fn func(error)) {
	p.errorLogger = &fn
}

func (s *TAMQPServer) reportError(err error) {
	if s.errorLogger != nil {
		(*s.errorLogger)(err)
	} else {
		log.Errorf("thrift/transport/amqp: error processing request:", err)
	}
}
func (s *TAMQPServer) AcceptLoop() error {
	for {
		select {
		case delivery, ok := <-s.deliveries:
			if !ok {
				return errClosedChannel
			}

			if len(delivery.Body) == 0 {
				delivery.Ack(false)
				s.reportError(errEmptyMessage)
			} else {
				client, _ := NewTAMQPDelivery(delivery, s.channel)

				go func() {
					if err := s.processRequest(client); err != nil {
						s.reportError(err)
					}
				}()
			}
		case <-s.quit:
			return nil
		}
	}
}

func (p *TAMQPServer) Serve() error {
	if err := p.Listen(); err != nil {
		return err
	}

	return p.AcceptLoop()
}

func (s *TAMQPServer) executeProcessor(client thrift.TTransport) error {
	var (
		tFactory  = thrift.NewTTransportFactory()
		processor = s.processorFactory.GetProcessor(client)

		inputTransport  = tFactory.GetTransport(client)
		outputTransport = tFactory.GetTransport(client)

		inputProtocol  = s.inputProtocolFactory.GetProtocol(inputTransport)
		outputProtocol = s.outputProtocolFactory.GetProtocol(outputTransport)
	)

	_, err := processor.Process(inputProtocol, outputProtocol)

	return err
}

func (s *TAMQPServer) processRequest(client thrift.TTransport) error {
	defer func() {
		if e := recover(); e != nil {
			s.reportError(
				fmt.Errorf(
					"thrift/transport/amqp: panic in processor: %s: %s",
					e,
					debug.Stack(),
				),
			)
			client.(*TAMQPDelivery).Delivery.Ack(false)
		}
	}()

	var resultChan = make(chan error)

	go func() { resultChan <- s.executeProcessor(client) }()

	if s.options.Timeout != 0 {
		time.AfterFunc(
			s.options.Timeout,
			func() { resultChan <- errTimeout },
		)
	}

	err := <-resultChan

	if errAck := client.(*TAMQPDelivery).Delivery.Ack(false); errAck != nil {
		log.Errorf("thrift/trasnport/amqp: ack: %s", err.Error())
	}

	if errThrift, ok := err.(thrift.TTransportException); ok && errThrift.TypeId() == thrift.END_OF_FILE {
		return nil
	}

	return err
}
