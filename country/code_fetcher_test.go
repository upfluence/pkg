package country

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeFetcher(t *testing.T) {
	mustFetch := func(k string) CountryCode {
		cc, ok := Alpha2CodeFetcher.Fetch(k)

		if !ok {
			panic("not found")
		}

		return cc
	}

	for _, tt := range []struct {
		k    string
		want CountryCode
	}{
		{
			k:    "fr",
			want: mustFetch("FR"),
		},
		{
			k:    "fR",
			want: mustFetch("FR"),
		},
		{k: "zz"},
		{
			k:    "USA",
			want: mustFetch("US"),
		},
		{
			k:    "United states",
			want: mustFetch("US"),
		},
		{
			k:    "Côte d'Ivoire",
			want: mustFetch("CI"),
		},
		{
			k:    "cote d'ivoire",
			want: mustFetch("CI"),
		},
		{k: ""},
	} {
		cc, ok := DefaultCodeFetcher.Fetch(tt.k)

		assert.Equal(t, tt.want != CountryCode{}, ok)
		assert.Equal(t, tt.want, cc)
	}
}
