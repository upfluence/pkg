package iopool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type entity struct {
	opened  int32
	openErr error
	isOpen  bool

	closed   int32
	closeErr error
}

func (e *entity) Close() error {
	atomic.AddInt32(&e.closed, 1)
	return e.closeErr
}

func (e *entity) Open(context.Context) error {
	atomic.AddInt32(&e.opened, 1)
	return e.openErr
}

func (e *entity) IsOpen() bool {
	return e.isOpen
}

func TestGarbageIdleConnections(t *testing.T) {
	e := &entity{isOpen: true}
	p := NewPool[*entity](
		func(context.Context) (*entity, error) { return e, nil },
		WithIdleTimeout(100*time.Millisecond),
	)

	v, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, v, e)

	assert.Nil(t, p.Put(v))

	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&e.closed))

	assert.Equal(t, 9, len(p.createc))
	assert.Equal(t, 1, len(p.poolc))

	v, err = p.Get(context.Background())
	assert.Nil(t, err)

	assert.Equal(t, 9, len(p.createc))
	assert.Equal(t, 0, len(p.poolc))

	assert.Nil(t, p.Discard(v))

	assert.Equal(t, 10, len(p.createc))
	assert.Equal(t, 0, len(p.poolc))

	assert.Nil(t, p.Close())
}

func TestReuseConnections(t *testing.T) {
	e := &entity{isOpen: true}
	p := NewPool[*entity](
		func(context.Context) (*entity, error) { return e, nil },
		WithIdleTimeout(100*time.Millisecond),
	)

	v1, err := p.Get(context.Background())
	assert.Nil(t, err)

	assert.Nil(t, p.Put(v1))

	v2, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, v1, v2)

	assert.Nil(t, p.Put(v2))

	assert.Nil(t, p.Close())
}

func TestCloseSync(t *testing.T) {
	e := &entity{isOpen: true}
	p := NewPool[*entity](
		func(context.Context) (*entity, error) { return e, nil },
		WithIdleTimeout(100*time.Millisecond),
	)

	assert.True(t, p.IsOpen())

	p.Get(context.Background())

	go func() {
		time.Sleep(time.Millisecond)
		assert.Nil(t, p.Put(e))
	}()

	assert.Nil(t, p.Close())

	assert.False(t, p.IsOpen())
	assert.Equal(t, int32(1), atomic.LoadInt32(&e.closed))

	eg, err := p.Get(context.Background())
	assert.Nil(t, eg)
	assert.Equal(t, ErrClosed, err)
}

func TestLimitedIdleSize(t *testing.T) {
	i := 0
	es := []*entity{
		&entity{isOpen: true},
		&entity{isOpen: true},
		&entity{isOpen: true},
	}

	p := NewPool[*entity](
		func(context.Context) (*entity, error) {
			i++
			return es[i-1], nil
		},
		WithMaxIdle(1),
	)

	e1, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, es[0], e1)

	e2, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, es[1], e2)

	assert.Nil(t, p.Put(e1))
	assert.Nil(t, p.Put(e2))

	assert.Equal(t, int32(0), atomic.LoadInt32(&es[0].closed))
	assert.Equal(t, int32(1), atomic.LoadInt32(&es[1].closed))

	p.Close()
}

func TestConcurrentPut(t *testing.T) {
	i := 0
	es := []*entity{
		&entity{isOpen: true},
	}

	p := NewPool[*entity](
		func(context.Context) (*entity, error) {
			i++
			return es[i-1], nil
		},
		WithSize(1),
	)

	wg := sync.WaitGroup{}
	wg.Add(2)
	e1, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, es[0], e1)

	go func() {
		e2, err := p.Get(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, es[0], e2)
		wg.Done()
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		assert.Nil(t, p.Put(e1))
		wg.Done()
	}()

	wg.Wait()

	p.Put(es[0])
	p.Close()
}

func TestConcurrentDiscard(t *testing.T) {
	i := 0
	es := []*entity{
		&entity{isOpen: true},
		&entity{isOpen: true},
	}

	p := NewPool[*entity](
		func(context.Context) (*entity, error) {
			i++
			return es[i-1], nil
		},
		WithSize(1),
	)

	wg := sync.WaitGroup{}
	wg.Add(2)
	e1, err := p.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, es[0], e1)

	go func() {
		e2, err := p.Get(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, es[1], e2)
		wg.Done()
	}()

	go func() {
		assert.Nil(t, p.Discard(e1))
		wg.Done()
	}()

	wg.Wait()
	p.Discard(es[1])
	p.Close()
}
