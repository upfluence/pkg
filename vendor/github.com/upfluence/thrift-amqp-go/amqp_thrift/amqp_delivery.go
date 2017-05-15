package amqp_thrift

import (
	"bytes"

	"github.com/streadway/amqp"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type TAMQPDelivery struct {
	Delivery    amqp.Delivery
	Channel     *amqp.Channel
	ReadBuffer  *bytes.Buffer
	writeBuffer *bytes.Buffer
}

func NewTAMQPDelivery(delivery amqp.Delivery, channel *amqp.Channel) (thrift.TTransport, error) {
	rBuffer := bytes.NewBuffer(delivery.Body)
	wBuffer := bytes.NewBuffer(make([]byte, 0, 1024))

	return &TAMQPDelivery{delivery, channel, rBuffer, wBuffer}, nil
}

func (d *TAMQPDelivery) Open() error {
	return nil
}

func (d *TAMQPDelivery) IsOpen() bool {
	return d.Channel != nil
}

func (d *TAMQPDelivery) Close() error {
	return nil
}

func (d *TAMQPDelivery) Read(buf []byte) (int, error) {
	return d.ReadBuffer.Read(buf)
}

func (d *TAMQPDelivery) Write(buf []byte) (int, error) {
	return d.writeBuffer.Write(buf)
}

func (d *TAMQPDelivery) Flush() error {
	if d.Delivery.ReplyTo != "" {
		if err := d.Channel.Publish(
			"",
			d.Delivery.ReplyTo,
			false,
			false,
			amqp.Publishing{
				CorrelationId: d.Delivery.CorrelationId,
				Body:          d.writeBuffer.Bytes(),
			},
		); err != nil {
			return err
		}
	}

	return nil
}
