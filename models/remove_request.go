package data

type RemoveRequest struct {
	ServiceName string `json:"service_name"`
}

func (r *RemoveRequest) Validate() bool {
	name, ok := NormalizeServiceName(r.ServiceName)
	if ok {
		r.ServiceName = name
	}
	return ok
}
