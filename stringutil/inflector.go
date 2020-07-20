package stringutil

import "strings"

var DefaultInflector Inflector = multiInflector{
	mis: []matcherInflector{
		ExceptionInflector{singular: "medium", plural: "media"},
		suffixInflector{singular: "y", plural: "ies"},
	},
	di: InflectorFuncs{
		PluralizeFunc:   func(n string) string { return n + "s" },
		SingularizeFunc: func(n string) string { return strings.TrimSuffix(n, "s") },
	},
}

type Inflector interface {
	Pluralize(n string) string
	Singularize(n string) string
}

type matcherInflector interface {
	Inflector

	matchSingular(string) bool
	matchPlural(string) bool
}

type multiInflector struct {
	mis []matcherInflector

	di InflectorFuncs
}

func (mi multiInflector) Singularize(n string) string {
	for _, i := range mi.mis {
		if i.matchPlural(n) {
			return i.Singularize(n)
		}
	}

	return mi.di.Singularize(n)
}

func (mi multiInflector) Pluralize(n string) string {
	for _, i := range mi.mis {
		if i.matchSingular(n) {
			return i.Pluralize(n)
		}
	}

	return mi.di.Pluralize(n)
}

type suffixInflector struct {
	singular, plural string
}

func (si suffixInflector) matchSingular(n string) bool {
	return strings.HasSuffix(strings.ToLower(n), si.singular)
}

func (si suffixInflector) matchPlural(n string) bool {
	return strings.HasSuffix(strings.ToLower(n), si.plural)
}

func (si suffixInflector) Singularize(n string) string {
	if !si.matchPlural(n) {
		return n
	}

	return strings.TrimSuffix(strings.ToLower(n), si.plural) + si.singular
}

func (si suffixInflector) Pluralize(n string) string {
	if !si.matchSingular(n) {
		return n
	}

	return strings.TrimSuffix(strings.ToLower(n), si.singular) + si.plural
}

type ExceptionInflector struct {
	singular, plural string
}

func (ei ExceptionInflector) matchPlural(n string) bool {
	return strings.EqualFold(n, ei.plural)
}

func (ei ExceptionInflector) Singularize(n string) string {
	if !ei.matchPlural(n) {
		return n
	}

	return ei.singular
}

func (ei ExceptionInflector) matchSingular(n string) bool {
	return strings.EqualFold(n, ei.singular)
}

func (ei ExceptionInflector) Pluralize(n string) string {
	if !ei.matchSingular(n) {
		return n
	}

	return ei.plural
}

type InflectorFuncs struct {
	PluralizeFunc   func(string) string
	SingularizeFunc func(string) string
}

func (ifs InflectorFuncs) Pluralize(n string) string {
	return ifs.PluralizeFunc(n)
}

func (ifs InflectorFuncs) Singularize(n string) string {
	return ifs.SingularizeFunc(n)
}
