package c8y

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
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

	answerReceived := false
	// Answer receive-callback
	var authReceive MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		answerReceived = true
		answer := string(msg.Payload())
		if strings.HasPrefix(answer, "70") {
			log.Println("received authorization: " + answer)
			c8yAuth <- answer
		} else {
			log.Println("received unknown message:" + answer)
			c8yError <- errors.New(fmt.Sprintf("unknown message received: %v", msg))
		}
	}
	// Error receive-callback
	var errReceive MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		answerReceived = true
		answer := string(msg.Payload())
		log.Println("received error message:" + answer)
		c8yError <- errors.New(fmt.Sprintf("unknown message received: %v", msg))
	}

	// configure OnConnect callback: subscribe
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("MQTT client connected.")

		// subscribe to authorization message
		if token := c.Subscribe("s/dcr", 0, authReceive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, errReceive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
			return
		}

		// publish until answer received
		for !answerReceived {
			println("publishing...")
			c.Publish("s/ucr", 0, false, nil)
			time.Sleep(10 * time.Second)
		}
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
