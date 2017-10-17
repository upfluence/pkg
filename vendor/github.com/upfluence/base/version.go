package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 1, 2}
	gitVersion = &version.GitVersion{"18049e54efa4", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
