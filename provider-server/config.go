package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/hashicorp/hcl"
	hclobj "github.com/hashicorp/hcl/hcl"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	tfconfig "github.com/hashicorp/terraform/config"
)

type Config struct {
	Providers map[string]ProviderConfig
}

type ProviderConfig struct {
	Name      string
	Command   string                 `hcl:"command"`
	rawConfig map[string]interface{} `hcl:"config"`
}

func LoadConfigFromFile(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	obj, err := hcl.Parse(string(bytes))
	if err != nil {
		return nil, err
	}

	return makeConfig(obj)
}

func makeConfig(obj *hclobj.Object) (*Config, error) {
	config := &Config{
		Providers: map[string]ProviderConfig{},
	}

	if providers := obj.Get("provider", false); providers != nil {
		var objects []*hclobj.Object
		for _, o1 := range providers.Elem(false) {
			for _, o2 := range o1.Elem(true) {
				objects = append(objects, o2)
			}
		}

		for _, o := range objects {
			pConfig := ProviderConfig{}
			hcl.DecodeObject(&pConfig, o)
			pConfig.Name = o.Key

			if _, ok := config.Providers[o.Key]; ok {
				return nil, fmt.Errorf("duplicate definition of provider %s", o.Key)
			}

			config.Providers[o.Key] = pConfig
		}
	}

	return config, nil
}

func (c *Config) ProviderFactories() map[string]terraform.ResourceProviderFactory {
	result := make(map[string]terraform.ResourceProviderFactory)
	for k, v := range c.Providers {
		result[k] = v.Factory()
	}
	return result
}

func (pc *ProviderConfig) Factory() terraform.ResourceProviderFactory {
	var pluginConfig plugin.ClientConfig
	pluginConfig.Cmd = exec.Command(pc.Command)
	pluginConfig.Managed = true
	client := plugin.NewClient(&pluginConfig)

	return func() (terraform.ResourceProvider, error) {
		rpcClient, err := client.Client()
		if err != nil {
			return nil, err
		}
		return rpcClient.ResourceProvider()
	}
}

func (c *ProviderConfig) RawConfig() (*tfconfig.RawConfig, error) {
	return tfconfig.NewRawConfig(c.rawConfig)
}
