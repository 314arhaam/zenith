package data

type AddRequest struct {
	ServiceName string `json:"service_name"`
}

func (r *AddRequest) Validate() bool {
	name, ok := NormalizeServiceName(r.ServiceName)
	if ok {
		r.ServiceName = name
	}
	return ok
}
