package cfg

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

type Configurator interface {
	Populate(context.Context, interface{}) error
}

type configurator struct {
	providers []provider.Provider
	factory   setter.Factory
}

func NewDefaultConfigurator(providers ...provider.Provider) Configurator {
	cfg := newConfigurator(
		append(providers, env.NewDefaultProvider(), flags.NewDefaultProvider())...,
	)

	return &helpConfigurator{configurator: cfg, stderr: os.Stderr}
}

func NewConfigurator(providers ...provider.Provider) Configurator {
	return newConfigurator(providers...)
}

func newConfigurator(providers ...provider.Provider) *configurator {
	return &configurator{
		providers: providers,
		factory:   setter.DefaultFactory,
	}
}

func (c *configurator) Populate(ctx context.Context, in interface{}) error {
	return walker.Walk(
		in,
		func(f *walker.Field) error { return c.walkFunc(ctx, f) },
	)
}

func (c *configurator) walkFunc(ctx context.Context, f *walker.Field) error {
	s := c.factory.Build(f.Field)

	if s == nil {
		return nil
	}

	for _, p := range c.providers {
		var (
			v   string
			ok  bool
			err error
		)

		for _, k := range walker.BuildFieldKeys(p.StructTag(), f) {
			v, ok, err = p.Provide(ctx, k)

			if err != nil {
				return errors.Wrapf(
					err,
					"Populate {struct: %T field: %s}",
					v,
					f.Field.Name,
				)
			}

			if ok {
				break
			}
		}

		if !ok {
			continue
		}

		if err := s.Set(v, f.Value.Interface()); err != nil {
			return err
		}
	}

	if f.Value.Type().Implements(setter.ValueType) {
		return walker.SkipStruct
	}

	return nil
}
