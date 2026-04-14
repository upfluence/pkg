package size

import (
	"container/list"
	"sync"

	"github.com/upfluence/pkg/v2/cache/policy"
)

type Policy[K comparable] struct {
	mu sync.Mutex

	size int

	l  *list.List
	ks map[K]*list.Element

	closed bool

	fn func(K)

	// sends tracks the number of goroutines currently blocked on a channel
	// send. Close() waits for this to reach zero before closing the channel,
	// ensuring no send-on-closed-channel panic can occur.
	sends sync.WaitGroup

	ch chan K
}

func NewLRUPolicy[K comparable](size int) *Policy[K] {
	p := Policy[K]{
		l:    list.New(),
		size: size,
		ks:   make(map[K]*list.Element),
		ch:   make(chan K, 1),
	}

	p.fn = p.move

	return &p
}

func (p *Policy[K]) C() <-chan K {
	return p.ch
}

func (p *Policy[K]) Op(k K, op policy.OpType) error {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return policy.ErrClosed
	}

	var evicted K
	var hasEviction bool

	switch op {
	case policy.Set:
		evicted, hasEviction = p.insert(k)
	case policy.Get:
		p.fn(k)
	case policy.Evict:
		p.evict(k)
	}

	if hasEviction {
		// Register the send before releasing the lock so Close() cannot
		// observe sends==0 and close the channel between Unlock and the
		// actual send below.
		p.sends.Add(1)
	}

	p.mu.Unlock()

	if hasEviction {
		p.ch <- evicted
		p.sends.Done()
	}

	return nil
}

func (p *Policy[K]) move(k K) {
	e, ok := p.ks[k]
	if !ok {
		return
	}

	p.l.MoveToBack(e)
}

// insert adds k to the LRU list. If the list exceeds its size limit the
// least-recently-used key is removed and returned.
// Must be called with p.mu held.
func (p *Policy[K]) insert(k K) (evicted K, ok bool) {
	if _, exists := p.ks[k]; exists {
		return evicted, false
	}

	p.ks[k] = p.l.PushBack(k)

	if p.l.Len() > p.size {
		v := p.l.Remove(p.l.Front()).(K)
		delete(p.ks, v)
		return v, true
	}

	return evicted, false
}

// evict removes k from the list and map. Must be called with p.mu held.
func (p *Policy[K]) evict(k K) {
	e, ok := p.ks[k]
	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy[K]) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	// Wait for any sends that were registered before we set closed=true to
	// complete. After this, no new sends can be registered (Op returns
	// ErrClosed), so it is safe to close the channel.
	p.sends.Wait()

	close(p.ch)
	return nil
}
