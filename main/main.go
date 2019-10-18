package main

import (
	"fmt"
	"github.com/ubirch/ubirch-go-c8y-client/c8y"
)

func main() {
	tenant := "ubirch"
	c8yPassword := "---"
	c8yAuth, err := c8y.C8yBootstrap(tenant, c8yPassword)
	if err != nil {
		panic(fmt.Sprintf("unable to bootstrap device: %v", err))
	}

	fmt.Printf(c8yAuth)
}
