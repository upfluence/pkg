package timetest

import (
	"sync"
	"time"

	"github.com/upfluence/pkg/timeutil"
)

type Clock struct {
	mu sync.RWMutex

	now time.Time

	timers  []*timer
	tickers []*ticker
}

func (c *Clock) MoveTo(t time.Time) {
	c.mu.Lock()
	c.moveTo(t)
	c.mu.Unlock()
}

func (c *Clock) MoveBy(d time.Duration) {
	c.mu.Lock()
	c.moveTo(c.now.Add(d))
	c.mu.Unlock()
}

func (c *Clock) moveTo(n time.Time) {
	c.now = n

	for _, t := range c.tickers {
		t.moveTo(n)
	}

	for _, t := range c.timers {
		t.moveTo(n)
	}
}

func (c *Clock) newTimer(d time.Duration, fn func()) *timer {
	c.mu.Lock()
	t := timer{
		c:        make(chan time.Time, 1),
		now:      c.now,
		fn:       fn,
		deadline: c.now.Add(d),
	}

	c.timers = append(c.timers, &t)
	c.mu.Unlock()

	return &t
}

func (c *Clock) Timer(d time.Duration) timeutil.Timer {
	return c.newTimer(d, nil)
}

func (c *Clock) TimerFunc(d time.Duration, fn func()) timeutil.Timer {
	return c.newTimer(d, fn)
}

func (c *Clock) Ticker(d time.Duration) timeutil.Ticker {
	c.mu.Lock()
	t := ticker{
		c: make(chan time.Time, 1),
		d: d,
		t: c.now.Add(d),
	}

	c.tickers = append(c.tickers, &t)
	c.mu.Unlock()

	return &t
}

func (c *Clock) Now() time.Time {
	c.mu.RLock()
	t := c.now
	c.mu.RUnlock()

	return t
}

type timer struct {
	sync.Mutex
	stopped bool
	fired   bool

	fn func()
	c  chan time.Time

	now      time.Time
	deadline time.Time
}

func (t *timer) moveTo(n time.Time) {
	t.Lock()
	defer t.Unlock()

	t.now = n

	if t.stopped || t.fired {
		return
	}

	if !t.deadline.After(n) {
		t.fired = true
		if t.fn != nil {
			t.fn()
		}

		select {
		case t.c <- t.deadline:
		default:
		}
	}
}

func (t *timer) Reset(d time.Duration) bool {
	t.Lock()
	s, f := t.stopped, t.fired

	t.fired = false
	t.stopped = false
	t.deadline = t.now.Add(d)
	t.Unlock()

	return !s && !f
}

func (t *timer) C() <-chan time.Time {
	return t.c
}

func (t *timer) Stop() bool {
	t.Lock()

	s, f := t.stopped, t.fired

	t.stopped = true
	t.Unlock()

	return !s && !f
}

type ticker struct {
	c chan time.Time
	d time.Duration

	sync.Mutex
	s bool
	t time.Time
}

func (t *ticker) moveTo(n time.Time) {
	t.Lock()
	s := t.s
	t.Unlock()

	if s {
		return
	}

	for {
		if t.t.After(n) {
			return
		} else {
			select {
			case t.c <- t.t:
			default:
			}

			t.t = t.t.Add(t.d)
		}
	}
}

func (t *ticker) C() <-chan time.Time {
	return t.c
}

func (t *ticker) Stop() {
	t.Lock()

	if !t.s {
		t.s = true
	}

	t.Unlock()
}
