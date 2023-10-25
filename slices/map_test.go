//go:build go1.21

package slices

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors"
)

var errMock = errors.New("test")

func TestMapWithContextError(t *testing.T) {
	var (
		ctx = context.Background()
	)

	for _, tt := range []struct {
		name    string
		in      []string
		fn      func(context.Context, string) (int64, error)
		want    []int64
		wantErr error
	}{
		{
			name: "error",
			in:   []string{"fail"},
			fn: func(context.Context, string) (int64, error) {
				return 0, errMock
			},
			wantErr: errMock,
		},
		{
			name: "success",
			in:   []string{"1", "22"},
			fn: func(_ context.Context, s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			want: []int64{1, 22},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapWithContextError(ctx, tt.in, tt.fn)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapWithError(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      []string
		fn      func(string) (int64, error)
		want    []int64
		wantErr error
	}{
		{
			name: "error",
			in:   []string{"fail"},
			fn: func(string) (int64, error) {
				return 0, errMock
			},
			wantErr: errMock,
		},
		{
			name: "success",
			in:   []string{"1", "22"},
			fn: func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			want: []int64{1, 22},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapWithError(tt.in, tt.fn)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMap(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      []string
		fn      func(string) int64
		want    []int64
		wantErr error
	}{
		{
			name: "success",
			in:   []string{"1", "22"},
			fn: func(s string) int64 {
				i, _ := strconv.ParseInt(s, 10, 64)
				return i
			},
			want: []int64{1, 22},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.in, tt.fn)
			assert.Equal(t, tt.want, got)
		})
	}
}
