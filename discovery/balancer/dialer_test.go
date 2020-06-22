package balancer_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/balancer/roundrobin"
	"github.com/upfluence/pkg/discovery/resolver/static"
)

func TestDialer(t *testing.T) {
	s1 := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			io.WriteString(rw, "s1")
		}),
	)

	s2 := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			io.WriteString(rw, "s2")
		}),
	)

	r := static.NewResolverFromStrings(
		[]string{s1.Listener.Addr().String(), s2.Listener.Addr().String()},
	)

	d := balancer.Dialer{
		Builder: balancer.ResolverBuilder{
			Builder:      r,
			BalancerFunc: roundrobin.BalancerFunc,
		},
	}

	defer d.Close()
	defer s1.Close()
	defer s2.Close()

	cl := http.Client{Transport: &http.Transport{DialContext: d.DialContext}}

	req, err := http.NewRequest("GET", "http://example.com/foo", http.NoBody)
	assert.Nil(t, err)

	for _, want := range []string{"s1", "s2", "s1"} {
		resp, err := cl.Do(req)
		assert.Nil(t, err)

		buf, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		resp.Body.Close()
		assert.Equal(t, want, string(buf))

		cl.CloseIdleConnections()
	}
}
