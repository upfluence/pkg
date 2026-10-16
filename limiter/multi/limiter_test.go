//nolint:goconst
package multi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/limiter"
)

var errBoom = errors.New("boom")

type fakeReserver struct {
	name string
	err  error
	log  *[]string
}

func (fr fakeReserver) Reserve(_ context.Context, _ limiter.AllowOptions) (limiter.DoneFunc, limiter.CancelFunc, error) {
	*fr.log = append(*fr.log, "reserve "+fr.name)

	if fr.err != nil {
		return nil, nil, fr.err
	}

	return func() { *fr.log = append(*fr.log, "done "+fr.name) },
		func() { *fr.log = append(*fr.log, "cancel "+fr.name) },
		nil
}

func TestReserve(t *testing.T) {
	for _, tt := range []struct {
		name     string
		haveErrs map[string]error
		wantErr  error
		wantLog  []string
	}{
		{
			name:    "all succeed",
			wantLog: []string{"reserve a", "reserve b", "reserve c"},
		},
		{
			name:     "middle fails",
			haveErrs: map[string]error{"b": errBoom},
			wantErr:  errBoom,
			wantLog:  []string{"reserve a", "reserve b", "cancel a"},
		},
		{
			name:     "last fails",
			haveErrs: map[string]error{"c": limiter.ErrLimited},
			wantErr:  limiter.ErrLimited,
			wantLog:  []string{"reserve a", "reserve b", "reserve c", "cancel b", "cancel a"},
		},
		{
			name:     "first fails",
			haveErrs: map[string]error{"a": errBoom},
			wantErr:  errBoom,
			wantLog:  []string{"reserve a"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				log []string
				rs  []limiter.Reserver
			)

			for _, n := range []string{"a", "b", "c"} {
				rs = append(rs, fakeReserver{name: n, err: tt.haveErrs[n], log: &log})
			}

			done, cancel, err := NewLimiter(rs...).Reserve(context.Background(), limiter.AllowOptions{N: 1})

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantLog, log)

			if tt.wantErr != nil {
				assert.Nil(t, done)
				assert.Nil(t, cancel)
			}
		})
	}
}

func TestReserveEnd(t *testing.T) {
	for _, tt := range []struct {
		name    string
		haveEnd func(limiter.DoneFunc, limiter.CancelFunc)
		wantLog []string
	}{
		{
			name:    "done",
			haveEnd: func(d limiter.DoneFunc, _ limiter.CancelFunc) { d() },
			wantLog: []string{"done a", "done b"},
		},
		{
			name:    "cancel in reverse",
			haveEnd: func(_ limiter.DoneFunc, c limiter.CancelFunc) { c() },
			wantLog: []string{"cancel b", "cancel a"},
		},
		{
			name: "done twice",
			haveEnd: func(d limiter.DoneFunc, _ limiter.CancelFunc) {
				d()
				d()
			},
			wantLog: []string{"done a", "done b"},
		},
		{
			name: "cancel after done",
			haveEnd: func(d limiter.DoneFunc, c limiter.CancelFunc) {
				d()
				c()
			},
			wantLog: []string{"done a", "done b"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var log []string

			done, cancel, err := NewLimiter(
				fakeReserver{name: "a", log: &log},
				fakeReserver{name: "b", log: &log},
			).Reserve(context.Background(), limiter.AllowOptions{N: 1})

			require.NoError(t, err)

			log = nil

			tt.haveEnd(done, cancel)

			assert.Equal(t, tt.wantLog, log)
		})
	}
}

func TestAllow(t *testing.T) {
	var log []string

	done, err := NewLimiter(
		fakeReserver{name: "a", log: &log},
		fakeReserver{name: "b", err: errBoom, log: &log},
	).Allow(context.Background(), limiter.AllowOptions{})

	require.ErrorIs(t, err, errBoom)
	assert.Nil(t, done)
	assert.Equal(t, []string{"reserve a", "reserve b", "cancel a"}, log)

	done, err = NewLimiter().Allow(context.Background(), limiter.AllowOptions{})

	require.NoError(t, err)
	assert.NotNil(t, done)
}
