package api

import (
	"github.com/hashicorp/terraform/terraform"
)

type ResourceInfoMessage struct {
	Name string `json:"name"`
}

func NewResourceInfoMessage(resource *terraform.ResourceType) *ResourceInfoMessage {
	return &ResourceInfoMessage{
		Name: resource.Name,
	}
}
