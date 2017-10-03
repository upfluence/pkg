package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 1, 0}
	gitVersion = &version.GitVersion{"255c893ecd37", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
