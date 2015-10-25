package api

import (
	"github.com/hashicorp/terraform/terraform"
)

type ProviderInfoMessage struct {
	Resources []ResourceInfoMessage `json:"resources,omitempty"`
}

func NewProviderInfoMessage(provider terraform.ResourceProvider) *ProviderInfoMessage {
	ret := &ProviderInfoMessage{}
	resources := provider.Resources()
	for _, resource := range resources {
		ret.Resources = append(ret.Resources, *NewResourceInfoMessage(&resource))
	}
	return ret
}
