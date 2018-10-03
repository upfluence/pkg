package proxyutil

import (
	"net/http"
	"time"

	thttp "github.com/upfluence/proxy/http"
	"github.com/upfluence/uthrift-go/uthrift/client/pool"
	"github.com/upfluence/uthrift-go/uthrift/clientprovider"
	"github.com/upfluence/uthrift-go/uthrift/middleware/prometheus"
	"github.com/upfluence/uthrift-go/uthrift/middleware/stapler"
	"github.com/upfluence/uthrift-go/uthrift/transport/http/binary"

	"github.com/upfluence/pkg/proxyutil/local"
	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/pool/bounded"
)

const (
	prefetch     = 20
	maxIdleConns = 20
)

var (
	proxyURL = cfg.FetchString(
		"PROXY_HTTP_SERVICE_URL",
		"http://localhost:8080/proxy",
	)

	trans = &http.Transport{
		MaxIdleConnsPerHost: maxIdleConns,
		MaxIdleConns:        maxIdleConns,
		IdleConnTimeout:     2 * time.Minute,
	}
)

func BuildRequester() (thttp.HTTPProxy, error) {
	if cfg.FetchBool("LOCAL_REQUESTER", false) {
		return local.NewRequester(&http.Client{Transport: trans}), nil
	}

	return thttp.NewHTTPProxyClientFactoryProvider(
		clientprovider.NewBareProvider(
			clientprovider.WithClient(
				pool.NewFactory(bounded.NewPoolFactory(5*prefetch)),
			),
			clientprovider.WithTransport(binary.NewSingleEndpointFactory(proxyURL)),
			clientprovider.WithMiddlewares(
				prometheus.NewFactory(prometheus.Client),
				stapler.NewDefaultFactory(),
			),
		),
	)
}
