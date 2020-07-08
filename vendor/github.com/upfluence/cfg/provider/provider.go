package provider

import "context"

type Provider interface {
	StructTag() string
	Provide(context.Context, string) (string, bool, error)
}

func ProvideError(tag string, err error) Provider {
	return &faultyProvider{tag: tag, err: err}
}

type faultyProvider struct {
	tag string
	err error
}

func (p *faultyProvider) Err() error        { return p.err }
func (p *faultyProvider) StructTag() string { return p.tag }
func (p *faultyProvider) Provide(context.Context, string) (string, bool, error) {
	return "", false, p.err
}

type KeyFn func(string) string

func NewStaticProvider(tag string, vs map[string]string, kfn KeyFn) Provider {
	return &staticProvider{vs: vs, tag: tag, kfn: kfn}
}

type staticProvider struct {
	vs  map[string]string
	tag string
	kfn KeyFn
}

func (sp *staticProvider) StructTag() string { return sp.tag }

func (sp *staticProvider) Provide(_ context.Context, k string) (string, bool, error) {
	v, ok := sp.vs[sp.kfn(k)]

	return v, ok, nil
}

type KeyFormatterProvider interface {
	Provider

	FormatKey(string) string
}

type KeyFormatterFunc struct {
	Provider

	KeyFormatFunc func(string) string
}

func (kff KeyFormatterFunc) FormatKey(n string) string {
	return kff.KeyFormatFunc(n)
}
