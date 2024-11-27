package country

import (
	"strings"
	"sync"

	"github.com/upfluence/pkg/stringutil"
)

var (
	Alpha2CodeFetcher = &IndexedCodeFetcher{
		NormalizeKey: strings.ToLower,
		ExtractKeys:  func(cc CountryCode) []string { return []string{cc.Alpha2} },
	}

	Alpha3CodeFetcher = &IndexedCodeFetcher{
		NormalizeKey: strings.ToLower,
		ExtractKeys:  func(cc CountryCode) []string { return []string{cc.Alpha3} },
	}

	NameCodeFetcher = &IndexedCodeFetcher{
		NormalizeKey: func(k string) string {
			return strings.ToLower(stringutil.DecodeToASCII(k))
		},
		ExtractKeys: func(cc CountryCode) []string { return []string{cc.Name} },
	}

	AlternateNameCodeFetcher = &IndexedCodeFetcher{
		NormalizeKey: func(k string) string {
			return strings.ToLower(stringutil.DecodeToASCII(k))
		},
		ExtractKeys: func(cc CountryCode) []string { return cc.AlternateNames },
	}

	DefaultCodeFetcher = MultiCodeFetcher{
		Alpha2CodeFetcher,
		Alpha3CodeFetcher,
		NameCodeFetcher,
		AlternateNameCodeFetcher,
	}
)

type CodeFetcher interface {
	Fetch(string) (CountryCode, bool)
}

type IndexedCodeFetcher struct {
	NormalizeKey func(string) string
	ExtractKeys  func(CountryCode) []string

	// If left nil it will use DefaultCountryCodes
	CountryCodes []CountryCode

	once                sync.Once
	indexedCountryCodes map[string]CountryCode
}

func (icf *IndexedCodeFetcher) Fetch(key string) (CountryCode, bool) {
	if key == "" {
		return CountryCode{}, false
	}

	icf.once.Do(func() {
		ccs := icf.CountryCodes

		if ccs == nil {
			ccs = DefaultCountryCodes
		}

		icf.indexedCountryCodes = make(map[string]CountryCode, len(ccs))

		for _, cc := range ccs {
			for _, k := range icf.ExtractKeys(cc) {
				icf.indexedCountryCodes[icf.NormalizeKey(k)] = cc
			}
		}
	})

	cc, ok := icf.indexedCountryCodes[icf.NormalizeKey(key)]

	return cc, ok
}

type MultiCodeFetcher []CodeFetcher

func (cfs MultiCodeFetcher) Fetch(key string) (CountryCode, bool) {
	if key == "" {
		return CountryCode{}, false
	}

	for _, cf := range cfs {
		if cc, ok := cf.Fetch(key); ok {
			return cc, true
		}
	}

	return CountryCode{}, false
}
