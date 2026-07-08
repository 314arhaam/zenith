package handlers

import "zenith/core"

type Handler struct {
	Core core.System
}

func NewHandler() Handler {
	return Handler{
		Core: core.NewSystem(),
	}
}
