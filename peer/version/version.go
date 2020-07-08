package version

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultVersion = "v0.0.0-dirty"

func ParseSemanticVersion(v string) SemanticVersion {
	var (
		suffix string

		vs = strings.SplitN(v, "-", 2)
	)

	if len(vs) == 2 {
		suffix = vs[1]
	}

	splittedVersion := strings.Split(vs[0], ".")

	if len(splittedVersion) != 3 {
		return SemanticVersion{}
	}

	major, _ := strconv.Atoi(splittedVersion[0][1:])
	minor, _ := strconv.Atoi(splittedVersion[1])
	patch, _ := strconv.Atoi(splittedVersion[2])

	return SemanticVersion{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Suffix: suffix,
	}
}

type GitVersion struct {
	Commit string
	Remote string
	Branch string
}

func (gv GitVersion) String() string {
	return fmt.Sprintf("v.0.0.0+git-%s", gv.Commit[:7])
}

type SemanticVersion struct {
	Major int
	Minor int
	Patch int

	Suffix string
}

func (sv SemanticVersion) String() string {
	v := fmt.Sprintf("v%d.%d.%d", sv.Major, sv.Minor, sv.Patch)

	if sv.Suffix != "" {
		v += "-" + sv.Suffix
	}

	return v
}

type Version struct {
	Semantic SemanticVersion
	Git      GitVersion
}

func (v Version) String() string {
	if v.Semantic != (SemanticVersion{}) {
		return v.Semantic.String()
	}

	if v.Git != (GitVersion{}) {
		return v.Git.String()
	}

	return defaultVersion
}
