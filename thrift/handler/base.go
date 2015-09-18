package handler

import "github.com/upfluence/base/base_service"

type Handler struct {
	UnitName  string
	Version   string
	SpawnDate int64
}

func (h *Handler) GetName() (string, error) {
	return h.UnitName, nil
}

func (h *Handler) GetVersion() (string, error) {
	return h.Version, nil
}

func (h *Handler) GetStatus() (base_service.Status, error) {
	return base_service.Status_ALIVE, nil
}

func (h *Handler) AliveSince() (int64, error) {
	return h.SpawnDate, nil
}
