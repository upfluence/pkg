package amqp

import (
	"sync"

	"github.com/upfluence/pkg/amqp/channelpool"
	"github.com/upfluence/pkg/amqp/connectionpicker"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/balancer/roundrobin"
	"github.com/upfluence/pkg/discovery/resolver"
	"github.com/upfluence/pkg/multierror"
)

type Cluster struct {
	Picker connectionpicker.Picker
	Pool   channelpool.Pool

	b balancer.Balancer
}

func (c *Cluster) Close() error {
	return multierror.Combine(
		c.Pool.Close(),
		c.Picker.Close(),
		c.b.Close(),
	)
}

type ClusterDialer struct {
	BalancerBuilder balancer.Builder
	PickerOptions   []connectionpicker.Option
	PoolOptions     []channelpool.Option

	mu sync.Mutex
	cs map[string]*Cluster
}

func NewClusterDialerFromResolver(rb resolver.Builder) *ClusterDialer {
	return &ClusterDialer{
		BalancerBuilder: balancer.ResolverBuilder{
			Builder:      rb,
			BalancerFunc: roundrobin.BalancerFunc,
		},
	}
}

func (cd *ClusterDialer) Dial(n string) *Cluster {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if cd.cs == nil {
		cd.cs = make(map[string]*Cluster, 1)
	}

	c, ok := cd.cs[n]

	if ok {
		return c
	}

	b := cd.BalancerBuilder.Build(n)

	picker := connectionpicker.NewPicker(
		append(
			[]connectionpicker.Option{connectionpicker.WithBalancer(b)},
			cd.PickerOptions...,
		)...,
	)

	c = &Cluster{
		Picker: picker,
		Pool: channelpool.NewPool(
			append(
				[]channelpool.Option{channelpool.WithPicker(picker)},
				cd.PoolOptions...,
			)...,
		),
	}

	cd.cs[n] = c

	return c
}

func (cd *ClusterDialer) Close() error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	var errs []error

	for _, c := range cd.cs {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	cd.cs = nil

	return multierror.Wrap(errs)
}
