package c8y

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func C8yBootstrap(tenant string, password string) (string, error) {
	// configure MQTT client
	opts := MQTT.NewClientOptions().AddBroker("tcps://" + tenant + ".cumulocity.com:8883") // scheme://host:port
	opts.SetUsername("management/devicebootstrap")
	opts.SetUsername(password)
	opts.SetClientID(fmt.Sprintf("%s-c8y-mqtt-client-%d", tenant, rand.Uint32()))

	c8yError := make(chan error)
	c8yAuth := make(chan string)
	var receive MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		authorization := string(msg.Payload())
		log.Println("received authorization: " + authorization)
		if strings.HasPrefix(authorization, "70") {
			c8yAuth <- authorization
		}
		c8yError <- errors.New(fmt.Sprintf("unknown message received: %v", msg))
	}

	// configure OnConnect callback: subscribe
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("connected")
		if token := c.Subscribe("s/dcr", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		c.Publish("s/ucr", 0, false, nil)
	}

	client := MQTT.NewClient(opts)
	log.Println("connecting...")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return "", token.Error()
	}

	defer client.Disconnect(0)
	select {
	case auth := <-c8yAuth:
		return auth, nil
	case err := <-c8yError:
		return "", err
	}
}
