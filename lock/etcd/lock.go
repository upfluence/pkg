package etcd

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/upfluence/pkg/backoff"
	"github.com/upfluence/pkg/backoff/exponential"
	"github.com/upfluence/pkg/lock"
)

type lockManager struct {
	cl client.KeysAPI

	s      backoff.Strategy
	key    string
	prefix string

	sync.Mutex
	ls map[string]*lockImpl
}

func NewLockManager(url, key, prefix string) (lock.LockManager, error) {
	cl, err := client.New(client.Config{Endpoints: []string{url}})

	if err != nil {
		return nil, err
	}

	return &lockManager{
		cl:     client.NewKeysAPI(cl),
		s:      exponential.NewDefaultBackoff(5*time.Millisecond, time.Second),
		key:    key,
		prefix: prefix,
	}, nil
}

func (lm *lockManager) Lock(n string) lock.Lock {
	return &lockImpl{lm: lm, n: n}
}

type lockImpl struct {
	lm  *lockManager
	n   string
	cnt int32
}

func (l *lockImpl) String() string { return l.n }

func (l *lockImpl) Acquire(ctx context.Context, opts lock.AcquireOptions) (lock.Lease, error) {
	if time.Since(opts.Deadline) > 0 {
		return nil, lock.ErrPastTime
	}

	inc := atomic.AddInt32(&l.cnt, 1)

	le := lease{
		l: l,
		d: opts.Deadline,
		k: fmt.Sprintf("%s/%s", l.lm.prefix, l.n),
		v: fmt.Sprintf("%s-%d", l.lm.key, inc),
	}

	le.Context, le.cancel = context.WithCancel(context.Background())

	var i int

	for {
		err := le.acquire(ctx, opts.Deadline)
		switch terr := err.(type) {
		case nil:
			le.t = time.AfterFunc(
				time.Until(opts.Deadline),
				func() { le.Release(context.Background()) },
			)

			return &le, nil
		case client.Error:
			if terr.Code != client.ErrorCodeNodeExist {
				return nil, err
			}

			if opts.NoWait {
				return nil, lock.ErrAlreadyAcquired
			}
		default:
			return nil, err
		}

		d, err := l.lm.s.Backoff(i)

		if err != nil {
			return nil, err
		}

		time.Sleep(d)
		i++
	}
}

type lease struct {
	l *lockImpl

	context.Context
	cancel context.CancelFunc

	ctxfn func() (context.Context, context.CancelFunc)

	k, v string

	sync.Mutex
	t *time.Timer
	d time.Time
}

func (l *lease) acquire(ctx context.Context, d time.Time) error {
	_, err := l.l.lm.cl.Set(
		ctx,
		l.k,
		l.v,
		&client.SetOptions{PrevExist: client.PrevNoExist, TTL: time.Until(d) + time.Second},
	)

	return err
}

func (l *lease) release(ctx context.Context) error {
	_, err := l.l.lm.cl.Delete(ctx, l.k, &client.DeleteOptions{PrevValue: l.v})

	return err
}

func (l *lease) renew(ctx context.Context, d time.Time) error {
	_, err := l.l.lm.cl.Set(
		ctx,
		l.k,
		l.v,
		&client.SetOptions{PrevValue: l.v, TTL: time.Until(d) + time.Second},
	)

	return err
}

func (l *lease) Release(ctx context.Context) error {
	l.cancel()
	d := l.Deadline()

	if time.Until(d) < 0 {
		return l.Err()
	}

	err := l.release(ctx)
	l.t.Stop()

	return err
}

func (l *lease) KeepAlive(ctx context.Context, d time.Time) error {
	if time.Until(d) <= 0 {
		return lock.ErrPastTime
	}

	if err := l.Err(); err != nil {
		return err
	}

	if l.Deadline().After(d) {
		return nil
	}

	err := l.renew(ctx, d)

	if err == nil {
		l.Lock()
		l.t.Reset(time.Until(d))
		l.d = d
		l.Unlock()
	}

	return err
}

func (l *lease) Deadline() time.Time {
	l.Lock()
	d := l.d
	l.Unlock()

	return d
}
