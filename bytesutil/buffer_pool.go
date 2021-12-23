package bytesutil

import (
	"bytes"
	"sync"
)

const DefaultMaxBufferSize = 1 << 22 // 4MB

type BufferPool struct {
	pool          sync.Pool
	maxBufferSize int
}

func NewBufferPool() *BufferPool {
	return NewBufferPoolWithMaxBufferSize(DefaultMaxBufferSize)
}

func NewBufferPoolWithMaxBufferSize(maxBufferSize int) *BufferPool {
	return &BufferPool{
		pool:          sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
		maxBufferSize: maxBufferSize,
	}
}

func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	if buf.Cap() > bp.maxBufferSize {
		// let it be garbage collected
		return
	}

	buf.Reset()
	bp.pool.Put(buf)
}
