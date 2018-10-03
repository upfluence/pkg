package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/prometheus/metricutil"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/pkg/responseutil"
)

const (
	namespace = "uthrift"
	subsystem = "service"

	Client RPCSide = "client"
	Server RPCSide = "server"
)

type RPCSide string

var (
	env = cfg.FetchString("ENV", "development")

	requestTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Histogram of processed items",
		},
		[]string{"rpc_side", "namespace", "service", "method", "type", "env", "status"},
	)

	requestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_duration_second",
			Help:      "Histogram of processing time",
		},
		[]string{"rpc_side", "namespace", "service", "method", "type", "env"},
	)
)

func init() {
	prometheus.MustRegister(requestTotalCount, requestHistogram)
}

type Factory struct {
	env  string
	side RPCSide
}

func NewFactory(side RPCSide) *Factory {
	return &Factory{env: env, side: side}
}

func (f *Factory) GetMiddleware(ns string, svc string) thrift.TMiddleware {
	return NewMiddleware(ns, svc, f.side, f.env)
}

func NewMiddleware(namespace, service string, side RPCSide, env string) *Middleware {
	return &Middleware{
		namespace: namespace,
		service:   service,
		side:      side,
		env:       env,
	}
}

type Middleware struct {
	namespace, service, env string
	side                    RPCSide
}

func (m *Middleware) HandleBinaryRequest(ctx thrift.Context, mth string, seqID int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) (thrift.TResponse, error)) (thrift.TResponse, error) {
	var (
		err error
		res thrift.TResponse
	)

	m.handle(
		mth,
		"binary",
		func() error {
			res, err = next(ctx, req)

			return responseutil.FetchError(res, err)
		},
	)

	return res, err
}

func (m *Middleware) HandleUnaryRequest(ctx thrift.Context, mth string, seqID int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) error) error {
	var err error

	m.handle(
		mth,
		"unary",
		func() error {
			err = next(ctx, req)
			return err
		},
	)

	return err
}

func (m *Middleware) handle(mth, mType string, next func() error) {
	var (
		t0  = time.Now()
		err = next()
	)

	requestTotalCount.WithLabelValues(
		string(m.side),
		m.namespace,
		m.service,
		mth,
		mType,
		m.env,
		metricutil.ResultStatus(err),
	).Inc()

	requestHistogram.WithLabelValues(
		string(m.side),
		m.namespace,
		m.service,
		mth,
		mType,
		env,
	).Observe(time.Since(t0).Seconds())
}
