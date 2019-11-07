package timetest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	var c Clock

	assert.Equal(t, time.Time{}, c.Now())

	t0 := time.Now()
	c.MoveTo(t0)

	assert.Equal(t, t0, c.Now())
}

func TestTimer(t *testing.T) {
	var (
		c Clock

		timer = c.Timer(2 * time.Second)
	)

	assertEmptyChan(t, timer)
	c.MoveBy(1 * time.Second)
	assertEmptyChan(t, timer)
	c.MoveBy(2 * time.Second)

	t0 := <-timer.C()
	assert.Equal(t, time.Time{}.Add(2*time.Second), t0)

	timer.Reset(time.Second)

	c.MoveBy(2 * time.Second)
	t0 = <-timer.C()
	assert.Equal(t, time.Time{}.Add(4*time.Second), t0)

	timer.Reset(time.Second)
	timer.Stop()

	c.MoveBy(2 * time.Second)
	assertEmptyChan(t, timer)
}

func TestTimerFunc(t *testing.T) {
	var (
		c Clock

		ch = make(chan struct{})
	)

	c.TimerFunc(2*time.Second, func() { close(ch) })
	c.MoveBy(5 * time.Second)

	<-ch
}

type timeChanner interface {
	C() <-chan time.Time
}

func assertEmptyChan(t *testing.T, tc timeChanner) {
	select {
	case <-tc.C():
		t.Error("No tick were supposed to arrive here")
	default:
	}
}

func TestTicker(t *testing.T) {
	var (
		c Clock

		ticker = c.Ticker(2 * time.Second)
	)

	assertEmptyChan(t, ticker)
	c.MoveBy(1 * time.Second)
	assertEmptyChan(t, ticker)
	c.MoveBy(2 * time.Second)

	t0 := <-ticker.C()
	assert.Equal(t, time.Time{}.Add(2*time.Second), t0)

	ticker.Stop()

	c.MoveBy(5 * time.Second)

	assertEmptyChan(t, ticker)
}
