package main

import (
	"github.com/ubirch/ubirch-go-c8y-client/c8y"
	"log"
)

const ConfigFile = "config.json"

func main() {
	log.Println("UBIRCH Golang Cumulocity client")

	// read configuration
	conf := Config{}
	conf.Load(ConfigFile)

	err := c8y.Send(conf.Tenant, conf.User, conf.Password)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// bootstrap
	c8yAuth, err := c8y.C8yBootstrap(conf.Tenant, conf.Bootstrap)
	if err != nil {
		log.Printf("tenant: %s, password: %s\n", conf.Tenant, conf.Password)
		log.Fatalf("unable to bootstrap device: %v", err)
	}
	log.Printf(c8yAuth)
}
