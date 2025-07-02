package currency

import (
	"context"

	"github.com/upfluence/errors"
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
	exm, err := e.exchange(ctx, m, c)

	return exm, errors.WithTags(err, map[string]interface{}{
		"target_currency": c,
		"source_currency": m.Currency,
		"source_cents":    m.Cents,
	})
}

func (e *Exchanger) exchange(ctx context.Context, m Money, c Currency) (Money, error) {
	if c == m.Currency {
		return m, nil
	}

	if m.Cents == .0 {
		return Money{Currency: c}, nil
	}

	var (
		err error

		fromRate = 1.
		toRate   = 1.
	)

	if m.Currency != e.BaseCurrency() {

		fromRate, err = e.Rate(ctx, m.Currency)

		if err != nil {
			return Money{}, errors.Wrap(err, "could not get from rate")
		}
	}

	if c != e.BaseCurrency() {
		toRate, err = e.Rate(ctx, c)

		if err != nil {
			return Money{}, errors.Wrap(err, "could not get to rate")
		}
	}

	if fromRate == .0 || toRate == .0 {
		return Money{Currency: c}, nil
	}

	cents := int(float64(m.Cents) * toRate / fromRate)
	return Money{Cents: cents, Currency: c}, nil
}
