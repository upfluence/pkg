package closer

import (
	"context"
	"sync"
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
	return func(m *Monitor) { m.cp = cp }
}

type Monitor struct {
	s  State
	cp ClosingPolicy

	Ctx    context.Context
	cancel context.CancelFunc

	mu   sync.Mutex
	cond *sync.Cond

	count int
}

func NewMonitor(opts ...MonitorOption) *Monitor {
	var m Monitor

	for _, opt := range opts {
		opt(&m)
	}

	m.cond = sync.NewCond(&m.mu)
	m.Ctx, m.cancel = context.WithCancel(context.Background())

	return &m
}

func (m *Monitor) Run(fn func(context.Context)) {
	m.mu.Lock()
	m.count++
	m.mu.Unlock()

	go func() {
		fn(m.Ctx)

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
	m.mu.Lock()
	m.s = Closing
	m.mu.Unlock()

	m.cancel()
	done := make(chan struct{})
	cancelled := false

	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		for {
			if m.count == 0 || m.s == Closed {
				close(done)
				m.s = Closed
				break
			}

			if cancelled {
				close(done)
				break
			}

			m.cond.Wait()
		}
	}()

	select {
	case <-ctx.Done():
		m.mu.Lock()
		cancelled = true
		m.cond.Broadcast()
		m.mu.Unlock()

		return ctx.Err()
	case <-done:
		return nil
	}
}

func (m *Monitor) Close() error {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cp == NoWait {
		m.s = Closed
	}

	for {
		if m.count == 0 || m.s == Closed {
			return nil
		}

		m.cond.Wait()
	}
}
