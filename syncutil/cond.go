package syncutil

import (
	"context"
	"sync"
)

type Cond struct {
	Locker sync.Locker

	cond *sync.Cond
	once sync.Once
}

func (c *Cond) init() {
	c.once.Do(func() { c.cond = &sync.Cond{L: c.Locker} })
}

func (c *Cond) Signal() {
	c.init()
	c.cond.Signal()
}

func (c *Cond) Broadcast() {
	c.init()
	c.cond.Broadcast()
}

func (c *Cond) Wait(ctx context.Context, fn func() bool) error {
	var (
		cancelled bool

		done = make(chan struct{})
	)

	c.init()

	go func() {
		c.Locker.Lock()
		defer c.Locker.Unlock()

		for {
			if fn() {
				close(done)
				break
			}

			if cancelled {
				close(done)
				break
			}

			c.cond.Wait()
		}
	}()

	select {
	case <-ctx.Done():
		c.Locker.Lock()
		cancelled = true
		c.cond.Broadcast()
		c.Locker.Unlock()

		return ctx.Err()
	case <-done:
		return nil
	}
}
