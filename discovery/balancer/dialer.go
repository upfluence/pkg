package balancer

import (
	"context"
	"net"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/syncutil"
)

type Dialer struct {
	Builder Builder
	Dialer  *net.Dialer
	Options GetOptions

	mu  sync.Mutex
	lds map[string]*localDialer
}

func (d *Dialer) dialer() *net.Dialer {
	if d.Dialer == nil {
		d.Dialer = &net.Dialer{}
	}

	return d.Dialer
}

func (d *Dialer) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d.mu.Lock()

	if d.lds == nil {
		d.lds = make(map[string]*localDialer)
	}

	ld, ok := d.lds[addr]

	if !ok {
		ld = &localDialer{d: d, b: d.Builder.Build(addr)}
		d.lds[addr] = ld
	}
	d.mu.Unlock()

	return ld.dial(ctx, network)
}

func (d *Dialer) Close() error {
	var errs []error

	d.mu.Lock()

	for _, ld := range d.lds {
		if err := ld.close(); err != nil {
			errs = append(errs)
		}
	}

	d.lds = nil
	d.mu.Unlock()

	return errors.WrapErrors(errs)
}

type localDialer struct {
	d *Dialer
	b Balancer

	opened bool
	sf     syncutil.Singleflight[struct{}]
}

func (ld *localDialer) open(ctx context.Context) (struct{}, error) {
	if !ld.b.IsOpen() {
		if err := ld.b.Open(ctx); err != nil {
			return struct{}{}, err
		}
	}

	ld.opened = true

	return struct{}{}, nil
}

func (ld *localDialer) dial(ctx context.Context, network string) (net.Conn, error) {
	if !ld.opened {
		if _, _, err := ld.sf.Do(ctx, ld.open); err != nil {
			return nil, err
		}
	}

	p, err := ld.b.Get(ctx, ld.d.Options)

	if err != nil {
		return nil, err
	}

	return ld.d.dialer().DialContext(ctx, network, p.Addr())
}

func (ld *localDialer) close() error {
	return errors.Combine(ld.sf.Close(), ld.b.Close())
}
