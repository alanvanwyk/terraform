package api

import (
	"net/http"
)

type RefreshRequest struct {
	InstanceInfo *InstanceInfoMessage  `json:"instanceInfo"`
	CurrentState *InstanceStateMessage `json:"currentState"`
}

func (r *RefreshRequest) UnmarshalRequest(req *http.Request) error {
	return unmarshalRequestBody(req, r)
}

type RefreshResponse struct {
	NewState *InstanceStateMessage `json:"newState,omitempty"`
	Error    string                `json:"error,omitempty"`
}

func (r *RefreshResponse) WriteResponse(resp http.ResponseWriter) {
	status := 200
	if r.Error != "" {
		status = 500
	}
	writeResponse(resp, status, r)
}
