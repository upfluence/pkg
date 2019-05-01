package group

import "context"

type throttledGroup struct {
	Group

	ch chan struct{}
}

func (tg *throttledGroup) Do(r Runner) {
	tg.Group.Do(wrapRunner(r, tg.ch))
}

func SharedThrottledGroup(g Group, ch chan struct{}) Group {
	return &throttledGroup{Group: g, ch: ch}
}

func ThrottledGroup(g Group, cap int) Group {
	return SharedThrottledGroup(g, make(chan struct{}, cap))
}

func wrapRunner(r Runner, ch chan struct{}) Runner {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- struct{}{}:
		}

		err := r(ctx)

		select {
		case <-ch:
		default:
		}

		return err
	}
}
