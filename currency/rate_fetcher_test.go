package currency

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type staticRateFetcher map[string]float64

func (srf staticRateFetcher) BaseCurrency() Currency {
	return Currency("MOCK")
}

func (srf staticRateFetcher) Rate(_ context.Context, cur Currency) (float64, error) {
	if srf == nil {
		return .0, ErrCurrencyNotHandled
	}

	r, ok := srf[string(cur)]

	if !ok {
		return .0, ErrCurrencyNotHandled
	}

	return r, nil
}

func TestExchange(t *testing.T) {
	for _, tt := range []struct {
		rates map[string]float64

		in Money
		t  Currency

		wantMoney Money
		wantErr   error
	}{
		{in: Money{Currency: "foo"}, t: "foo", wantMoney: Money{Currency: "foo"}},
		{in: Money{Currency: "foo"}, t: "bar", wantMoney: Money{Currency: "bar"}},
		{
			in:      Money{Currency: "foo", Cents: 1},
			t:       "bar",
			wantErr: ErrCurrencyNotHandled,
		},
		{
			rates:     map[string]float64{"foo": 2., "bar": 4.},
			in:        Money{Currency: "foo", Cents: 100},
			t:         "bar",
			wantMoney: Money{Currency: "bar", Cents: 200},
		},
		{
			rates:     map[string]float64{"foo": 2., "bar": .0},
			in:        Money{Currency: "foo", Cents: 100},
			t:         "bar",
			wantMoney: Money{Currency: "bar"},
		},
	} {
		e := Exchanger{RateFetcher: staticRateFetcher(tt.rates)}

		res, err := e.Exchange(context.Background(), tt.in, tt.t)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantMoney, res)
	}
}
