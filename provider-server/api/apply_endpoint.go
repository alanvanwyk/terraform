package api

import (
	"net/http"
)

type ApplyRequest struct {
	InstanceInfo *InstanceInfoMessage  `json:"instanceInfo"`
	CurrentState *InstanceStateMessage `json:"currentState"`
	Diff         *DiffMessage          `json:"diff"`
}

func (r *ApplyRequest) UnmarshalRequest(req *http.Request) error {
	return unmarshalRequestBody(req, r)
}

type ApplyResponse struct {
	NewState *InstanceStateMessage `json:"newState,omitempty"`
	Error    string                `json:"error,omitempty"`
}

func (r *ApplyResponse) WriteResponse(resp http.ResponseWriter) {
	status := 200
	if r.Error != "" {
		status = 500
	}
	writeResponse(resp, status, r)
}
