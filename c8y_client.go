package ubirch_go_c8y_client

import (
	"fmt"
	"github.com/google/uuid"
	"log"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func C8yBootstrap(tenant string, password string, handler MQTT.MessageHandler) (string, error) {
	// configure MQTT client
	opts := MQTT.NewClientOptions().AddBroker("tcp://" + tenant + ".cumulocity.com:8883") // scheme://host:port
	opts.SetUsername("management/devicebootstrap")
	opts.SetUsername(password)
	opts.SetClientID(tenant + "-c8y-mqtt-client")
	if handler != nil {
		opts.SetDefaultPublishHandler(handler)
	}

	// configure OnConnect callback: subscribe
	subTopic := "s/dcr"
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(subTopic, 0, handler); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	// configure OnMessage callback -> if message.payload starts with b'70': authorization = message.payload -> publish
	c8yAuth := "TODO"

	// enable SSL/TLS support

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	} else {
		log.Printf("connected to mqtt server: %v", client)
	}

	// wait for answer (OnMessage callback)

	// publish
	pubTopic := "s/ucr"
	client.Publish(pubTopic, 0, false, nil) // TODO check if that's right

	client.Disconnect(999)

	return c8yAuth, nil
}
