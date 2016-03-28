package base

import "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/version"

var (
	semVersion  = &version.SemanticVersion{0, 0, 11}
	gitVersion  = &version.GitVersion{"9a6a5f55db41", "https://github.com/upfluence/upfluence-if", "master"}
	baseVersion = &version.Version{semVersion, gitVersion}
)

type Base struct{}

func (p *Base) Version() *version.Version {
	return baseVersion
}
