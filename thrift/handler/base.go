package handler

import (
	"time"

	"github.com/upfluence/base"
	"github.com/upfluence/base/base_service"
	"github.com/upfluence/base/version"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/peer"
)

func NewBaseHandler(p *peer.Peer) *Base {
	return &Base{
		Peer:      p,
		SpawnDate: time.Now().Unix(),
		StatusFn:  func() base_service.Status { return base_service.Status_ALIVE },
	}
}

type Base struct {
	*peer.Peer

	SpawnDate int64
	StatusFn  func() base_service.Status
}

func (h *Base) GetName(thrift.Context) (string, error) {
	return h.InstanceName, nil
}

func (h *Base) GetVersion(thrift.Context) (*version.Version, error) {
	return h.Version, nil
}

func (h *Base) GetInterfaceVersions(thrift.Context) (map[string]*version.Version, error) {
	versions := make(map[string]*version.Version)

	for _, i := range h.Interfaces {
		versions[i.Name()] = i.Version()
	}

	if _, ok := versions["base"]; !ok {
		versions["base"] = (&base.Base{}).Version()
	}

	return versions, nil
}

func (h *Base) GetStatus(thrift.Context) (base_service.Status, error) {
	return h.StatusFn(), nil
}

func (h *Base) AliveSince(thrift.Context) (int64, error) {
	return h.SpawnDate, nil
}
