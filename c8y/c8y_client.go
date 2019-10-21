package c8y

import (
	"errors"
	"fmt"
	"log"
	"strings"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func Bootstrap(tenant string, uuid string, password string) (string, error) {
	// configure MQTT client
	address := "tcps://" + tenant + ".cumulocity.com:8883/"
	log.Println(address)
	opts := MQTT.NewClientOptions().AddBroker(address) // scheme://host:port
	opts.SetUsername("management/devicebootstrap")
	opts.SetPassword(password)
	opts.SetClientID(uuid)
	c8yError := make(chan error)
	c8yAuth := make(chan string)
	var receive MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		authorization := string(msg.Payload())
		if strings.HasPrefix(authorization, "70") {
			c8yAuth <- authorization
		}
		c8yError <- errors.New(fmt.Sprintf("unknown message received: %v", msg))
	}

	// configure OnConnect callback: subscribe
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe("s/dcr", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		c.Publish("s/ucr", 0, false, nil)
	}

	client := MQTT.NewClient(opts)
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

func Send(tenant string, uuid string, user string, password string) error {
	address := "tcps://" + tenant + ".cumulocity.com:8883/"
	opts := MQTT.NewClientOptions().AddBroker(address) // scheme://host:port
	opts.SetUsername(tenant + "/" + user)
	opts.SetPassword(password)
	opts.SetClientID(uuid)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	token := client.Publish("s/us", 0, false, "211,25")
	token.Wait()
	return token.Error()

}
