package csv

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/upfluence/cfg"
)

type Reader struct {
	r *csv.Reader

	c cfg.Configurator
	p *provider
}

func NewReader(r io.Reader) (*Reader, error) {
	cr := csv.NewReader(r)

	headers, err := cr.Read()

	if err != nil {
		return nil, err
	}

	p := newProvider(headers)

	return &Reader{
		r: cr,
		p: p,
		c: cfg.NewConfigurator(p),
	}, nil
}

func (r *Reader) Read(v interface{}) (map[string]string, error) {
	line, err := r.r.Read()

	if err != nil {
		return nil, err
	}

	r.p.reset(line)

	if err := r.c.Populate(context.Background(), v); err != nil {
		return nil, err
	}

	return r.p.leftover(), nil
}

type provider struct {
	headers []string
	idxs    map[string]int

	read map[int]struct{}
	line []string
}

func newProvider(headers []string) *provider {
	idxs := make(map[string]int, len(headers))

	for i, v := range headers {
		idxs[v] = i
	}

	return &provider{headers: headers, idxs: idxs}
}

func (p *provider) reset(line []string) {
	p.read = make(map[int]struct{})
	p.line = line
}

func (p *provider) leftover() map[string]string {
	if len(p.read) == len(p.line) {
		return nil
	}

	vs := make(map[string]string, len(p.line)-len(p.read))

	for i, h := range p.headers {
		if _, ok := p.read[i]; !ok {
			vs[h] = p.line[i]
		}
	}

	return vs
}

func (p *provider) StructTag() string { return "csv" }

func (p *provider) Provide(_ context.Context, k string) (string, bool, error) {
	idx, ok := p.idxs[k]

	if !ok {
		return "", false, nil
	}

	p.read[idx] = struct{}{}

	return p.line[idx], true, nil
}
