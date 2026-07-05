package handlers

import "zenith/core"

type Handler struct {
	Core core.ServiceData
}

func NewHandler() Handler {
	return Handler{
		Core: core.NewServiceData(),
	}
}
