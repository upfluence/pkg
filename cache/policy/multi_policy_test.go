package policy

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopPolicy(t *testing.T) {
	p := CombinePolicies[string]()

	select {
	case <-p.C():
		t.Error("should not be able to pull from channel yet")
	default:
	}

	assert.Nil(t, p.Op("foo", Set))

	assert.Nil(t, p.Close())

	assert.Equal(t, ErrClosed, p.Op("foo", Set))

	_, ok := <-p.C()
	assert.False(t, ok)
}

type op struct {
	k  string
	op OpType
}

type mockPolicy struct {
	sync.Mutex

	closed bool
	ops    []op

	ch chan string

	oerr, cerr error
}

func (mp *mockPolicy) C() <-chan string {
	return mp.ch
}

func (mp *mockPolicy) Op(k string, ot OpType) error {
	mp.Lock()
	mp.ops = append(mp.ops, op{k: k, op: ot})
	mp.Unlock()

	return mp.oerr
}

func (mp *mockPolicy) Close() error {
	mp.Lock()
	defer mp.Unlock()

	if mp.closed {
		return nil
	}

	mp.closed = true
	close(mp.ch)

	return mp.cerr
}

func TestMultiPolicy(t *testing.T) {
	var (
		m1 = &mockPolicy{ch: make(chan string)}
		m2 = &mockPolicy{ch: make(chan string)}
		m3 = &mockPolicy{ch: make(chan string)}

		wg sync.WaitGroup
	)

	p := CombinePolicies[string](m1, m2, m3)

	wg.Add(1)

	go func() {
		m2.ch <- "foo"
		wg.Done()
	}()

	k := <-p.C()

	wg.Wait()

	assert.Equal(t, "foo", k)
	assert.Equal(t, []op{{"foo", Evict}}, m1.ops)
	assert.Equal(t, []op{{"foo", Evict}}, m3.ops)

	assert.Nil(t, p.Close())

	// After Close(), Op is forwarded to child mocks which are also closed.
	// multiPolicy has no closed check of its own; children return nil from
	// their mock Op regardless of closed state, so this should not error.
	err := p.Op("bar", Set)
	assert.Nil(t, err)
	assert.Equal(t, []op{{"foo", Evict}, {"bar", Set}}, m1.ops)
	assert.Equal(t, []op{{"bar", Set}}, m2.ops)
	assert.Equal(t, []op{{"foo", Evict}, {"bar", Set}}, m3.ops)
}
