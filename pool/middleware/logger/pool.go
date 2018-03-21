package logger

import (
	"context"
	"time"

	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/pool"
)

type PoolFactory struct {
	name string
	log  func(string, ...interface{})
	next pool.IntrospectablePoolFactory
}

func NewDebugPoolFactory(name string, next pool.IntrospectablePoolFactory) *PoolFactory {
	return &PoolFactory{
		name: name,
		log:  func(fmt string, vs ...interface{}) { log.Debugf(fmt, vs...) },
		next: next,
	}
}

func (f *PoolFactory) GetPool(factory pool.Factory) pool.Pool {
	return &Pool{
		name: f.name,
		log:  f.log,
		next: f.next.GetIntrospectablePool(
			func(ctx context.Context) (interface{}, error) {
				var t0 = time.Now()

				defer func() {
					f.log("%s: Factory called: %v", f.name, time.Since(t0))
				}()

				return factory(ctx)
			},
		),
	}
}

type Pool struct {
	name string
	log  func(string, ...interface{})
	next pool.IntrospectablePool
}

func (p *Pool) logAction(method string, t0 time.Time) {
	var in, out = p.next.GetStats()

	p.log(
		"%s: %s: [ checked in: %d, checked out: %d ]: %v",
		p.name,
		method,
		in,
		out,
		time.Since(t0),
	)
}

func (p *Pool) Get(ctx context.Context) (interface{}, error) {
	var t0 = time.Now()

	defer p.logAction("Get", t0)

	return p.next.Get(ctx)
}

func (p *Pool) Put(e interface{}) error {
	var t0 = time.Now()

	defer p.logAction("Put", t0)

	return p.next.Put(e)
}

func (p *Pool) Discard(e interface{}) error {
	var t0 = time.Now()

	defer p.logAction("Discard", t0)

	return p.next.Discard(e)
}
