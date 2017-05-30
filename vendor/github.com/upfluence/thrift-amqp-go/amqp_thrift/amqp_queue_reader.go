package amqp_thrift

import (
	"bytes"
	"sync"

	"github.com/streadway/amqp"
)

type AMQPQueueReader struct {
	Channel     *amqp.Channel
	QueueName   string
	consumerTag string
	exitChan    chan bool
	newData     chan bool
	isClosing   chan bool
	closeChan   chan *amqp.Error
	buffer      *bytes.Buffer
	mutex       *sync.Mutex
	deliveries  <-chan amqp.Delivery
}

func NewAMQPQueueReader(channel *amqp.Channel, queueName string, consumerTag string, exitChan chan bool) (*AMQPQueueReader, error) {
	return &AMQPQueueReader{
		channel,
		queueName,
		consumerTag,
		exitChan,
		make(chan bool, 256),
		make(chan bool, 1),
		make(chan *amqp.Error),
		bytes.NewBuffer(make([]byte, 0, 1024)),
		&sync.Mutex{},
		nil,
	}, nil
}

func (r *AMQPQueueReader) Open() error {
	var err error

	r.Channel.NotifyClose(r.closeChan)

	r.deliveries, err = r.Channel.Consume(
		r.QueueName,   // name
		r.consumerTag, // consumerTag,
		true,          // noAck
		false,         // exclusive
		false,         //            noLocal
		false,         // noWait
		nil,           // arguments
	)

	return err
}

func (r *AMQPQueueReader) Consume() {
	for {
		select {
		case <-r.closeChan:
			r.newData <- true
			return
		case delivery, ok := <-r.deliveries:
			if !ok {
				r.newData <- true
				return
			}

			shouldNotify := false

			if r.buffer.Len() == 0 {
				shouldNotify = true
			}

			r.buffer.Write(delivery.Body)

			if shouldNotify {
				r.newData <- true
			}
		case <-r.exitChan:
			return
		}
	}

	return
}

func (r *AMQPQueueReader) Read(b []byte) (int, error) {
	if r.buffer.Len() == 0 {
		select {
		case <-r.newData:
		case <-r.isClosing:
		}
	}

	return r.buffer.Read(b)
}
