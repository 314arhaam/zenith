package data

type RemoveRequest struct {
	ServiceName string `json:"service_name"`
}

func (r *RemoveRequest) Validate() bool {
	if r.ServiceName == "" {
		return false
	}
	return true
}
