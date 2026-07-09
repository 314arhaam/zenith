package data

type AddRequest struct {
	ServiceName string `json:"service_name"`
}

func (r *AddRequest) Validate() bool {
	if r.ServiceName == "" {
		return false
	}
	return true
}
