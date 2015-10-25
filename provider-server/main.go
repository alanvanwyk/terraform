package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-filename>\n", os.Args[0])
		os.Exit(1)
	}
	configFilename := os.Args[1]
	config, err := LoadConfigFromFile(configFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %s\n", err)
		os.Exit(2)
	}

	log.Println("Config is", config)

	err = RunServer(config)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		os.Exit(3)
	}
}
