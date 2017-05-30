package amqp_thrift

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/streadway/amqp"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	errOpenTimeout = errors.New("Open timeout errror")
	errOneWay      = errors.New("Can't read from it, one way mode only")
)

type TAMQPClient struct {
	URI                   string
	Connection            *amqp.Connection
	Channel               *amqp.Channel
	QueueName             string
	ExchangeName          string
	RoutingKey            string
	consumerTag           string
	requestBuffer         *bytes.Buffer
	responseReader        io.Reader
	stopConnectionOnClose bool
	deliveries            <-chan amqp.Delivery
	exitChan              chan bool
	openTimeout           time.Duration
	isOneway              bool
}

func NewTAMQPClientFromConnAndQueue(
	conn *amqp.Connection,
	channel *amqp.Channel,
	exchangeName,
	routingKey string,
	consumerTag string,
	queueName string,
	openTimeout time.Duration,
	isOneway bool,
) (thrift.TTransport, error) {
	buf := make([]byte, 0, 1024)
	cTag := fmt.Sprintf("%s-%d", consumerTag, rand.Uint32())

	return &TAMQPClient{
		Connection:    conn,
		Channel:       channel,
		ExchangeName:  exchangeName,
		RoutingKey:    routingKey,
		requestBuffer: bytes.NewBuffer(buf),
		consumerTag:   cTag,
		exitChan:      make(chan bool, 1),
		QueueName:     queueName,
		openTimeout:   openTimeout,
		isOneway:      isOneway,
	}, nil
}

func NewTAMQPClientFromConn(
	conn *amqp.Connection,
	channel *amqp.Channel,
	exchangeName,
	routingKey string,
	consumerTag string,
	openTimeout time.Duration,
	isOneway bool,
) (thrift.TTransport, error) {
	return NewTAMQPClientFromConnAndQueue(
		conn,
		channel,
		exchangeName,
		routingKey,
		consumerTag,
		"",
		openTimeout,
		isOneway,
	)
}

func NewTAMQPClient(
	amqpURI,
	exchangeName,
	routingKey string,
	openTimeout time.Duration,
) (thrift.TTransport, error) {
	buf := make([]byte, 0, 1024)

	return &TAMQPClient{
		URI:                   amqpURI,
		requestBuffer:         bytes.NewBuffer(buf),
		ExchangeName:          exchangeName,
		RoutingKey:            routingKey,
		stopConnectionOnClose: true,
		exitChan:              make(chan bool, 1),
		openTimeout:           openTimeout,
	}, nil
}

func (c *TAMQPClient) Open() error {
	var err error
	errChan := make(chan error)

	go func() { errChan <- c.open() }()

	if c.openTimeout > 0 {
		select {
		case err = <-errChan:
		case <-time.After(c.openTimeout):
			err = errOpenTimeout
		}
	} else {
		err = <-errChan
	}

	return err
}

func (c *TAMQPClient) open() error {
	var err error

	if c.Connection == nil {
		if c.Connection, err = amqp.Dial(c.URI); err != nil {
			return err
		}
	}

	if c.Channel == nil {
		if c.Channel, err = c.Connection.Channel(); err != nil {
			return err
		}
	}

	if c.isOneway {
		return nil
	}

	if c.QueueName == "" {
		var queue amqp.Queue

		queue, err = c.Channel.QueueDeclare(
			"",    // name of the queue
			true,  // durable
			true,  // delete when usused
			true,  // exclusive
			false, // noWait
			nil,   // arguments
		)

		if err != nil {
			return err
		}

		c.QueueName = queue.Name
	}

	r, err := NewAMQPQueueReader(
		c.Channel,
		c.QueueName,
		c.consumerTag,
		c.exitChan,
	)

	if err != nil {
		return err
	}

	c.responseReader = r

	if err = r.Open(); err != nil {
		return err
	}

	go r.Consume()

	return nil
}

func (c *TAMQPClient) IsOpen() bool {
	return c.Connection != nil && c.Channel != nil
}

func (c *TAMQPClient) Close() error {
	if c.isOneway {
		return nil
	}

	if c.consumerTag != "" {
		c.Channel.Cancel(c.consumerTag, true)
	}

	if c.stopConnectionOnClose {
		if c.Connection == nil {
			return errors.New("The connection is not opened")
		}

		defer func() {
			c.Channel = nil
			c.Connection = nil
		}()

		c.Connection.Close()
	}

	c.exitChan <- true
	return nil
}

func (c *TAMQPClient) Read(buf []byte) (int, error) {
	if c.isOneway {
		return 0, errOneWay
	}

	return c.responseReader.Read(buf)
}

func (c *TAMQPClient) Write(buf []byte) (int, error) {
	return c.requestBuffer.Write(buf)
}

func (c *TAMQPClient) Flush() error {
	err := c.Channel.Publish(
		c.ExchangeName,
		c.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Body:          c.requestBuffer.Bytes(),
			ReplyTo:       c.QueueName,
			CorrelationId: generateUUID(),
		},
	)

	buf := make([]byte, 0, 1024)
	c.requestBuffer = bytes.NewBuffer(buf)
	return err
}

func generateUUID() string {
	f, _ := os.Open("/dev/urandom")
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
