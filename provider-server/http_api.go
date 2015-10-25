package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	//tfconfig "github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/provider-server/api"
	"github.com/hashicorp/terraform/terraform"
)

type APIServer struct {
	providers map[string]terraform.ResourceProvider
	listener  net.Listener
	mux       *http.ServeMux
}

type ProviderAPI struct {
	name     string
	provider terraform.ResourceProvider
}

func startHTTPAPI(providers map[string]terraform.ResourceProvider) error {
	// TODO: Make listen port configurable, and support HTTPS
	listenAddr := ":8080"
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %s", listenAddr, err)
	}

	mux := http.NewServeMux()

	srv := &APIServer{
		mux:       mux,
		providers: providers,
		listener:  ln,
	}
	srv.registerHandlers()

	log.Println("[INFO] Listening on", listenAddr)
	http.Serve(ln, mux)

	return err
}

func (s *APIServer) registerHandlers() {
	s.mux.HandleFunc("/", s.Index)

	for name, provider := range s.providers {
		api := &ProviderAPI{name, provider}
		s.mux.HandleFunc(api.OperationPath("validate"), api.Validate)
		s.mux.HandleFunc(api.OperationPath("diff"), api.Diff)
		s.mux.HandleFunc(api.OperationPath("refresh"), api.Refresh)
		s.mux.HandleFunc(api.OperationPath("apply"), api.Apply)
	}
}

func (s *APIServer) Index(resp http.ResponseWriter, req *http.Request) {
	ret := &api.IndexResponse{
		Providers: map[string]api.ProviderInfoMessage{},
	}

	for name, provider := range s.providers {
		ret.Providers[name] = *api.NewProviderInfoMessage(provider)
	}

	ret.WriteResponse(resp)
}

func (pa *ProviderAPI) OperationPath(name string) string {
	if name != "" {
		return fmt.Sprintf("/%s/%s", pa.name, name)
	} else {
		return fmt.Sprintf("/%s/%s", pa.name)
	}
}

func (pa *ProviderAPI) Validate(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	data := &api.ValidateRequest{}
	err := data.UnmarshalRequest(req)
	if err != nil {
		log.Println("[ERROR] Bad Validate request:", err)
		resp.WriteHeader(400)
		return
	}

	config, err := data.Config.ResourceConfig()
	if err != nil {
		respData := &api.ValidateResponse{
			Errors: []string{err.Error()},
		}
		respData.WriteResponse(resp)
		return
	}

	log.Printf("[DEBUG] Validating config for %s resource %s", pa.name, data.ResourceName)
	warns, errs := pa.provider.ValidateResource(data.ResourceName, config)
	var errsStr []string
	if len(errs) > 0 {
		errsStr = make([]string, len(errs))
		for i, err := range errs {
			errsStr[i] = err.Error()
		}
	}
	respData := &api.ValidateResponse{
		Warnings: warns,
		Errors:   errsStr,
	}
	respData.WriteResponse(resp)
}

func (pa *ProviderAPI) Diff(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	data := &api.DiffRequest{}
	err := data.UnmarshalRequest(req)
	if err != nil {
		log.Println("[ERROR] Bad Diff request:", err)
		resp.WriteHeader(400)
		return
	}

	info := data.InstanceInfo.InstanceInfo()
	state := data.CurrentState.InstanceState()
	config, err := data.NewConfig.ValidResourceConfig(pa.provider, data.InstanceInfo.Type)
	if err != nil {
		respData := &api.BadRequestResponse{
			Error: err.Error(),
		}
		respData.WriteResponse(resp)
		return
	}

	log.Printf("[DEBUG] Diffing %s resource %s", pa.name, data.InstanceInfo.Type)
	diff, err := pa.provider.Diff(info, state, config)
	respData := &api.DiffResponse{}
	if err != nil {
		respData.Error = err.Error()
	} else {
		if diff != nil {
			respData.Diff = api.NewDiffMessage(diff)
		}
	}
	respData.WriteResponse(resp)
}

func (pa *ProviderAPI) Refresh(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	data := &api.RefreshRequest{}
	err := data.UnmarshalRequest(req)
	if err != nil {
		log.Println("[ERROR] Bad Refresh request:", err)
		resp.WriteHeader(400)
		return
	}

	info := data.InstanceInfo.InstanceInfo()
	currentState := data.CurrentState.InstanceState()

	log.Printf("[DEBUG] Refreshing %s resource %s", pa.name, data.InstanceInfo.Type)
	newState, err := pa.provider.Refresh(info, currentState)
	respData := &api.RefreshResponse{}
	if err != nil {
		respData.Error = err.Error()
	} else {
		if newState != nil {
			respData.NewState = api.NewInstanceStateMessage(newState)
		}
	}
	respData.WriteResponse(resp)
}

func (pa *ProviderAPI) Apply(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	data := &api.ApplyRequest{}
	err := data.UnmarshalRequest(req)
	if err != nil {
		log.Println("[ERROR] Bad Apply request:", err)
		resp.WriteHeader(400)
		return
	}

	info := data.InstanceInfo.InstanceInfo()
	currentState := data.CurrentState.InstanceState()
	diff := data.Diff.InstanceDiff()

	log.Printf("[DEBUG] Applying %s resource %s", pa.name, data.InstanceInfo.Type)
	newState, err := pa.provider.Apply(info, currentState, diff)
	respData := &api.ApplyResponse{}
	if err != nil {
		respData.Error = err.Error()
	} else {
		if newState != nil {
			respData.NewState = api.NewInstanceStateMessage(newState)
		}
	}
	respData.WriteResponse(resp)
}
