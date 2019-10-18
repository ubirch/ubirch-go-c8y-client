package main

import (
	// "github.com/google/uuid"
	"github.com/ubirch/ubirch-go-c8y-client"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func main() {
	tenant := "ubirch"
	c8yPassword := "xyz"
	handler := MQTT.MessageHandler{}
	c8yAuth, err := ubirch_go_c8y_client.C8yBootstrap(tenant, c8yPassword, handler)
}
