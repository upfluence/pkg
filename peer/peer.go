package peer

import (
	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/peer/version"
)

var (
	GitCommit = ""
	GitBranch = ""
	GitRemote = ""

	Version = "v0.0.0"
)

func buildGitVersion() version.GitVersion {
	commit := cfg.FetchString("GIT_COMMIT", GitCommit)

	if commit == "" {
		return version.GitVersion{}
	}

	return version.GitVersion{
		Commit: commit,
		Remote: cfg.FetchString("GIT_REMOTE", GitRemote),
		Branch: cfg.FetchString("GIT_BRANCH", GitBranch),
	}
}

type Peer struct {
	InstanceName string
	AppName      string
	ProjectName  string
	Environment  string

	Version version.Version
}

func FromEnv() *Peer {
	sv := version.ParseSemanticVersion(cfg.FetchString("VERSION", Version))

	return &Peer{
		InstanceName: cfg.FetchString("UNIT_NAME", "unknow-service"),
		AppName:      cfg.FetchString("APP_NAME", "unknown-app"),
		ProjectName:  cfg.FetchString("PROJECT_NAME", "unknown-project"),
		Environment:  cfg.FetchString("ENV", "development"),
		Version:      version.Version{Git: buildGitVersion(), Semantic: sv},
	}
}
