package eucentralbank

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"

	"github.com/upfluence/pkg/currency"
	"github.com/upfluence/pkg/syncutil"
)

const apiURL = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

type RateFetcher struct {
	sf syncutil.Singleflight[map[currency.Currency]float64]

	cl  *http.Client
	req *http.Request

	duration time.Duration

	rates     map[currency.Currency]float64
	expiresAt time.Time

	currency currency.Currency
}

type cube struct {
	Currency string  `xml:"currency,attr"`
	Rate     float64 `xml:"rate,attr"`
}

type ratePayload struct {
	Cube []cube `xml:"Cube>Cube>Cube"`
}

func NewRateFetcher() *RateFetcher {
	req, _ := http.NewRequest("GET", apiURL, http.NoBody)

	return &RateFetcher{
		cl:       http.DefaultClient,
		req:      req,
		duration: time.Hour,
		currency: "EUR",
	}
}

func (rf *RateFetcher) BaseCurrency() currency.Currency { return rf.currency }

func (rf *RateFetcher) Rate(ctx context.Context, c currency.Currency) (float64, error) {
	if c == rf.currency {
		return 1., nil
	}

	_, rates, err := rf.sf.Do(ctx, func(ctx context.Context) (map[currency.Currency]float64, error) {
		if rf.expiresAt.After(time.Now()) && rf.rates != nil {
			return rf.rates, nil
		}

		resp, err := rf.cl.Do(rf.req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		var p ratePayload

		if err := xml.NewDecoder(resp.Body).Decode(&p); err != nil {
			return nil, err
		}

		res := make(map[currency.Currency]float64, len(p.Cube))

		for _, cube := range p.Cube {
			res[currency.Currency(cube.Currency)] = cube.Rate
		}

		rf.rates = res
		rf.expiresAt = time.Now().Add(rf.duration)

		return res, nil
	})

	if err != nil {
		return .0, err
	}

	if r, ok := rates[c]; ok {
		return r, nil
	}

	return .0, currency.ErrCurrencyNotHandled
}
