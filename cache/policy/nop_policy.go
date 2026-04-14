package policy

import (
	"sync"
)

// NopPolicy is an eviction policy that never proactively evicts entries.
// All Op calls are no-ops; C() returns a channel that is only closed when
// Close() is called.
type NopPolicy[K comparable] struct {
	mu     sync.Mutex
	closed bool
	ch     chan K
}

func (np *NopPolicy[K]) C() <-chan K {
	np.mu.Lock()
	defer np.mu.Unlock()

	if np.ch == nil {
		np.ch = make(chan K)
		if np.closed {
			close(np.ch)
		}
	}

	return np.ch
}

func (np *NopPolicy[K]) Op(_ K, _ OpType) error {
	np.mu.Lock()
	defer np.mu.Unlock()

	if np.closed {
		return ErrClosed
	}

	return nil
}

func (np *NopPolicy[K]) Close() error {
	np.mu.Lock()
	defer np.mu.Unlock()

	if np.closed {
		return nil
	}

	np.closed = true

	if np.ch != nil {
		close(np.ch)
	}

	return nil
}
