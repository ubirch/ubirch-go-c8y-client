package main

import (
	"github.com/ubirch/ubirch-go-c8y-client/c8y"
	"log"
	"time"
)

const ConfigFile = "config.json"

func main() {
	log.Println("UBIRCH Golang Cumulocity client")

	// read configuration
	conf := Config{}
	conf.Load(ConfigFile)

	UUID := conf.UUID
	c8yTenant := conf.Tenant
	c8yBootstrap := conf.Bootstrap

	log.Println("UUID: " + UUID)

	client, err := c8y.GetClient(UUID, c8yTenant, c8yBootstrap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer client.Disconnect(0)

	value := true
	for {
		err = c8y.Send(client, value)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		value = !value
		time.Sleep(2 * time.Second)
	}
}
