package timeutil

import "time"

var defaultClock = clock{}

func Background() Clock { return defaultClock }

type Clock interface {
	Now() time.Time

	Timer(time.Duration) Timer
	TimerFunc(time.Duration, func()) Timer

	Ticker(time.Duration) Ticker
}

type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(time.Duration) bool
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type clock struct{}

func (clock) Now() time.Time                { return time.Now() }
func (clock) Timer(d time.Duration) Timer   { return timer{time.NewTimer(d)} }
func (clock) Ticker(d time.Duration) Ticker { return ticker{time.NewTicker(d)} }

func (clock) TimerFunc(d time.Duration, fn func()) Timer {
	return timer{time.AfterFunc(d, fn)}
}

type timer struct {
	*time.Timer
}

func (t timer) C() <-chan time.Time {
	return t.Timer.C
}

type ticker struct {
	*time.Ticker
}

func (t ticker) C() <-chan time.Time {
	return t.Ticker.C
}
