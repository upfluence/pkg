package amqp_thrift

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/streadway/amqp"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"
)

const (
	exchangeType        = "direct"
	DefaultAMQPURI      = "amqp://guest:guest@localhost:5672/%2F"
	DefaultQueueName    = "rpc-server-queue"
	DefaultRoutingKey   = "thrift"
	DefaultExchangeName = "rpc-server-exchange"
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
		log.Println("error processing request:", err)
	}
}
func (s *TAMQPServer) AcceptLoop() error {
	for {
		select {
		case delivery, ok := <-s.deliveries:
			if !ok {
				return errors.New("Channel closed")
			}

			if len(delivery.Body) == 0 {
				delivery.Ack(false)

				s.reportError(errors.New("Message empty"))
			} else {
				client, _ := NewTAMQPDelivery(delivery, s.channel)

				go func() {
					if err := s.processRequests(client); err != nil {
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

func (p *TAMQPServer) processRequests(client thrift.TTransport) error {
	tFactory := thrift.NewTTransportFactory()
	processor := p.processorFactory.GetProcessor(client)
	inputTransport := tFactory.GetTransport(client)
	outputTransport := tFactory.GetTransport(client)
	inputProtocol := p.inputProtocolFactory.GetProtocol(inputTransport)

	buf := bytes.NewBuffer(inputTransport.(*TAMQPDelivery).ReadBuffer.Bytes())
	inputDupProtocol := p.inputProtocolFactory.GetProtocol(
		&TAMQPDelivery{ReadBuffer: buf},
	)

	outputProtocol := p.outputProtocolFactory.GetProtocol(outputTransport)

	defer func() {
		if e := recover(); e != nil {
			p.reportError(
				errors.New(fmt.Sprintf("panic in processor: %s: %s", e, debug.Stack())),
			)
			client.(*TAMQPDelivery).Delivery.Ack(false)
		}
	}()

	resultChan := make(chan error)

	go func() {
		_, err := processor.Process(inputProtocol, outputProtocol)

		if _, t, _, _ := inputDupProtocol.ReadMessageBegin(); t == thrift.ONEWAY {
			client.(*TAMQPDelivery).Delivery.Ack(false)
		}

		if err != nil {
			if errThrift, ok := err.(thrift.TTransportException); ok && errThrift.TypeId() != thrift.END_OF_FILE {
				log.Println("error processing request:", err)
				resultChan <- err
				return
			}
		}
		resultChan <- nil
	}()

	if p.options.Timeout != 0 {
		go func() {
			time.Sleep(p.options.Timeout)

			resultChan <- errors.New("Timeout reached")
		}()
	}

	return <-resultChan
}
