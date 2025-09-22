package eucentralbank

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/currency"
)

const bodyStub = `
<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<gesmes:subject>Reference rates</gesmes:subject>
	<gesmes:Sender>
		<gesmes:name>European Central Bank</gesmes:name>
	</gesmes:Sender>
	<Cube>
		<Cube time='2020-05-21'>
			<Cube currency='USD' rate='1.1000'/>
			<Cube currency='JPY' rate='118.42'/>
			<Cube currency='BGN' rate='1.9558'/>
			<Cube currency='CZK' rate='27.212'/>
			<Cube currency='DKK' rate='7.4563'/>
			<Cube currency='GBP' rate='0.89943'/>
			<Cube currency='HUF' rate='348.59'/>
			<Cube currency='PLN' rate='4.5298'/>
			<Cube currency='RON' rate='4.8423'/>
			<Cube currency='SEK' rate='10.5300'/>
			<Cube currency='CHF' rate='1.0628'/>
			<Cube currency='ISK' rate='156.30'/>
			<Cube currency='NOK' rate='10.9030'/>
			<Cube currency='HRK' rate='7.5805'/>
			<Cube currency='RUB' rate='77.9883'/>
			<Cube currency='TRY' rate='7.4781'/>
			<Cube currency='AUD' rate='1.6710'/>
			<Cube currency='BRL' rate='6.2532'/>
			<Cube currency='CAD' rate='1.5310'/>
			<Cube currency='CNY' rate='7.8153'/>
			<Cube currency='HKD' rate='8.5298'/>
			<Cube currency='IDR' rate='16205.75'/>
			<Cube currency='ILS' rate='3.8659'/>
			<Cube currency='INR' rate='83.0545'/>
			<Cube currency='KRW' rate='1354.60'/>
			<Cube currency='MXN' rate='25.5043'/>
			<Cube currency='MYR' rate='4.7801'/>
			<Cube currency='NZD' rate='1.7949'/>
			<Cube currency='PHP' rate='55.649'/>
			<Cube currency='SGD' rate='1.5561'/>
			<Cube currency='THB' rate='34.991'/>
			<Cube currency='ZAR' rate='19.6577'/>
		</Cube>
	</Cube>
</gesmes:Envelope>
`

type staticTripper struct {
	body string
}

func (st *staticTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(st.body)),
	}, nil
}

func TestRateFetcher(t *testing.T) {
	rf := NewRateFetcher()
	rf.cl.Transport = &staticTripper{body: bodyStub}

	for _, tt := range []struct {
		cur currency.Currency

		wantRate float64
		wantErr  error
	}{
		{cur: "USD", wantRate: 1.1},
		{cur: "EUR", wantRate: 1.},
		{cur: "XYZ", wantErr: currency.ErrCurrencyNotHandled},
	} {
		rate, err := rf.Rate(context.Background(), tt.cur)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantRate, rate)
	}
}
