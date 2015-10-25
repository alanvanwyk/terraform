package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/terraform"
)

func RunServer(config *Config) error {

	providers := map[string]terraform.ResourceProvider{}

	for name, pConfig := range config.Providers {
		log.Println("[DEBUG] Initializing provider", name)
		factory := pConfig.Factory()
		provider, err := factory()
		if err != nil {
			return fmt.Errorf("error initializing %s: %s", name, err)
		}

		pRawConfig, err := pConfig.RawConfig()
		if err != nil {
			return fmt.Errorf("error decoding %s config: %s", name, err)
		}

		pResConfig := terraform.NewResourceConfig(pRawConfig)
		warns, errs := provider.Validate(pResConfig)
		for _, warning := range warns {
			log.Println("[WARNING] %s: %s", name, warning)
		}
		if len(errs) > 0 {
			fmt.Errorf("error in %s config: %s", name, errs[0])
		}

		err = provider.Configure(pResConfig)
		if err != nil {
			fmt.Errorf("error configuring %s: %s", name, err)
		}

		providers[name] = provider
	}

	return startHTTPAPI(providers)
}
