package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	tfconfig "github.com/hashicorp/terraform/config"
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

type resourceInfoResponse struct {
	Name string `json:"name"`
}

type providerResponse struct {
	Resources []resourceInfoResponse `json:"resources,omitempty"`
}

type indexResponse struct {
	Providers map[string]providerResponse `json:"providers"`
}

type instanceInfoRequest struct {
	Id         string   `json:"id"`
	ModulePath []string `json:"modulePath"`
	Type       string   `json:"resourceName"`
}

func (i *instanceInfoRequest) InstanceInfo() *terraform.InstanceInfo {
	return &terraform.InstanceInfo{
		Id:         i.Id,
		ModulePath: i.ModulePath,
		Type:       i.Type,
	}
}

type instanceEphemeralMessage struct {
	ConnInfo map[string]string `json:"connectionInfo"`
}

func (i *instanceEphemeralMessage) EphemeralState() terraform.EphemeralState {
	return terraform.EphemeralState{
		ConnInfo: i.ConnInfo,
	}
}

type instanceStateMessage struct {
	ID         string                   `json:"id"`
	Attributes map[string]string        `json:"attributes"`
	Ephemeral  instanceEphemeralMessage `json:"ephemeral"`
	Meta       map[string]string        `json:"meta"`
}

func newInstanceStateMessage(s *terraform.InstanceState) *instanceStateMessage {
	return &instanceStateMessage{
		ID:         s.ID,
		Attributes: s.Attributes,
		Ephemeral: instanceEphemeralMessage{
			ConnInfo: s.Ephemeral.ConnInfo,
		},
		Meta: s.Meta,
	}
}

func (i *instanceStateMessage) InstanceState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID:         i.ID,
		Attributes: i.Attributes,
		//Ephemeral: i.Ephemeral.EphemeralState(),
		Meta: i.Meta,
	}
}

type diffMessage struct {
	Attributes     map[string]attrDiffMessage `json:"attributes"`
	Destroy        bool                       `json:"destroy"`
	DestroyTainted bool                       `json:"destroyTainted"`
}

type attrDiffMessage struct {
	Old         string                 `json:"oldValue"`
	New         string                 `json:"newValue"`
	NewComputed bool                   `json:"newIsComputed"`
	NewRemoved  bool                   `json:"newIsRemoved"`
	NewExtra    interface{}            `json:"newExtra,omitempty"`
	RequiresNew bool                   `json:"requiresNew"`
	Type        terraform.DiffAttrType `json:"type"`
}

type validateRequest struct {
	ResourceName string                 `json:"resourceName"`
	Config       map[string]interface{} `json:"config"`
}

type validateResponse struct {
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

type refreshRequest struct {
	InstanceInfo instanceInfoRequest  `json:"instanceInfo"`
	CurrentState instanceStateMessage `json:"currentState"`
}

type refreshResponse struct {
	NewState *instanceStateMessage `json:"newState,omitempty"`
	Error    string                `json:"error,omitempty"`
}

type diffRequest struct {
	InstanceInfo instanceInfoRequest    `json:"instanceInfo"`
	CurrentState instanceStateMessage   `json:"currentState"`
	NewConfig    map[string]interface{} `json:"newConfig"`
}

type diffResponse struct {
	Diff  *diffMessage `json:"diff,omitempty"`
	Error string       `json:"error,omitempty"`
}

type applyRequest struct {
	InstanceInfo instanceInfoRequest  `json:"instanceInfo"`
	CurrentState instanceStateMessage `json:"currentState"`
	Diff         diffMessage          `json:"diff"`
}

type applyResponse struct {
	NewState *instanceStateMessage `json:"newState,omitempty"`
	Error    string                `json:"error,omitempty"`
}

func (s *APIServer) Index(resp http.ResponseWriter, req *http.Request) {
	ret := &indexResponse{
		Providers: map[string]providerResponse{},
	}

	for name, provider := range s.providers {
		presp := providerResponse{}
		resources := provider.Resources()
		for _, resource := range resources {
			presp.Resources = append(presp.Resources, resourceInfoResponse{
				Name: resource.Name,
			})
		}
		ret.Providers[name] = presp
	}

	buf, _ := json.Marshal(ret)
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(buf)
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

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(500)
		return
	}

	data := &validateRequest{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		resp.WriteHeader(400)
	}

	rawConfig, err := tfconfig.NewRawConfig(data.Config)
	if err != nil {
		respData := &validateResponse{
			Errors: []string{err.Error()},
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(400)
		respRawData, _ := json.Marshal(respData)
		resp.Write(respRawData)
		return
	}
	config := terraform.NewResourceConfig(rawConfig)

	log.Printf("[DEBUG] Validating config for %s resource %s", pa.name, data.ResourceName)
	warns, errs := pa.provider.ValidateResource(data.ResourceName, config)
	var errsStr []string
	if len(errs) > 0 {
		errsStr = make([]string, len(errs))
		for i, err := range errs {
			errsStr[i] = err.Error()
		}
	}
	respData := &validateResponse{
		Warnings: warns,
		Errors:   errsStr,
	}
	resp.Header().Set("Content-Type", "application/json")
	if len(errs) > 0 {
		resp.WriteHeader(400)
	}
	respRawData, _ := json.Marshal(respData)
	resp.Write(respRawData)
}

func (pa *ProviderAPI) Diff(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(500)
		return
	}

	data := &diffRequest{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		resp.WriteHeader(400)
	}

	iinfo := data.InstanceInfo.InstanceInfo()
	currentState := data.CurrentState.InstanceState()

	rawConfig, err := tfconfig.NewRawConfig(data.NewConfig)
	if err != nil {
		respData := &diffResponse{
			Error: err.Error(),
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(400)
		respRawData, _ := json.Marshal(respData)
		resp.Write(respRawData)
		return
	}
	config := terraform.NewResourceConfig(rawConfig)

	// Validate the config just to make sure, since Diff implementations
	// aren't robust against invalid data.
	_, errs := pa.provider.ValidateResource(data.InstanceInfo.Type, config)
	if len(errs) > 0 {
		respData := &diffResponse{
			Error: errs[0].Error(),
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(400)
		respRawData, _ := json.Marshal(respData)
		resp.Write(respRawData)
		return
	}

	log.Printf("[DEBUG] Diffing %s resource %s", pa.name, data.InstanceInfo.Type)
	diff, err := pa.provider.Diff(iinfo, currentState, config)
	log.Printf("[DEBUG] result is", diff)
	resp.Header().Set("Content-Type", "application/json")
	respData := &diffResponse{}
	if err != nil {
		resp.WriteHeader(500)
		respData.Error = err.Error()
	} else {
		if diff != nil {
			diffMsg := &diffMessage{}
			diffMsg.Destroy = diff.Destroy
			diffMsg.DestroyTainted = diff.DestroyTainted
			if diff.Attributes != nil {
				diffMsg.Attributes = map[string]attrDiffMessage{}
				for k, ad := range diff.Attributes {
					diffMsg.Attributes[k] = attrDiffMessage{
						Old:         ad.Old,
						New:         ad.New,
						NewComputed: ad.NewComputed,
						NewRemoved:  ad.NewRemoved,
						NewExtra:    ad.NewExtra,
						RequiresNew: ad.RequiresNew,
						Type:        ad.Type,
					}
				}
			}
			respData.Diff = diffMsg
		}
	}
	respRawData, _ := json.Marshal(respData)
	resp.Write(respRawData)
}

func (pa *ProviderAPI) Refresh(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(500)
		return
	}

	data := &refreshRequest{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		resp.WriteHeader(400)
	}

	iinfo := data.InstanceInfo.InstanceInfo()
	currentState := data.CurrentState.InstanceState()

	log.Printf("[DEBUG] Refreshing %s resource %s", pa.name, data.InstanceInfo.Type)
	newState, err := pa.provider.Refresh(iinfo, currentState)
	resp.Header().Set("Content-Type", "application/json")
	respData := &refreshResponse{}
	if err != nil {
		resp.WriteHeader(500)
		respData.Error = err.Error()
	} else {
		if newState != nil {
			respData.NewState = newInstanceStateMessage(newState)
		}
	}
	respRawData, _ := json.Marshal(respData)
	resp.Write(respRawData)
}

func (pa *ProviderAPI) Apply(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}

	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(500)
		return
	}

	data := &applyRequest{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		resp.WriteHeader(400)
	}

	iinfo := data.InstanceInfo.InstanceInfo()
	currentState := data.CurrentState.InstanceState()
	diffMsg := data.Diff
	diff := &terraform.InstanceDiff{
		Attributes:     map[string]*terraform.ResourceAttrDiff{},
		Destroy:        diffMsg.Destroy,
		DestroyTainted: diffMsg.DestroyTainted,
	}
	for k, ad := range diffMsg.Attributes {
		diff.Attributes[k] = &terraform.ResourceAttrDiff{
			Old:         ad.Old,
			New:         ad.New,
			NewComputed: ad.NewComputed,
			NewRemoved:  ad.NewRemoved,
			NewExtra:    ad.NewExtra,
			RequiresNew: ad.RequiresNew,
			Type:        ad.Type,
		}
	}

	log.Printf("[DEBUG] Applying %s resource %s", pa.name, data.InstanceInfo.Type)
	newState, err := pa.provider.Apply(iinfo, currentState, diff)
	log.Printf("[DEBUG] result is", diff)
	resp.Header().Set("Content-Type", "application/json")
	respData := &applyResponse{}
	if err != nil {
		resp.WriteHeader(500)
		respData.Error = err.Error()
	} else {
		if newState != nil {
			respData.NewState = newInstanceStateMessage(newState)
		}
	}
	respRawData, _ := json.Marshal(respData)
	resp.Write(respRawData)
}
