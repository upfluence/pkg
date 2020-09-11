package currency_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/currency"
	"github.com/upfluence/pkg/currency/currencytest"
)

func TestExchange(t *testing.T) {
	for _, tt := range []struct {
		rates map[string]float64

		in currency.Money
		t  currency.Currency

		wantMoney currency.Money
		wantErr   error
	}{
		{
			in:        currency.Money{Currency: "foo"},
			t:         "foo",
			wantMoney: currency.Money{Currency: "foo"},
		},
		{
			in:        currency.Money{Currency: "foo"},
			t:         "bar",
			wantMoney: currency.Money{Currency: "bar"},
		},
		{
			in:      currency.Money{Currency: "foo", Cents: 1},
			t:       "bar",
			wantErr: currency.ErrCurrencyNotHandled,
		},
		{
			rates:     map[string]float64{"foo": 2., "bar": 4.},
			in:        currency.Money{Currency: "foo", Cents: 100},
			t:         "bar",
			wantMoney: currency.Money{Currency: "bar", Cents: 200},
		},
		{
			rates:     map[string]float64{"foo": 2., "bar": .0},
			in:        currency.Money{Currency: "foo", Cents: 100},
			t:         "bar",
			wantMoney: currency.Money{Currency: "bar"},
		},
	} {
		e := currency.Exchanger{
			RateFetcher: currencytest.FakeRateFetcher{Rates: tt.rates},
		}

		res, err := e.Exchange(context.Background(), tt.in, tt.t)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantMoney, res)
	}
}
