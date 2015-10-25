package api

import (
	"net/http"
)

type DiffRequest struct {
	InstanceInfo InstanceInfoMessage  `json:"instanceInfo"`
	CurrentState InstanceStateMessage `json:"currentState"`
	NewConfig    ConfigMessage        `json:"newConfig"`
}

func (r *DiffRequest) UnmarshalRequest(req *http.Request) error {
	return unmarshalRequestBody(req, r)
}

type DiffResponse struct {
	Diff  *DiffMessage `json:"diff,omitempty"`
	Error string       `json:"error,omitempty"`
}

func (r *DiffResponse) WriteResponse(resp http.ResponseWriter) {
	status := 200
	if r.Error != "" {
		status = 500
	}
	writeResponse(resp, status, r)
}
