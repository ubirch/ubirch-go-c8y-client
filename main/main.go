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

	client, err := c8y.GetClient(UUID, c8yTenant, c8yBootstrap, "")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer client.Disconnect(0)

	switchName1 := UUID + "-A"
	switchName2 := UUID + "-B"
	var value1 byte = 1
	var value2 byte = 1
	for {
		now := time.Now().UTC()
		err = c8y.Send(client, switchName1, value1, now)
		if err != nil {
			log.Printf("error: %v", err)
		}
		err = c8y.Send(client, switchName2, value2, now)
		if err != nil {
			log.Printf("error: %v", err)
		}
		value1 ^= 1
		if value1 == 1 {
			value2 ^= 1
		}
		time.Sleep(30 * time.Second)
	}
}
