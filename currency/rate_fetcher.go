package currency

import (
	"context"
	"errors"
)

type Currency string

var ErrCurrencyNotHandled = errors.New("currency: Currency not handled")

type RateFetcher interface {
	BaseCurrency() Currency
	Rate(context.Context, Currency) (float64, error)
}

type Money struct {
	Cents    int
	Currency Currency
}

type Exchanger struct {
	RateFetcher
}

func (e *Exchanger) Exchange(ctx context.Context, m Money, c Currency) (Money, error) {
	if c == m.Currency {
		return m, nil
	}

	if m.Cents == .0 {
		return Money{Currency: c}, nil
	}

	var (
		fromRate = 1.
		toRate   = 1.
	)

	if m.Currency != e.BaseCurrency() {
		var err error

		fromRate, err = e.Rate(ctx, m.Currency)

		if err != nil {
			return Money{}, err
		}
	}

	if c != e.BaseCurrency() {
		var err error

		toRate, err = e.Rate(ctx, c)

		if err != nil {
			return Money{}, err
		}
	}

	if fromRate == .0 || toRate == .0 {
		return Money{Currency: c}, nil
	}

	cents := int(float64(m.Cents) * toRate / fromRate)
	return Money{Cents: cents, Currency: c}, nil
}
