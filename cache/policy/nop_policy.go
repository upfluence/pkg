package policy

import (
	"sync"
	"sync/atomic"
)

type NopPolicy[K comparable] struct {
	sync.Once
	sync.Mutex

	closed int32
	ch     chan K
}

func (np *NopPolicy[K]) C() <-chan K {
	np.Do(func() {
		np.Lock()
		defer np.Unlock()

		if np.ch == nil {
			np.ch = make(chan K)

			if atomic.LoadInt32(&np.closed) == 1 {
				close(np.ch)
			}
		}
	})

	return np.ch
}

func (np *NopPolicy[K]) Op(K, OpType) error {
	if atomic.LoadInt32(&np.closed) == 1 {
		return ErrClosed
	}

	return nil
}

func (np *NopPolicy[K]) Close() error {
	if atomic.CompareAndSwapInt32(&np.closed, 0, 1) {
		np.Lock()
		if np.ch != nil {
			close(np.ch)
		}
		np.Unlock()
	}

	return nil
}
