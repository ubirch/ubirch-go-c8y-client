package c8y

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)
import MQTT "github.com/eclipse/paho.mqtt.golang"

func Bootstrap(tenant string, uuid string, password string) (string, error) {
	// configure MQTT client
	address := "tcps://" + tenant + ".cumulocity.com:8883/" // scheme://host:port
	log.Println(address)
	opts := MQTT.NewClientOptions().AddBroker(address)
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
			log.Println("publishing...")
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

func GetClient(uuid string, tenant string, user string, password string) (MQTT.Client, error) {
	address := "tcps://" + tenant + ".cumulocity.com:8883/"
	opts := MQTT.NewClientOptions().AddBroker(address) // scheme://host:port
	opts.SetClientID(uuid)
	opts.SetUsername(tenant + "/" + user)
	opts.SetPassword(password)

	c8yError := make(chan error)
	c8yReady := make(chan bool)

	// receive-callback
	var receive MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
		answer := string(msg.Payload())
		log.Println("received error message:" + answer)
	}

	// configure OnConnect callback: subscribe
	opts.OnConnect = func(c MQTT.Client) {
		log.Println("MQTT client connected.")

		// subscribe to error messages
		if token := c.Subscribe("s/e", 0, receive); token.Wait() && token.Error() != nil {
			c8yError <- token.Error()
		} else {
			c8yReady <- true
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	select {
	case _ = <-c8yReady:
		return client, nil
	case err := <-c8yError:
		return nil, err
	}
}

func Send(c MQTT.Client, valueToSend int) error {
	message := "200,c8y_Value,V," + strconv.Itoa(valueToSend)
	log.Println("publishing...")
	if token := c.Publish("s/us", 0, false, message); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
