package consumer

import (
	"context"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/testutil"
)

func TestIntegartion(t *testing.T) {
	testutil.FetchEnvVariable(t, "RABBITMQ_URL")
	c := NewConsumer()
	ctx := context.Background()

	err := c.Open(ctx)
	assert.Nil(t, err)

	q, err := c.QueueName(ctx)
	assert.Nil(t, err)

	dc, ec, err := c.Consume()
	assert.Nil(t, err)

	p := c.(*consumer).opts.pool
	ch, err := p.Get(ctx)
	assert.Nil(t, err)

	err = ch.Publish("", q, false, false, amqp.Publishing{Body: []byte("foo")})
	assert.Nil(t, err)

	assert.Nil(t, p.Put(ch))

	select {
	case <-ec:
		t.Errorf("did not expect message on the error chan")
	default:
	}

	d := <-dc
	assert.Equal(t, []byte("foo"), d.Body)

	assert.Nil(t, c.Close())

	_, ok := <-dc
	assert.False(t, ok)

	_, ok = <-ec
	assert.False(t, ok)

	_, _, err = c.Consume()
	assert.Equal(t, ErrCancelled, err)

	assert.Nil(t, c.Close())
}
