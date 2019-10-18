package ubirch_go_c8y_client

import (
	"fmt"
	"github.com/google/uuid"
	"log"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func C8yBootstrap(tenant string, password string, handler MQTT.MessageHandler) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(tenant + ".cumulocity.com:8883") // scheme://host:port
	opts.SetUsername("management/devicebootstrap")
	opts.SetUsername(password)
	opts.SetClientID(tenant + "-c8y-mqtt-client")
	if handler != nil {
		opts.SetDefaultPublishHandler(handler)
	}
	topic := "s/dcr"

	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	} else {
		log.Printf("connected to mqtt server: %v", client)
	}

	return client, nil
}
