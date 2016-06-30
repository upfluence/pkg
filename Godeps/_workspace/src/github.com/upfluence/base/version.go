package base

import "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/version"

var (
	semVersion  = &version.SemanticVersion{0, 0, 13}
	gitVersion  = &version.GitVersion{"06e06196cb2c", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
)

type Base struct{}

func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
	return baseVersion
}
