package baseutil

import (
	"fmt"

	"github.com/upfluence/base/base_service"
	"github.com/upfluence/base/version"

	"github.com/upfluence/pkg/log"
)

const defaultVersion = "v0.0.0-dirty"

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

func IntrospectService(h base_service.BaseService) {
	var (
		version, _ = h.GetVersion()
		name, _    = h.GetName()
		ifaces, _  = h.GetInterfaceVersions()
	)

	log.Noticef(
		"Service %s %s",
		name,
		SerializeVersion(version),
	)

	if len(ifaces) > 0 {
		log.Noticef("Interface versions:")

		for n, v := range ifaces {
			log.Noticef("* %s %s", n, SerializeVersion(v))
		}
	}
}
