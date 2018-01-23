package peer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/upfluence/base"
	"github.com/upfluence/base/version"

	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/log"
)

const defaultVersion = "v0.0.0-dirty"

var (
	GitCommit = ""
	GitBranch = ""
	GitRemote = ""

	Version = "v0.0.0"
)

func parseSemanticVersion(v string) *version.SemanticVersion {
	splittedVersion := strings.Split(v, ".")

	if len(splittedVersion) != 3 {
		return nil
	}

	major, _ := strconv.Atoi(splittedVersion[0][1:])
	minor, _ := strconv.Atoi(splittedVersion[1])
	patch, _ := strconv.Atoi(splittedVersion[2])

	return &version.SemanticVersion{
		Major: int16(major),
		Minor: int16(minor),
		Patch: int16(patch),
	}
}

func buildGitVersion() *version.GitVersion {
	commit := cfg.FetchString("GIT_COMMIT", GitCommit)

	if commit == "" {
		return nil
	}

	return &version.GitVersion{
		Commit: commit,
		Remote: cfg.FetchString("GIT_REMOTE", GitRemote),
		Branch: cfg.FetchString("GIT_BRANCH", GitBranch),
	}
}

func SerializeVersion(v *version.Version) string {
	if v == nil {
		return defaultVersion
	}

	if sv := v.SemanticVersion; sv != nil {
		return fmt.Sprintf("v%d.%d.%d", sv.Major, sv.Minor, sv.Patch)
	}

	if gv := v.GitVersion; gv != nil {
		return fmt.Sprintf("v.0.0.0+git-%s", gv.Commit[:7])
	}

	return defaultVersion
}

type Interface interface {
	Version() *version.Version
	Name() string
}

type Peer struct {
	InstanceName, AppName, ProjectName string
	Version                            *version.Version
	Interfaces                         []Interface
}

func FromEnv() *Peer {
	return &Peer{
		InstanceName: cfg.FetchString("UNIT_NAME", "unknow-service"),
		AppName:      cfg.FetchString("APP_NAME", ""),
		ProjectName:  cfg.FetchString("PROJECT_NAME", ""),
		Version: &version.Version{
			GitVersion: buildGitVersion(),
			SemanticVersion: parseSemanticVersion(
				cfg.FetchString("VERSION", Version),
			),
		},
		Interfaces: []Interface{&base.Base{}},
	}
}

func (p *Peer) Introspect() {
	log.Noticef("Service %s %s", p.InstanceName, SerializeVersion(p.Version))

	if len(p.Interfaces) > 0 {
		log.Noticef("Interface versions:")

		for _, iface := range p.Interfaces {
			log.Noticef("* %s %s", iface.Name(), SerializeVersion(iface.Version()))
		}
	}
}
