package eucentralbank

import (
	"context"
	"encoding/xml"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/upfluence/pkg/currency"
)

const apiURL = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

type RateFetcher struct {
	mu sync.RWMutex
	g  singleflight.Group

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

	rf.mu.RLock()

	if rf.expiresAt.After(time.Now()) && rf.rates != nil {
		r, ok := rf.rates[c]
		rf.mu.RUnlock()

		if ok {
			return r, nil
		}

		return .0, currency.ErrCurrencyNotHandled
	}

	rf.mu.RUnlock()

	resc := rf.g.DoChan("", func() (interface{}, error) {
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

		rf.mu.Lock()

		rf.rates = res
		rf.expiresAt = time.Now().Add(rf.duration)

		rf.mu.Unlock()

		return res, nil
	})

	select {
	case <-ctx.Done():
		return .0, ctx.Err()
	case res := <-resc:
		if res.Err != nil {
			return .0, res.Err
		}

		rates := res.Val.(map[currency.Currency]float64)
		r, ok := rates[c]

		if ok {
			return r, nil
		}

		return .0, currency.ErrCurrencyNotHandled
	}
}
