package api

import (
	"encoding/json"
	"net/http"
)

type BadRequestResponse struct {
	Error string
}

func (r *BadRequestResponse) WriteResponse(resp http.ResponseWriter) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(400)
	bytes, err := json.Marshal(r)
	if err != nil {
		// Should never happen
		panic(err)
	}
	resp.Write(bytes)
}
