package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Request interface {
	UnmarshalRequest(req *http.Request) error
}

type Response interface {
	WriteResponse(resp http.ResponseWriter)
}

func unmarshalRequestBody(req *http.Request, target interface{}) error {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, target)
}

func writeResponse(resp http.ResponseWriter, status int, data interface{}) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)
	bytes, err := json.Marshal(data)
	if err != nil {
		// Should never happen
		panic(err)
	}
	resp.Write(bytes)
}
