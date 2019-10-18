package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// configuration of the client
type Config struct {
	Tenant   string `json:"tenant"`
	Password string `json:"password"`
}

func (c *Config) Load(filename string) {
	// read the configuration file
	contextBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("ERROR: unable to read configuration: %v", err)
	}

	// parse json configuration
	err = json.Unmarshal(contextBytes, c)
	if err != nil {
		log.Fatalf("ERROR: unable to parse configuration: %v", err)
	}

	log.Printf("configuration found")
}
