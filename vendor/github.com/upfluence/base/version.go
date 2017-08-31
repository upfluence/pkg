package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 0, 22}
	gitVersion = &version.GitVersion{"d76c1b08cc0b", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
