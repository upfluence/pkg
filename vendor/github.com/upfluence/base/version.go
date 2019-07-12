package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 1, 30}
	gitVersion = &version.GitVersion{"ce0b522acb3d", "git@github.com:upfluence/upfluence-if.git", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
