package closer

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/syncutil"
)

type State uint8

const (
	Open State = iota
	Closing
	Closed
)

type ClosingPolicy uint8

const (
	NoWait ClosingPolicy = iota
	Wait
)

type MonitorOption func(*Monitor)

func WithClosingPolicy(cp ClosingPolicy) MonitorOption {
	return func(m *Monitor) { m.ClosingPolicy = cp }
}

type Monitor struct {
	ClosingPolicy ClosingPolicy

	ctx    context.Context
	cancel context.CancelFunc

	once sync.Once
	mu   sync.Mutex
	cond *syncutil.Cond

	s     State
	count int
}

func NewMonitor(opts ...MonitorOption) *Monitor {
	var m Monitor

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

func (m *Monitor) Context() context.Context {
	m.init()
	return m.ctx
}

func (m *Monitor) init() {
	m.once.Do(func() {
		m.cond = &syncutil.Cond{Locker: &m.mu}
		m.ctx, m.cancel = context.WithCancel(context.Background())
	})
}

func (m *Monitor) Run(fn func(context.Context)) {
	m.init()

	if !m.IsOpen() {
		return
	}

	m.mu.Lock()
	m.count++
	m.mu.Unlock()

	go func() {
		fn(m.ctx)

		m.mu.Lock()
		m.count--
		m.cond.Broadcast()
		m.mu.Unlock()
	}()
}

func (m *Monitor) State() State {
	m.mu.Lock()
	s := m.s
	m.mu.Unlock()

	return s
}

func (m *Monitor) IsOpen() bool {
	return m.State() == Open
}

func (m *Monitor) Shutdown(ctx context.Context) error {
	m.init()

	m.mu.Lock()
	m.s = Closing
	m.mu.Unlock()

	if err := m.cond.Wait(ctx, func() bool { return m.count == 0 }); err != nil {
		return err
	}

	m.mu.Lock()
	m.s = Closed
	m.mu.Unlock()

	return nil
}

func (m *Monitor) Close() error {
	m.init()
	m.cancel()

	m.mu.Lock()

	if m.ClosingPolicy == NoWait {
		m.s = Closed
	}

	m.mu.Unlock()

	return m.cond.Wait(
		context.Background(),
		func() bool { return m.count == 0 || m.s == Closed },
	)
}
