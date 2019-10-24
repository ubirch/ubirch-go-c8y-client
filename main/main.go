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

	switchName1 := "switch1"
	var value byte = 1
	for {
		now := time.Now().UTC()
		err = c8y.Send(client, switchName1, value, now)
		if err != nil {
			log.Printf("error: %v", err)
		}
		value ^= 1
		time.Sleep(5 * time.Second)
	}
}
