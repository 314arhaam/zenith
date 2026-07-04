package handlefuncs

import data "zenith/models"

type Handler struct {
	Core data.ServiceData
}

func NewHandler() Handler {
	return Handler{
		Core: data.NewServiceData(),
	}
}
