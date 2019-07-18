package policy

import (
	"sync"
	"sync/atomic"
)

type NopPolicy struct {
	sync.Once
	sync.Mutex

	closed int32
	ch     chan string
}

func (np *NopPolicy) C() <-chan string {
	np.Do(func() {
		np.Lock()
		defer np.Unlock()

		if np.ch == nil {
			np.ch = make(chan string)

			if atomic.LoadInt32(&np.closed) == 1 {
				close(np.ch)
			}
		}
	})

	return np.ch
}

func (np *NopPolicy) Op(string, OpType) error {
	if atomic.LoadInt32(&np.closed) == 1 {
		return ErrClosed
	}

	return nil
}

func (np *NopPolicy) Close() error {
	if atomic.CompareAndSwapInt32(&np.closed, 0, 1) {
		np.Lock()
		if np.ch != nil {
			close(np.ch)
		}
		np.Unlock()
	}

	return nil
}
