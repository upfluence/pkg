package base

import "github.com/upfluence/base/version"




var (
	semVersion = &version.SemanticVersion{0, 1, 51}
	gitVersion = &version.GitVersion{"de1d7e1346329ce9ff10281a1222f4d4641dd323", "git@github.com:upfluence/upfluence-if.git", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
	)


type Base struct {}
func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
  return baseVersion
}
