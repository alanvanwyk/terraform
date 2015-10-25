package api

import (
	"net/http"
)

type ValidateRequest struct {
	ResourceName string        `json:"resourceName"`
	Config       ConfigMessage `json:"config"`
}

func (r *ValidateRequest) UnmarshalRequest(req *http.Request) error {
	return unmarshalRequestBody(req, r)
}

type ValidateResponse struct {
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

func (r *ValidateResponse) WriteResponse(resp http.ResponseWriter) {
	writeResponse(resp, 200, r)
}
