package handler

import "github.com/upfluence/base/base_service"

const (
	BASE_EXCHANGE_NAME = "monitoring-exchange"
)

type Base struct {
	UnitName  string
	Version   string
	SpawnDate int64
}

func (h *Base) GetName() (string, error) {
	return h.UnitName, nil
}

func (h *Base) GetVersion() (string, error) {
	return h.Version, nil
}

func (h *Base) GetStatus() (base_service.Status, error) {
	return base_service.Status_ALIVE, nil
}

func (h *Base) AliveSince() (int64, error) {
	return h.SpawnDate, nil
}
