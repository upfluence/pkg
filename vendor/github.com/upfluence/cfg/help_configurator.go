package cfg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
)

type helpConfig struct {
	Help bool `flag:"h,help" env:"HELP"`
}

var (
	defaultHeaders = []byte("Arguments:\n")
)

type helpConfigurator struct {
	*configurator

	stderr io.Writer
}

func (hc *helpConfigurator) Populate(ctx context.Context, in interface{}) error {
	var cfg helpConfig

	if err := hc.configurator.Populate(ctx, &cfg); err != nil {
		return err
	}

	if cfg.Help {
		hc.printDefaults(in)
		os.Exit(2)
	}

	return hc.configurator.Populate(ctx, in)
}

func (hc *helpConfigurator) printDefaults(in interface{}) error {
	hc.stderr.Write(defaultHeaders)

	return walker.Walk(
		in,
		func(f *walker.Field) error {
			s := hc.factory.Build(f.Field)

			if s == nil {
				return nil
			}

			fks := walker.BuildFieldKeys("", f)

			if len(fks) == 0 {
				return nil
			}

			var b bytes.Buffer

			b.WriteString("\t- ")

			b.WriteString(fks[0])

			b.WriteString(": ")
			b.WriteString(s.String())

			if h, ok := f.Field.Tag.Lookup("help"); ok {
				b.WriteString(" ")
				b.WriteString(h)
			}

			fv := reflectutil.IndirectedValue(f.Value).FieldByName(f.Field.Name)
			if !reflectutil.IsZero(fv) {
				v := reflectutil.IndirectedValue(fv).Interface()

				b.WriteString(" (default: ")

				if ss, ok := v.(fmt.Stringer); ok {
					b.WriteString(ss.String())
				} else {
					fmt.Fprintf(&b, "%+v", v)
				}

				b.WriteString(")")
			}

			var providedKeys []string

			for _, p := range hc.providers {
				if kf, ok := p.(provider.KeyFormatterProvider); ok {
					var ks []string

					for _, k := range walker.BuildFieldKeys(p.StructTag(), f) {
						ks = append(ks, kf.FormatKey(k))
					}

					if len(ks) > 0 {
						providedKeys = append(
							providedKeys,
							fmt.Sprintf("%s: %s", p.StructTag(), strings.Join(ks, ", ")),
						)
					}
				}
			}

			if len(providedKeys) > 0 {
				b.WriteString(" (")
				b.WriteString(strings.Join(providedKeys, ", "))
				b.WriteString(")")
			}

			b.WriteRune('\n')
			b.WriteTo(hc.stderr)

			return nil
		},
	)
}
