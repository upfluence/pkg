package resolver

import (
	"context"
	"fmt"

	"github.com/upfluence/pkg/log"
)

type Puller struct {
	resolver  Resolver
	closeChan <-chan interface{}

	updateFn func(Update)
}

func NewPuller(r Resolver, fn func(Update)) (*Puller, chan<- interface{}) {
	var ch = make(chan interface{})

	return &Puller{
		resolver:  r,
		updateFn:  fn,
		closeChan: ch,
	}, ch
}

func (p *Puller) String() string {
	return fmt.Sprintf("%v", p.resolver)
}

func (p *Puller) Open(ctx context.Context) error {
	if err := p.resolver.Open(ctx); err != nil {
		return err
	}

	go p.pull()

	return nil
}

func (p *Puller) pull() {
	for {
		var (
			channelOpen = true
			ch, err     = p.resolver.Resolve(context.Background())
		)

		if err != nil {
			log.Errorf("resolver: %+v", err)
		}

		for channelOpen {
			select {
			case <-p.closeChan:
				p.resolver.Close()
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
