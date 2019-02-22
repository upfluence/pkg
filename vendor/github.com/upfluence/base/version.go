package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 1, 26}
	gitVersion = &version.GitVersion{"12e1e202b411", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
