package base

import "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/version"

var (
	semVersion  = &version.SemanticVersion{0, 0, 19}
	gitVersion  = &version.GitVersion{"86d62e2b8121", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
)

type Base struct{}

func (p *Base) Name() string { return "base" }

func (p *Base) Version() *version.Version {
	return baseVersion
}
