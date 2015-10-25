package api

import (
	"net/http"
)

type IndexResponse struct {
	Providers map[string]ProviderInfoMessage `json:"providers"`
}

func (r *IndexResponse) WriteResponse(resp http.ResponseWriter) {
	writeResponse(resp, 200, r)
}
