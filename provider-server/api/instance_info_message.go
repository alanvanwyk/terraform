package api

import (
	"github.com/hashicorp/terraform/terraform"
)

type InstanceInfoMessage struct {
	Id         string   `json:"id"`
	ModulePath []string `json:"modulePath"`
	Type       string   `json:"resourceName"`
}

func (i *InstanceInfoMessage) InstanceInfo() *terraform.InstanceInfo {
	return &terraform.InstanceInfo{
		Id:         i.Id,
		ModulePath: i.ModulePath,
		Type:       i.Type,
	}
}
