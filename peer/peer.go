package peer

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/upfluence/pkg/envutil"
	"github.com/upfluence/pkg/peer/version"
)

var (
	GitCommit = ""
	GitBranch = ""
	GitRemote = ""

	Version = "v0.0.0"
)

func buildGitVersion() version.GitVersion {
	commit := envutil.FetchString("GIT_COMMIT", GitCommit)

	if commit == "" {
		return version.GitVersion{}
	}

	return version.GitVersion{
		Commit: commit,
		Remote: envutil.FetchString("GIT_REMOTE", GitRemote),
		Branch: envutil.FetchString("GIT_BRANCH", GitBranch),
	}
}

type Peer struct {
	Authority    string
	InstanceName string
	AppName      string
	ProjectName  string
	Environment  string

	Version version.Version
}

func ParsePeerURL(u string) (*Peer, error) {
	uu, err := url.Parse(u)

	if err != nil {
		return nil, err
	}

	if uu.Scheme != "peer" {
		return nil, fmt.Errorf("invalid scheme: %q", uu.Scheme)
	}

	unitName := strings.TrimPrefix(uu.Path, "/")

	p := Peer{
		InstanceName: unitName,
		Authority:    uu.Host,
		AppName:      unitName,
		ProjectName:  unitName,
	}

	if uu.User != nil {
		p.Environment = uu.User.Username()
	}

	vs := uu.Query()

	if v := vs.Get("app-name"); v != "" {
		p.AppName = v
		p.ProjectName = v
	}

	if v := vs.Get("project-name"); v != "" {
		p.ProjectName = v
	}

	if v := vs.Get("semantic-version"); v != "" {
		p.Version.Semantic = version.ParseSemanticVersion(v)
	}

	if v := vs.Get("git-version"); v != "" {
		p.Version.Git.Commit = v
	}

	return &p, nil
}

func (p *Peer) URL() *url.URL {
	vs := url.Values{}

	if p.InstanceName != p.AppName {
		vs.Add("app-name", p.AppName)
	}

	if p.AppName != p.ProjectName {
		vs.Add("project-name", p.ProjectName)
	}

	if p.Version.Semantic.Valid() {
		vs.Add("semantic-version", p.Version.Semantic.String())
	}

	if p.Version.Git.Valid() {
		vs.Add("git-version", p.Version.Git.Commit)
	}

	var u *url.Userinfo

	if p.Environment != "" {
		u = url.User(p.Environment)
	}

	return &url.URL{
		Scheme:   "peer",
		Host:     p.Authority,
		User:     u,
		Path:     p.InstanceName,
		RawQuery: vs.Encode(),
	}
}

func FromEnv() *Peer {
	sv := version.ParseSemanticVersion(envutil.FetchString("VERSION", Version))

	return &Peer{
		Authority:    envutil.FetchString("AUTHORITY", "local"),
		InstanceName: envutil.FetchString("UNIT_NAME", "unknow-service"),
		AppName:      envutil.FetchString("APP_NAME", "unknown-app"),
		ProjectName:  envutil.FetchString("PROJECT_NAME", "unknown-app"),
		Environment:  envutil.FetchString("ENV", "development"),
		Version:      version.Version{Git: buildGitVersion(), Semantic: sv},
	}
}
