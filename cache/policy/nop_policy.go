package policy

import (
	"sync"
	"sync/atomic"
)

type NopPolicy struct {
	closed int32

	chonce sync.Once
	ch     chan string
}

func (np *NopPolicy) C() <-chan string {
	np.chonce.Do(func() {
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
		if np.ch != nil {
			close(np.ch)
		}
	}

	return nil
}
