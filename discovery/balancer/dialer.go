package balancer

import (
	"context"
	"net"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/syncutil"
)

type Dialer[T peer.Peer] struct {
	Builder Builder[T]
	Dialer  *net.Dialer
	Options GetOptions

	mu  sync.Mutex
	lds map[string]*localDialer[T]
}

func (d *Dialer[T]) dialer() *net.Dialer {
	if d.Dialer == nil {
		d.Dialer = &net.Dialer{}
	}

	return d.Dialer
}

func (d *Dialer[T]) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *Dialer[T]) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d.mu.Lock()

	if d.lds == nil {
		d.lds = make(map[string]*localDialer[T])
	}

	ld, ok := d.lds[addr]

	if !ok {
		ld = &localDialer[T]{d: d, b: d.Builder.Build(addr)}
		d.lds[addr] = ld
	}
	d.mu.Unlock()

	return ld.dial(ctx, network)
}

func (d *Dialer[T]) Close() error {
	var errs []error

	d.mu.Lock()

	for _, ld := range d.lds {
		if err := ld.close(); err != nil {
			errs = append(errs, err)
		}
	}

	d.lds = nil
	d.mu.Unlock()

	return errors.WrapErrors(errs)
}

type localDialer[T peer.Peer] struct {
	d *Dialer[T]
	b Balancer[T]

	opened bool
	sf     syncutil.Singleflight[struct{}]
}

func (ld *localDialer[T]) open(ctx context.Context) (struct{}, error) {
	if !ld.b.IsOpen() {
		if err := ld.b.Open(ctx); err != nil {
			return struct{}{}, err
		}
	}

	ld.opened = true

	return struct{}{}, nil
}

func (ld *localDialer[T]) dial(ctx context.Context, network string) (net.Conn, error) {
	if !ld.opened {
		if _, _, err := ld.sf.Do(ctx, ld.open); err != nil {
			return nil, err
		}
	}

	p, done, err := ld.b.Get(ctx, ld.d.Options)

	if err != nil {
		return nil, err
	}

	conn, err := ld.d.dialer().DialContext(ctx, network, p.Addr())

	if err != nil {
		done(err)

		return nil, err
	}

	return &doneCloserConn{Conn: conn, done: done}, nil
}

func (ld *localDialer[T]) close() error {
	return errors.Combine(ld.sf.Close(), ld.b.Close())
}

type doneCloserConn struct {
	net.Conn

	done func(error)
}

func (dcc *doneCloserConn) Close() error {
	err := dcc.Conn.Close()

	dcc.done(err)

	return err
}
