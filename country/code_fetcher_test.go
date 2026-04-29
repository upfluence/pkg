package country

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustFetch(k string) CountryCode {
	cc, ok := Alpha2CodeFetcher.Fetch(k)

	if !ok {
		panic("not found")
	}

	return cc
}

func TestCodeFetcher_Fetch(t *testing.T) {
	var zeroValue CountryCode

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
		{
			k:    "russia",
			want: mustFetch("RU"),
		},
		{
			k:    "Türkiye",
			want: mustFetch("TR"),
		},
		{k: ""},
	} {
		cc, ok := DefaultCodeFetcher.Fetch(tt.k)

		assert.Equal(t, tt.want, cc)

		if ok {
			assert.NotEqual(t, zeroValue, cc)
		} else {
			assert.Equal(t, zeroValue, cc)
		}
	}
}

func TestCodeFetcher_Search(t *testing.T) {
	for _, tt := range []struct {
		searchTerm      string
		searchOperatior SearchOperator
		want            []CountryCode
	}{
		{
			searchTerm:      "United",
			searchOperatior: SearchOperatorContains,
			want: []CountryCode{
				mustFetch("US"),
				mustFetch("UM"),
				mustFetch("AE"),
				mustFetch("TZ"),
				mustFetch("UK"),
			},
		},
		{
			searchTerm:      "fRa",
			searchOperatior: SearchOperatorContains,
			want: []CountryCode{
				mustFetch("FR"),
				mustFetch("FX"),
			},
		},
		{
			searchTerm:      "DE",
			searchOperatior: SearchOperatorContains,
			want: []CountryCode{
				mustFetch("CD"),
				mustFetch("FM"),
				mustFetch("GP"),
				mustFetch("RU"),
				mustFetch("DE"),
				mustFetch("LA"),
				mustFetch("CV"),
				mustFetch("BD"),
				mustFetch("KP"),
				mustFetch("DK"),
				mustFetch("SE"),
			},
		},
		{
			searchTerm:      "States Mi",
			searchOperatior: SearchOperatorContains,
			want: []CountryCode{
				mustFetch("UM"),
			},
		},
		{
			searchTerm:      "States Mi",
			searchOperatior: SearchOperatorMatchBoolPrefix,
			want: []CountryCode{
				mustFetch("UM"), // "United States Minor Outlying Islands"
				mustFetch("FM"), // "Micronesia, Federated States of"
			},
		},
	} {
		assert.ElementsMatch(t, tt.want, DefaultCodeFetcher.Search(tt.searchTerm, tt.searchOperatior))
	}
}
