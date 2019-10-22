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

	UUID := conf.UUID
	//c8yBootstrap := conf.Bootstrap
	c8yTenant := conf.Tenant
	c8yUser := conf.User
	c8yPassword := conf.Password

	println("UUID: " + UUID)

	//// bootstrap
	//c8yAuth, err := c8y.Bootstrap(c8yTenant, UUID, c8yBootstrap)
	//if err != nil {
	//	log.Printf("tenant: %s, password: %s\n", c8yTenant, c8yBootstrap)
	//	log.Fatalf("unable to bootstrap device: %v", err)
	//}
	//log.Println(c8yAuth)

	done, err := c8y.GetClient(UUID, c8yTenant, c8yUser, c8yPassword)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	println(done)

}
