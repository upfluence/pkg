package country

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"

	uslices "github.com/upfluence/pkg/slices"
	"github.com/upfluence/pkg/v2/stringutil"
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

type SearchOperator func(key, searchTerm string) bool

func SearchOperatorContains(key, searchTerm string) bool {
	return strings.Contains(strings.ToLower(key), strings.ToLower(searchTerm))
}

func SearchOperatorMatchBoolPrefix(key, searchTerm string) bool {
	words := strings.FieldsFunc(searchTerm, unicode.IsSpace)

	if len(words) == 0 {
		return true
	}

	for _, word := range words[:len(words)-1] {
		// We use `(?i)` to make the match case-insensitive
		// We use `regexp.QuoteMeta` to escape any special characters in the word
		pattern := `(?i)\b` + regexp.QuoteMeta(word) + `\b`
		matched, err := regexp.MatchString(pattern, key)

		if err != nil {
			// In the unlikely event of a regex compilation error, return false
			return false
		}

		if !matched {
			return false
		}
	}

	regexp.QuoteMeta(words[len(words)-1])
	pattern := `(?i)\b` + regexp.QuoteMeta(words[len(words)-1])
	matched, err := regexp.MatchString(pattern, key)

	if err != nil {
		// In the unlikely event of a regex compilation error, return false
		return false
	}

	return matched
}

type CodeFetcher interface {
	Fetch(string) (CountryCode, bool)
	Search(string, SearchOperator) []CountryCode
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

	icf.once.Do(icf.prepareIndex)

	cc, ok := icf.indexedCountryCodes[icf.NormalizeKey(key)]

	return cc, ok
}

func (icf *IndexedCodeFetcher) Search(searchTerm string, operator SearchOperator) []CountryCode {
	if searchTerm == "" {
		return nil
	}

	icf.once.Do(icf.prepareIndex)

	return uslices.Reduce(
		slices.Collect(maps.Keys(icf.indexedCountryCodes)),
		func(acc []CountryCode, key string) []CountryCode {
			if operator(key, searchTerm) {
				return append(acc, icf.indexedCountryCodes[key])
			}

			return acc
		},
	)
}

func (icf *IndexedCodeFetcher) prepareIndex() {
	ccs := icf.CountryCodes

	if ccs == nil {
		ccs = DefaultCountryCodes
	}

	icf.indexedCountryCodes = make(map[string]CountryCode, len(ccs))

	for _, cc := range ccs {
		for _, k := range icf.ExtractKeys(cc) {
			normalizedKey := icf.NormalizeKey(k)

			if assigned, ok := icf.indexedCountryCodes[normalizedKey]; ok && assigned.Assignment < cc.Assignment {
				continue
			}

			icf.indexedCountryCodes[normalizedKey] = cc
		}
	}
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

func (cfs MultiCodeFetcher) Search(key string, operator SearchOperator) []CountryCode {
	if key == "" {
		return nil
	}

	countryCodeByAlpha2 := map[string]CountryCode{}

	for _, cf := range cfs {
		for _, cc := range cf.Search(key, operator) {
			countryCodeByAlpha2[cc.Alpha2] = cc
		}
	}

	return slices.Collect(maps.Values(countryCodeByAlpha2))
}
