package handler

import (
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/base_service"
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/base/version"
)

const (
	BASE_EXCHANGE_NAME = "monitoring-exchange"
)

type Interface interface {
	Version() *version.Version
	Name() string
}

type Base struct {
	UnitName   string
	SpawnDate  int64
	Version    *version.Version
	Interfaces []Interface
}

func (h *Base) GetName() (string, error) {
	return h.UnitName, nil
}

func (h *Base) GetVersion() (*version.Version, error) {
	return h.Version, nil
}

func (h *Base) GetInterfaceVersions() (map[string]*version.Version, error) {
	versions := make(map[string]*version.Version)

	for _, i := range h.Interfaces {
		versions[i.Name()] = i.Version()
	}

	if _, ok := versions["base"]; !ok {
		versions["base"] = (&base.Base{}).Version()
	}

	return versions, nil
}

func (h *Base) GetStatus() (base_service.Status, error) {
	return base_service.Status_ALIVE, nil
}

func (h *Base) AliveSince() (int64, error) {
	return h.SpawnDate, nil
}
