package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/log"
)

type Puller struct {
	resolver Resolver

	openOnce sync.Once

	closed bool
	closeL sync.RWMutex

	closeOnce sync.Once
	closeChan chan struct{}

	updateFn func(Update)
}

func NewPuller(r Resolver, fn func(Update)) (*Puller, func()) {
	var p = &Puller{
		resolver:  r,
		updateFn:  fn,
		closeChan: make(chan struct{}),
	}

	return p, func() { p.closeOnce.Do(func() { close(p.closeChan) }) }
}

func (p *Puller) IsOpen() bool {
	p.closeL.RLock()
	defer p.closeL.RUnlock()

	return !p.closed
}

func (p *Puller) String() string {
	return fmt.Sprintf("%v", p.resolver)
}

func (p *Puller) Open(ctx context.Context) error {
	var err error

	p.openOnce.Do(func() {
		if errO := p.resolver.Open(ctx); errO != nil {
			err = errO
		}

		go p.pull()
	})

	return err
}

func (p *Puller) close() {
	p.resolver.Close()

	p.closeL.Lock()
	defer p.closeL.Unlock()

	p.closed = true
}

func (p *Puller) pull() {
	for {
		var (
			channelOpen = true
			ch, err     = p.resolver.Resolve(context.Background())
		)

		if err != nil {
			log.Errorf("resolver: %+v", err)

			p.close()
			return
		}

		for channelOpen {
			select {
			case <-p.closeChan:
				p.close()

				return
			case update, ok := <-ch:
				if !ok {
					channelOpen = false
				} else {
					p.updateFn(update)
				}
			}
		}
	}
}
