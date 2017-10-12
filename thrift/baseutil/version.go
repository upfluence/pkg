package baseutil

import (
	"fmt"

	"github.com/upfluence/base/version"
)

const defaultVersion = "v0.0.0-dirty"

func ToString(v *version.Version) string {
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
