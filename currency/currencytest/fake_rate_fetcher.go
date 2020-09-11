package currencytest

import (
	"context"

	"github.com/upfluence/pkg/currency"
)

type FakeRateFetcher struct {
	Currency string
	Rates    map[string]float64
	Err      error
}

func (frf FakeRateFetcher) BaseCurrency() currency.Currency {
	if frf.Currency == "" {
		return currency.Currency("USD")
	}

	return currency.Currency(frf.Currency)
}

func (frf FakeRateFetcher) Rate(_ context.Context, cur currency.Currency) (float64, error) {
	if frf.Err != nil {
		return .0, frf.Err
	}

	if cur == frf.BaseCurrency() {
		return 1., nil
	}

	if frf.Rates == nil {
		return .0, currency.ErrCurrencyNotHandled
	}

	r, ok := frf.Rates[string(cur)]

	if !ok {
		return .0, currency.ErrCurrencyNotHandled
	}

	return r, nil
}
