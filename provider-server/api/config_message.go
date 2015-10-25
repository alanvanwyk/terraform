package api

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
)

type ConfigMessage map[string]interface{}

func (m ConfigMessage) ResourceConfig() (*terraform.ResourceConfig, error) {
	rawConfig, err := config.NewRawConfig(map[string]interface{}(m))
	if err != nil {
		return nil, err
	}
	return terraform.NewResourceConfig(rawConfig), nil
}

// ValidResourceConfig converts the config emssage into a ResourceConfig
// and asks the given provider to validate it. The config will be returned
// only if the provider says it is valid.
func (m ConfigMessage) ValidResourceConfig(provider terraform.ResourceProvider, resourceName string) (*terraform.ResourceConfig, error) {
	config, err := m.ResourceConfig()
	if err != nil {
		return nil, err
	}

	_, errs := provider.ValidateResource(resourceName, config)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return config, nil
}
